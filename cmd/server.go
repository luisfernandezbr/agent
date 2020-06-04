package cmd

import (
	"context"
	"encoding/json"
	"os"
	"plugin"
	"regexp"

	"github.com/go-redis/redis/v8"
	emanager "github.com/pinpt/agent.next/internal/manager/eventapi"
	"github.com/pinpt/agent.next/internal/server"
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
		log.Debug(logger, "looking for integrations", "dir", integrationsDir)
		integrationFiles, err := fileutil.FindFiles(integrationsDir, regexp.MustCompile(".so$"))
		if err != nil {
			log.Fatal(logger, "error loading integrations", "err", err)
		}
		if len(integrationFiles) == 0 {
			log.Fatal(logger, "no integrations found", "dir", integrationsDir)
		}
		var state sdk.State
		intconfig := sdk.Config{}
		integrations := make(map[string]*server.IntegrationContext)
		for _, fn := range integrationFiles {
			log.Debug(logger, "attempt to load plugin", "file", fn)
			plug, err := plugin.Open(fn)
			if err != nil {
				log.Fatal(logger, "couldn't open integration plugin", "err", err, "file", fn)
			}
			sym, err := plug.Lookup("Integration")
			if err != nil {
				log.Fatal(logger, "couldn't integration plugin entrypoint", "err", err)
			}
			instance := sym.(sdk.Integration)
			descriptor, err := sdk.LoadDescriptorFromPlugin(plug)
			if err != nil {
				log.Fatal(logger, "error loading integration", "err", err, "file", fn)
			}
			integrations[descriptor.RefType] = &server.IntegrationContext{
				Integration: instance,
				Descriptor:  descriptor,
			}
			log.Info(logger, "loaded integration", "name", descriptor.Name, "ref_type", descriptor.RefType, "build", descriptor.BuildCommitSHA, "date", descriptor.BuildDate)
		}
		var channel, uuid, apikey string
		var subchannel *event.SubscriptionChannel
		var redisClient *redis.Client
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
		manager := emanager.New(logger, channel)
		for _, instance := range integrations {
			if err := instance.Integration.Start(logger, intconfig, manager); err != nil {
				log.Fatal(logger, "error starting integration", "err", err, "name", instance.Descriptor.Name)
			}
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

		server, err := server.New(server.Config{
			Ctx:                 ctx,
			Dir:                 tmpdir,
			Logger:              logger,
			State:               state,
			SubscriptionChannel: subchannel,
			RedisClient:         redisClient,
			Integrations:        integrations,
			UUID:                uuid,
			Channel:             channel,
			APIKey:              apikey,
			Secret:              secret,
		})
		if err != nil {
			log.Fatal(logger, "error starting server", "err", err)
		}

		<-done
		log.Info(logger, "stopping")
		server.Close()
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
	serverCmd.Flags().String("redis", pos.Getenv("PP_REDIS_URL", "0.0.0.0:6379"), "the redis endpoint url")
	serverCmd.Flags().Int("redisDB", 0, "the redis db")
	serverCmd.Flags().MarkHidden("secret")
	serverCmd.Flags().MarkHidden("channel")
	serverCmd.Flags().MarkHidden("tempdir")
	serverCmd.Flags().MarkHidden("redis")
	serverCmd.Flags().MarkHidden("redisDB")
}
