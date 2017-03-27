package god

import (
	"runtime"

	t "github.com/corpusc/viscript/god/terminal"

	igl "github.com/corpusc/viscript/god/gl" //internal gl
)

//glfw
//glfw.PollEvents()
//only remaining

var (
	CloseWindow bool            = false
	Terms       t.TerminalStack = t.TerminalStack{}
)

func Init() {
	println("<god.go>.Init()")
	// GLFW event handling must run on the main OS thread
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()

	initScreen()
	initEvents()
	initTerms()
}

func initScreen() {
	println("<god.go>.initScreen()")
	igl.InitCanvas()
	igl.WindowInit()
	igl.LoadTextures()
	igl.InitRenderer()
}

func initEvents() {
	println("<god.go>.initEvents()")
	igl.InitInputEvents(igl.GlfwWindow)
	igl.InitMiscEvents(igl.GlfwWindow)
}

func initTerms() {
	Terms.Init()
	Terms.AddTerminal()
	// Terms.AddTerminal()
	// Terms.AddTerminal()
}

func TeardownScreen() {
	println("<god.go>.TeardownScreen()")
	igl.ScreenTeardown()
}

func PollUiInputEvents() {
	igl.PollEvents() //move to gl
}

//could be in messages
func DispatchEvents() []byte {
	message := []byte{}

	for len(igl.InputEvents) > 0 {
		v := <-igl.InputEvents
		message = UnpackMessage(v)
	}

	return message
}

func Tick() {
	igl.Curs.Tick()
	Terms.Tick()
}

func UpdateDrawBuffer() {
	igl.DrawBegin()
	Terms.Draw()
	igl.DrawEnd()
}

func SwapDrawBuffer() {
	igl.SwapDrawBuffer()
}
