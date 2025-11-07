package handlers

import (
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
		log.V(3).Info("Added directory watch", "path", cls.OldName)
		h.watcher.Add(cls.OldName)
		for _, entry := range cls.DirEntries() {
			h.Create(entry)
		}
	}

	//targetRoot := path.Dir(cls.NewName)
	//if _, err := h.Stat(targetRoot); !errors.Is(err, fs.ErrExist) {
	//	if err = h.MkdirAll(targetRoot, PermissionsRwxRxRx); err != nil {
	//		log.V(4).Info("errors.Is(err, fs.ErrPermission)", "value", errors.Is(err, fs.ErrPermission), "err", err)
	//		switch {
	//		case errors.Is(err, fs.ErrPermission):
	//			panic(fmt.Sprintf("Permission error trying to create directory for namespace: %v, dir: %s", err, targetRoot))
	//		default:
	//			log.Error(err, "Failed to create directory for namespace", "dir", targetRoot, "err", err)
	//			return
	//		}
	//	}
	//	log.V(3).Info("Created directory for namespace", "dir", targetRoot)
	//}
	//if err := h.Link(cls.OldName, cls.NewName); err != nil {
	//	log.V(4).Info("errors.Is(err, fs.ErrPermission)", "value", errors.Is(err, fs.ErrPermission), "err", err)
	//	switch {
	//	case errors.Is(err, fs.ErrExist):
	//		log.V(4).Info("Link already exists", "target", cls.NewName)
	//	case errors.Is(err, fs.ErrPermission):
	//		panic(fmt.Sprintf("Permission error trying to create link for container log directory: %v, cls: %s", err, cls))
	//	default:
	//		log.Error(err, "Failed to create link", "cls", cls, "target", cls.NewName)
	//		return
	//	}
	//}
	//log.V(3).Info("Created link", "cls", cls, "target", cls.NewName)
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
	log.V(3).Info("Handling remove event", "path", cls.OldName)
	if cls.IsDir() {
		if err := h.watcher.Remove(cls.NewName); err != nil {
			log.V(3).Error(err, "Error removing watch", "path", cls.NewName)
		}
	}
	cls.RemoveNew()
	log.V(3).Info("Removed directory", "path", cls.OldName)
}
