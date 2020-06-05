package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	devexport "github.com/pinpt/agent.next/internal/export/dev"
	emanager "github.com/pinpt/agent.next/internal/manager/eventapi"
	"github.com/pinpt/agent.next/internal/pipe/console"
	"github.com/pinpt/agent.next/internal/pipe/file"
	"github.com/pinpt/agent.next/internal/server"
	devstate "github.com/pinpt/agent.next/internal/state/file"
	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/event"
	"github.com/pinpt/go-common/v10/fileutil"
	"github.com/pinpt/go-common/v10/log"
	pos "github.com/pinpt/go-common/v10/os"
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

// Main is the main entrypoint for an integration
func Main(integration sdk.Integration, args ...string) {
	descriptor, err := sdk.LoadDescriptor(args[0], args[1], args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error loading integration descriptor:", err)
		os.Exit(1)
	}
	var serverCmd = &cobra.Command{
		Use:   descriptor.RefType,
		Short: fmt.Sprintf("run the %s integration listening for events", descriptor.RefType),
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
			var state sdk.State
			var kv map[string]interface{}
			setargs, _ := cmd.Flags().GetStringSlice("set")
			if len(setargs) > 0 {
				kv = make(map[string]interface{})
				for _, setarg := range setargs {
					tok := strings.Split(setarg, "=")
					kv[tok[0]] = tok[1]
				}
			}
			devMode, _ := cmd.Flags().GetBool("dev")
			intconfig := sdk.NewConfig(kv)
			var channel, uuid, apikey string
			var subchannel *event.SubscriptionChannel
			var redisClient *redis.Client

			if !devMode {
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
						Headers: map[string]string{
							"integration": descriptor.RefType,
						},
					})
					if err != nil {
						log.Fatal(logger, "error creating subscription", "err", err)
					}
					subchannel = ch

					// we must connect to redis in multi mode
					redisURL, _ := cmd.Flags().GetString("redis")
					redisDb, _ := cmd.Flags().GetInt("redisDB")
					redisClient = redis.NewClient(&redis.Options{
						Addr: redisURL,
						DB:   redisDb,
					})
					log.Debug(logger, "attempt to ping redis", "url", redisURL)
					err = redisClient.Ping(ctx).Err()
					if err != nil {
						log.Fatal(logger, "error connecting to redis", "url", redisURL, "db", redisDb, "err", err)
					}
					defer redisClient.Close()
					log.Info(logger, "redis ping OK", "url", redisURL, "db", redisDb)
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
							"uuid":        config.DeviceID,
							"integration": descriptor.RefType,
						},
						DisablePing: true,
					})
					if err != nil {
						log.Fatal(logger, "error creating subscription", "err", err)
					}
					subchannel = ch
					log.Info(logger, "running in single agent mode", "uuid", config.DeviceID, "customer_id", config.CustomerID, "channel", config.Channel)
				}
			} else {
				channel, _ = cmd.Flags().GetString("channel")
			}

			manager := emanager.New(logger, channel)
			if err := integration.Start(logger, intconfig, manager); err != nil {
				log.Fatal(logger, "error starting integration", "err", err, "name", descriptor.Name)
			}
			done := make(chan bool, 1)
			shutdown := make(chan bool)
			pos.OnExit(func(_ int) {
				log.Info(logger, "shutting down")
				done <- true
				<-shutdown
			})

			if subchannel != nil {
				log.Info(logger, "waiting to connect")
				subchannel.WaitForReady()
				log.Info(logger, "running")
			}

			// get our temp folder to place in progress files
			tmpdir, _ := cmd.Flags().GetString("tempdir")
			if tmpdir == "" {
				tmpdir = os.TempDir()
			}
			os.MkdirAll(tmpdir, 0700)

			serverConfig := server.Config{
				Ctx:                 ctx,
				Dir:                 tmpdir,
				Logger:              logger,
				State:               state,
				SubscriptionChannel: subchannel,
				RedisClient:         redisClient,
				Integration: &server.IntegrationContext{
					Integration: integration,
					Descriptor:  descriptor,
				},
				UUID:    uuid,
				Channel: channel,
				APIKey:  apikey,
				Secret:  secret,
			}
			if devMode {
				serverConfig.DevMode = true
				completion := make(chan devexport.Completion, 1)
				statedir := filepath.Join(os.TempDir(), "agent.state")
				statefn := filepath.Join(statedir, "state.json")
				stateobj, err := devstate.New(statefn)
				if err != nil {
					log.Fatal(logger, "error opening state file", "err", err)
				}
				var pipe sdk.Pipe
				outdir, _ := cmd.Flags().GetString("dir")
				if outdir != "" {
					os.MkdirAll(outdir, 0700)
					pipe = file.New(logger, outdir)
				} else {
					pipe = console.New(logger)
				}
				exp, err := devexport.New(logger, intconfig, stateobj, "9999", "1234", pipe, completion)
				if err != nil {
					log.Fatal(logger, "export failed", "err", err)
				}
				_ctx, cancel := context.WithCancel(ctx)
				pos.OnExit(func(_ int) {
					log.Info(logger, "shutting down")
					cancel()
					go func() {
						time.Sleep(time.Second)
						os.Exit(1) // force exit if not already stopped
					}()
				})
				serverConfig.Ctx = _ctx
				serverConfig.State = stateobj
				serverConfig.DevExport = exp
				serverConfig.DevPipe = pipe
			}
			server, err := server.New(serverConfig)
			if err != nil {
				log.Fatal(logger, "error starting server", "err", err)
			}

			if !devMode {
				<-done
				log.Info(logger, "stopping")
				server.Close()
				subchannel.Close()
				shutdown <- true
				log.Info(logger, "stopped")
			}
		},
	}
	log.RegisterFlags(serverCmd)
	serverCmd.Flags().String("config", "", "the config file location")
	serverCmd.Flags().String("secret", pos.Getenv("PP_AUTH_SHARED_SECRET", "fa0s8f09a8sd09f8iasdlkfjalsfm,.m,xf"), "the secret which is only useful when running in the cloud")
	serverCmd.Flags().String("channel", pos.Getenv("PP_CHANNEL", "dev"), "the channel configuration")
	serverCmd.Flags().String("tempdir", "dist/export", "the directory to place files")
	serverCmd.Flags().String("redis", pos.Getenv("PP_REDIS_URL", "0.0.0.0:6379"), "the redis endpoint url")
	serverCmd.Flags().Int("redisDB", 0, "the redis db")
	serverCmd.Flags().Bool("dev", false, "running in dev mode, do a fake integration")
	serverCmd.Flags().String("dir", "", "directory to place files when in dev mode")
	serverCmd.Flags().StringSlice("set", []string{}, "set a config value from the command line")
	serverCmd.Flags().MarkHidden("secret")
	serverCmd.Flags().MarkHidden("channel")
	serverCmd.Flags().MarkHidden("tempdir")
	serverCmd.Flags().MarkHidden("redis")
	serverCmd.Flags().MarkHidden("redisDB")
	serverCmd.Flags().MarkHidden("dev")
	serverCmd.Flags().MarkHidden("dir")
	if err := serverCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
