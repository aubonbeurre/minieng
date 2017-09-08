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
	fmt.Printf("COUCOU\n")
	if t, err := glplus.NewRGBATexture(img, true, false); err != nil {
		panic(err)
	} else {
		return TextureResource{
			Texture: t,
		}
	}
}

// Texture represents a texture loaded in the GPU RAM (by using OpenGL), which defined dimensions and viewport
type Texture struct {
	id     *glplus.Texture
	width  float32
	height float32
}

// Width returns the width of the texture.
func (t Texture) Width() float32 {
	return t.width
}

// Height returns the height of the texture.
func (t Texture) Height() float32 {
	return t.height
}

// Texture ...
func (t Texture) Texture() *glplus.Texture {
	return t.id
}

// Close ...
func (t Texture) Close() {
	glplus.Gl.DeleteTexture(t.id)
}

func init() {
	minieng.Files.Register(".jpg", &imageLoader{images: make(map[string]TextureResource)})
	minieng.Files.Register(".png", &imageLoader{images: make(map[string]TextureResource)})
	minieng.Files.Register(".gif", &imageLoader{images: make(map[string]TextureResource)})
}
