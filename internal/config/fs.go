package config

import (
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
)

type FileSystem interface {
	IsDir(path string) (bool, error)
	ReadFile(path string) ([]byte, error)
}

type defaultFS struct{}

func (fs defaultFS) IsDir(path string) (bool, error) {
	logrus.Debugf("Checking is directory: '%s'", path)
	file, err := os.Open(path)
	if err != nil {
		return false, nil
	}
	defer file.Close()

	stats, err := file.Stat()
	if err != nil {
		return false, err
	}

	return stats.IsDir(), nil
}

func (fs defaultFS) ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

var fsInstance FileSystem = defaultFS{}

func SetFS(replacement FileSystem) {
	fsInstance = replacement
}
