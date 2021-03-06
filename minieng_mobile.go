//+build android

package minieng

import (
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/aubonbeurre/glplus"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/asset"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/app/debug"
	"golang.org/x/mobile/exp/gl/glutil"
	mobilegl "golang.org/x/mobile/gl"
)

var (
	// Gl is the current OpenGL context
	Gl *glplus.Context
	sz size.Event

	canvasWidth, canvasHeight float32

	msaaPreference int

	ResizeXOffset = float32(0)
	ResizeYOffset = float32(0)

	Backend string = "Mobile"
)

// CreateWindow creates a window with the specified parameters
func CreateWindow(title string, width, height int) {
	msaaPreference = 0
}

// WindowSize returns the width and height of the current window
func WindowSize() (w, h int) {
	return sz.WidthPx, sz.HeightPx
}

func warning(msg string) {
	log.Println("[WARNING] " + msg)
}

func notImplemented(msg string) {
	warning(msg + "is not yet implemented on this platform")
}

// CursorPos returns the current cursor position
func CursorPos() (x, y float32) {
	notImplemented("CursorPos")
	return 0, 0
}

// WindowWidth returns the current window width
func WindowWidth() float32 {
	return windowWidth
}

// WindowHeight returns the current window height
func WindowHeight() float32 {
	return windowHeight
}

// CanvasWidth returns the current canvas width
func CanvasWidth() float32 {
	return canvasWidth
}

// CanvasHeight returns the current canvas height
func CanvasHeight() float32 {
	return canvasHeight
}

func CanvasScale() float32 {
	return CanvasWidth() / WindowWidth()
}

func DestroyWindow() { /* nothing to do here? */ }

func runLoop(defaultScene Scene) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		CloseEvent()
	}()

	app.Main(func(a app.App) {
		var (
			images *glutil.Images
			fps    *debug.FPS
		)

		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					Gl = glplus.NewContext(e.DrawContext)
					glplus.Gl = Gl
					RunPreparation(defaultScene)

					images = glutil.NewImages(e.DrawContext.(mobilegl.Context))
					fps = debug.NewFPS(images)

					// Start tick, minimize the delta
					Time.Tick()

					// Let the device know we want to start painting :-)
					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					CloseEvent()
					Gl = nil
				}

			case size.Event:
				sz = e
				windowWidth = float32(sz.WidthPx)
				windowHeight = float32(sz.HeightPx)
				canvasWidth = float32(sz.WidthPx)
				canvasHeight = float32(sz.HeightPx)
				Gl.Viewport(0, 0, sz.WidthPx, sz.HeightPx)
			case paint.Event:
				if e.External {
					// As we are actively painting as fast as
					// we can (usually 60 FPS), skip any paint
					// events sent by the system.
					continue
				}

				RunIteration()

				fps.Draw(sz)

				Input.Mouse.Action = Neutral
				a.Publish() // same as SwapBuffers

				// Drive the animation by preparing to paint the next frame
				// after this one is shown. - FPS is ignored here!
				a.Send(paint.Event{})
			case touch.Event:
				Input.Mouse.X = e.X
				Input.Mouse.Y = e.Y
				switch e.Type {
				case touch.TypeBegin:
					Input.Mouse.Action = Press
				case touch.TypeMove:
					Input.Mouse.Action = Move
				case touch.TypeEnd:
					Input.Mouse.Action = Release
				}
			}
		}
	})
}

// RunPreparation is called only once, and is called automatically when calling Open
// It is only here for benchmarking in combination with OpenHeadlessNoRun
func RunPreparation(defaultScene Scene) {
	Time = NewClock()
	SetScene(defaultScene, false)
}

// RunIteration runs one iteration / frame
func RunIteration() {
	Time.Tick()

	Input.update()

	// Then update the world and all Systems
	currentWorld.Update(Time.Delta())

}

// SetCursor changes the cursor - not yet implemented
func SetCursor(Cursor) {
	notImplemented("SetCursor")
}

//SetCursorVisibility sets the visibility of the cursor.
//If true the cursor is visible, if false the cursor is not.
//Does nothing in mobile since there's no visible cursor to begin with
func SetCursorVisibility(visible bool) {}

// SetTitle has no effect on mobile
func SetTitle(title string) {}

// openFile is the mobile-specific way of opening a file
func openFile(url string) (io.ReadCloser, error) {
	usedUrl := url
	if strings.HasPrefix(url, "assets/") {
		usedUrl = usedUrl[7:]
	}

	return asset.Open(usedUrl)
}
