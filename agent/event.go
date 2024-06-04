package agent

import (
	"net"

	"github.com/blorticus-go/diameter"
)

// PeerEventType provides various events that occur during the lifetime of a peer connection.
type PeerEventType int

const (
	ListenerAcceptedTransportEvent PeerEventType = iota
	PeerClosedTransportEvent
	ClosedTransportToPeerEvent
	DiameterConnectionEstablishedEvent
	DiameterConnectionClosedEvent
	StateMachineMessageReceivedFromPeerEvent
	StateMachineMessageSentToPeerEvent
	MessageReceivedFromPeerEvent
	ErrorEvent
)

type PeerStateEvent struct {
	Type        PeerEventType
	RemotePeer  *DiameterEntity
	Conn        net.Conn
	Error       error
	Message     *diameter.Message
	PeerHandler *PeerStateManager
	Peer        *Peer
}

type PeerStateNotifier struct {
	eventChannel chan<- *PeerStateEvent
	transport    net.Conn
	peer         *Peer
}

func NewPeerStateNotifier(eventChannel chan<- *PeerStateEvent) *PeerStateNotifier {
	return &PeerStateNotifier{
		eventChannel: eventChannel,
	}
}

func (n *PeerStateNotifier) SetPeer(p *Peer) *PeerStateNotifier {
	n.peer = p
	return n
}

func (n *PeerStateNotifier) SetTransport(c net.Conn) *PeerStateNotifier {
	n.transport = c
	return n
}

func (n *PeerStateNotifier) NotifyThatListenerAcceptedTransportFromAPeer(c net.Conn) {
	n.SetTransport(c)
	n.eventChannel <- &PeerStateEvent{
		Type: ListenerAcceptedTransportEvent,
		Conn: n.transport,
		Peer: n.peer,
	}
}

func (n *PeerStateNotifier) NotifyThatThePeerClosedTheTransport() {
	n.eventChannel <- &PeerStateEvent{
		Type: PeerClosedTransportEvent,
		Conn: n.transport,
		Peer: n.peer,
	}
}

func (n *PeerStateNotifier) ThatTheTransportToThePeerWasClosed() {
	n.eventChannel <- &PeerStateEvent{
		Type: ClosedTransportToPeerEvent,
		Conn: n.transport,
		Peer: n.peer,
	}
}

func (n *PeerStateNotifier) NotifyThatDiameterConnectionHasBeenEstablished() {
	n.eventChannel <- &PeerStateEvent{
		Type: DiameterConnectionEstablishedEvent,
		Conn: n.transport,
		Peer: n.peer,
	}
}

func (n *PeerStateNotifier) NotifyThatDiameterConnectionHasBeenClosed() {
	n.eventChannel <- &PeerStateEvent{
		Type: DiameterConnectionClosedEvent,
		Conn: n.transport,
		Peer: n.peer,
	}
}

func (n *PeerStateNotifier) NotifyThatAnErrorOccurred(err error) {
	n.eventChannel <- &PeerStateEvent{
		Type:  ErrorEvent,
		Conn:  n.transport,
		Peer:  n.peer,
		Error: err,
	}
}

func (n *PeerStateNotifier) NotifyThatAStateMachineMessageWasReceivedFromThePeer(m *diameter.Message) {
	n.eventChannel <- &PeerStateEvent{
		Type:    StateMachineMessageReceivedFromPeerEvent,
		Conn:    n.transport,
		Peer:    n.peer,
		Message: m,
	}
}

func (n *PeerStateNotifier) NotifyThatAStateMachineMessageWasSentToThePeer(m *diameter.Message) {
	n.eventChannel <- &PeerStateEvent{
		Type:    StateMachineMessageSentToPeerEvent,
		Conn:    n.transport,
		Peer:    n.peer,
		Message: m,
	}
}

func (n *PeerStateNotifier) NotifyThatAMessageWasReceivedFromThePeer(m *diameter.Message) {
	n.eventChannel <- &PeerStateEvent{
		Type:    MessageReceivedFromPeerEvent,
		Conn:    n.transport,
		Peer:    n.peer,
		Message: m,
	}
}

type ConnectionError struct {
	errStr string
}

func NewConnectionError(fromError error) *ConnectionError {
	return &ConnectionError{fromError.Error()}
}

func (e *ConnectionError) Error() string {
	return e.errStr
}

type DiameterStateMachineError struct {
	errStr string
}

func NewDiameterConnectionStateMachineError(fromError error) *DiameterStateMachineError {
	return &DiameterStateMachineError{fromError.Error()}
}

func (e *DiameterStateMachineError) Error() string {
	return e.errStr
}

type MessageProcessingError struct {
	errStr string
}

func NewMessageProcessingError(err error) *MessageProcessingError {
	return &MessageProcessingError{err.Error()}
}

func (e *MessageProcessingError) Error() string {
	return e.errStr
}

type TransportError struct {
	errStr string
}

func NewTransportError(fromError error) *TransportError {
	return &TransportError{fromError.Error()}

}

func (e *TransportError) Error() string {
	return e.errStr
}

type ReceiverError struct {
	errStr string
}

func NewReceiverError(fromError error) *ReceiverError {
	return &ReceiverError{fromError.Error()}
}

func (e *ReceiverError) Error() string {
	return e.errStr
}

type DiameterConnectionTimedOutError struct{}

func NewConnectionTimedOutError(c net.Conn) *DiameterConnectionTimedOutError {
	return &DiameterConnectionTimedOutError{}
}

func (e *DiameterConnectionTimedOutError) Error() string {
	return "diameter connection timed out"
}
