package cmd

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"plugin"
	"regexp"
	"sync"
	"time"

	"github.com/pinpt/agent.next/internal/export/eventapi"
	emanager "github.com/pinpt/agent.next/internal/manager/eventapi"
	pipe "github.com/pinpt/agent.next/internal/pipe/eventapi"
	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/event"
	"github.com/pinpt/go-common/fileutil"
	"github.com/pinpt/go-common/log"
	pos "github.com/pinpt/go-common/os"
	pstr "github.com/pinpt/go-common/strings"
	"github.com/pinpt/integration-sdk/agent"
	"github.com/spf13/cobra"
)

type configFile struct {
	Channel    string `json:"channel"`
	CustomerID string `json:"customer_id"`
	DeviceID   string `json:"device_id"`
	SystemID   string `json:"system_id"`
	APIKey     string `json:"api_key"`
}

type integrationData struct {
	Integration sdk.Integration
	Descriptor  *sdk.Descriptor
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "run an integration server listening for events",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewCommandLogger(cmd)
		defer logger.Close()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cfg, _ := cmd.Flags().GetString("config")
		secret, _ := cmd.Flags().GetString("secret")
		if cfg == "" && secret == "" {
			log.Fatal(logger, "missing --config")
		}
		integrationsDir, _ := cmd.Flags().GetString("integrations")
		integrationFiles, err := fileutil.FindFiles(integrationsDir, regexp.MustCompile(".so$"))
		if err != nil {
			log.Fatal(logger, "error loading integrations", "err", err)
		}
		manager := emanager.New(logger)
		intconfig := sdk.Config{}
		integrations := make(map[string]*integrationData)
		for _, fn := range integrationFiles {
			plug, err := plugin.Open(fn)
			if err != nil {
				log.Fatal(logger, "couldn't open integration plugin", "err", err, "file", fn)
			}
			sym, err := plug.Lookup("Integration")
			if err != nil {
				log.Fatal(logger, "couldn't integration plugin entrypoint", "err", err)
			}
			instance := sym.(sdk.Integration)
			if err := instance.Start(logger, intconfig, manager); err != nil {
				log.Fatal(logger, "error starting integration", "err", err, "file", fn)
			}
			descriptor, err := sdk.LoadDescriptorFromPlugin(plug)
			if err != nil {
				log.Fatal(logger, "error loading integration", "err", err, "file", fn)
			}
			integrations[descriptor.RefType] = &integrationData{instance, descriptor}
			log.Info(logger, "loaded integration", "name", descriptor.Name, "ref_type", descriptor.RefType, "build", descriptor.BuildCommitSHA, "date", descriptor.BuildDate)
		}
		if len(integrations) == 0 {
			log.Fatal(logger, "no integrations found", "dir", integrationsDir)
		}
		var channel, uuid, apikey string
		var subchannel *event.SubscriptionChannel
		if secret != "" {
			channel, _ = cmd.Flags().GetString("channel")
			// running in multi agent mode
			ch, err := event.NewSubscription(ctx, event.Subscription{
				Logger:            logger,
				Topics:            []string{agent.ExportRequestModelName.String()},
				GroupID:           "agent",
				HTTPHeaders:       map[string]string{"x-api-key": secret},
				DisableAutoCommit: true,
				Channel:           channel,
				DisablePing:       true,
			})
			if err != nil {
				log.Fatal(logger, "error creating subscription", "err", err)
			}
			subchannel = ch
			log.Info(logger, "running in multi agent mode", "channel", channel)
		} else {
			// running in single agent mode
			if !fileutil.FileExists(cfg) {
				log.Fatal(logger, "couldn't find config file at "+cfg)
			}
			of, err := os.Open(cfg)
			if err != nil {
				log.Fatal(logger, "error loading config file at "+cfg, "err", err)
			}
			var config configFile
			if err := json.NewDecoder(of).Decode(&config); err != nil {
				of.Close()
				log.Fatal(logger, "error parsing config file at "+cfg, "err", err)
			}
			of.Close()
			channel = config.Channel
			uuid = config.DeviceID
			apikey = config.APIKey
			ch, err := event.NewSubscription(ctx, event.Subscription{
				Logger: logger,
				Topics: []string{
					agent.ExportRequestModelName.String(),
				},
				GroupID:           "agent-" + config.DeviceID,
				DisableAutoCommit: true,
				Channel:           config.Channel,
				Headers: map[string]string{
					"uuid": config.DeviceID,
				},
				DisablePing: true,
			})
			if err != nil {
				log.Fatal(logger, "error creating subscription", "err", err)
			}
			subchannel = ch
			log.Info(logger, "running in single agent mode", "uuid", config.DeviceID, "customer_id", config.CustomerID, "channel", config.Channel)
		}
		done := make(chan bool, 1)
		shutdown := make(chan bool)
		pos.OnExit(func(_ int) {
			log.Info(logger, "shutting down")
			done <- true
			<-shutdown
		})
		subchannel.WaitForReady()
		log.Info(logger, "running")

		// get our temp folder to place in progress files
		tmpdir, _ := cmd.Flags().GetString("tempdir")
		if tmpdir == "" {
			tmpdir = os.TempDir()
		}
		os.MkdirAll(tmpdir, 0700)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Debug(logger, "starting subscription channel")
			for evt := range subchannel.Channel() {
				log.Debug(logger, "received event", "evt", evt)
				switch evt.Model {
				case agent.ExportRequestModelName.String():
					var req agent.ExportRequest
					json.Unmarshal([]byte(evt.Data), &req)
					dir := filepath.Join(tmpdir, req.JobID)
					os.MkdirAll(dir, 0700)
					var iwg sync.WaitGroup
					started := time.Now()
					for _, i := range req.Integrations {
						integrationdata := integrations[i.Name]
						if integrationdata == nil {
							// FIXME: send an error response
							log.Error(logger, "couldn't find integration named", "name", i.Name)
							continue
						}
						integration := integrationdata.Integration
						descriptor := integrationdata.Descriptor
						// build the sdk config for the integration
						sdkconfig := sdk.Config{}
						for k, v := range i.Authorization.ToMap() {
							sdkconfig[k] = pstr.Value(v)
						}
						iwg.Add(1)
						// start the integration in it's own thread
						go func(integration sdk.Integration, descriptor *sdk.Descriptor) {
							defer iwg.Done()
							completion := make(chan eventapi.Completion, 1)
							p := pipe.New(pipe.Config{
								Ctx:        ctx,
								Logger:     logger,
								Dir:        dir,
								CustomerID: req.CustomerID,
								UUID:       uuid,
								JobID:      req.JobID,
								Channel:    channel,
								APIKey:     apikey,
								Secret:     secret,
							})
							c, err := eventapi.New(eventapi.Config{
								Ctx:        ctx,
								Logger:     logger,
								Config:     sdkconfig,
								CustomerID: req.CustomerID,
								JobID:      req.JobID,
								UUID:       uuid,
								Pipe:       p,
								Completion: completion,
								Channel:    channel,
								APIKey:     apikey,
								Secret:     secret,
							})
							if err != nil {
								log.Fatal(logger, "error creating event api export", "err", err)
							}
							ts := time.Now()
							if err := integration.Export(c); err != nil {
								log.Fatal(logger, "error running export", "err", err)
							}
							<-completion
							log.Debug(logger, "export completed", "integration", descriptor.RefType, "duration", time.Since(ts))
						}(integration, descriptor)
					}
					log.Debug(logger, "waiting for export to complete")
					iwg.Wait()
					log.Info(logger, "export completed", "duration", time.Since(started), "jobid", req.JobID, "customer_id", req.CustomerID)
				}
				evt.Commit()
			}
		}()
		<-done
		log.Info(logger, "stopping")
		subchannel.Close()
		shutdown <- true
		log.Info(logger, "stopped")
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().String("config", "", "the config file location")
	serverCmd.Flags().String("secret", pos.Getenv("PP_AUTH_SHARED_SECRET", "fa0s8f09a8sd09f8iasdlkfjalsfm,.m,xf"), "the secret which is only useful when running in the cloud")
	serverCmd.Flags().String("channel", pos.Getenv("PP_CHANNEL", "dev"), "the channel configuration")
	serverCmd.Flags().String("integrations", "dist", "the directory where our integration plugins are stored")
	serverCmd.Flags().String("tempdir", "dist/export", "the directory to place files")
	serverCmd.Flags().MarkHidden("secret")
	serverCmd.Flags().MarkHidden("channel")
	serverCmd.Flags().MarkHidden("tempdir")
}
