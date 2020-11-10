package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ViaQ/logerr/log"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers"
)

// HACK - This command is for development use only
func main() {

	image := flag.String("image", "quay.io/openshift/origin-logging-fluentd:latest", "The image to use to run the benchmark")
	totalMessages := flag.Uint64("totMessages", 10000, "The number of messages to write")
	msgSize := flag.Uint64("size", 1024, "The message size in bytes")
	verbosity := flag.Int("verbosity", 0, "")
	doCleanup := flag.Bool("docleanup", true, "set to false to preserve the namespace")
	sample := flag.Bool("sample", false, "set to true to dump a sample message")

	flag.Parse()

	log.MustInit("functional-benchmark")
	log.SetLogLevel(*verbosity)
	log.V(1).Info("Args: %v", os.Args)

	if err := os.Setenv(constants.FluentdImageEnvVar, *image); err != nil {
		log.Error(err, "Error setting fluent image env var")
		os.Exit(1)
	}
	testclient := client.NewHackTest()
	framework := functional.NewFluentdFunctionalFrameworkUsing(&testclient.Test, testclient.Close, *verbosity)
	if *doCleanup {
		defer framework.Cleanup()
	}

	functional.NewClusterLogForwarderBuilder(framework.Forwarder).
		FromInput(logging.InputNameApplication).
		ToFluentForwardOutput()
	if err := framework.Deploy(); err != nil {
		log.Error(err, "Error deploying test pod")
		os.Exit(1)
	}
	startTime := time.Now()
	var (
		logs    []string
		readErr error
	)
	done := make(chan bool)
	go func() {
		logs, readErr = framework.ReadNApplicationLogsFrom(*totalMessages, functional.ForwardOutputName)
		done <- true
	}()
	//defer reader to get logs
	if err := framework.WritesNApplicationLogsOfSize(*totalMessages, *msgSize); err != nil {
		log.Error(err, "Error writing logs to test pod")
		os.Exit(1)
	}
	<-done
	endTime := time.Now()
	if readErr != nil {
		log.Error(readErr, "Error reading logs")
		os.Exit(1)
	}
	jsonlogs, err := helpers.ParseLogs(fmt.Sprintf("[%s]", strings.Join(logs, ",")))
	if err != nil {
		log.Error(err, "Error parsing logs")
		os.Exit(1)
	}
	if *sample {
		fmt.Printf("Sample:\n%s\n", test.JSONString(jsonlogs[0]))
	}
	fmt.Printf("  Total Msg: %d\n", *totalMessages)
	fmt.Printf("Size(bytes): %d\n", *msgSize)
	fmt.Printf(" Elapsed(s): %s\n", endTime.Sub(startTime))
	fmt.Printf("    Mean(s): %f\n", mean(jsonlogs))
	fmt.Printf("     Min(s): %f\n", min(jsonlogs))
	fmt.Printf("     Max(s): %f\n", max(jsonlogs))
	fmt.Printf("  Median(s): %f\n", median(jsonlogs))
	fmt.Printf(" Mean Bloat: %f\n", meanBloat(jsonlogs))
}

func meanBloat(logs helpers.Logs) float64 {
	if len(logs) == 0 {
		return 0
	}
	total := float64(0)
	for _, e := range logs {
		total += e.Bloat()
	}
	return total / float64(len(logs))
}

func mean(logs helpers.Logs) float64 {
	total := float64(0)
	for _, e := range logs {
		total += e.Difference()
	}
	return total / float64(len(logs))
}

func median(logs helpers.Logs) float64 {
	diffs := sortLogsByTimeDiff(logs)
	return diffs[(len(diffs)/2)+1]
}
func min(logs helpers.Logs) float64 {
	diffs := sortLogsByTimeDiff(logs)
	return diffs[0]
}
func max(logs helpers.Logs) float64 {
	diffs := sortLogsByTimeDiff(logs)
	return diffs[len(diffs)-1]
}
func sortLogsByTimeDiff(logs helpers.Logs) []float64 {
	diffs := make([]float64, len(logs))
	for i, e := range logs {
		diffs[i] = e.Difference()
	}
	sort.Slice(diffs, func(i, j int) bool { return diffs[i] < diffs[j] })
	return diffs
}
