package common

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/gif" // jus because
	_ "image/jpeg"
	_ "image/png"
	"io"

	"github.com/aubonbeurre/glplus"
	"github.com/aubonbeurre/minieng"
)

// TextureResource ...
type TextureResource struct {
	Texture *glplus.GPTexture
	Img     *image.RGBA
	url     string
}

// URL ...
func (t TextureResource) URL() string {
	return t.url
}

type imageLoader struct {
	images map[string]TextureResource
}

func (i *imageLoader) Load(url string, data io.Reader) error {
	img, _, err := image.Decode(data)
	if err != nil {
		return err
	}

	b := img.Bounds()
	newm := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(newm, newm.Bounds(), img, b.Min, draw.Src)

	i.images[url] = NewTextureResource(newm)

	return nil
}

func (i *imageLoader) Unload(url string) error {
	delete(i.images, url)
	return nil
}

func (i *imageLoader) Resource(url string) (minieng.Resource, error) {
	texture, ok := i.images[url]
	if !ok {
		return nil, fmt.Errorf("resource not loaded by `FileLoader`: %q", url)
	}

	return texture, nil
}

// NewTextureResource sends the image to the GPU and returns a `TextureResource` for easy access
func NewTextureResource(img *image.RGBA) TextureResource {
	if t, err := glplus.NewRGBATexture(img, true, false); err != nil {
		panic(err)
	} else {
		return TextureResource{
			Texture: t,
			Img:     img,
		}
	}
}

func init() {
	minieng.Files.Register(".jpg", &imageLoader{images: make(map[string]TextureResource)})
	minieng.Files.Register(".png", &imageLoader{images: make(map[string]TextureResource)})
	minieng.Files.Register(".gif", &imageLoader{images: make(map[string]TextureResource)})
}
