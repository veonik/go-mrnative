// +build !debug

package mrnative

//go:generate go-bindata -pkg mrnative -tags "!debug" tpl/...

import (
	"github.com/tyler-sommer/stick"
)

func newTemplateLoader() stick.Loader {
	return &assetLoader{}
}

type assetLoader struct{}

func (l *assetLoader) Load(name string) (string, error) {
	res, err := Asset(name)
	if err != nil {
		return "", err
	}
	return string(res), nil
}
