package diameter

import (
	"net"
)

// agent := NewAgent(Identity{OriginHost: $oh, OriginRealm: $or})
// listener := net.Listener("tcp", "0.0.0.0:5060")
// go agent.Start([]net.Listener{listener})
// c := agent.EventChannel()
//
// for {
//    event := <- c
//	  ...
//	  conn := net.Dial("tcp", "203.0.113.1:5060")
//	  p := NewPeer(...)
//	  agent.StartSessionWithPeer(p, conn)
// }

type EventType int

const (
	IncomingTransportConnectionEstablished EventType = iota
	PeerConnectionEsablished
	MessageReceivedFromPeer
	MessageSentToPeer
	ConnectionTimedOut
	ConnectionError
	StateMachineError
	MessageError
	TransportError
	ListenerError
	PeerConnectionTerminated
	TransportClosed
)

type AgentEvent struct {
	Type       EventType
	Peer       *DiameterEntity
	Error      error
	Message    *Message
	Connection net.Conn
	Listener   net.Listener
}

type Agent struct {
	eventChannel chan *AgentEvent
}

func NewAgent() *Agent {
	return &Agent{
		eventChannel: make(chan *AgentEvent, 20),
	}
}

func (agent *Agent) EstablishDiameterConnectionTo(peer *DiameterEntity, conn net.Conn) {

}

func (agent *Agent) Start(listeners []net.Listener) {
	for _, l := range listeners {
		go agent.listenerHandler(l)
	}
}

func (agent *Agent) EventChannel() <-chan *AgentEvent {
	return agent.eventChannel
}

func (agent *Agent) listenerHandler(listener net.Listener) {
	for {
		c, err := listener.Accept()
		if err != nil {
			agent.eventChannel <- &AgentEvent{
				Type:       ListenerError,
				Error:      err,
				Listener:   listener,
				Connection: c,
			}
		} else {
			go agent.incomingPeerHandler(c)
		}
	}
}
