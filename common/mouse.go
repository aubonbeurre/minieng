package common

import (
	"image"

	"github.com/aubonbeurre/minieng"
)

// Cursor is a reference to a GLFW-cursor - to be used with the `SetCursor` method.
type Cursor uint8

const (
	// CursorNone ...
	CursorNone = iota
	// CursorArrow ...
	CursorArrow
	// CursorCrosshair ...
	CursorCrosshair
	// CursorHand ...
	CursorHand
	// CursorIBeam ...
	CursorIBeam
	// CursorHResize ...
	CursorHResize
	// CursorVResize ...
	CursorVResize
)

// MouseSystemPriority ...
const MouseSystemPriority = 100

// Mouse is the representation of the physical mouse
type Mouse struct {
	// X is the current x position of the mouse in the game
	X float32
	// Y is the current y position of the mouse in the game
	Y float32
	// ScrollX is the current scrolled position on the x component
	ScrollX float32
	// ScrollY is the current scrolled position on the y component
	ScrollY float32
	// Action is the currently active Action
	Action minieng.Action
	// Button is which button is being pressed on the mouse
	Button minieng.MouseButton
	// Modifier is whether any modifier mouse buttons are being pressed
	Modifer minieng.Modifier
}

// MouseComponent is the location for the MouseSystem to store its results;
// to be used / viewed by other Systems
type MouseComponent struct {
	// Clicked is true whenever the Mouse was clicked over
	// the entity space in this frame
	Clicked bool
	// Released is true whenever the left mouse button is released over the
	// entity space in this frame
	Released bool
	// Hovered is true whenever the Mouse is hovering
	// the entity space in this frame. This does not necessarily imply that
	// the mouse button was pressed down in your entity space.
	Hovered bool
	// Dragged is true whenever the entity space was left-clicked,
	// and then the mouse started moving (while holding)
	Dragged bool
	// RightClicked is true whenever the entity space was right-clicked
	// in this frame
	RightClicked bool
	// RightDragged is true whenever the entity space was right-clicked,
	// and then the mouse started moving (while holding)
	RightDragged bool
	// RightReleased is true whenever the right mouse button is released over
	// the entity space in this frame. This does not necessarily imply that
	// the mouse button was pressed down in your entity space.
	RightReleased bool
	// Enter is true whenever the Mouse entered the entity space in that frame,
	// but wasn't in that space during the previous frame
	Enter bool
	// Leave is true whenever the Mouse was in the space on the previous frame,
	// but now isn't
	Leave bool
	// Position of the mouse at any moment this is generally used
	// in conjunction with Track = true
	MouseX float32
	MouseY float32
	// Set manually this to true and your mouse component will track the mouse
	// and your entity will always be able to receive an updated mouse
	// component even if its space is not under the mouse cursor
	// WARNING: you MUST know why you want to use this because it will
	// have serious performance impacts if you have many entities with
	// a MouseComponent in tracking mode.
	// This is ideally used for a really small number of entities
	// that must really be aware of the mouse details event when the
	// mouse is not hovering them
	Track bool
	// Modifier is used to store the eventual modifiers that were pressed during
	// the same time the different click events occurred
	Modifier minieng.Modifier

	// startedDragging is used internally to see if *this* is the object that is being dragged
	startedDragging bool
	// startedRightDragging is used internally to see if *this* is the object that is being right-dragged
	rightStartedDragging bool
}

// SpaceComponent ...
type SpaceComponent struct {
	Bounds image.Rectangle
}

// Contains ...
func (m *SpaceComponent) Contains(pt image.Point) bool {
	return image.Rect(pt.X, pt.Y, pt.X+1, pt.Y+1).In(m.Bounds)
}

type mouseEntity struct {
	*minieng.BasicEntity
	*MouseComponent
	*SpaceComponent
	*RenderComponent
}

// MouseSystem listens for mouse events, and changes value for MouseComponent accordingly
type MouseSystem struct {
	entities []mouseEntity
	world    *minieng.World

	mouseX    float32
	mouseY    float32
	mouseDown bool
}

// Priority returns a priority higher than most, to ensure that this System runs before all others
func (m *MouseSystem) Priority() int { return MouseSystemPriority }

// New ...
func (m *MouseSystem) New(w *minieng.World) {
	m.world = w
}

// Add adds a new entity to the MouseSystem.
// * RenderComponent is only required if you're using the HUDShader on this Entity.
// * SpaceComponent is required whenever you want to know specific mouse-events on this Entity (like hover,
//   click, etc.). If you don't need those, then you can omit the SpaceComponent.
// * MouseComponent is always required.
// * BasicEntity is always required.
func (m *MouseSystem) Add(basic *minieng.BasicEntity, mouse *MouseComponent, space *SpaceComponent, render *RenderComponent) {
	m.entities = append(m.entities, mouseEntity{basic, mouse, space, render})
}

// Remove ...
func (m *MouseSystem) Remove(basic minieng.BasicEntity) {
	var delete = -1
	for index, entity := range m.entities {
		if entity.ID() == basic.ID() {
			delete = index
			break
		}
	}
	if delete >= 0 {
		m.entities = append(m.entities[:delete], m.entities[delete+1:]...)
	}
}

// Update ...
func (m *MouseSystem) Update(dt float32) {
	// Translate Mouse.X and Mouse.Y into "game coordinates"
	switch minieng.Backend {
	case "GLFW":
		m.mouseX = minieng.Input.Mouse.X
		m.mouseY = minieng.Input.Mouse.Y
	case "Mobile":
		m.mouseX = minieng.Input.Mouse.X
		m.mouseY = minieng.Input.Mouse.Y
	case "Web":
		m.mouseX = minieng.Input.Mouse.X
		m.mouseY = minieng.Input.Mouse.Y
	}

	for _, e := range m.entities {
		// Reset all values except these
		*e.MouseComponent = MouseComponent{
			Track:                e.MouseComponent.Track,
			Hovered:              e.MouseComponent.Hovered,
			startedDragging:      e.MouseComponent.startedDragging,
			rightStartedDragging: e.MouseComponent.rightStartedDragging,
		}

		if e.MouseComponent.Track {
			// track mouse position so that systems that need to stay on the mouse
			// position can do it (think an RTS when placing a new building and
			// you get a ghost building following your mouse until you click to
			// place it somewhere in your world.
			e.MouseComponent.MouseX = m.mouseX
			e.MouseComponent.MouseY = m.mouseY
		}

		mx := m.mouseX
		my := m.mouseY

		if e.SpaceComponent == nil {
			continue // with other entities
		}

		if e.RenderComponent != nil {
			if e.RenderComponent.Hidden {
				continue // skip hidden components
			}
		}

		// If the Mouse component is a tracker we always update it
		// Check if the X-value is within range
		// and if the Y-value is within range

		if e.MouseComponent.Track || e.MouseComponent.startedDragging ||
			e.SpaceComponent.Contains(image.Point{int(mx), int(my)}) {

			e.MouseComponent.Enter = !e.MouseComponent.Hovered
			e.MouseComponent.Hovered = true
			e.MouseComponent.Released = false

			if !e.MouseComponent.Track {
				// If we're tracking, we've already set these
				e.MouseComponent.MouseX = mx
				e.MouseComponent.MouseY = my
			}

			switch minieng.Input.Mouse.Action {
			case minieng.Press:
				switch minieng.Input.Mouse.Button {
				case minieng.MouseButtonLeft:
					e.MouseComponent.Clicked = true
					e.MouseComponent.startedDragging = true
				case minieng.MouseButtonRight:
					e.MouseComponent.RightClicked = true
					e.MouseComponent.rightStartedDragging = true
				}

				m.mouseDown = true
			case minieng.Release:
				switch minieng.Input.Mouse.Button {
				case minieng.MouseButtonLeft:
					e.MouseComponent.Released = true
				case minieng.MouseButtonRight:
					e.MouseComponent.RightReleased = true
				}
			case minieng.Move:
				if m.mouseDown && e.MouseComponent.startedDragging {
					e.MouseComponent.Dragged = true
				}
				if m.mouseDown && e.MouseComponent.rightStartedDragging {
					e.MouseComponent.RightDragged = true
				}
			}
		} else {
			if e.MouseComponent.Hovered {
				e.MouseComponent.Leave = true
			}

			e.MouseComponent.Hovered = false
		}

		if minieng.Input.Mouse.Action == minieng.Release {
			// dragging stops as soon as one of the currently pressed buttons
			// is released
			e.MouseComponent.Dragged = false
			e.MouseComponent.startedDragging = false
			// TODO maybe separate out the release into left-button release and right-button release
			e.MouseComponent.rightStartedDragging = false
			// mouseDown goes false as soon as one of the pressed buttons is
			// released. Effectively ending any dragging
			m.mouseDown = false
		}

		// propagate the modifiers to the mouse component so that game
		// implementers can take different decisions based on those
		e.MouseComponent.Modifier = minieng.Input.Mouse.Modifer
	}
}
