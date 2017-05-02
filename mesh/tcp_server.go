package mesh

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/skycoin/skycoin/src/mesh/messages"
)

type MeshServer struct {
	address          string
	lock             *sync.Mutex
	connections      map[uint32]net.Conn
	responseChannels map[uint32]chan []byte
	sequence         uint32
}

func NewMeshServer(address string) *MeshServer {
	server := &MeshServer{address: address}
	server.lock = &sync.Mutex{}
	server.responseChannels = make(map[uint32]chan []byte)
	server.connections = make(map[uint32]net.Conn)
	return server
}

func (self *MeshServer) Serve() {
	address := self.address

	l, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}

	log.Println("Listening for incoming messages on", self.address)

	for {
		appConn, err := l.Accept() // accept a connection which is created by an app
		if err != nil {
			log.Println("Cannot accept client's connection:", err)
			return
		}
		defer appConn.Close()

		remoteAddr := appConn.RemoteAddr().String()
		go func() { // run listening the connection for user command exchange between viscript and app (ping, shutdown, resources request etc.)
			for {
				message := make([]byte, 16384)

				n, err := appConn.Read(message)
				if err != nil {
					return
					if err == io.EOF {
						continue
					} else {
						log.Printf("error while reading message from %s: %s\n", remoteAddr, err)
						break
					}
				}

				uc := &messages.UserCommand{}
				err = messages.Deserialize(message[:n], uc)
				if err != nil {
					panic(err)
				}
				log.Println("received message for sequence", uc.Sequence)

				appId := uc.AppId
				sequence := uc.Sequence

				self.lock.Lock()
				if _, ok := self.connections[appId]; !ok { // if viscript already created an app, this connection is already in the map
					self.connections[appId] = appConn // if no, then add current connection to the map; so next iterations will skip this step
				}
				respChan, ok0 := self.responseChannels[sequence] // take response channel for responding to it
				self.lock.Unlock()

				if !ok0 {
					log.Println("no response channel", err)
					continue
				}
				respChan <- uc.Payload // respond to it
			}
		}()
	}
}

func (self *MeshServer) Send(appId uint32, msg []byte) ([]byte, error) {

	respChan, sequence := self.MakeResponseChannel()

	self.lock.Lock()
	conn, ok := self.connections[appId]
	self.lock.Unlock()

	if !ok {
		return nil, errors.New(fmt.Sprintf("no connection to meshnet with id %d\n", appId))
	}

	uc := &messages.UserCommand{sequence, appId, msg}
	ucS := messages.Serialize(messages.MsgUserCommand, uc)
	_, err := conn.Write(ucS)
	if err != nil {
		return nil, err
	}

	response, err := self.Wait(respChan, sequence)
	return response, err
}

func (self *MeshServer) MakeResponseChannel() (chan []byte, uint32) {

	respChan := make(chan []byte)

	self.lock.Lock()
	sequence := self.sequence
	self.responseChannels[sequence] = respChan
	self.sequence++
	self.lock.Unlock()

	return respChan, sequence
}

func (self *MeshServer) Wait(respChan chan []byte, sequence uint32) ([]byte, error) {
	select {
	case response := <-respChan:
		return response, nil
	case <-time.After(time.Second * 10):
		return nil, errors.New(fmt.Sprintf("Timeout when receiving response for %d\n", sequence))
	}
}