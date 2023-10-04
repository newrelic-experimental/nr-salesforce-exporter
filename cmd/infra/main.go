package main

import (
	"context"
	"fmt"
	"newrelic/multienv/pkg/env"
	"newrelic/multienv/pkg/env/infra"
	"os"
	"runtime"

	sdk_args "github.com/newrelic/infra-integrations-sdk/v4/args"
	"github.com/newrelic/infra-integrations-sdk/v4/integration"
	sdk_log "github.com/newrelic/infra-integrations-sdk/v4/log"
	"github.com/spf13/viper"
)

type argumentList struct {
	sdk_args.DefaultArgumentList
	ConfigPath  string `help:"Path to configuration"`
	ShowVersion bool   `default:"false" help:"Print build information and exit"`
}

var (
	gArgs argumentList

	/* Args below are populated via ldflags at build time */
	gIntegrationID      = ""
	gIntegrationName    = ""
	gIntegrationVersion = "0.0.0"
	gGitCommit          = ""
	gBuildDate          = ""
)

func main() {
	logger := env.NewLogrusLogger()

	logger.Debugf("starting %s integration", gIntegrationName)
	defer logger.Debugf("%s integration exited", gIntegrationName)

	buildInfo := &env.BuildInfo{
		Id:        gIntegrationID,
		Name:      gIntegrationName,
		Version:   gIntegrationVersion,
		GitCommit: gGitCommit,
		BuildDate: gBuildDate,
	}

	i, err := createIntegration(buildInfo, logger)
	fatalIfErr(logger, err)

	if gArgs.ShowVersion {
		fmt.Print(version(buildInfo))
		os.Exit(0)
	}

	logger.Debugf(version(buildInfo))

	if gArgs.ConfigPath == "" {
		fatalIfErr(logger, fmt.Errorf("no config path specified"))
	}

	env, err := env.NewEnvironment(
		buildInfo,
		env.WithLogger(logger),
		env.WithConfigFile(gArgs.ConfigPath),
	)
	fatalIfErr(logger, err)

	entity, err := getOrCreateEntity(i)
	fatalIfErr(logger, err)

	err = infra.Run(context.Background(), env, i, entity)
	fatalIfErr(logger, err)

	err = i.Publish()
	fatalIfErr(logger, err)
}

func createIntegration(
	buildInfo *env.BuildInfo,
	logger sdk_log.Logger,
) (*integration.Integration, error) {
	return integration.New(
		buildInfo.Id,
		buildInfo.Version,
		integration.Args(&gArgs),
		integration.Logger(logger),
	)
}

func getOrCreateEntity(i *integration.Integration) (*integration.Entity, error) {
	entityName := viper.GetString("entity_name")
	if entityName == "" {
		return i.HostEntity, nil
	}

	entityType := viper.GetString("entity_type")
	if entityType == "" {
		return nil, fmt.Errorf("missing entity type")
	}

	entityDisplay := viper.GetString("entity_display")
	if entityDisplay == "" {
		return nil, fmt.Errorf("missing entity display name")
	}

	// Create entity
	entity, err := i.NewEntity(entityName, entityType, entityDisplay)
	if err != nil {
		return nil, fmt.Errorf("error creating entity")
	}

	i.AddEntity(entity)

	return entity, nil
}

func version(buildInfo *env.BuildInfo) string {
	return fmt.Sprintf(
		"New Relic %s Infrastructure Integration Version: %s, Platform: %s, GoVersion: %s, GitCommit: %s, BuildDate: %s\n",
		buildInfo.Name,
		buildInfo.Version,
		fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		runtime.Version(),
		buildInfo.GitCommit,
		buildInfo.BuildDate,
	)
}

func fatalIfErr(logger sdk_log.Logger, err error) {
	if err != nil {
		logger.Errorf("integration failed: %v", err)
		os.Exit(1)
	}
}
