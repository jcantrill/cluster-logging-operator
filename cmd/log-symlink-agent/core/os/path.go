package os

import (
	"os"
)

type PathOperator interface {
	Stat(name string) (os.FileInfo, error)
	ReadDir(name string) ([]os.DirEntry, error)
	Remove(name string) error
	MkdirAll(path string, perm os.FileMode) error
	Link(oldName, newName string) error
}

type CoreOSPathOperator struct {}

func (c CoreOSPathOperator) Link(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

func (c CoreOSPathOperator) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (c CoreOSPathOperator) ReadDir(name string) ([]os.DirEntry, error) {
	return os.ReadDir(name)
}

func (c CoreOSPathOperator) Remove(name string) error {
	return os.Remove(name)
}

func (c CoreOSPathOperator) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}
