package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/pflag"

	"github.com/openshift/cluster-logging-operator/internal/pkg/generator/forwarder"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/sirupsen/logrus"
)

// HACK - This command is for development use only
func main() {

	yamlFile := flag.String("file", "", "ClusterLogForwarder yaml file. - for stdin")
	includeDefaultLogStore := flag.Bool("include-default-store", true, "Include the default storage when generating the config")
	help := flag.Bool("help", false, "This message")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		level, err := logrus.ParseLevel(logLevel)
		if err != nil {
			logrus.Errorf("Unable to evaluate the LOG_LEVEL: %s %v", logLevel, err)
			os.Exit(1)
		}
		logrus.SetLevel(level)
	}
	logger.Debugf("Args: %v", os.Args)

	if *help || len(os.Args) == 0 {
		pflag.Usage()
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		logrus.Error("Need to pass the logging forwarding yaml as an arg")
		os.Exit(1)
	}

	var reader func() ([]byte, error)
	if *yamlFile != "-" {
		reader = func() ([]byte, error) { return ioutil.ReadFile(*yamlFile) }
	} else {
		reader = func() ([]byte, error) {
			stdin := bufio.NewReader(os.Stdin)
			return ioutil.ReadAll(stdin)
		}
	}

	content, err := reader()
	if err != nil {
		logger.Errorf("Error reading file %s: %v", *yamlFile, err)
		os.Exit(1)
	}

	generatedConfig, err := forwarder.Generate(string(content), *includeDefaultLogStore)
	if err != nil {
		logger.Warnf("Unable to generate log configuration: %v", err)
		os.Exit(1)
	}
	fmt.Println(generatedConfig)
}
