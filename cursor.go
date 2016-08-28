package main

import (
	//"fmt"
	"github.com/go-gl/glfw/v3.1/glfw"
	"time"
)

var curs Cursor = Cursor{time.Now(), true, 0, 0, 0, 0}

// there's 2 cursors actually....
// 1) for mouse: keeps updated to current pointer position
// 2) for text: keyboard typing will be inserted here

type Cursor struct {
	NextBlinkChange time.Time
	Visible         bool
	X               int // text insert position (at the character level)
	Y               int
	MouseX          int // current mouse position (at the character level)
	MouseY          int
}

func (c *Cursor) Draw() {
	if c.NextBlinkChange.Before(time.Now()) {
		c.NextBlinkChange = time.Now().Add(time.Millisecond * 170)
		c.Visible = !c.Visible
	}

	if c.Visible == true {
		textRend.DrawCharAt('_', c.X, c.Y)
	}
}

func (c *Cursor) ConvertMouseClickToTextCursorPosition(button uint8, action uint8) {
	if glfw.MouseButton(button) == glfw.MouseButtonLeft &&
		glfw.Action(action) == glfw.Press {

		if c.MouseY < len(document) {
			curs.Y = c.MouseY

			if c.MouseX <= len(document[curs.Y]) {
				curs.X = c.MouseX
			} else {
				curs.X = len(document[curs.Y])
			}
		} else {
			curs.Y = len(document) - 1
		}
	}
}
