package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pinpt/agent/v4/internal/util"
	"github.com/pinpt/agent/v4/runner"
	"github.com/pinpt/agent/v4/sdk"
	"github.com/pinpt/agent/v4/sysinfo"
	"github.com/pinpt/go-common/v10/api"
	"github.com/pinpt/go-common/v10/datetime"
	"github.com/pinpt/go-common/v10/event"
	"github.com/pinpt/go-common/v10/fileutil"
	"github.com/pinpt/go-common/v10/graphql"
	pjson "github.com/pinpt/go-common/v10/json"
	"github.com/pinpt/go-common/v10/log"
	pos "github.com/pinpt/go-common/v10/os"
	pstrings "github.com/pinpt/go-common/v10/strings"
	"github.com/pinpt/integration-sdk/agent"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// DBChange event
type DBChange struct {
	// Action the action that was taken
	Action string `json:"action" codec:"action" bson:"action" yaml:"action" faker:"-"`
	// Data the data payload of the change
	Data string `json:"data" codec:"data" bson:"data" yaml:"data" faker:"-"`
}

// Integration A registry integration
type Integration struct {
	// RefType the reference type
	RefType string `json:"ref_type" codec:"ref_type" bson:"ref_type" yaml:"ref_type" faker:"-"`
	// UpdatedAt the date the integration was last updated
	UpdatedAt int64 `json:"updated_ts" codec:"updated_ts" bson:"updated_ts" yaml:"updated_ts" faker:"-"`
	// Version the latest version that was published
	Version string `json:"version" codec:"version" bson:"version" yaml:"version" faker:"-"`
}

func getIntegration(ctx context.Context, logger log.Logger, channel string, dir string, publisher string, integration string, version string, cmdargs []string, force bool) (*exec.Cmd, error) {
	if publisher == "" {
		return nil, fmt.Errorf("missing publisher")
	}
	if integration == "" {
		return nil, fmt.Errorf("missing integration")
	}
	longName := fmt.Sprintf("%s/%s", publisher, integration)
	if version != "" {
		longName += "/" + version
	}
	integrationExecutable, _ := filepath.Abs(filepath.Join(dir, integration))
	if force || !fileutil.FileExists(integrationExecutable) {
		log.Info(logger, "need to download integration", "integration", longName, "force", force)
		var err error
		integrationExecutable, err = downloadIntegration(logger, channel, dir, publisher, integration, version)
		if err != nil {
			return nil, fmt.Errorf("error downloading integration %s: %w", longName, err)
		}
		log.Info(logger, "downloaded", "integration", integrationExecutable)
	}
	return startIntegration(ctx, logger, integrationExecutable, cmdargs)
}

func startIntegration(ctx context.Context, logger log.Logger, integrationExecutable string, cmdargs []string) (*exec.Cmd, error) {
	log.Info(logger, "starting", "file", integrationExecutable)
	cm := exec.CommandContext(ctx, integrationExecutable, cmdargs...)
	cm.Stdout = os.Stdout
	cm.Stderr = os.Stderr
	cm.Stdin = os.Stdin
	cm.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cm.Start(); err != nil {
		return nil, err
	}
	return cm, nil
}

func configFilename(cmd *cobra.Command) (string, error) {
	fn, _ := cmd.Flags().GetString("config")
	if fn == "" {
		fn = filepath.Join(os.Getenv("HOME"), ".pinpoint-agent/config.json")
	}
	return filepath.Abs(fn)
}

// clientFromConfig will use the contents of ConfigFile to make a client
func clientFromConfig(config *runner.ConfigFile) (graphql.Client, error) {
	gclient, err := graphql.NewClient(config.CustomerID, "", "", api.BackendURL(api.GraphService, config.Channel))
	if err != nil {
		return nil, err
	}
	gclient.SetHeader("Authorization", config.APIKey)

	return gclient, nil
}

func validateConfig(config *runner.ConfigFile, channel string, force bool) (bool, error) {
	var resp struct {
		Expired bool `json:"expired"`
		Valid   bool `json:"valid"`
	}
	if !force {
		res, err := api.Get(context.Background(), channel, api.AuthService, "/validate?customer_id="+config.CustomerID, config.APIKey)
		if err != nil {
			return false, err
		}
		defer res.Body.Close()
		if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
			return false, err
		}
	} else {
		resp.Expired = true
	}
	if resp.Expired {
		newToken, err := util.RefreshOAuth2Token(http.DefaultClient, channel, "pinpoint", config.RefreshKey)
		if err != nil {
			return false, err
		}
		config.APIKey = newToken // update the new token
		return true, nil
	}
	if resp.Valid {
		return false, fmt.Errorf("the apikey or refresh token is no longer valid")
	}
	return false, nil
}

func saveConfig(cmd *cobra.Command, config *runner.ConfigFile) error {
	cfg, err := configFilename(cmd)
	if err != nil {
		return err
	}
	of, err := os.Open(cfg)
	if err != nil {
		return err
	}
	defer of.Close()
	// save our channels back to the config
	if err := json.NewEncoder(of).Encode(config); err != nil {
		return err
	}
	return nil
}

func loadConfig(cmd *cobra.Command, logger log.Logger, channel string) (string, *runner.ConfigFile) {
	cfg, err := configFilename(cmd)
	if err != nil {
		log.Fatal(logger, "error getting config file name", "err", err)
	}
	if fileutil.FileExists(cfg) {
		var config runner.ConfigFile
		of, err := os.Open(cfg)
		if err != nil {
			log.Fatal(logger, "error opening config file at "+cfg, "err", err)
		}
		defer of.Close()
		if err := json.NewDecoder(of).Decode(&config); err != nil {
			log.Fatal(logger, "error parsing config file at "+cfg, "err", err)
		}
		of.Close()
		updated, err := validateConfig(&config, channel, false)
		if err != nil {
			log.Fatal(logger, "error validating the apikey", "err", err)
		}
		if updated {
			// save our changes back to the config
			if err := saveConfig(cmd, &config); err != nil {
				log.Fatal(logger, "error opening config file for writing at "+cfg, "err", err)
			}
		}
		client, err := clientFromConfig(&config)
		if err != nil {
			log.Fatal(logger, "error creating client", "err", err)
		}
		exists, err := enrollmentExists(client, config.EnrollmentID)
		if err != nil {
			log.Fatal(logger, "error checking enrollment", "err", err)
		}
		if exists {
			return cfg, &config
		}
		log.Info(logger, "agent configuration found, but not known to Pinpoint, re-enrolling now", "path", cfg)
	} else {
		log.Info(logger, "no agent configuration found, enrolling now", "path", cfg)
	}
	config, err := enrollAgent(logger, channel, cfg)
	if err != nil {
		log.Fatal(logger, "error enrolling new agent", "err", err)
	}
	return cfg, config
}

type integrationInstruction int

const (
	doNothing integrationInstruction = iota
	shouldStart
	shouldStop
)

func vetDBChange(evt event.SubscriptionEvent, enrollmentID string) (integrationInstruction, *agent.IntegrationInstance, error) {
	var dbchange DBChange
	if err := json.Unmarshal([]byte(evt.Data), &dbchange); err != nil {
		return 0, nil, fmt.Errorf("error decoding dbchange: %w", err)
	}
	var instance agent.IntegrationInstance
	if err := json.Unmarshal([]byte(dbchange.Data), &instance); err != nil {
		return 0, nil, fmt.Errorf("error decoding integration instance: %w", err)
	}
	if instance.EnrollmentID == nil || *instance.EnrollmentID == "" || *instance.EnrollmentID != enrollmentID {
		return doNothing, nil, nil
	}
	if instance.Active == true && instance.Setup == agent.IntegrationInstanceSetupReady {
		return shouldStart, &instance, nil
	}
	if instance.Active == false && instance.Deleted == true {
		return shouldStop, &instance, nil
	}
	return doNothing, nil, nil
}

type integrationResult struct {
	Data *struct {
		Integration struct {
			RefType   string `json:"ref_type"`
			Publisher struct {
				Identifier string `json:"identifier"`
			} `json:"publisher"`
		} `json:"Integration"`
	} `json:"registry"`
}

var integrationQuery = `query findIntegration($id: ID!) {
	registry {
	  Integration(_id: $id) {
		ref_type
		publisher {
		  identifier
		}
	  }
	}
}`

func pingEnrollment(logger log.Logger, client graphql.Client, enrollmentID string, datefield string, active bool) error {
	log.Info(logger, "updating enrollment", "setting", datefield, "enrollment_id", enrollmentID, "active", active)
	now := datetime.NewDateNow()
	vars := make(graphql.Variables)
	if datefield != "" {
		vars[datefield] = now
		vars[agent.EnrollmentModelRunningColumn] = active
	}
	vars[agent.EnrollmentModelLastPingDateColumn] = now
	return agent.ExecEnrollmentSilentUpdateMutation(client, enrollmentID, vars, false)
}

func setIntegrationRunning(client graphql.Client, integrationInstanceID string) error {
	vars := graphql.Variables{
		agent.IntegrationInstanceModelSetupColumn: agent.IntegrationInstanceSetupRunning,
	}
	if err := agent.ExecIntegrationInstanceSilentUpdateMutation(client, integrationInstanceID, vars, false); err != nil {
		return fmt.Errorf("error updating integration instance to running: %w", err)
	}
	return nil
}

func runIntegrationMonitor(ctx context.Context, logger log.Logger, cmd *cobra.Command) {
	channel, _ := cmd.Flags().GetString("channel")
	args := []string{}
	cmd.Flags().Visit(func(f *pflag.Flag) {
		args = append(args, "--"+f.Name, f.Value.String())
	})
	var gclient graphql.Client
	integrations := make(map[string]string) // id -> identifier/ref_type
	processes := make(map[string]*exec.Cmd)
	var processLock sync.Mutex
	getIntegration := func(id string) (string, error) {
		processLock.Lock()
		val := integrations[id]
		if val != "" {
			processLock.Unlock()
			return val, nil
		}
		var res integrationResult
		if err := gclient.Query(integrationQuery, graphql.Variables{"id": id}, &res); err != nil {
			processLock.Unlock()
			return "", err
		}
		if res.Data == nil {
			processLock.Unlock()
			return "", fmt.Errorf("couldn't find integration with id: %s", id)
		}
		val = res.Data.Integration.Publisher.Identifier + "/" + res.Data.Integration.RefType
		integrations[id] = val
		processLock.Unlock()
		return val, nil
	}
	cfg, config := loadConfig(cmd, logger, channel)
	if channel == "" {
		channel = config.Channel
	}
	args = append(args, "--config", cfg)
	gclient, err := graphql.NewClient(config.CustomerID, "", "", api.BackendURL(api.GraphService, channel))
	if err != nil {
		log.Fatal(logger, "error creating graphql client", "err", err)
	}
	gclient.SetHeader("Authorization", config.APIKey)

	errors := make(chan error)
	go func() {
		for err := range errors {
			if err != nil {
				log.Fatal(logger, err.Error())
			}
		}
	}()
	groupID := "agent-run-monitor"
	if config.EnrollmentID != "" {
		// if self managed, we need to use a different group id than the cloud
		groupID += "-" + config.EnrollmentID
	}
	ch, err := event.NewSubscription(ctx, event.Subscription{
		GroupID:           "agent-run-monitor-" + config.EnrollmentID,
		Topics:            []string{"ops.db.Change"},
		Channel:           channel,
		APIKey:            config.APIKey,
		DisablePing:       true,
		Logger:            logger,
		Errors:            errors,
		DisableAutoCommit: true,
		Filter: &event.SubscriptionFilter{
			ObjectExpr: `model:"agent.IntegrationInstance" AND action:"update"`,
		},
	})
	if err != nil {
		log.Fatal(logger, "error creating montior subscription", "err", err)
	}
	ch.WaitForReady()
	defer ch.Close()

	// set startup date
	if err := pingEnrollment(logger, gclient, config.EnrollmentID, agent.EnrollmentModelLastStartupDateColumn, true); err != nil {
		log.Error(logger, "unable to update enrollment", "enrollment_id", config.EnrollmentID, "err", err)
	}

	runIntegration := func(name string) {
		log.Info(logger, "running integration", "name", name)
		processLock.Lock()
		startFile, _ := ioutil.TempFile("", "")
		defer os.Remove(startFile.Name())
		args = append(args, "--start-file", startFile.Name())
		c := exec.CommandContext(ctx, os.Args[0], append([]string{"run", name}, args...)...)
		var wg sync.WaitGroup
		wg.Add(1)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		if err := c.Start(); err != nil {
			processLock.Unlock()
			log.Fatal(logger, "error starting "+name, "err", err)
		}
		exited := make(chan bool)
		pos.OnExit(func(_ int) {
			log.Debug(logger, "exit")
			close(exited)
		})
		go func() {
			for {
				select {
				case <-exited:
					wg.Done()
					return
				case <-time.After(time.Second):
					if fileutil.FileExists(startFile.Name()) {
						wg.Done()
						os.Remove(startFile.Name())
						return
					}
				case <-time.After(5 * time.Minute):
					log.Fatal(logger, "failed to start integration "+name+" after 5 minutes")
				}
			}
		}()
		processes[name] = c
		processLock.Unlock()
		log.Debug(logger, "waiting for integration to start")
		wg.Wait()
		if c != nil && c.ProcessState != nil && c.ProcessState.Exited() {
			log.Info(logger, "integration is not running")
		} else {
			log.Info(logger, "integration is running")
		}
	}

	// find all the integrations we have setup
	query := &agent.IntegrationInstanceQuery{
		Filters: []string{
			agent.IntegrationInstanceModelDeletedColumn + " = ?",
			agent.IntegrationInstanceModelEnrollmentIDColumn + " = ?",
		},
		Params: []interface{}{
			false,
			config.EnrollmentID,
		},
	}
	q := &agent.IntegrationInstanceQueryInput{Query: query}
	instances, err := agent.FindIntegrationInstances(gclient, q)
	if err != nil {
		log.Fatal(logger, "error finding integration instances", "err", err)
	}
	if instances != nil {
		for _, edge := range instances.Edges {
			name, err := getIntegration(edge.Node.IntegrationID)
			if err != nil {
				log.Fatal(logger, "error fetching integration name for integration", "err", err, "integration_id", edge.Node.IntegrationID, "id", edge.Node.ID)
			}
			runIntegration(name)
			if edge.Node.Setup == agent.IntegrationInstanceSetupReady {
				log.Info(logger, "setting integration to running 🏃‍♀️", "integration_instance_id", edge.Node.ID)
				if err := setIntegrationRunning(gclient, edge.Node.ID); err != nil {
					log.Fatal(logger, "error updating integration instance", "err", err, "integration_id", edge.Node.IntegrationID, "id", edge.Node.ID)
				}
			}
		}
	}

	var restarting bool
	done := make(chan bool)
	finished := make(chan bool)
	pos.OnExit(func(_ int) {
		if !restarting {
			done <- true
			<-finished
			log.Info(logger, "👯")
		}
	})

	if err := pingEnrollment(logger, gclient, config.EnrollmentID, "", true); err != nil {
		log.Error(logger, "unable to update enrollment", "enrollment_id", config.EnrollmentID, "err", err)
	}

	// calculate the duration of time left before the
	refreshDuration := config.Expires.Sub(time.Now().Add(time.Minute * 30))

	var shutdownWg sync.WaitGroup

	// run a loop waiting for exit or an updated integration instance
completed:
	for {
		select {
		case <-time.After(refreshDuration):
			// if we get here, we need to refresh our expiring apikey and restart all the integrations
			log.Info(logger, "need to update apikey before it expires and restart 🤞🏽")

			// 1. fetch our new apikey
			if _, err := validateConfig(config, channel, true); err != nil {
				log.Fatal(logger, "error refreshing the expiring apikey", "err", err)
			}
			// 2. save the config file changes
			if err := saveConfig(cmd, config); err != nil {
				log.Fatal(logger, "error saving the new apikey", "err", err)
			}
			// 3. stop all of our current integrations
			restarting = true
			done <- true

			// 4. restart ourself which should re-entrant this function
			go runIntegrationMonitor(ctx, logger, cmd)

		case <-time.After(time.Minute * 5):
			if err := pingEnrollment(logger, gclient, config.EnrollmentID, "", true); err != nil {
				log.Error(logger, "unable to update enrollment", "enrollment_id", config.EnrollmentID, "err", err)
			}
		case <-done:
			processLock.Lock()
			for k, c := range processes {
				log.Debug(logger, "stopping "+k, "pid", c.Process.Pid)
				shutdownWg.Add(1)
				go func(c *exec.Cmd, name string) {
					defer shutdownWg.Done()
					syscall.Kill(-c.Process.Pid, syscall.SIGINT)
					exited := make(chan bool)
					go func() {
						c.Wait()
						log.Debug(logger, "exited "+name)
						exited <- true
					}()
					select {
					case <-time.After(time.Second * 15):
						log.Debug(logger, "timed out on exit for "+name)
						if c.Process != nil {
							c.Process.Kill()
						}
						return
					case <-exited:
						return
					}
				}(c, k)
				delete(processes, k)
			}
			processLock.Unlock()
			break completed
		case evt := <-ch.Channel():
			instruction, instance, err := vetDBChange(evt, config.EnrollmentID)
			if err != nil {
				log.Fatal(logger, "error decoding db change", "err", err)
			}
			switch instruction {
			case shouldStart:
				log.Info(logger, "db change received, need to create a new process", "id", instance.ID)
				if err := setIntegrationRunning(gclient, instance.ID); err != nil {
					log.Fatal(logger, "error updating integration", "err", err)
				}
				name, err := getIntegration(instance.ID)
				if err != nil {
					log.Fatal(logger, "error fetching integration detail", "err", err)
				}
				processLock.Lock()
				c := processes[name]
				if c == nil {
					processLock.Unlock()
					runIntegration(name)
				} else {
					processLock.Unlock()
				}
			case shouldStop:
				log.Info(logger, "db change delete received, need to delete process", "id", instance.ID)
				name, err := getIntegration(instance.ID)
				if err != nil {
					log.Fatal(logger, "error fetching integration detail", "err", err)
				}
				processLock.Lock()
				c := processes[name]
				if c != nil {
					c.Process.Kill()
					delete(processes, instance.RefType)
				}
				processLock.Unlock()
			}
			evt.Commit()
		}
	}

	shutdownWg.Wait()

	// if restarting, dont send a shutdown or block on finished
	if restarting {
		return
	}

	if err := pingEnrollment(logger, gclient, config.EnrollmentID, agent.EnrollmentModelLastShutdownDateColumn, false); err != nil {
		log.Error(logger, "unable to update enrollment", "enrollment_id", config.EnrollmentID, "err", err)
	}
	finished <- true
}

func enrollmentExists(client graphql.Client, enrollmentID string) (bool, error) {
	enrollment, err := agent.FindEnrollment(client, enrollmentID)
	return enrollment != nil, err
}

func enrollAgent(logger log.Logger, channel string, configFileName string) (*runner.ConfigFile, error) {
	var config runner.ConfigFile
	config.Channel = channel

	url := sdk.JoinURL(api.BackendURL(api.AppService, channel), "/enroll")

	var userID string
	err := util.WaitForRedirect(url, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		config.APIKey = q.Get("apikey")
		config.RefreshKey = q.Get("refresh_token")
		config.CustomerID = q.Get("customer_id")
		expires := q.Get("expires")
		if expires != "" {
			e, _ := strconv.ParseInt(expires, 10, 64)
			config.Expires = datetime.DateFromEpoch(e)
		} else {
			config.Expires = time.Now().Add(time.Hour * 23)
		}
		userID = q.Get("user_id")
		w.WriteHeader(http.StatusOK)
	})
	if err != nil {
		return nil, fmt.Errorf("error waiting for browser: %w", err)
	}
	log.Info(logger, "logged in", "customer_id", config.CustomerID)

	log.Info(logger, "enrolling agent...", "customer_id", config.CustomerID)
	client, err := graphql.NewClient(config.CustomerID, "", "", api.BackendURL(api.GraphService, channel))
	if err != nil {
		return nil, fmt.Errorf("error creating graphql client: %w", err)
	}
	client.SetHeader("Authorization", config.APIKey)
	info, err := sysinfo.GetSystemInfo()
	if err != nil {
		return nil, fmt.Errorf("error getting system info: %w", err)
	}
	config.SystemID = info.ID
	created := agent.EnrollmentCreatedDate(datetime.NewDateNow())
	enr := agent.Enrollment{
		AgentVersion: commit, // TODO(robin): when we start versioning, switch this to version
		CreatedDate:  created,
		SystemID:     info.ID,
		Hostname:     info.Hostname,
		NumCPU:       info.NumCPU,
		OS:           info.OS,
		Architecture: info.Architecture,
		GoVersion:    info.GoVersion,
		CustomerID:   config.CustomerID,
		UserID:       userID,
		ID:           agent.NewEnrollmentID(config.CustomerID, info.ID),
	}
	config.EnrollmentID = enr.ID
	if err := agent.CreateEnrollment(client, enr); err != nil {
		if strings.Contains(err.Error(), "duplicate key error") {
			log.Info(logger, "looks like this system has already been enrolled, recreating local config")
		} else {
			return nil, fmt.Errorf("error creating enrollment: %w", err)
		}
	}
	os.MkdirAll(filepath.Dir(configFileName), 0700)
	if err := ioutil.WriteFile(configFileName, []byte(pjson.Stringify(config)), 0644); err != nil {
		return nil, fmt.Errorf("error writing config file: %w", err)
	}
	log.Info(logger, "agent enrolled 🎉", "customer_id", config.CustomerID)
	return &config, nil
}

// enrollAgentCmd will authenticate with pinpoint and create an agent enrollment
var enrollAgentCmd = &cobra.Command{
	Use:    "enroll",
	Short:  "connect this agent to Pinpoint's backend",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewCommandLogger(cmd)
		defer logger.Close()
		channel, _ := cmd.Flags().GetString("channel")
		fn, err := configFilename(cmd)
		if err != nil {
			log.Fatal(logger, "error getting config file name", "err", err)
		}
		if _, err := enrollAgent(logger, channel, fn); err != nil {
			log.Fatal(logger, "error enrolling this agent", "err", err)
		}
	},
}

func copyFile(from, to string) error {
	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(to)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run <integration> <version>",
	Short: "run a published integration",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		_logger := log.NewCommandLogger(cmd)
		defer _logger.Close()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if len(args) == 0 {
			log.Info(_logger, "starting main process")
			// need to figure out all our configured integrations and run each one
			runIntegrationMonitor(ctx, _logger, cmd)
			return
		}

		fullIntegration := args[0]
		var version string
		if len(args) > 1 {
			version = args[1]
		}
		tok := strings.Split(fullIntegration, "/")
		if len(tok) != 2 {
			log.Fatal(_logger, "integration should be in the format: publisher/integration such as pinpt/github")
		}
		publisher := tok[0]
		integration := tok[1]
		logger := log.With(_logger, "pkg", integration)
		channel, _ := cmd.Flags().GetString("channel")
		dir, _ := cmd.Flags().GetString("dir")
		secret, _ := cmd.Flags().GetString("secret")
		dir, err := filepath.Abs(dir)
		if err != nil {
			log.Fatal(logger, "error getting dir absolute path", "err", err)
		}

		var selfManaged bool
		var ch *event.SubscriptionChannel
		var cmdargs []string
		if secret != "" {
			log.Debug(logger, "creating internal subscription")
			if channel == "" {
				channel = "stable"
			}
			// each replica agent should recieve updates
			groupSuffix, err := os.Hostname()
			if err != nil {
				groupSuffix = pstrings.NewUUIDV4()
				log.Warn(logger, "unable to get hostname, using random uuid", "uuid", groupSuffix, "err", err)
			}
			ch, err = event.NewSubscription(ctx, event.Subscription{
				GroupID:     "agent-run-" + publisher + "-" + integration + "-" + groupSuffix,
				Topics:      []string{"ops.db.Change"},
				Channel:     channel,
				HTTPHeaders: map[string]string{"x-api-key": secret},
				DisablePing: true,
				Temporary:   true,
				Logger:      logger,
				Filter: &event.SubscriptionFilter{
					ObjectExpr: `model:"registry.Integration"`,
				},
			})
			cmdargs = append(cmdargs, "--secret="+secret, "--channel="+channel)
		} else {
			selfManaged = true
			log.Debug(logger, "creating external subscription")
			cfg, config := loadConfig(cmd, logger, channel)
			apikey := config.APIKey
			if channel == "" {
				channel = config.Channel
			}
			ch, err = event.NewSubscription(ctx, event.Subscription{
				GroupID:     "agent-run-" + publisher + "-" + integration + "-" + config.EnrollmentID,
				Topics:      []string{"ops.db.Change"},
				Channel:     channel,
				APIKey:      apikey,
				DisablePing: true,
				Logger:      logger,
				Filter: &event.SubscriptionFilter{
					ObjectExpr: `model:"registry.Integration"`,
				},
			})
			cmdargs = append(cmdargs, "--config="+cfg, "--channel="+channel)
		}
		if err != nil {
			log.Fatal(logger, "error creating subscription", "err", err)
		}

		// start file is used to signal to the monitor when we're running
		startfile, _ := cmd.Flags().GetString("start-file")
		if startfile != "" {
			os.Remove(startfile)
			cmdargs = append(cmdargs, "--start-file", startfile)
			defer os.Remove(startfile)
		}

		log.Info(logger, "waiting for subscription to be ready", "channel", channel)
		ch.WaitForReady()
		log.Info(logger, "subscription is ready")

		defer ch.Close()

		var stopped, restarting bool
		var stoppedLock, restartLock sync.Mutex
		done := make(chan bool)
		finished := make(chan bool, 1)
		restart := make(chan bool, 2)
		exited := make(chan bool)
		var currentCmd *exec.Cmd
		var restarted int

		pos.OnExit(func(_ int) {
			stoppedLock.Lock()
			stopped = true
			stoppedLock.Unlock()
			done <- true
			<-finished
		})

		integrationBinary := filepath.Join(dir, integration)
		previousIntegrationBinary := filepath.Join(dir, "old-"+integration)

		restart <- true // start it up

	exit:
		for {
			stoppedLock.Lock()
			s := stopped
			stoppedLock.Unlock()
			if s {
				break
			}
			select {
			case evt := <-ch.Channel():
				var dbchange DBChange
				json.Unmarshal([]byte(evt.Data), &dbchange)
				var instance Integration
				json.Unmarshal([]byte(dbchange.Data), &instance)
				log.Debug(logger, "db change event received", "ref_type", instance.RefType, "integration", integration)
				if instance.RefType == integration {
					switch dbchange.Action {
					case "update", "UPDATE", "upsert", "UPSERT":
						// copy the binary so we can rollback if needed
						if err := copyFile(integrationBinary, previousIntegrationBinary); err != nil {
							log.Error(logger, "error copying integration", "err", err)
							break exit
						}
						log.Info(logger, "new integration detected, will restart in 15s", "version", instance.Version)
						restartLock.Lock()
						restarting = true
						restarted = 0
						version = instance.Version
						time.Sleep(time.Second * 15)
						restart <- true // force a new download
						restartLock.Unlock()
					case "delete", "DELETE":
						// TODO -- exit with a special code to indicate we don't need to restart this integration
					}
				}
				go evt.Commit() // we need to put in a separate thread since we're holding the sub thread
			case <-done:
				if currentCmd != nil {
					syscall.Kill(-currentCmd.Process.Pid, syscall.SIGINT)
					select {
					case <-time.After(time.Second * 10):
						break
					case <-exited:
						break
					}
				}
				break
			case force := <-restart:
				log.Info(logger, "restart requested")
				if currentCmd != nil && currentCmd.Process != nil {
					currentCmd.Process.Kill()
					currentCmd = nil
				}
				stoppedLock.Lock()
				s := stopped
				stoppedLock.Unlock()
				log.Info(logger, "need to start", "stopped", s)
				if !s {
					restarted++
					c, err := getIntegration(ctx, logger, channel, dir, publisher, integration, version, cmdargs, force)
					if err != nil {
						log.Error(logger, "error running integration", "err", err)
						if !fileutil.FileExists(previousIntegrationBinary) {
							break exit
						} else {
							log.Info(logger, "attempting to roll back to previous version of integration", "integration", integration)
							os.Remove(integrationBinary)
							os.Rename(previousIntegrationBinary, integrationBinary)
							os.Chmod(integrationBinary, 0775)
							c, err = startIntegration(ctx, logger, integrationBinary, cmdargs)
							if err != nil {
								log.Error(logger, "error running rolled back integration", "err", err)
								break exit
							}
						}
					}
					currentCmd = c
					os.Remove(previousIntegrationBinary)
					log.Info(logger, "started", "pid", c.Process.Pid)
					go func() {
						// monitor the exit
						err := currentCmd.Wait()
						if err != nil {
							if currentCmd != nil && currentCmd.ProcessState != nil {
								if currentCmd.ProcessState.ExitCode() != 0 {
									if !selfManaged {
										// in cloud we should just end the process
										log.Fatal(logger, "integration has exited", "restarted", restarted, "code", currentCmd.ProcessState.ExitCode(), "err", err)
									} else {
										log.Error(logger, "integration has exited", "restarted", restarted, "code", currentCmd.ProcessState.ExitCode(), "err", err)
									}

								}
							}
							log.Info(logger, "pausing", "duration", time.Second*time.Duration(restarted))
							time.Sleep(time.Second * time.Duration(restarted))
						} else {
							restarted = 0
						}
						// try and restart if we're not in the stopping mode
						stoppedLock.Lock()
						s := stopped
						stoppedLock.Unlock()
						if !s {
							restartLock.Lock()
							r := restarting
							restartLock.Unlock()
							if !r {
								restart <- false // restart but don't force a new download
							}
						} else {
							exited <- true
						}
					}()
				}
			}
		}

		log.Info(logger, "👋")
		finished <- true
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().String("channel", pos.Getenv("PP_CHANNEL", ""), "the channel which can be set")
	runCmd.Flags().String("config", "", "the location of the config file")
	runCmd.Flags().StringP("dir", "d", "", "directory inside of which to run the integration")
	runCmd.Flags().String("secret", pos.Getenv("PP_AUTH_SHARED_SECRET", ""), "internal shared secret")
	runCmd.Flags().String("start-file", "", "the start file to write once running")
	runCmd.Flags().MarkHidden("secret")
	runCmd.Flags().MarkHidden("start-file")

	rootCmd.AddCommand(enrollAgentCmd)
	enrollAgentCmd.Flags().String("channel", pos.Getenv("PP_CHANNEL", ""), "the channel which can be set")
	enrollAgentCmd.Flags().String("config", "", "the location of the config file")
}
