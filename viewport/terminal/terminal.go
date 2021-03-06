package terminal

import (
	"github.com/skycoin/viscript/app"
	"github.com/skycoin/viscript/config"
	"github.com/skycoin/viscript/hypervisor"
	//"github.com/skycoin/viscript/hypervisor/input/keyboard"
	"github.com/skycoin/viscript/msg"
)

const (
	NumPromptLines = 2
	MinimumColumns = 16 //don't allow resizing smaller than this
	path           = "viewport/terminal/terminal"
)

var numOOB int //number of out of bound characters

type Terminal struct {
	TerminalId     msg.TerminalId
	FixedSize      bool
	AttachedTask   msg.TaskId
	OutChannelId   uint32 //id of pubsub channel
	InChannel      chan []byte
	ResizingRight  bool
	ResizingBottom bool

	//int/character grid space
	Curr     app.Vec2I //current flow insert position
	Cursor   app.Vec2I //user controlled position (within command prompt row/s)
	GridSize app.Vec2I //number of characters across
	Chars    [][]uint32

	//float/GL space
	//(mouse pos events & frame buffer sizes are the only things that use pixels)
	BorderSize float32
	CharSize   app.Vec2F
	Bounds     *app.Rectangle
	Depth      float32 //0 for lowest
}

func (t *Terminal) Init() {
	println("<Terminal>.Init()")

	t.TerminalId = msg.RandTerminalId()
	t.InChannel = make(chan []byte, msg.ChannelCapacity)
	t.BorderSize = 0.013
	t.GridSize = app.Vec2I{80, 32}
	t.setupNewGrid()
	t.CharSize.X = (t.Bounds.Width() - t.BorderSize*2) / float32(t.GridSize.X)
	t.CharSize.Y = (t.Bounds.Height() - t.BorderSize*2) / float32(t.GridSize.Y)

	t.PutString(app.HelpText)
	t.SetStringAt(0, 1, ">")
	t.SetCursor(1, 1)
	t.Curr.X = 1
	t.Curr.Y = 1
	t.ResizingRight = false
	t.ResizingBottom = false

	//push down window by size of id tab
	tabOffset := t.BorderSize + t.CharSize.Y
	t.Bounds.Top -= tabOffset
	t.Bounds.Bottom -= tabOffset
}

func (t *Terminal) Tick() {
	//println("Terminal.Tick()")

	for len(t.InChannel) > 0 {
		t.UnpackMessage(<-t.InChannel)
	}
}

func (t *Terminal) IsResizing() bool {
	return t.ResizingRight || t.ResizingBottom
}

func (t *Terminal) SetResizingOff() {
	t.ResizingRight = false
	t.ResizingBottom = false
}

func (t *Terminal) ResizeHorizontally(newRight float32) {
	t.ResizingRight = true
	delta := newRight - t.Bounds.Right
	sx := t.GridSize.X

	// if keyboard.AltKeyIsDown {
	// 	//if we re-enable holding CTRL for pixel resizing, will need to adjust GridSize too
	// 	t.Bounds.Right = newRight
	// } else {
	for delta > t.CharSize.X {
		delta -= t.CharSize.X

		t.Bounds.Right += t.CharSize.X
		t.GridSize.X++
	}

	for delta < -t.CharSize.X {
		delta += t.CharSize.X

		if t.GridSize.X > MinimumColumns {
			t.Bounds.Right -= t.CharSize.X
			t.GridSize.X--
		}
	}
	// }

	if /* x changed */ sx != t.GridSize.X {
		t.setupNewGrid()
	}
}

func (t *Terminal) ResizeVertically(newBottom float32) {
	t.ResizingBottom = true
	delta := newBottom - t.Bounds.Bottom
	sy := t.GridSize.Y

	// if keyboard.AltKeyIsDown {
	// 	//if we re-enable holding CTRL for pixel resizing, will need to adjust GridSize too
	// 	t.Bounds.Bottom = newBottom
	// } else {
	for delta > t.CharSize.Y {
		delta -= t.CharSize.Y
		t.Bounds.Bottom += t.CharSize.Y
		t.GridSize.Y--
	}

	for delta < -t.CharSize.Y {
		delta += t.CharSize.Y
		t.Bounds.Bottom -= t.CharSize.Y
		t.GridSize.Y++
	}
	// }

	if /* y changed */ sy != t.GridSize.Y {
		t.setupNewGrid()
	}
}

func (t *Terminal) RelayToTask(message []byte) {
	hypervisor.DbusGlobal.PublishTo(t.OutChannelId, message)
}

func (t *Terminal) MoveRight() {
	t.Curr.X++

	if t.Curr.X >= t.GridSize.X {
		t.NewLine()
	}
}

func (t *Terminal) NewLine() {
	t.Curr.X = 0
	t.Curr.Y++

	//reserve space along bottom to allow for max prompt size
	if t.Curr.Y > t.GridSize.Y-NumPromptLines {
		t.Curr.Y--

		//shift everything up by one line
		for y := 0; y < t.GridSize.Y-1; y++ {
			for x := 0; x < t.GridSize.X; x++ {
				t.Chars[y][x] = t.Chars[y+1][x]
			}
		}
	}

	if config.Global.Settings.RunHeadless {
		Terms.DrawTextMode()
	}
}

func (t *Terminal) GetVisualInfo() msg.MessageVisualInfo {
	return msg.MessageVisualInfo{
		uint32(t.GridSize.X),
		uint32(t.GridSize.Y),
		uint32(t.Curr.X),
		uint32(t.Curr.Y),
		uint32(NumPromptLines)}
}

//
//
//private
//
//

func (t *Terminal) clear() {
	for y := 0; y < t.GridSize.Y; y++ {
		for x := 0; x < t.GridSize.X; x++ {
			t.Chars[y][x] = 0
		}
	}
}

func (t *Terminal) updateCommandPrompt(m msg.MessageCommandPrompt) {
	for i := 0; i < t.GridSize.X*2; i++ {
		var char uint32
		x := i % t.GridSize.X
		y := i / t.GridSize.X
		y += int(t.Curr.Y)

		if i == int(m.CursorOffset) {
			t.SetCursor(x, y)
		}

		if i < len(m.CommandLine) {
			char = uint32(m.CommandLine[i])
		} else {
			char = 0
		}

		t.SetCharacterAt(x, y, char)
	}
}

func (t *Terminal) posIsValidElsePrint(X, Y int) bool { //...any errors to OS box

	if X < 0 || X >= t.GridSize.X ||
		Y < 0 || Y >= t.GridSize.Y {
		numOOB++

		if numOOB == 1 {
			println("****** ATTEMPTED OUT OF BOUNDS CHARACTER PLACEMENT! ******")
		}

		return false
	}

	return true
}

func (t *Terminal) setupNewGrid() {
	t.Curr = app.Vec2I{0, 0}
	t.Chars = [][]uint32{}

	//allocate every grid position in the "Chars" multi-dimensional slice
	for y := 0; y < t.GridSize.Y; y++ {
		t.Chars = append(t.Chars, []uint32{})

		for x := 0; x < t.GridSize.X; x++ {
			t.Chars[y] = append(t.Chars[y], 0)
		}
	}

	t.updateVisualInfoOfTask()
}

func (t *Terminal) updateVisualInfoOfTask() {
	if t.OutChannelId != 0 {
		m := msg.Serialize(msg.TypeVisualInfo, t.GetVisualInfo())
		hypervisor.DbusGlobal.PublishTo(t.OutChannelId, m)
	}
}
