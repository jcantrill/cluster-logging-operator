package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/ViaQ/logerr/log"

	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
)

// HACK - This command is for development use only
func main() {

	image := flag.String("image", "quay.io/openshift/origin-logging-fluentd:latest", "The image to use to run the benchmark")
	totalMessages := flag.Int("totMessages", 10000, "The number of messages to write per stressor")
	msgSize := flag.Int("size", 1024, "The message size in bytes")
	verbosity := flag.Int("verbosity", 0, "")
	doCleanup := flag.Bool("docleanup", true, "set to false to preserve the namespace")
	sample := flag.Bool("sample", false, "set to true to dump a sample message")
	platform := flag.String("platform", "cluster", "The runtime environment: cluster (default), local (experimental). local requires podman")
	output := flag.String("output", "default", "The output format: default, csv")

	totLogStressors := flag.Int("exp-tot-stressors", 1, "Experimental. Total log stressors (platform=local)")
	collectorConfigPath := flag.String("exp-collector-config", "", "Experimental. The collector config to use (platform=local")

	flag.Parse()

	log.MustInit("functional-benchmark")
	log.SetLogLevel(*verbosity)
	log.V(1).Info("Starting functional benchmarker", "args", os.Args)

	if err := os.Setenv(constants.FluentdImageEnvVar, *image); err != nil {
		log.Error(err, "Error setting fluent image env var")
		os.Exit(1)
	}

	collectorConfig := ReadConfig(*collectorConfigPath)
	log.V(1).Info(collectorConfig)

	runs := map[string]string{
		"baseline": fluentdBaselineConf,
		"config":   collectorConfig,
	}

	reporter := NewReporter(*output)
	for name, config := range runs {
		log.V(1).Info("Executing", "run", name)
		stats, metrics := NewRun(config, *platform, *totLogStressors, *verbosity, *msgSize, *totalMessages, *sample, *doCleanup)()
		reporter.Add(name, stats, metrics)
	}
	reporter.Print()
}
func NewRun(collectorConfig, platform string, totLogStressors, verbosity, msgSize, totMessages int, sample, doCleanup bool) func() (Statistics, Metrics) {
	return func() (Statistics, Metrics) {
		runner := NewBencharker(collectorConfig, platform, totLogStressors, verbosity, msgSize, totMessages)
		runner.Deploy()
		if doCleanup {
			log.V(2).Info("Deferring cleanup", "doCleanup", doCleanup)
			defer runner.Cleanup()
		}

		startTime := time.Now()
		var (
			logs    []string
			readErr error
			metrics Metrics
		)
		done := make(chan bool)
		go func() {
			logs, readErr = runner.ReadApplicationLogs()
			metrics = runner.Metrics()
			done <- true
		}()
		//defer reader to get logs
		if err := runner.WritesApplicationLogsOfSize(msgSize); err != nil {
			log.Error(err, "Error writing application logs")
			os.Exit(1)
		}
		<-done
		endTime := time.Now()
		if readErr != nil {
			log.Error(readErr, "Error reading logs")
			os.Exit(1)
		}
		log.V(4).Info("Read logs", "raw", logs)
		perflogs := types.PerfLogs{}
		err := json.Unmarshal([]byte(utils.ToJsonLogs(logs)), &perflogs)
		if err != nil {
			log.Error(err, "Error parsing logs")
			os.Exit(1)
		}
		log.V(4).Info("Read logs", "parsed", perflogs)
		log.V(4).Info("Read logs", "parsed", perflogs)
		if sample {
			fmt.Printf("Sample:\n%s\n", test.JSONString(perflogs[0]))
		}
		return *NewStatisics(perflogs, msgSize, endTime.Sub(startTime)), metrics
	}
}

type Runner interface {
	Deploy()
	WritesApplicationLogsOfSize(msgSize int) error
	ReadApplicationLogs() ([]string, error)
	Metrics() Metrics
	Cleanup()
}

func NewBencharker(collectorConfig, platform string, totLogStressors, verbosity, msgSize, totMessages int) Runner {
	if platform == "local" {
		return NewPodmanRunner(collectorConfig, totLogStressors, verbosity, msgSize, totMessages)
	}
	return &ContainerRunner{verbosity: verbosity, totalMessages: totMessages}
}

func ReadConfig(configFile string) string {
	var reader func() ([]byte, error)
	switch configFile {
	case "-":
		log.V(1).Info("Reading from stdin")
		reader = func() ([]byte, error) {
			stdin := bufio.NewReader(os.Stdin)
			return ioutil.ReadAll(stdin)
		}
	case "":
		log.V(1).Info("received empty configFile. Generating from CLF")
		return ""
	default:
		log.V(1).Info("reading configfile", "filename", configFile)
		reader = func() ([]byte, error) { return ioutil.ReadFile(configFile) }
	}
	content, err := reader()
	if err != nil {
		log.Error(err, "Error reading config")
		os.Exit(1)
	}
	return string(content)
}

type Metrics struct {
	cpuUserTicks     string
	cpuKernelTicks   string
	memVirtualPeakKB string
}
