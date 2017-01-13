package process

import (
	"github.com/corpusc/viscript/msg"
)

//example of a gui API and calling out
//sending messages back to hypervisor to set terminal
func SetTerminal(out chan []byte, x uint32, y uint32, Rune uint32) {
	var m msg.MessageSetTerminal
	m.TermId = 0
	m.X = x
	m.Y = y
	m.Rune = Rune
	//write event to out channel
	b := msg.Serialize(msg.TypeSetTerminal, m) //serialize as byte string
	out <- b                                   //write to output channel
}

//example of a gui API and calling out
//sending messages back to hypervisor to set terminal
func SetCursor(out chan []byte, x uint32, y uint32) {
	var m msg.MessageSetCursor
	m.TermId = 0
	m.X = x
	m.Y = y
	//write event to out channel
	b := msg.Serialize(msg.MessageSetCursor, m) //serialize as byte string
	out <- b                                    //write to output channel
}