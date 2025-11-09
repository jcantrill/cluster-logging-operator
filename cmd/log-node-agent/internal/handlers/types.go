package handlers

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"

	internalos "github.com/openshift/cluster-logging-operator/cmd/log-node-agent/internal/core/os"
)

var (
	streamRE       = regexp.MustCompile(`(?P<root>^.*/)(?P<stream>(?P<namespace>[a-z][a-z0-9\-]*[a-z])_(.*)_(.*)$)`)
	namespaceIndex = streamRE.SubexpIndex("namespace")
	streamIndex    = streamRE.SubexpIndex("stream")
)

func NewContainerLogStream(targetPathRoot, basePath string) ContainerLogStream {
	log.V(5).Info("Creating ContainerLogStream object", "basePath", basePath)
	ns := extractNamespace(basePath)
	cls := ContainerLogStream{
		TargetPathRoot: targetPathRoot,
		Namespace:      ns,
		OldName:        basePath,
		NewName:        path.Join(targetPathRoot, ns, stream(basePath)),
		Os:             internalos.CoreOSPathOperator{},
	}
	log.V(5).Info("Created ContainerLogStream object", "cls", cls)
	return cls
}

type ContainerLogStream struct {
	isDir          bool
	TargetPathRoot string
	Namespace      string
	OldName        string
	NewName        string
	Os             internalos.PathOperator
}

func (l ContainerLogStream) IsDir() bool {
	if info, err := l.Os.Stat(l.OldName); err == nil {
		l.isDir = info.IsDir()
	} else {
		if !errors.Is(err, fs.ErrNotExist) {
			log.V(3).Info("Failed to stat old directory", "err", err)
		}
	}
	return l.isDir
}

func (l ContainerLogStream) DirEntries() (streams []ContainerLogStream) {
	if !l.IsDir() {
		return streams
	}
	log.V(3).Info("Reading directory entries", "path", l.OldName)
	entries, err := l.Os.ReadDir(l.OldName)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		cls := NewContainerLogStream(l.TargetPathRoot, path.Join(l.OldName, entry.Name()))
		streams = append(streams, cls)
	}
	return streams
}

func (l ContainerLogStream) RemoveNew() {
	if err := l.Os.Remove(l.NewName); err != nil {
		log.V(3).Error(err, "Error removing", "path", l.NewName)
	}
}

func (l ContainerLogStream) MakeNew() {
	log.V(5).Info("MakeNew...", "cls", l)
	if l.IsDir() {
		log.V(3).Info("Creating...", "newName", l.NewName, "permissions", PermissionsRwxRxRx)
		err := l.Os.MkdirAll(l.NewName, PermissionsRwxRxRx)
		switch {
		case errors.Is(err, fs.ErrPermission):
			panic(fmt.Sprintf("Permission error trying to create NewName: %v, cls: %v", err, l))
		case errors.Is(err, fs.ErrExist):
			log.V(4).Info("NewName already exists", "newName", l.NewName)
			return
		case err != nil:
			log.Error(err, "Failed to create NewName", "NewName", l.NewName)
		default:
			log.V(3).Info("Created", "newName", l.NewName)
			return
		}
	}
	if err := l.Os.Link(l.OldName, l.NewName); err != nil {
		switch {
		case errors.Is(err, fs.ErrExist):
			log.V(4).Info("Link already exists", "target", l.NewName)
		case errors.Is(err, fs.ErrPermission):
			panic(fmt.Sprintf("Permission error trying to create link : %v, cls: %s", err, l.NewName))
		default:
			log.Error(err, "Failed to create link", "cls", l)
		}
	}
	log.V(3).Info("Created link", "newName", l.NewName)
}

// extractNamespace returns the namespace from the path created by CRIO for a container log stream
// i.e. /var/log/pods/openshift-operator-lifecycle-manager_packageserver-5f89579495-7g57f_706603d5-69bf-4e76-9f1a-7664d919afa3
func extractNamespace(path string) string {
	if matches := streamRE.FindStringSubmatch(path); len(matches) > namespaceIndex {
		return matches[namespaceIndex]
	}
	panic(fmt.Sprintf("Unable to determine namespace from path: %s", path))
}

// stream extracts the log stream information from the log directory root (i.e. /var/log/pods)
func stream(path string) string {
	if matches := streamRE.FindStringSubmatch(path); len(matches) > streamIndex {
		return matches[streamIndex]
	}
	panic(fmt.Sprintf("Unable to determine stream from path: %s", path))
}
