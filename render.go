package minieng

import (
	"image/color"
	"sort"

	"github.com/aubonbeurre/glplus"
)

const (
	// RenderSystemPriority ...
	RenderSystemPriority = -1000
)

type renderChangeMessage struct{}

func (renderChangeMessage) Type() string {
	return "renderChangeMessage"
}

// Drawable ...
type Drawable interface {
	Setup()
	Draw(td float32)
	Delete()
}

// RenderComponent ...
type RenderComponent struct {
	Drawable Drawable
	zIndex   float32
}

// Component ...
type Component struct {
	BasicEntity
	RenderComponent
}

// SetZIndex ...
func (r *RenderComponent) SetZIndex(index float32) {
	r.zIndex = index
	Mailbox.Dispatch(&renderChangeMessage{})
}

type renderEntity struct {
	*BasicEntity
	*RenderComponent
}

type renderEntityList []renderEntity

func (r renderEntityList) Len() int {
	return len(r)
}

func (r renderEntityList) Less(i, j int) bool {
	if r[i].RenderComponent.zIndex == r[j].RenderComponent.zIndex {
		return r[i].ID() < r[j].ID()
	}

	return r[i].RenderComponent.zIndex < r[j].RenderComponent.zIndex
}

func (r renderEntityList) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

// RenderSystem ...
type RenderSystem struct {
	entities renderEntityList
	world    *World

	sortingNeeded bool
	//currentShader Shader
}

// Priority ...
func (*RenderSystem) Priority() int { return RenderSystemPriority }

// New ...
func (rs *RenderSystem) New(w *World) {
	rs.world = w

	//addCameraSystemOnce(w)

	//initShaders(w)
	//engo.Gl.Enable(engo.Gl.MULTISAMPLE)

	Mailbox.Listen("renderChangeMessage", func(Message) {
		rs.sortingNeeded = true
	})
}

// Add ...
func (rs *RenderSystem) Add(basic *BasicEntity, render *RenderComponent) {
	rs.entities = append(rs.entities, renderEntity{BasicEntity: basic, RenderComponent: render})
	render.Drawable.Setup()
	rs.sortingNeeded = true
}

// RemoveAll ...
func (rs *RenderSystem) RemoveAll() {
	for len(rs.entities) > 0 {
		rs.Remove(*rs.entities[len(rs.entities)-1].GetBasicEntity())
	}
}

// Remove ...
func (rs *RenderSystem) Remove(basic BasicEntity) {
	var delete = -1
	for index, entity := range rs.entities {
		if entity.ID() == basic.ID() {
			delete = index
			break
		}
	}
	if delete >= 0 {
		rs.entities[delete].Drawable.Delete()
		rs.entities = append(rs.entities[:delete], rs.entities[delete+1:]...)
		rs.sortingNeeded = true
	}
}

// Update ...
func (rs *RenderSystem) Update(dt float32) {
	if rs.sortingNeeded {
		sort.Sort(rs.entities)
		rs.sortingNeeded = false
	}
	Gl := glplus.Gl
	Gl.Clear(Gl.COLOR_BUFFER_BIT | Gl.DEPTH_BUFFER_BIT)

	for _, e := range rs.entities {
		if e.Drawable != nil {
			e.Drawable.Draw(dt)
		}
	}
}

// SetBackground ...
func SetBackground(c color.Color) {
	r, g, b, a := c.RGBA()

	Gl := glplus.Gl
	Gl.ClearColor(float32(r)/0xffff, float32(g)/0xffff, float32(b)/0xffff, float32(a)/0xffff)
}
