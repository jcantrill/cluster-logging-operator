package main

import (
	"errors"
	"os"
	"path"

	"github.com/fsnotify/fsnotify"
	"github.com/openshift/cluster-logging-operator/cmd/log-node-agent/internal/config"
	"github.com/openshift/cluster-logging-operator/cmd/log-node-agent/internal/handlers"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

var (
	log = utils.InitStaticLogger("log-symlink-agent")
)

func main() {

	options := config.InitOptions(log)
	log.V(0).Info("Starting agent...", "options", options)

	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error(err, "Failed creating file watcher")
		os.Exit(1)
	}
	defer watcher.Close()

	// Start listening for events.
	handler := handlers.New(watcher, options.ExcludeNamespace)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					log.V(3).Info("Unable to get watcher events", "ok", ok, "event", event)
					break
				}
				log.V(3).Info("Received event", "event", event, "ok", ok)

				if event.Op.Has(fsnotify.Create) || event.Op.Has(fsnotify.Rename) || event.Op.Has(fsnotify.Remove) {
					log.V(5).Info("Processing event notification", "event", event.String())
				} else {
					log.V(5).Info("Ignoring event", "event", event.String())
					break
				}
				//create, //rename //remove
				// assume single operation
				logStream := handlers.NewContainerLogStream(options.DestPath, event.Name)
				switch {
				case event.Op.Has(fsnotify.Create):
					handler.Create(logStream)
				case event.Op.Has(fsnotify.Rename):
					handler.Rename(logStream)
				case event.Op.Has(fsnotify.Remove):
					handler.Remove(logStream)
				default:
					log.Info("Unhandle event", "event", event, "cls", logStream)
				}
			case watchError, ok := <-watcher.Errors:
				evaluateWatchError(watchError, ok)
			}
		}
	}()

	// Add a path.
	err = watcher.Add(options.SourcePath) //may need buffer configurability?
	if err != nil {
		log.Error(err, "Failed adding the src-path to the watcher", "path", options.SourcePath)
	}

	initTargetDir(handler, options.SourcePath, options.DestPath, 2, 0)

	<-make(chan struct{})
}

func evaluateWatchError(err error, ok bool) {
	if !ok {
		log.V(3).Info("Unable to get watcher errors", "ok", ok, "error", err)
		return
	}
	switch {
	case errors.Is(err, fsnotify.ErrEventOverflow):
		log.Error(err, "Too many queued events. Consider increasing 'fs.inotify.max_queued_events' using 'sysctl'")
	default:
		panic(err)
	}
	log.V(1).Error(err, "watch error")
}

func initTargetDir(handler handlers.Handler, sourcePath, destPath string, maxRecurse, currentDepth int) {
	log.V(3).Info("Initializing target directory", "path", sourcePath)
	if currentDepth > maxRecurse {
		log.V(3).Info("Achieved max depth", "currentDepth", currentDepth, "maxRecurse", maxRecurse, "sourcePath", sourcePath)
		return
	}
	entries, err := os.ReadDir(sourcePath)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		logStream := handlers.NewContainerLogStream(destPath, path.Join(sourcePath, entry.Name()))
		handler.Create(logStream)
		currentDepth += 1
		initTargetDir(handler, path.Join(sourcePath, entry.Name()), destPath, maxRecurse, currentDepth)
	}
}
