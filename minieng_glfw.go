//+build !netgo,!android

package minieng

import (
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/aubonbeurre/glplus"
	"github.com/go-gl/glfw3/v3.2/glfw"
	"github.com/inkyblackness/imgui-go"
)

// GLFW implements a platform based on github.com/go-gl/glfw (v3.2).
type GLFW struct {
	imguiIO imgui.IO

	window      *glfw.Window
	imguirender *glplus.OpenGL3

	time             float64
	mouseJustPressed [3]bool
}

var (
	// Backend ...
	Backend = "GLFW"

	// ResizeXOffset ...
	ResizeXOffset = float32(0)

	// ResizeYOffset ...
	ResizeYOffset = float32(0)

	window   *glfw.Window
	context  *imgui.Context
	platform *GLFW

	canvasWidth  float32
	canvasHeight float32
	retinaScale  float32 = 1
)

const (
	targetFPS = 30
)

// WindowWidth ...
func WindowWidth() float32 {
	return windowWidth
}

// WindowHeight ...
func WindowHeight() float32 {
	return windowHeight
}

// CanvasWidth ...
func CanvasWidth() float32 {
	return canvasWidth
}

// CanvasHeight ...
func CanvasHeight() float32 {
	return canvasHeight
}

// CanvasScale ...
func CanvasScale() float32 {
	return retinaScale
}

// handle GLFW errors by printing them out
func errorCallback(err glfw.ErrorCode, desc string) {
	fmt.Printf("%v: %v\n", err, desc)
}

// key events are a way to get input from GLFW.
// here we check for the escape key being pressed. if it is pressed,
// request that the window be closed
func keyCallback(w *glfw.Window, k glfw.Key, scancode int, a glfw.Action, mods glfw.ModifierKey) {
	platform.keyChange(w, k, scancode, a, mods)

	if platform.imguiIO.WantCaptureKeyboard() {
		return
	}

	key := Key(k)
	if a == glfw.Press {
		Input.keys.Set(key, true)
	} else if a == glfw.Release {
		Input.keys.Set(key, false)
	}
}

func mouseDownCallback(w *glfw.Window, b glfw.MouseButton, a glfw.Action, m glfw.ModifierKey) {
	platform.mouseButtonChange(w, b, a, m)

	if platform.imguiIO.WantCaptureMouse() {
		return
	}

	x, y := window.GetCursorPos()
	Input.Mouse.X, Input.Mouse.Y = float32(x)*retinaScale, float32(y)*retinaScale

	// this is only valid because we use an internal structure that is
	// 100% compatible with glfw3.h
	Input.Mouse.Button = MouseButton(b)
	Input.Mouse.Modifer = Modifier(m)

	if a == glfw.Press {
		Input.Mouse.Action = Press
	} else {
		Input.Mouse.Action = Release
	}
}

func mouseMoveCallback(w *glfw.Window, x float64, y float64) {
	if platform.imguiIO.WantCaptureMouse() {
		return
	}

	Input.Mouse.X, Input.Mouse.Y = float32(x)*retinaScale, float32(y)*retinaScale
	if Input.Mouse.Action != Release && Input.Mouse.Action != Press {
		Input.Mouse.Action = Move
	}
}

func mouseWheelCallback(w *glfw.Window, xoff float64, yoff float64) {
	platform.mouseScrollChange(w, xoff, yoff)

	if platform.imguiIO.WantCaptureMouse() {
		return
	}

	Input.Mouse.ScrollX = float32(xoff)
	Input.Mouse.ScrollY = float32(yoff)
}

func onSizeCallback(w *glfw.Window, width int, height int) {
	message := WindowResizeMessage{
		OldWidth:  int(windowWidth),
		OldHeight: int(windowHeight),
		NewWidth:  width,
		NewHeight: height,
	}

	windowWidth = float32(width)
	windowHeight = float32(height)

	// get the texture of the window because it may have changed since creation
	x, y := w.GetFramebufferSize()
	canvasWidth = float32(x)
	canvasHeight = float32(y)
	retinaScale = canvasWidth / float32(width)

	Mailbox.Dispatch(message)
}

func (p *GLFW) setKeyMapping() {
	// Keyboard mapping. ImGui will use those indices to peek into the io.KeysDown[] array.
	p.imguiIO.KeyMap(imgui.KeyTab, int(glfw.KeyTab))
	p.imguiIO.KeyMap(imgui.KeyLeftArrow, int(glfw.KeyLeft))
	p.imguiIO.KeyMap(imgui.KeyRightArrow, int(glfw.KeyRight))
	p.imguiIO.KeyMap(imgui.KeyUpArrow, int(glfw.KeyUp))
	p.imguiIO.KeyMap(imgui.KeyDownArrow, int(glfw.KeyDown))
	p.imguiIO.KeyMap(imgui.KeyPageUp, int(glfw.KeyPageUp))
	p.imguiIO.KeyMap(imgui.KeyPageDown, int(glfw.KeyPageDown))
	p.imguiIO.KeyMap(imgui.KeyHome, int(glfw.KeyHome))
	p.imguiIO.KeyMap(imgui.KeyEnd, int(glfw.KeyEnd))
	p.imguiIO.KeyMap(imgui.KeyInsert, int(glfw.KeyInsert))
	p.imguiIO.KeyMap(imgui.KeyDelete, int(glfw.KeyDelete))
	p.imguiIO.KeyMap(imgui.KeyBackspace, int(glfw.KeyBackspace))
	p.imguiIO.KeyMap(imgui.KeySpace, int(glfw.KeySpace))
	p.imguiIO.KeyMap(imgui.KeyEnter, int(glfw.KeyEnter))
	p.imguiIO.KeyMap(imgui.KeyEscape, int(glfw.KeyEscape))
	p.imguiIO.KeyMap(imgui.KeyA, int(glfw.KeyA))
	p.imguiIO.KeyMap(imgui.KeyC, int(glfw.KeyC))
	p.imguiIO.KeyMap(imgui.KeyV, int(glfw.KeyV))
	p.imguiIO.KeyMap(imgui.KeyX, int(glfw.KeyX))
	p.imguiIO.KeyMap(imgui.KeyY, int(glfw.KeyY))
	p.imguiIO.KeyMap(imgui.KeyZ, int(glfw.KeyZ))
}

var glfwButtonIndexByID = map[glfw.MouseButton]int{
	glfw.MouseButton1: 0,
	glfw.MouseButton2: 1,
	glfw.MouseButton3: 2,
}

var glfwButtonIDByIndex = map[int]glfw.MouseButton{
	0: glfw.MouseButton1,
	1: glfw.MouseButton2,
	2: glfw.MouseButton3,
}

func (p *GLFW) mouseButtonChange(window *glfw.Window, rawButton glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	buttonIndex, known := glfwButtonIndexByID[rawButton]

	if known && (action == glfw.Press) {
		p.mouseJustPressed[buttonIndex] = true
	}
}

func (p *GLFW) mouseScrollChange(window *glfw.Window, x, y float64) {
	p.imguiIO.AddMouseWheelDelta(float32(x), float32(y))
}

func (p *GLFW) keyChange(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		p.imguiIO.KeyPress(int(key))
	}
	if action == glfw.Release {
		p.imguiIO.KeyRelease(int(key))
	}

	// Modifiers are not reliable across systems
	p.imguiIO.KeyCtrl(int(glfw.KeyLeftControl), int(glfw.KeyRightControl))
	p.imguiIO.KeyShift(int(glfw.KeyLeftShift), int(glfw.KeyRightShift))
	p.imguiIO.KeyAlt(int(glfw.KeyLeftAlt), int(glfw.KeyRightAlt))
	p.imguiIO.KeySuper(int(glfw.KeyLeftSuper), int(glfw.KeyRightSuper))
}

func (p *GLFW) charChange(window *glfw.Window, char rune) {
	p.imguiIO.AddInputCharacters(string(char))
}

// DisplaySize returns the dimension of the display.
func (p *GLFW) DisplaySize() [2]float32 {
	w, h := p.window.GetSize()
	return [2]float32{float32(w), float32(h)}
}

// FramebufferSize returns the dimension of the framebuffer.
func (p *GLFW) FramebufferSize() [2]float32 {
	w, h := p.window.GetFramebufferSize()
	return [2]float32{float32(w), float32(h)}
}

// NewFrame marks the begin of a render pass. It forwards all current state to imgui IO.
func (p *GLFW) NewFrame() {
	// Setup display size (every frame to accommodate for window resizing)
	if platform.imguirender == nil {
		var err error
		if platform.imguirender, err = glplus.NewOpenGL3(platform.imguiIO); err != nil {
			panic(err)
		}
	}
	displaySize := p.DisplaySize()
	p.imguiIO.SetDisplaySize(imgui.Vec2{X: displaySize[0], Y: displaySize[1]})

	// Setup time step
	currentTime := glfw.GetTime()
	if p.time > 0 {
		p.imguiIO.SetDeltaTime(float32(currentTime - p.time))
	}
	p.time = currentTime

	// Setup inputs
	if p.window.GetAttrib(glfw.Focused) != 0 {
		x, y := p.window.GetCursorPos()
		p.imguiIO.SetMousePosition(imgui.Vec2{X: float32(x), Y: float32(y)})
	} else {
		p.imguiIO.SetMousePosition(imgui.Vec2{X: -math.MaxFloat32, Y: -math.MaxFloat32})
	}

	for i := 0; i < len(p.mouseJustPressed); i++ {
		down := p.mouseJustPressed[i] || (p.window.GetMouseButton(glfwButtonIDByIndex[i]) == glfw.Press)
		p.imguiIO.SetMouseButtonDown(i, down)
		p.mouseJustPressed[i] = false
	}
}

// ClipboardText returns the current clipboard text, if available.
func (p *GLFW) ClipboardText() (string, error) {
	return p.window.GetClipboardString()
}

// SetClipboardText sets the text as the current clipboard text.
func (p *GLFW) SetClipboardText(text string) {
	p.window.SetClipboardString(text)
}

// CreateWindow ...
func CreateWindow(title string, width, height int) {
	context = imgui.CreateContext(nil)

	// make sure that we display any errors that are encountered
	//glfw.SetErrorCallback(errorCallback)

	// the GLFW library has to be initialized before any of the methods
	// can be invoked
	var err error
	if err = glfw.Init(); err != nil {
		panic(err)
	}

	// hints are the way you configure the features requested for the
	// window and are required to be set before calling glfw.CreateWindow().

	// desired number of samples to use for mulitsampling
	//glfw.WindowHint(glfw.Samples, 4)

	// request a OpenGL 4.1 core context
	if false && runtime.GOOS == "darwin" {
		glfw.WindowHint(glfw.ContextVersionMajor, 3)
		glfw.WindowHint(glfw.ContextVersionMinor, 3)
	} else {
		glfw.WindowHint(glfw.ContextVersionMajor, 4)
		glfw.WindowHint(glfw.ContextVersionMinor, 1)
	}
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.Resizable, glfw.True)

	glfw.WindowHint(glfw.Samples, 4)

	// do the actual window creation
	windowWidth = float32(width)
	windowHeight = float32(height)
	window, err = glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		// we legitimately cannot recover from a failure to create
		// the window in this sample, so just bail out
		panic(err)
	}

	// set the callback function to get all of the key input from the user
	window.SetKeyCallback(keyCallback)
	window.SetMouseButtonCallback(mouseDownCallback)
	window.SetScrollCallback(mouseWheelCallback)
	window.SetCursorPosCallback(mouseMoveCallback)
	window.SetSizeCallback(onSizeCallback)

	// GLFW3 can work with more than one window, so make sure we set our
	// new window as the current context to operate on
	window.MakeContextCurrent()

	// make sure that GLEW initializes all of the GL functions
	glplus.Gl = glplus.NewContext()
	fmt.Println("OpenGL version", glplus.Gl.Version())

	x, y := window.GetFramebufferSize()
	canvasWidth = float32(x)
	canvasHeight = float32(y)
	retinaScale = canvasWidth / windowWidth

	platform = &GLFW{
		imguiIO: imgui.CurrentIO(),
		window:  window,
	}
	platform.setKeyMapping()
	window.SetCharCallback(platform.charChange)

	lasttime = glfw.GetTime()

}

// DestroyWindow ...
func DestroyWindow() {
	glfw.Terminate()
	context.Destroy()
}

// RunPreparation is called automatically when calling Open. It should only be called once.
func RunPreparation(defaultScene Scene) {
	Time = NewClock()

	SetScene(defaultScene, false)
}

// RunIteration runs one iteration per frame
func RunIteration() {
	Time.Tick()

	Input.update()
	glfw.PollEvents()

	// Signal start of a new frame
	platform.NewFrame()
	imgui.NewFrame()

	// Then update the world and all Systems
	currentWorld.Update(Time.Delta())

	// Rendering
	imgui.Render() // This call only creates the draw data list. Actual rendering to framebuffer is done below.

	platform.imguirender.Render(platform.DisplaySize(), platform.FramebufferSize(), imgui.RenderedDrawData())

	// Lastly, forget keypresses and swap buffers
	// reset values to avoid catching the same "signal" twice
	Input.Mouse.ScrollX, Input.Mouse.ScrollY = 0, 0
	Input.Mouse.Action = Neutral

	window.SwapBuffers()

	for glfw.GetTime() < lasttime+1.0/targetFPS {
		time.Sleep(10 * time.Millisecond)
	}

	lasttime += 1.0 / targetFPS
}

func runLoop(defaultScene Scene) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		CloseEvent()
	}()

	RunPreparation(defaultScene)
	ticker := time.NewTicker(time.Duration(int(time.Second) / 60))

	// Start tick, minimize the delta
	Time.Tick()

Outer:
	for {
		select {
		case <-ticker.C:
			RunIteration()
			if closeGame {
				break Outer
			}
			if window.ShouldClose() {
				CloseEvent()
			}
		}
	}
	ticker.Stop()
}

func init() {
	runtime.LockOSThread()

	Grave = Key(glfw.KeyGraveAccent)
	Dash = Key(glfw.KeyMinus)
	Apostrophe = Key(glfw.KeyApostrophe)
	Semicolon = Key(glfw.KeySemicolon)
	Equals = Key(glfw.KeyEqual)
	Comma = Key(glfw.KeyComma)
	Period = Key(glfw.KeyPeriod)
	Slash = Key(glfw.KeySlash)
	Backslash = Key(glfw.KeyBackslash)
	Backspace = Key(glfw.KeyBackspace)
	Tab = Key(glfw.KeyTab)
	CapsLock = Key(glfw.KeyCapsLock)
	Space = Key(glfw.KeySpace)
	Enter = Key(glfw.KeyEnter)
	Escape = Key(glfw.KeyEscape)
	Insert = Key(glfw.KeyInsert)
	PrintScreen = Key(glfw.KeyPrintScreen)
	Delete = Key(glfw.KeyDelete)
	PageUp = Key(glfw.KeyPageUp)
	PageDown = Key(glfw.KeyPageDown)
	Home = Key(glfw.KeyHome)
	End = Key(glfw.KeyEnd)
	Pause = Key(glfw.KeyPause)
	ScrollLock = Key(glfw.KeyScrollLock)
	ArrowLeft = Key(glfw.KeyLeft)
	ArrowRight = Key(glfw.KeyRight)
	ArrowDown = Key(glfw.KeyDown)
	ArrowUp = Key(glfw.KeyUp)
	LeftBracket = Key(glfw.KeyLeftBracket)
	LeftShift = Key(glfw.KeyLeftShift)
	LeftControl = Key(glfw.KeyLeftControl)
	LeftSuper = Key(glfw.KeyLeftSuper)
	LeftAlt = Key(glfw.KeyLeftAlt)
	RightBracket = Key(glfw.KeyRightBracket)
	RightShift = Key(glfw.KeyRightShift)
	RightControl = Key(glfw.KeyRightControl)
	RightSuper = Key(glfw.KeyRightSuper)
	RightAlt = Key(glfw.KeyRightAlt)
	Zero = Key(glfw.Key0)
	One = Key(glfw.Key1)
	Two = Key(glfw.Key2)
	Three = Key(glfw.Key3)
	Four = Key(glfw.Key4)
	Five = Key(glfw.Key5)
	Six = Key(glfw.Key6)
	Seven = Key(glfw.Key7)
	Eight = Key(glfw.Key8)
	Nine = Key(glfw.Key9)
	F1 = Key(glfw.KeyF1)
	F2 = Key(glfw.KeyF2)
	F3 = Key(glfw.KeyF3)
	F4 = Key(glfw.KeyF4)
	F5 = Key(glfw.KeyF5)
	F6 = Key(glfw.KeyF6)
	F7 = Key(glfw.KeyF7)
	F8 = Key(glfw.KeyF8)
	F9 = Key(glfw.KeyF9)
	F10 = Key(glfw.KeyF10)
	F11 = Key(glfw.KeyF11)
	F12 = Key(glfw.KeyF12)
	A = Key(glfw.KeyA)
	B = Key(glfw.KeyB)
	C = Key(glfw.KeyC)
	D = Key(glfw.KeyD)
	E = Key(glfw.KeyE)
	F = Key(glfw.KeyF)
	G = Key(glfw.KeyG)
	H = Key(glfw.KeyH)
	I = Key(glfw.KeyI)
	J = Key(glfw.KeyJ)
	K = Key(glfw.KeyK)
	L = Key(glfw.KeyL)
	M = Key(glfw.KeyM)
	N = Key(glfw.KeyN)
	O = Key(glfw.KeyO)
	P = Key(glfw.KeyP)
	Q = Key(glfw.KeyQ)
	R = Key(glfw.KeyR)
	S = Key(glfw.KeyS)
	T = Key(glfw.KeyT)
	U = Key(glfw.KeyU)
	V = Key(glfw.KeyV)
	W = Key(glfw.KeyW)
	X = Key(glfw.KeyX)
	Y = Key(glfw.KeyY)
	Z = Key(glfw.KeyZ)
	NumLock = Key(glfw.KeyNumLock)
	NumMultiply = Key(glfw.KeyKPMultiply)
	NumDivide = Key(glfw.KeyKPDivide)
	NumAdd = Key(glfw.KeyKPAdd)
	NumSubtract = Key(glfw.KeyKPSubtract)
	NumZero = Key(glfw.KeyKP0)
	NumOne = Key(glfw.KeyKP1)
	NumTwo = Key(glfw.KeyKP2)
	NumThree = Key(glfw.KeyKP3)
	NumFour = Key(glfw.KeyKP4)
	NumFive = Key(glfw.KeyKP5)
	NumSix = Key(glfw.KeyKP6)
	NumSeven = Key(glfw.KeyKP7)
	NumEight = Key(glfw.KeyKP8)
	NumNine = Key(glfw.KeyKP9)
	NumDecimal = Key(glfw.KeyKPDecimal)
	NumEnter = Key(glfw.KeyKPEnter)
}

func openFile(url string) (io.ReadCloser, error) {
	return os.Open(url)
}
