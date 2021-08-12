package config

import "io/ioutil"

type FileSystem interface {
	ReadFile(path string) ([]byte, error)
}

type defaultFS struct{}

func (fs defaultFS) ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

var fsInstance FileSystem = defaultFS{}

func SetFS(replacement FileSystem) {
	fsInstance = replacement
}
