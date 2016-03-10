package server

import (
	"os"
	"path"
)

func getDirectory(basedir string, id string, name string) string {
	return path.Join(basedir, name, id)
}

func prepareFilename(dir string, filename string) (string, error) {
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return "/tmp/dump", err
	}
	return path.Join(dir, filename), nil
}
