package minieng

var (
	// Input handles all input: mouse, keyboard and touch
	Input *InputManager

	// Mailbox is used by all Systems to communicate
	Mailbox *MessageManager

	currentWorld *World
	currentScene Scene

	closeGame bool

	// Time ...
	Time *Clock

	windowWidth  float32
	windowHeight float32

	lasttime float64
)
