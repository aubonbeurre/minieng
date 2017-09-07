package minieng

// RunOptions ...
type RunOptions struct {
	// Title is the Window title
	Title string

	// Width ...
	Width, Height int
}

// Exit is the safest way to close your game, as `engo` will correctly attempt to close all windows, handlers and contexts
func Exit() {
	closeGame = true
}

// CloseEvent ...
func CloseEvent() {
	for _, scenes := range scenes {
		if exiter, ok := scenes.scene.(Exiter); ok {
			exiter.Exit()
		}
	}
	Exit()
}

// Run is called to create a window, initialize everything, and start the main loop. Once this function returns,
// the game window has been closed already. You can supply a lot of options within `RunOptions`, and your starting
// `Scene` should be defined in `defaultScene`.
func Run(o RunOptions, defaultScene Scene) {

	// Create input
	Input = NewInputManager()

	CreateWindow(o.Title, o.Width, o.Height)
	defer DestroyWindow()

	runLoop(defaultScene)
}
