package agent

import (
	"fmt"
	"net"

	"github.com/blorticus-go/diameter"
)

type AgentReceiver struct {
	Listener         net.Listener
	IdentityToAssert *DiameterEntity
}

type AgentEvent struct {
	Type       PeerEventType
	Peer       *Peer
	Error      error
	Message    *diameter.Message
	Connection net.Conn
	Receiver   *AgentReceiver
}

type Agent struct {
	outgoingEventChannel             chan *AgentEvent
	peerHandlersIncomingEventChannel chan *PeerStateEvent
}

func New() *Agent {
	return &Agent{
		outgoingEventChannel:             make(chan *AgentEvent, 20),
		peerHandlersIncomingEventChannel: make(chan *PeerStateEvent, 100),
	}
}

func (agent *Agent) EstablishDiameterConnectionTo(conn net.Conn, assertIdentity *DiameterEntity) {
	go NewInitiatorPeerStateManager(assertIdentity, conn, agent.peerHandlersIncomingEventChannel).NewRun()
}

func (agent *Agent) AcceptDiameterConnectionFrom(conn net.Conn, assertIdentity *DiameterEntity) {
	go NewInitiatedPeerStateManager(assertIdentity, conn, agent.peerHandlersIncomingEventChannel).NewRun()
}

func (agent *Agent) Run(receiver []*AgentReceiver) {
	for _, r := range receiver {
		go agent.runReceiverHandler(r)
	}

	for {
		peerHandlerEvent := <-agent.peerHandlersIncomingEventChannel
		agent.outgoingEventChannel <- &AgentEvent{
			Type:       peerHandlerEvent.Type,
			Peer:       peerHandlerEvent.Peer,
			Error:      peerHandlerEvent.Error,
			Message:    peerHandlerEvent.Message,
			Connection: peerHandlerEvent.Conn,
		}
	}
}

func (agent *Agent) EventChannel() <-chan *AgentEvent {
	return agent.outgoingEventChannel
}

func extractIPFromNetConn(c net.Conn) net.IP {
	switch addr := c.LocalAddr().(type) {
	case *net.TCPAddr:
		return addr.IP
	default:
		return nil
	}
}

func (agent *Agent) runReceiverHandler(receiver *AgentReceiver) {
	for {
		c, err := receiver.Listener.Accept()
		if err != nil {
			agent.notifyOfReceiverError(receiver, c, err)
			return
		}

		agent.notifyOfIncomingTransportConnectionOnListener(c)

		identityToAssert := *receiver.IdentityToAssert
		if len(identityToAssert.HostIPAddresses) == 0 {
			hostAddr := extractIPFromNetConn(c)
			if hostAddr == nil {
				agent.notifyOfReceiverError(receiver, c, fmt.Errorf("cannot extract local IP address from connection: %s", c.LocalAddr().String()))
				c.Close()
				return
			}

			identityToAssert.HostIPAddresses = []*net.IP{&hostAddr}
		}

		go NewInitiatedPeerStateManager(&identityToAssert, c, agent.peerHandlersIncomingEventChannel).NewRun()
	}
}

func (agent *Agent) notifyOfReceiverError(receiver *AgentReceiver, connection net.Conn, err error) {
	agent.outgoingEventChannel <- &AgentEvent{
		Type:       ErrorEvent,
		Error:      NewReceiverError(err),
		Receiver:   receiver,
		Connection: connection,
	}
}

func (agent *Agent) notifyOfIncomingTransportConnectionOnListener(connection net.Conn) {
	agent.outgoingEventChannel <- &AgentEvent{
		Type:       ListenerAcceptedTransportEvent,
		Connection: connection,
	}
}
