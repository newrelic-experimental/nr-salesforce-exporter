package env

import (
	"context"
	"fmt"
	"newrelic/multienv/pkg/config"
	"os"
	"time"

	nr_cli "github.com/newrelic/newrelic-client-go/newrelic"
	nr_cl_config "github.com/newrelic/newrelic-client-go/pkg/config"
	"github.com/newrelic/newrelic-client-go/pkg/logging"
	"github.com/newrelic/newrelic-client-go/pkg/region"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type envOpt func(env *Environment)

type BuildInfo struct {
	Id        string
	Name      string
	Version   string
	GitCommit string
	BuildDate string
}

type Environment struct {
	logger        *log.Logger
	nrClient      *nr_cli.NewRelic
	config        *config.Config
	buildInfo     BuildInfo
	configFile    string
	configPaths   []string
	eventsEnabled bool
}

func configLicenseKey(licenseKey string) nr_cli.ConfigOption {
	return func(cfg *nr_cl_config.Config) error {
		cfg.LicenseKey = licenseKey
		return nil
	}
}

func NewLogrusLogger() *log.Logger {
	logger := log.New()
	logger.SetLevel(log.WarnLevel)

	return logger
}

func NewEnvironment(
	buildInfo *BuildInfo,
	opts ...envOpt,
) (*Environment, error) {
	env := &Environment{}
	env.buildInfo = *buildInfo

	for _, opt := range opts {
		opt(env)
	}

	if env.logger == nil {
		env.logger = NewLogrusLogger()
	}

	err := env.setupConfig()
	if err != nil {
		return nil, err
	}

	err = env.setupClient()
	if err != nil {
		return nil, err
	}

	return env, nil
}

func WithLogger(log *log.Logger) envOpt {
	return func(env *Environment) { env.logger = log }
}

func WithConfigFile(configFile string) envOpt {
	return func(env *Environment) { env.configFile = configFile }
}

func (e *Environment) BuildInfo() BuildInfo {
	return e.buildInfo
}

func (e *Environment) Dispose() {
	e.maybeFlushEvents()
}

func (env *Environment) EnableEvents(accountID int) error {
	if env.eventsEnabled {
		return nil
	}

	// Start batch mode
	if err := env.nrClient.Events.BatchMode(
		context.Background(),
		accountID,
	); err != nil {
		return fmt.Errorf("error starting batch events mode: %v", err)
	}

	env.eventsEnabled = true

	return nil
}

func (env *Environment) maybeFlushEvents() error {
	if !env.eventsEnabled {
		return nil
	}

	err := env.nrClient.Events.Flush()
	if err != nil {
		return err
	}

	<-time.After(3 * time.Second)

	return nil
}

func (e *Environment) setupConfig() error {
	if e.configFile != "" {
		config, err := config.NewConfigWithFile(e.configFile)
		if err != nil {
			return err
		}

		e.config = config

		return nil
	}

	config, err := config.NewConfigWithPaths(e.configPaths)
	if err != nil {
		return err
	}

	e.config = config

	return nil
}

func (e *Environment) setupClient() error {
	licenseKey := viper.GetString("licenseKey")
	if licenseKey == "" {
		licenseKey = os.Getenv("NEW_RELIC_LICENSE_KEY")
		if licenseKey == "" {
			return fmt.Errorf("missing New Relic license key")
		}
	}

	apiKey := viper.GetString("apiKey")
	if apiKey == "" {
		apiKey = os.Getenv("NEW_RELIC_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("missing New Relic API key")
		}
	}

	nrRegion := viper.GetString("region")
	if nrRegion == "" {
		nrRegion = os.Getenv("NEW_RELIC_REGION")
		if nrRegion == "" {
			nrRegion = string(region.Default)
		}
	}

	// Initialize the New Relic Go Client
	client, err := nr_cli.New(
		configLicenseKey(licenseKey),
		nr_cli.ConfigPersonalAPIKey(apiKey),
		nr_cli.ConfigRegion(nrRegion),
		nr_cli.ConfigLogger(
			logging.NewLogrusLogger(logging.ConfigLoggerInstance(e.logger)),
		),
	)
	if err != nil {
		return fmt.Errorf("error creating New Relic client: %v", err)
	}

	e.nrClient = client

	return nil
}

func (c *Environment) Logger() *log.Logger {
	return c.logger
}

func (c *Environment) Client() *nr_cli.NewRelic {
	return c.nrClient
}

func (c *Environment) Config() *config.Config {
	return c.config
}
