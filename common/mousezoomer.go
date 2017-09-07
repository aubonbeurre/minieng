package common

import (
	"github.com/aubonbeurre/minieng"
)

const (
	// MouseZoomerPriority ...
	MouseZoomerPriority = 110
)

// MouseZoomerMessage ...
type MouseZoomerMessage struct {
	ScrollY float32
}

// Type ...
func (MouseZoomerMessage) Type() string {
	return "MouseZoomerMessage"
}

// MouseZoomer is a System that allows for zooming when the scroll wheel is used
type MouseZoomer struct {
}

// Priority ...
func (*MouseZoomer) Priority() int { return MouseZoomerPriority }

// Remove ...
func (*MouseZoomer) Remove(minieng.BasicEntity) {}

// Update ...
func (c *MouseZoomer) Update(float32) {
	if minieng.Input.Mouse.ScrollY != 0 {
		minieng.Mailbox.Dispatch(MouseZoomerMessage{ScrollY: minieng.Input.Mouse.ScrollY})
	}
}
