package common

import (
	"bytes"
	"fmt"
	"io"

	"github.com/aubonbeurre/minieng"
)

// BytesResource ...
type BytesResource struct {
	Buffer *bytes.Buffer
	url    string
}

// URL ...
func (t BytesResource) URL() string {
	return t.url
}

type stuffLoader struct {
	bytes map[string]BytesResource
}

func (i *stuffLoader) Load(url string, data io.Reader) error {
	buf := &bytes.Buffer{}
	if _, err := buf.ReadFrom(data); err != nil {
		return err
	}

	i.bytes[url] = NewBytesResource(buf)

	return nil
}

func (i *stuffLoader) Unload(url string) error {
	delete(i.bytes, url)
	return nil
}

func (i *stuffLoader) Resource(url string) (minieng.Resource, error) {
	stuff, ok := i.bytes[url]
	if !ok {
		return nil, fmt.Errorf("resource not loaded by `FileLoader`: %q", url)
	}

	return stuff, nil
}

// NewBytesResource sends the image to the GPU and returns a `BytesResource` for easy access
func NewBytesResource(buf *bytes.Buffer) BytesResource {
	return BytesResource{
		Buffer: buf,
	}
}

func init() {
	minieng.Files.Register(".json", &stuffLoader{bytes: make(map[string]BytesResource)})
	minieng.Files.Register(".obj", &stuffLoader{bytes: make(map[string]BytesResource)})
}
