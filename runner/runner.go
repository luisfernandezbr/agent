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
	"github.com/pinpt/go-common/v10/fileutil"
	"github.com/pinpt/go-common/v10/log"
	pos "github.com/pinpt/go-common/v10/os"
	"github.com/spf13/cobra"
)

// ConfigFile is the configuration file for the r
type ConfigFile struct {
	Channel    string `json:"channel"`
	CustomerID string `json:"customer_id"`
	DeviceID   string `json:"device_id"`
	SystemID   string `json:"system_id"`
	APIKey     string `json:"apikey"`
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
			dev, _ := cmd.Flags().GetBool("config")
			secret, _ := cmd.Flags().GetString("secret")
			if cfg == "" && secret == "" && !dev {
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
			var uuid, apikey, groupid string
			var redisClient *redis.Client
			channel, _ := cmd.Flags().GetString("channel")

			if !devMode {
				if secret != "" && cfg == "" {
					// running in multi agent mode
					if channel == "" {
						channel = "dev"
					}
					groupid = "agent"

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
					uuid = config.DeviceID
					apikey = config.APIKey
					if uuid == "" {
						config.DeviceID = config.CustomerID
					}
					groupid = "agent-" + config.DeviceID
					outdir, _ := cmd.Flags().GetString("dir")
					statefn := filepath.Join(outdir, descriptor.RefType+".state.json")
					stateobj, err := devstate.New(statefn)
					if err != nil {
						log.Fatal(logger, "error opening state file", "err", err)
					}
					state = stateobj
					defer stateobj.Close()
					log.Info(logger, "running in single agent mode", "uuid", config.DeviceID, "customer_id", config.CustomerID, "channel", channel)
				}
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
			if devMode {
				serverConfig.DevMode = true
				outdir, _ := cmd.Flags().GetString("dir")
				statefn := filepath.Join(outdir, descriptor.RefType+".state.json")

				stateobj, err := devstate.New(statefn)
				if err != nil {
					log.Fatal(logger, "error opening state file", "err", err)
				}
				var pipe sdk.Pipe
				if outdir != "" {
					os.MkdirAll(outdir, 0700)
					pipe = file.New(logger, outdir)
				} else {
					pipe = console.New(logger)
				}
				historical, _ := cmd.Flags().GetBool("historical")
				exp, err := devexport.New(logger, intconfig, stateobj, "9999", "1234", "1", historical, pipe)
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
				shutdown <- true
				log.Info(logger, "stopped")
			}
		},
	}
	log.RegisterFlags(serverCmd)
	serverCmd.Flags().String("config", "", "the config file location")
	serverCmd.Flags().String("secret", pos.Getenv("PP_AUTH_SHARED_SECRET", ""), "the secret which is only useful when running in the cloud")
	serverCmd.Flags().String("channel", pos.Getenv("PP_CHANNEL", ""), "the channel configuration")
	serverCmd.Flags().String("tempdir", "dist/export", "the directory to place files")
	serverCmd.Flags().String("redis", pos.Getenv("PP_REDIS_URL", "0.0.0.0:6379"), "the redis endpoint url")
	serverCmd.Flags().Int("redisDB", 15, "the redis db")
	serverCmd.Flags().Bool("dev", false, "running in dev mode, do a fake integration")
	serverCmd.Flags().String("dir", "", "directory to place files when in dev mode")
	serverCmd.Flags().StringSlice("set", []string{}, "set a config value from the command line")
	serverCmd.Flags().Bool("historical", false, "force a historical export")
	serverCmd.Flags().MarkHidden("dev")
	serverCmd.Flags().MarkHidden("dir")
	serverCmd.Flags().MarkHidden("historical")
	if err := serverCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
