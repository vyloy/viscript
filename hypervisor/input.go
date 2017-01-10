package hypervisor

import (
	"fmt"
	//_ "image/png"
	//"log"
	/*
		"go/build"
		"runtime"
	*/
	"bytes"
	"math"
	//"strconv"

	"encoding/binary"

	"github.com/corpusc/viscript/gfx"
	"github.com/corpusc/viscript/gl"
	"github.com/corpusc/viscript/msg"
	"github.com/corpusc/viscript/script"
	"github.com/go-gl/glfw/v3.2/glfw"
)

/*
var prevMousePixelX float64
var prevMousePixelY float64
var mousePixelDeltaX float64
var mousePixelDeltaY float64
*/

// this can also be triggered by onMouseButton
func onMouseCursorPos(m msg.MessageMousePos) {

	// gfx.Curs.UpdatePosition(float32(x), float32(y)) //state update

	// mousePixelDeltaX = x - prevMousePixelX
	// mousePixelDeltaY = y - prevMousePixelY
	// prevMousePixelX = x
	// prevMousePixelY = y

	// //rendering update
	// if /* LMB held */ w.GetMouseButton(glfw.MouseButtonLeft) == glfw.Press {
	// 	gfx.Rend.ScrollPanelThatIsHoveredOver(mousePixelDeltaX, mousePixelDeltaY)
	// }

}

func onMouseScroll(m msg.MessageMouseScroll) {
	/*
		var delta float64 = 30

		// if horizontal
		//state update
		if w.GetKey(glfw.KeyLeftShift) == glfw.Press || w.GetKey(glfw.KeyRightShift) == glfw.Press {
			gfx.Rend.ScrollPanelThatIsHoveredOver(yOff*-delta, 0)
		} else {
			gfx.Rend.ScrollPanelThatIsHoveredOver(xOff*delta, yOff*-delta)
		}
	*/
}

//Working
//func onChar(w *glfw.Window, char rune) {
func onChar(m msg.MessageOnCharacter) {
	InsertRuneIntoDocument("Rune", m.Rune)
	script.Process(false)
}

func eitherControlKeyHeld() bool {
	if gl.GlfwWindow.GetKey(glfw.KeyLeftControl) == glfw.Press || gl.GlfwWindow.GetKey(glfw.KeyRightControl) == glfw.Press {
		return true
	} else {
		return false
	}
}

// must be in range
func insert(slice []string, index int, value string) []string {
	slice = slice[0 : len(slice)+1]      // grow the slice by one element
	copy(slice[index+1:], slice[index:]) // move the upper part of the slice out of the way and open a hole
	slice[index] = value
	return slice
}

// similar to insert method, instead moves current slice element and appends to one above
func remove(slice []string, index int, value string) []string {
	slice = append(slice[:index], slice[index+1:]...)
	slice[index-1] = slice[index-1] + value
	return slice
}

func movedCursorSoUpdateDependents() {
	// --- Always-Visible-Cursor Autoscrolling ---
	//
	// TODO ^

	//
	// --- Selection Marking ---
	//
	// when SM is made functional,
	// we should probably detect whether cursor
	// position should update Start_ or End_ at this point.
	// rather than always making that the "end".
	// i doubt marking forwards or backwards will ever alter what is
	// done with the selection

	foc := gfx.Rend.Focused

	if foc.Selection.CurrentlySelecting {
		foc.Selection.EndX = foc.CursX
		foc.Selection.EndY = foc.CursY
	} else { // moving cursor without shift gets rid of selection
		foc.Selection.StartX = math.MaxUint32
		foc.Selection.StartY = math.MaxUint32
		foc.Selection.EndX = math.MaxUint32
		foc.Selection.EndY = math.MaxUint32
	}
}

func getSlice(wBuf *bytes.Buffer, err error) (data []byte) {
	data = make([]byte, 0)

	if err != nil {
		fmt.Println("binary.Write failed:", err)
	} else {
		b := wBuf.Bytes()

		for i := 0; i < wBuf.Len(); i++ {
			data = append(data, b[i])
		}
	}

	return
}
func getBytesOfRune(value rune) (data []byte) {
	wBuf := new(bytes.Buffer)
	err := binary.Write(wBuf, binary.LittleEndian, value)
	data = getSlice(wBuf, err)
	return
}
