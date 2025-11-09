package handlers_test

import (
	"io/fs"
	"os"
	"time"
)

type FakePathOperator struct {
	ReadDirFn func(name string) ([]os.DirEntry, error)
	StatFn    func(name string) (os.FileInfo, error)
}

func (f FakePathOperator) Stat(name string) (os.FileInfo, error) {
	return f.StatFn(name)
}

func (f FakePathOperator) ReadDir(name string) ([]os.DirEntry, error) {
	return f.ReadDirFn(name)
}

func (f FakePathOperator) Remove(name string) error {
	panic("implement me")
}

func (f FakePathOperator) MkdirAll(path string, perm os.FileMode) error {
	panic("implement me")
}

func (f FakePathOperator) Link(oldName, newName string) error {
	panic("implement me")
}

type FakeFileInfo struct {
	IsADir bool
}

func (f FakeFileInfo) Name() string {
	panic("implement me")
}

func (f FakeFileInfo) Size() int64 {
	panic("implement me")
}

func (f FakeFileInfo) Mode() fs.FileMode {
	panic("implement me")
}

func (f FakeFileInfo) ModTime() time.Time {
	panic("implement me")
}

func (f FakeFileInfo) IsDir() bool {
	return f.IsADir
}

func (f FakeFileInfo) Sys() any {
	panic("implement me")
}

type FakeDirectoryEntry struct {
	AName     string
	Directory bool
}

func (fd FakeDirectoryEntry) Name() string {
	return fd.AName
}

func (fd FakeDirectoryEntry) IsDir() bool {
	return fd.Directory
}

func (fd FakeDirectoryEntry) Type() os.FileMode {
	panic("implement me")
}

func (fd FakeDirectoryEntry) Info() (os.FileInfo, error) {
	panic("implement me")
}
