package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/openshift/cluster-logging-operator/internal/pkg/generator/forwarder"
	"github.com/spf13/pflag"

	log "github.com/openshift/cluster-logging-operator/pkg/logger"
)

// HACK - This command is for development use only
func main() {

	yamlFile := flag.String("file", "", "LogForwarding yaml file. - for stdin")
	includeDefaultLogStore := flag.Bool("include-default-store", true, "Include the default storage when generating the config")
	includeLegacyForward := flag.Bool("include-legacy-forward", true, "Include the legacy forward when generating the config")
	help := flag.Bool("help", false, "This message")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	log.Debugf("Args: %v", os.Args)

	if *help {
		pflag.Usage()
		os.Exit(1)
	}

	var reader func() ([]byte, error)
	switch *yamlFile {
	case "-":
		reader = func() ([]byte, error) { return ioutil.ReadFile(*yamlFile) }
	case "":
		reader = func() ([]byte, error) { return []byte{}, nil }
	default:
		reader = func() ([]byte, error) {
			stdin := bufio.NewReader(os.Stdin)
			return ioutil.ReadAll(stdin)
		}
	}

	content, err := reader()
	if err != nil {
		log.Error(err, "Error Unmarshalling file ", "file", yamlFile)
		os.Exit(1)
	}

	generatedConfig, err := forwarder.Generate(string(content), *includeDefaultLogStore, *includeLegacyForward)
	if err != nil {
		log.Error(err, "Unable to generate log configuration")
		os.Exit(1)
	}
	fmt.Println(generatedConfig)
}
