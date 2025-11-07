package handlers

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"
)

var (
	streamRE       = regexp.MustCompile(`(?P<root>^.*/)(?P<stream>(?P<namespace>[a-z][a-z0-9\-]*[a-z])_(.*)_(.*)$)`)
	namespaceIndex = streamRE.SubexpIndex("namespace")
	streamIndex    = streamRE.SubexpIndex("stream")
)

func NewLogContainerStream(targetPathRoot, basePath string) (ContainerLogStream, error) {
	log.V(5).Info("Creating ContainerLogStream", "basePath", basePath)
	ns := extractNamespace(basePath)
	cls := ContainerLogStream{
		Namespace: ns,
		OldName:   basePath,
		NewName:   path.Join(targetPathRoot, ns, stream(basePath)),
	}
	log.V(5).Info("Created ContainerLogStream", "cls", cls)
	return cls, nil
}

type ContainerLogStream struct {
	isDir          bool
	targetPathRoot string
	Namespace      string
	OldName        string
	NewName        string
}

func (l ContainerLogStream) IsDir() bool {
	if info, err := os.Stat(l.OldName); err == nil {
		l.isDir = info.IsDir()
	} else {
		log.V(3).Info("Failed to stat old directory", "err", err)
	}
	return l.isDir
}

func (l ContainerLogStream) DirEntries() (streams []ContainerLogStream) {
	entries, err := os.ReadDir(l.OldName)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		cls, err := NewLogContainerStream(l.NewName, entry.Name())
		if err != nil {
			log.Error(err, "Error creating containerLogStream", "path", entry.Name())
		} else {
			streams = append(streams, cls)
		}
	}
	return streams
}

func (l ContainerLogStream) RemoveNew() {
	if err := os.Remove(l.NewName); err != nil {
		log.V(3).Error(err, "Error removing", "path", l.NewName)
	}
}
func (l ContainerLogStream) MakeNew() {
	log.V(5).Info("MakeNew...", "cls", l)
	if l.IsDir() {
		_, err := os.Stat(l.NewName)
		switch {
		case errors.Is(err, fs.ErrExist):
			log.V(4).Info("NewName already exists", "newName", l.NewName)
		default:
			if err = os.MkdirAll(l.NewName, PermissionsRwxRxRx); err != nil {
				log.Error(err, "Failed to create NewName", "cls", l)
				if errors.Is(err, fs.ErrPermission) {
					panic(fmt.Sprintf("Permission error trying to create NewName: %v, cls: %v", err, l))
				}
			}
		}
		log.V(3).Info("Created", "newName", l.NewName)
		return
	}
	if err := os.Link(l.OldName, l.NewName); err != nil {
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
func stream(path string) string {
	if matches := streamRE.FindStringSubmatch(path); len(matches) > streamIndex {
		return matches[streamIndex]
	}
	panic(fmt.Sprintf("Unable to determine stream from path: %s", path))
}
