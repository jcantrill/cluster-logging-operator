package handlers

import (
	"errors"
	"os"

	"github.com/fsnotify/fsnotify"
)

type Handler struct {
	watcher   *fsnotify.Watcher
	excludeNS string
	destPath  string
	MkdirAll  func(string, os.FileMode) error
	Stat      func(string) (os.FileInfo, error)
	Link      func(string, string) error
	RemoveAll func(string) error
}

func New(watcher *fsnotify.Watcher, excludeNS string) Handler {
	return Handler{
		watcher:   watcher,
		excludeNS: excludeNS,
		MkdirAll:  os.MkdirAll,
		Stat:      os.Stat,
		Link:      os.Link,
		RemoveAll: os.RemoveAll,
	}
}

const (
	PermissionsRwxRxRx = 0755
)

// Create
// Dir-> container_dirs -> log files
func (h Handler) Create(cls ContainerLogStream) {
	if h.excludeNS == cls.Namespace {
		log.V(3).Info("Ignoring create event which matches exludedNS", "path", cls)
		return
	}
	log.V(3).Info("Handling create event", "path", cls)
	cls.MakeNew()
	if cls.IsDir() {
		log.V(3).Info("Adding watch", "path", cls.OldName)
		h.watcher.Add(cls.OldName)
	}
	log.V(5).Info("Watch list", "list", h.watcher.WatchList())
	for _, entry := range cls.DirEntries() {
		h.Create(entry)
	}

}

// Rename event is the same as remove.  The watcher creates a rename (old) and create (new) events
func (h Handler) Rename(cls ContainerLogStream) {
	if h.excludeNS == cls.Namespace {
		return
	}
	log.V(3).Info("Handling rename by removing", "path", cls)
	h.Remove(cls)
}

func (h Handler) Remove(cls ContainerLogStream) {
	if h.excludeNS == cls.Namespace {
		return
	}
	log.V(5).Info("Watch list", "list", h.watcher.WatchList())
	log.V(3).Info("Handling remove event", "path", cls.OldName)
	if err := h.watcher.Remove(cls.NewName); err != nil {
		if !errors.Is(err, fsnotify.ErrNonExistentWatch) {
			log.V(3).Error(err, "Error removing watch", "path", cls.NewName)
		}
	} else {
		log.V(3).Info("Removed watch", "path", cls.NewName)
	}
	cls.RemoveNew()
	log.V(3).Info("Removed directory", "path", cls.OldName)
}
