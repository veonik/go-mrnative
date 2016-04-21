// +build debug

package mrnative

import (
	"os"

	"github.com/tyler-sommer/stick"
)

var rootDir string

func init() {
	r, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	rootDir = r
}

func newTemplateLoader() stick.Loader {
	return stick.NewFilesystemLoader(rootDir)
}
