package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"

	log "github.com/sirupsen/logrus"

	"newrelic/multienv/pkg/env"
	"newrelic/multienv/pkg/env/standalone"
)

type args struct {
	ConfigFile  string
	ShowVersion bool
}

var (
	gArgs args

	/* Args below are populated via ldflags at build time */
	gIntegrationID      = ""
	gIntegrationName    = ""
	gIntegrationVersion = "0.0.0"
	gGitCommit          = ""
	gBuildDate          = ""
)

func init() {
	flag.StringVar(
		&gArgs.ConfigFile,
		"config_file",
		"",
		"path to integration configuration file",
	)

	flag.BoolVar(
		&gArgs.ShowVersion,
		"show_version",
		false,
		"display version information and exit",
	)
}

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

	flag.Parse()

	if gArgs.ShowVersion {
		fmt.Print(version(buildInfo))
		os.Exit(0)
	}

	logger.Debugf(version(buildInfo))

	configFile := gArgs.ConfigFile
	if configFile == "" {
		configFile = "./config.yaml"
	}

	env, err := env.NewEnvironment(
		buildInfo,
		env.WithLogger(logger),
		env.WithConfigFile(configFile),
	)
	fatalIfErr(logger, err)

	err = standalone.Start(context.Background(), env)
	fatalIfErr(logger, err)
}

func version(buildInfo *env.BuildInfo) string {
	return fmt.Sprintf(
		"New Relic %s Integration Version: %s, Platform: %s, GoVersion: %s, GitCommit: %s, BuildDate: %s\n",
		buildInfo.Name,
		buildInfo.Version,
		fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		runtime.Version(),
		buildInfo.GitCommit,
		buildInfo.BuildDate,
	)
}

func fatalIfErr(logger *log.Logger, err error) {
	if err != nil {
		logger.Errorf("integration failed: %v", err)
		os.Exit(1)
	}
}
