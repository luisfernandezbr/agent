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
	devwebhook "github.com/pinpt/agent.next/internal/webhook/dev"
	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/fileutil"
	"github.com/pinpt/go-common/v10/log"
	pos "github.com/pinpt/go-common/v10/os"
	"github.com/spf13/cobra"
)

// ConfigFile is the configuration file for the runner
type ConfigFile struct {
	Channel      string `json:"channel"`
	CustomerID   string `json:"customer_id"`
	SystemID     string `json:"system_id"`
	APIKey       string `json:"apikey"`
	EnrollmentID string `json:"enrollment_id"`
}

func cloudAgentGroupID(reftype string) string {
	return fmt.Sprintf("agent-%s", reftype)
}

func onPremiseAgentGroupID(reftype, systemID string) string {
	return fmt.Sprintf("agent-%s-%s", systemID, reftype)
}

func getIntegrationConfig(cmd *cobra.Command) sdk.Config {
	var kv map[string]interface{}
	setargs, _ := cmd.Flags().GetStringArray("set")
	if len(setargs) > 0 {
		kv = make(map[string]interface{})
		for _, setarg := range setargs {
			tok := strings.Split(setarg, "=")
			kv[tok[0]] = tok[1]
		}
	}
	return sdk.NewConfig(kv)
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
			log.Info(logger, "starting", "ref_type", descriptor.RefType, "version", descriptor.BuildCommitSHA)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			cfg, _ := cmd.Flags().GetString("config")
			channel, _ := cmd.Flags().GetString("channel")
			secret, _ := cmd.Flags().GetString("secret")
			groupid, _ := cmd.Flags().GetString("groupid")
			if cfg == "" && secret == "" {
				log.Fatal(logger, "missing --config")
			}
			intconfig := getIntegrationConfig(cmd)
			var state sdk.State
			var uuid, apikey string
			var redisClient *redis.Client
			var selfManaged bool

			if secret != "" && cfg == "" {
				// running in multi agent mode
				if channel == "" {
					channel = "dev"
				}
				if groupid == "" {
					groupid = cloudAgentGroupID(descriptor.RefType)
				}

				// we must connect to redis in multi mode
				redisURL, _ := cmd.Flags().GetString("redis")
				redisURL = strings.ReplaceAll(redisURL, "redis://", "")
				redisURL = strings.TrimPrefix(redisURL, "//")
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
				selfManaged = true
				// running in single agent mode
				if !fileutil.FileExists(cfg) {
					log.Fatal(logger, "couldn't find config file at "+cfg)
				}
				of, err := os.Open(cfg)
				if err != nil {
					log.Fatal(logger, "error loading config file at "+cfg, "err", err)
				}
				var config ConfigFile
				if err := json.NewDecoder(of).Decode(&config); err != nil {
					of.Close()
					log.Fatal(logger, "error parsing config file at "+cfg, "err", err)
				}
				of.Close()
				if channel == "" {
					channel = config.Channel
				}
				uuid = config.SystemID
				apikey = config.APIKey
				if uuid == "" {
					config.SystemID = config.CustomerID
				}
				if groupid == "" {
					groupid = onPremiseAgentGroupID(descriptor.RefType, config.SystemID)
				}
				outdir, _ := cmd.Flags().GetString("dir")
				statefn := filepath.Join(outdir, descriptor.RefType+".state.json")
				stateobj, err := devstate.New(statefn)
				if err != nil {
					log.Fatal(logger, "error opening state file", "err", err)
				}
				state = stateobj
				defer stateobj.Close()
				log.Info(logger, "running in single agent mode", "uuid", config.SystemID, "customer_id", config.CustomerID, "channel", channel)
			}

			manager := emanager.New(emanager.Config{
				Channel:     channel,
				Logger:      logger,
				Secret:      secret,
				APIKey:      apikey,
				SelfManaged: selfManaged,
			})
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

			// get our temp folder to place in progress files
			tmpdir, _ := cmd.Flags().GetString("tempdir")
			if tmpdir == "" {
				tmpdir = os.TempDir()
			}
			os.MkdirAll(tmpdir, 0700)

			serverConfig := server.Config{
				Ctx:         ctx,
				Dir:         tmpdir,
				Logger:      logger,
				State:       state,
				RedisClient: redisClient,
				Integration: &server.IntegrationContext{
					Integration: integration,
					Descriptor:  descriptor,
				},
				UUID:    uuid,
				Channel: channel,
				APIKey:  apikey,
				Secret:  secret,
				GroupID: groupid,
			}

			server, err := server.New(serverConfig)
			if err != nil {
				log.Fatal(logger, "error starting server", "err", err)
			}

			<-done
			log.Info(logger, "stopping")
			server.Close()
			shutdown <- true
			log.Info(logger, "stopped")
		},
	}

	var devExportCmd = &cobra.Command{
		Use:    "dev-export",
		Short:  fmt.Sprintf("run the %s integration export", descriptor.RefType),
		Args:   cobra.NoArgs,
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			logger := log.NewCommandLogger(cmd)
			defer logger.Close()
			log.Info(logger, "starting", "ref_type", descriptor.RefType, "version", descriptor.BuildCommitSHA)
			channel, _ := cmd.Flags().GetString("channel")
			secret, _ := cmd.Flags().GetString("secret")
			intconfig := getIntegrationConfig(cmd)
			webhookEnabled, _ := cmd.Flags().GetBool("webhook")
			manager := emanager.New(emanager.Config{
				Channel:        channel,
				Logger:         logger,
				Secret:         secret,
				WebhookEnabled: webhookEnabled,
			})
			if err := integration.Start(logger, intconfig, manager); err != nil {
				log.Fatal(logger, "error starting integration", "err", err, "name", descriptor.Name)
			}
			// get our temp folder to place in progress files
			tmpdir, _ := cmd.Flags().GetString("tempdir")
			if tmpdir == "" {
				tmpdir = os.TempDir()
			}
			os.MkdirAll(tmpdir, 0700)

			outdir, _ := cmd.Flags().GetString("dir")
			statefn := filepath.Join(outdir, descriptor.RefType+".state.json")

			stateobj, err := devstate.New(statefn)
			if err != nil {
				log.Fatal(logger, "error opening state file", "err", err)
			}
			defer stateobj.Close()
			var pipe sdk.Pipe
			if outdir != "" {
				os.MkdirAll(outdir, 0700)
				pipe = file.New(logger, outdir)
			} else {
				pipe = console.New(logger)
			}
			defer pipe.Close()
			historical, _ := cmd.Flags().GetBool("historical")
			exp, err := devexport.New(logger, intconfig, stateobj, "9999", "1234", "1", historical, pipe)
			if err != nil {
				log.Fatal(logger, "export failed", "err", err)
			}
			// TODO(robin): use context
			_, cancel := context.WithCancel(context.Background())
			pos.OnExit(func(_ int) {
				log.Info(logger, "shutting down")
				cancel()
				go func() {
					time.Sleep(time.Second)
					os.Exit(1) // force exit if not already stopped
				}()
			})
			if err := integration.Export(exp); err != nil {
				log.Fatal(logger, "error running export", "err", err)
			}
		},
	}

	var devWebhookCmd = &cobra.Command{
		Use:    "dev-webhook",
		Short:  fmt.Sprintf("run the %s integration webhook", descriptor.RefType),
		Args:   cobra.NoArgs,
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			logger := log.NewCommandLogger(cmd)
			defer logger.Close()
			log.Info(logger, "starting", "ref_type", descriptor.RefType, "version", descriptor.BuildCommitSHA)
			channel, _ := cmd.Flags().GetString("channel")
			secret, _ := cmd.Flags().GetString("secret")
			intconfig := getIntegrationConfig(cmd)
			manager := emanager.New(emanager.Config{
				Channel:        channel,
				Secret:         secret,
				Logger:         logger,
				WebhookEnabled: true,
			})
			if err := integration.Start(logger, intconfig, manager); err != nil {
				log.Fatal(logger, "error starting integration", "err", err, "name", descriptor.Name)
			}
			// get our temp folder to place in progress files
			tmpdir, _ := cmd.Flags().GetString("tempdir")
			if tmpdir == "" {
				tmpdir = os.TempDir()
			}
			os.MkdirAll(tmpdir, 0700)

			outdir, _ := cmd.Flags().GetString("dir")
			statefn := filepath.Join(outdir, descriptor.RefType+".state.json")

			stateobj, err := devstate.New(statefn)
			if err != nil {
				log.Fatal(logger, "error opening state file", "err", err)
			}
			defer stateobj.Close()
			var pipe sdk.Pipe
			if outdir != "" {
				os.MkdirAll(outdir, 0700)
				pipe = file.New(logger, outdir)
			} else {
				pipe = console.New(logger)
			}
			defer pipe.Close()

			datastr, _ := cmd.Flags().GetString("data")
			data := make(map[string]interface{})
			if err := json.Unmarshal([]byte(datastr), &data); err != nil {
				log.Fatal(logger, "unable to decode webhook paylaod", "err", err)
			}

			headers := make(map[string]string)
			headersArr, _ := cmd.Flags().GetStringArray("headers")
			if len(headersArr) > 0 {
				for _, setarg := range headersArr {
					tok := strings.Split(setarg, "=")
					headers[tok[0]] = tok[1]
				}
			}
			refID, _ := cmd.Flags().GetString("dir")
			headers["ref_id"] = refID
			headers["customer_id"] = "1234"
			headers["integration_instance_id"] = "1"

			webhook := devwebhook.New(
				logger,
				intconfig,
				stateobj,
				"1234",
				refID,
				"1",
				pipe,
				headers,
				data,
				[]byte(datastr),
			)
			// TODO(robin): use context
			_, cancel := context.WithCancel(context.Background())
			pos.OnExit(func(_ int) {
				log.Info(logger, "shutting down")
				cancel()
				go func() {
					time.Sleep(time.Second)
					os.Exit(1) // force exit if not already stopped
				}()
			})
			if err := integration.WebHook(webhook); err != nil {
				log.Fatal(logger, "error running export", "err", err)
			}
		},
	}

	// server command
	log.RegisterFlags(serverCmd)
	serverCmd.Flags().String("config", "", "the config file location")
	serverCmd.PersistentFlags().StringArray("set", []string{}, "set a config value from the command line")
	serverCmd.PersistentFlags().String("secret", pos.Getenv("PP_AUTH_SHARED_SECRET", ""), "the secret which is only useful when running in the cloud")
	serverCmd.PersistentFlags().String("channel", pos.Getenv("PP_CHANNEL", ""), "the channel configuration")
	serverCmd.PersistentFlags().String("tempdir", "dist/export", "the directory to place files")
	serverCmd.PersistentFlags().String("redis", pos.Getenv("PP_REDIS_URL", "0.0.0.0:6379"), "the redis endpoint url")
	serverCmd.PersistentFlags().Int("redisDB", 15, "the redis db")
	serverCmd.PersistentFlags().String("groupid", "", "override the group id")
	serverCmd.Flags().MarkHidden("groupid")
	serverCmd.AddCommand(devExportCmd)
	serverCmd.AddCommand(devWebhookCmd)

	// dev export command
	devExportCmd.Flags().String("dir", "", "directory to place files when in dev mode")
	devExportCmd.Flags().Bool("historical", false, "force a historical export")
	devExportCmd.Flags().Bool("webhook", false, "turn on webhooks")

	// dev webhook command
	devWebhookCmd.Flags().String("dir", "", "directory to place files when in dev mode")
	devWebhookCmd.Flags().String("data", "", "the json payload of the webhook")
	devWebhookCmd.Flags().String("headers", "", "the headers of the webhook")
	devWebhookCmd.Flags().String("ref-id", "", "the refid on the webhook")

	if err := serverCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
