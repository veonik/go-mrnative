// +build !debug

package mrnative

//go:generate go-bindata -pkg mrnative -tags "!debug" tpl/...

import (
	"github.com/tyler-sommer/stick"
	"io"
	"bytes"
)

func newTemplateLoader() stick.Loader {
	return &assetLoader{}
}

type assetLoader struct{}

type stringTemplate struct {
	name     string
	contents []byte
}

func (t *stringTemplate) Name() string {
	return t.name
}

func (t *stringTemplate) Contents() io.Reader {
	return bytes.NewBuffer(t.contents)
}

func (l *assetLoader) Load(name string) (stick.Template, error) {
	res, err := Asset(name)
	if err != nil {
		return nil, err
	}
	return &stringTemplate{name, res}, nil
}
