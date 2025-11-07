package config

import (
	"flag"
	"os"

	"github.com/go-logr/logr"
)

const (
	LogStressorImage = "quay.io/openshift-logging/cluster-logging-load-client:0.2"
	imageVector      = "quay.io/openshift-logging/vector:v0.47.0"
)

type Options struct {
	// SourcePath is the directory to watch
	SourcePath string

	//DestPath is the directory where to create symlinks based upon the entries in the sourcepath
	DestPath string

	ExcludeNamespace string
}

func InitOptions(log logr.Logger) Options {
	options := Options{}
	fs := flag.NewFlagSet("log-symlink-agent", flag.ExitOnError)

	fs.StringVar(&options.ExcludeNamespace, "exclude-namespace", "", "Optional. A namespace to exclude when handling symlinks")
	fs.StringVar(&options.SourcePath, "src-path", "", "Required. The source path to watch for directories")
	fs.StringVar(&options.DestPath, "dest-path", "", "Required. The destination path where to create symlinks")
	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Error(err, "Error parsing arguments")
		fs.Usage()
		os.Exit(1)
	}

	log.V(3).Info("Parsed options", "options", options)
	if options.SourcePath == "" || options.DestPath == "" {
		fs.Usage()
		os.Exit(1)
	}
	return options
}
