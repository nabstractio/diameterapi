package agent

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"

	"github.com/blorticus-go/diameter"
)

var cachedResponseCode2001 = diameter.NewTypedAVP(268, 0, true, diameter.Unsigned32, 2001)

type disconnectInitiation struct {
	returnChannel chan<- error
}

type PeerStateManager struct {
	localIdentity                 *DiameterEntity
	transport                     net.Conn
	messageReaderChannel          chan *messageReaderEvent
	disconnectNotificationChannel chan *disconnectInitiation
	eventChannel                  chan<- *PeerStateEvent
	cachedAVPs                    *diameterEntityCache
	sequenceGenerator             *diameter.SequenceGenerator
	quitChannel                   chan bool
	peer                          *Peer
	initialState                  InitialPeerState
}

func NewInitiatorPeerStateManager(localIdentity *DiameterEntity, conn net.Conn, eventChannel chan<- *PeerStateEvent) *PeerStateManager {
	return newPeerStateManager(localIdentity, PeerStateStartsWithTransportOpenedTowardPeer(), conn, eventChannel)
}

func NewInitiatedPeerStateManager(localIdentity *DiameterEntity, conn net.Conn, eventChannel chan<- *PeerStateEvent) *PeerStateManager {
	return newPeerStateManager(localIdentity, PeerStateStartsWithTransportOpenedByPeer(), conn, eventChannel)
}

func newPeerStateManager(localIdentity *DiameterEntity, initialState InitialPeerState, conn net.Conn, eventChannel chan<- *PeerStateEvent) *PeerStateManager {
	if localIdentity == nil {
		panic("self must not be null")
	}
	if conn == nil {
		panic("conn must not be nil")
	}
	if len(localIdentity.HostIPAddresses) == 0 {
		panic("there must be at least one Host-IP-Address")
	}

	messageReaderChannel := make(chan *messageReaderEvent)
	go incomingMessageStreamReceiver(conn, messageReaderChannel)

	return &PeerStateManager{
		localIdentity:                 localIdentity,
		transport:                     conn,
		eventChannel:                  eventChannel,
		messageReaderChannel:          messageReaderChannel,
		disconnectNotificationChannel: make(chan *disconnectInitiation),
		cachedAVPs: &diameterEntityCache{
			ResultCode:      diameter.NewTypedAVP(268, 0, true, diameter.Unsigned32, uint32(2000)),
			OriginHost:      diameter.NewTypedAVP(264, 0, true, diameter.DiamIdent, localIdentity.OriginHost),
			OriginRealm:     diameter.NewTypedAVP(296, 0, true, diameter.DiamIdent, localIdentity.OriginRealm),
			HostIPAddresses: []*diameter.AVP{diameter.NewTypedAVP(257, 0, true, diameter.Address, localIdentity.HostIPAddresses[0])},
			VendorId:        diameter.NewTypedAVP(266, 0, true, diameter.Unsigned32, localIdentity.VendorID),
			ProductName:     diameter.NewTypedAVP(269, 0, true, diameter.UTF8String, localIdentity.ProductName),
		},
		sequenceGenerator: diameter.NewSequenceGeneratorSet(),
		quitChannel:       make(chan bool),
		initialState:      initialState,
	}
}

func incomingMessageStreamReceiver(conn net.Conn, messageReaderChannel chan<- *messageReaderEvent) {
	messageStreamReader := diameter.NewMessageStreamReader(conn)

	for {
		msg, err := messageStreamReader.ReadNextMessage()
		if err != nil {
			messageReaderChannel <- &messageReaderEvent{
				IncomingMessage: msg,
				Error:           err,
			}
			return
		}

		messageReaderChannel <- &messageReaderEvent{
			IncomingMessage: msg,
		}
	}
}

func (manager *PeerStateManager) NewRun() {
	defer func() {
		manager.transport.Close()
		manager.eventChannel <- &PeerStateEvent{
			Type: ClosedTransportToPeerEvent,
			Conn: manager.transport,
			Peer: manager.peer,
		}
	}()

	watchdogTimer := StartNewWatchdogIntervalTimer(30)

	notifier := NewPeerStateNotifier(manager.eventChannel).SetTransport(manager.transport)

	peer, aFatalErrorOccured := manager.initialState.Execute(&InitialPeerStateBuilder{
		LocalEntity:             manager.localIdentity,
		PeerMessageEventChannel: manager.messageReaderChannel,
		Transport:               manager.transport,
		Notifier:                notifier,
		PeerFactory:             NewPeerFactory(manager.SendMessageViaPeer, manager.InitiateDisconnect),
		SequenceGenerator:       manager.sequenceGenerator,
	})

	if aFatalErrorOccured {
		return
	}

	messageBuilder := &MessageBuilder{
		CER: manager.generateCER,
		CEA: manager.generateCEA,
		DWR: manager.generateDWR,
		DWA: manager.generateDWA,
		DPR: manager.generateDPR,
		DPA: manager.generateDPA,
	}

	manager.peer = peer
	notifier.SetPeer(peer)
	notifier.NotifyThatDiameterConnectionHasBeenEstablished()

	nextState := PeerState(NewPeerStateConnected(notifier, manager.transport, peer))

	for {
		var messageToSend *diameter.Message
		var psErr *PeerStateError

		select {
		case disconnectInitiated := <-manager.disconnectNotificationChannel:
			switch nextState.CanInitiateDisconnectInThisState() {
			case true:
				if err := manager.SendStateMachineMessage(manager.generateDPR()); err != nil {
					disconnectInitiated.returnChannel <- err
					return
				}
				nextState = NewPeerStateHalfClosed(notifier, manager.transport, manager.peer)
				disconnectInitiated.returnChannel <- nil

			case false:
				disconnectInitiated.returnChannel <- fmt.Errorf("cannot initiate disconnect in the current state")
			}

		case messageReaderEvent := <-manager.messageReaderChannel:
			if messageReaderEvent.Error != nil {
				if messageReaderEvent.Error == io.EOF {
					notifier.NotifyThatThePeerClosedTheTransport()
				} else {
					notifier.NotifyThatAnErrorOccurred(messageReaderEvent.Error)
				}
				return
			}

			watchdogTimer.StopAndRestart()

			if messageType := stateMachineMessageTypeForMessage(messageReaderEvent.IncomingMessage); messageType != notAStateMachineMessage {
				notifier.NotifyThatAStateMachineMessageWasReceivedFromThePeer(messageReaderEvent.IncomingMessage)

				switch messageType {
				case cer:
					nextState, messageToSend, psErr = nextState.ProcessIncomingCER(messageReaderEvent.IncomingMessage, messageBuilder)
				case cea:
					nextState, messageToSend, psErr = nextState.ProcessIncomingCEA(messageReaderEvent.IncomingMessage, messageBuilder)
				case dwr:
					nextState, messageToSend, psErr = nextState.ProcessIncomingDWR(messageReaderEvent.IncomingMessage, messageBuilder)
				case dwa:
					nextState, messageToSend, psErr = nextState.ProcessIncomingDWA(messageReaderEvent.IncomingMessage, messageBuilder)
				case dpr:
					nextState, messageToSend, psErr = nextState.ProcessIncomingDPR(messageReaderEvent.IncomingMessage, messageBuilder)
				case dpa:
					nextState, messageToSend, psErr = nextState.ProcessIncomingDPA(messageReaderEvent.IncomingMessage, messageBuilder)
				}
			} else {
				notifier.NotifyThatAMessageWasReceivedFromThePeer(messageReaderEvent.IncomingMessage)
				nextState, psErr = nextState.ProcessIncomingNonStateMachineMessage(messageReaderEvent.IncomingMessage)
			}

			if psErr != nil {
				notifier.NotifyThatAnErrorOccurred(psErr.Error)
				if psErr.initiateDisconnectPeer {
					if err := manager.SendStateMachineMessage(manager.generateDPR()); err != nil {
						notifier.NotifyThatAnErrorOccurred(err)
					}
				}
				return
			}

			if messageToSend != nil {
				if err := manager.SendStateMachineMessage(messageToSend); err != nil {
					notifier.NotifyThatAnErrorOccurred(err)
					return
				}
			}

			if nextState.DiameterConnectionIsClosedInThisState() {
				return
			}

		case <-watchdogTimer.C:
			dwr := manager.generateDWR()
			if err := manager.SendStateMachineMessage(dwr); err != nil {
				notifier.NotifyThatAnErrorOccurred(err)
			}
			watchdogTimer.Restart()

		case <-manager.quitChannel:
			return
		}
	}
}

func (manager *PeerStateManager) InitiateDisconnect() error {
	c := make(chan error, 2)

	manager.disconnectNotificationChannel <- &disconnectInitiation{
		returnChannel: c,
	}

	return <-c
}

func (manager *PeerStateManager) SendMessageViaPeer(msg *diameter.Message) error {
	if MessageIsADiameterConnectionStateMessage(msg) {
		return fmt.Errorf("diameter connection state machine messages cannot be sent directly from client")
	}

	if msg.EndToEndID == 0 {
		msg.EndToEndID = manager.sequenceGenerator.NextEndToEndId()
	}
	if msg.HopByHopID == 0 {
		msg.HopByHopID = manager.sequenceGenerator.NextHopByHopId()
	}

	return manager.sendMessage(msg)
}

func (manager *PeerStateManager) SendStateMachineMessage(msg *diameter.Message) error {
	if err := manager.sendMessage(msg); err != nil {
		return err
	}

	manager.eventChannel <- &PeerStateEvent{
		Type:    StateMachineMessageSentToPeerEvent,
		Peer:    manager.peer,
		Conn:    manager.transport,
		Message: msg,
	}

	return nil
}

func (manager *PeerStateManager) sendMessage(msg *diameter.Message) error {
	_, err := manager.transport.Write(msg.Encode())
	if err != nil {
		if err == io.EOF {
			manager.eventChannel <- &PeerStateEvent{
				Type: PeerClosedTransportEvent,
				Peer: manager.peer,
				Conn: manager.transport,
			}
			return nil
		} else {
			return err
		}
	}

	return nil
}

type stateMachineMessageType int

const (
	cer stateMachineMessageType = iota
	cea
	dwr
	dwa
	dpr
	dpa
	notAStateMachineMessage
)

func stateMachineMessageTypeForMessage(m *diameter.Message) stateMachineMessageType {
	if m.AppID == 0 {
		switch m.Code {
		case CapabilitiesExchangeCode:
			if m.IsRequest() {
				return cer
			}
			return cea

		case DeviceWatchdogCode:
			if m.IsRequest() {
				return dwr
			}
			return dwa

		case DisconnectPeerCode:
			if m.IsRequest() {
				return dpr
			}
			return dpa
		}
	}

	return notAStateMachineMessage
}

func (manager *PeerStateManager) generateCER() *diameter.Message {
	return diameter.NewMessage(
		diameter.MsgFlagRequest,
		CapabilitiesExchangeCode,
		0,
		manager.sequenceGenerator.NextHopByHopId(),
		manager.sequenceGenerator.NextEndToEndId(),
		manager.localIdentity.CapabilitiesExchangeMandatoryAvps(),
		nil)
}

func (manager *PeerStateManager) generateCEA(forCER *diameter.Message) *diameter.Message {
	return forCER.GenerateMatchingResponseWithAvps(
		manager.localIdentity.CapabilitiesExchangeMandatoryAvpsWithResultCode(cachedResponseCode2001),
		nil,
	)
}

func (manager *PeerStateManager) generateDWR() *diameter.Message {
	return diameter.NewMessage(
		diameter.MsgFlagRequest,
		DeviceWatchdogCode,
		0,
		manager.sequenceGenerator.NextHopByHopId(),
		manager.sequenceGenerator.NextEndToEndId(),
		[]*diameter.AVP{
			manager.localIdentity.OriginHostAvp(),
			manager.localIdentity.OriginHostAvp(),
		},
		nil)
}

func (manager *PeerStateManager) generateDWA(forDWR *diameter.Message) *diameter.Message {
	return forDWR.GenerateMatchingResponseWithAvps(
		[]*diameter.AVP{
			cachedResponseCode2001,
			manager.localIdentity.OriginHostAvp(),
			manager.localIdentity.OriginHostAvp(),
		},
		nil,
	)
}

func (manager *PeerStateManager) generateDPR() *diameter.Message {
	return diameter.NewMessage(diameter.MsgFlagRequest, DisconnectPeerCode, 0, manager.sequenceGenerator.NextHopByHopId(), manager.sequenceGenerator.NextEndToEndId(),
		[]*diameter.AVP{
			manager.localIdentity.OriginHostAvp(),
			manager.localIdentity.OriginHostAvp(),
			diameter.NewTypedAVP(273, 0, true, diameter.Enumerated, int32(2)),
		},
		nil)
}

func (manager *PeerStateManager) generateDPA(forDPR *diameter.Message) *diameter.Message {
	return forDPR.GenerateMatchingResponseWithAvps(
		[]*diameter.AVP{
			cachedResponseCode2001,
			manager.localIdentity.OriginHostAvp(),
			manager.localIdentity.OriginHostAvp(),
		},
		nil,
	)
}

func MessageIsADiameterConnectionStateMessage(m *diameter.Message) bool {
	return m.AppID == 0 && (m.Code == CapabilitiesExchangeCode || m.Code == DeviceWatchdogCode || m.Code == DisconnectPeerCode)
}

func MessageIsNotADiameterConnectionStateMessage(m *diameter.Message) bool {
	return !MessageIsADiameterConnectionStateMessage(m)
}

type InitialPeerStateBuilder struct {
	LocalEntity             *DiameterEntity
	PeerMessageEventChannel <-chan *messageReaderEvent
	Transport               net.Conn
	Notifier                *PeerStateNotifier
	PeerFactory             *PeerFactory
	SequenceGenerator       *diameter.SequenceGenerator
}

type MessageBuilder struct {
	CER func() *diameter.Message
	DWR func() *diameter.Message
	DPR func() *diameter.Message

	CEA func(forCER *diameter.Message) *diameter.Message
	DWA func(forDWR *diameter.Message) *diameter.Message
	DPA func(forDPR *diameter.Message) *diameter.Message
}

type PeerStateError struct {
	Error                  error
	initiateDisconnectPeer bool
}

type InitialPeerState interface {
	Execute(b *InitialPeerStateBuilder) (peerEntityInformation *Peer, aFatalErrorOccurred bool)
}

type PeerState interface {
	ProcessIncomingCER(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, err *PeerStateError)
	ProcessIncomingCEA(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, err *PeerStateError)
	ProcessIncomingDWR(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, err *PeerStateError)
	ProcessIncomingDWA(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, err *PeerStateError)
	ProcessIncomingDPR(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, err *PeerStateError)
	ProcessIncomingDPA(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, err *PeerStateError)
	ProcessIncomingNonStateMachineMessage(m *diameter.Message) (nextState PeerState, err *PeerStateError)

	CanInitiateDisconnectInThisState() bool
	DiameterConnectionIsClosedInThisState() bool
}

type InitialPeerStatePeerOpenedTransport struct{}

func PeerStateStartsWithTransportOpenedByPeer() *InitialPeerStatePeerOpenedTransport {
	return &InitialPeerStatePeerOpenedTransport{}
}

func (s *InitialPeerStatePeerOpenedTransport) Execute(b *InitialPeerStateBuilder) (connectedPeer *Peer, aFatalErrorOccurred bool) {
	messageReaderEvent := <-b.PeerMessageEventChannel
	if messageReaderEvent.Error != nil {
		if messageReaderEvent.Error == io.EOF {
			b.Notifier.NotifyThatThePeerClosedTheTransport()
		} else {
			b.Notifier.NotifyThatAnErrorOccurred(messageReaderEvent.Error)
		}
		return nil, true
	}

	m := messageReaderEvent.IncomingMessage

	if MessageIsADiameterConnectionStateMessage(m) {
		b.Notifier.NotifyThatAStateMachineMessageWasReceivedFromThePeer(m)
	} else {
		b.Notifier.NotifyThatAMessageWasReceivedFromThePeer(m)
	}

	if m.AppID != 0 || m.Code != CapabilitiesExchangeCode || m.IsAnswer() {
		b.Notifier.NotifyThatAnErrorOccurred(fmt.Errorf("expected Capabilities-Exchange Request"))
		return nil, true
	}

	peerIdentity, err := DiameterEntityFromCapabilitiesExchangeMessage(m)
	if err != nil {
		b.Notifier.NotifyThatAnErrorOccurred(err)
		return nil, true
	}

	peer := b.PeerFactory.NewPeerFromDiameterEntity(peerIdentity)

	cea := m.GenerateMatchingResponseWithAvps(b.LocalEntity.CapabilitiesExchangeMandatoryAvpsWithResultCode(cachedResponseCode2001), nil)
	if _, err := b.Transport.Write(cea.Encode()); err != nil {
		b.Notifier.NotifyThatAnErrorOccurred(fmt.Errorf("failed to write Capabilities-Exchange Answer: %s", err))
		return nil, true
	}

	b.Notifier.NotifyThatAStateMachineMessageWasSentToThePeer(cea)

	return peer, false
}

type InitialPeerStatePeerTransportWasOpenedLocally struct{}

func PeerStateStartsWithTransportOpenedTowardPeer() *InitialPeerStatePeerTransportWasOpenedLocally {
	return &InitialPeerStatePeerTransportWasOpenedLocally{}
}

func (s *InitialPeerStatePeerTransportWasOpenedLocally) Execute(b *InitialPeerStateBuilder) (connectedPeer *Peer, aFatalErrorOccurred bool) {
	cer := diameter.NewMessage(diameter.MsgFlagRequest, CapabilitiesExchangeCode, 0, b.SequenceGenerator.NextHopByHopId(), b.SequenceGenerator.NextEndToEndId(), b.LocalEntity.CapabilitiesExchangeMandatoryAvps(), nil)

	if _, err := b.Transport.Write(cer.Encode()); err != nil {
		b.Notifier.NotifyThatAnErrorOccurred(err)
		return nil, true
	}

	b.Notifier.NotifyThatAStateMachineMessageWasSentToThePeer(cer)

	messageReaderEvent := <-b.PeerMessageEventChannel
	if messageReaderEvent.Error != nil {
		if messageReaderEvent.Error == io.EOF {
			b.Notifier.NotifyThatThePeerClosedTheTransport()
		} else {
			b.Notifier.NotifyThatAnErrorOccurred(messageReaderEvent.Error)
		}
		return nil, true
	}

	m := messageReaderEvent.IncomingMessage

	if MessageIsADiameterConnectionStateMessage(m) {
		b.Notifier.NotifyThatAStateMachineMessageWasReceivedFromThePeer(m)
	} else {
		b.Notifier.NotifyThatAMessageWasReceivedFromThePeer(m)
	}

	if m.AppID != 0 || m.Code != CapabilitiesExchangeCode || m.IsRequest() {
		b.Notifier.NotifyThatAnErrorOccurred(fmt.Errorf("expected Capabilities-Exchange Answer"))
		return nil, true
	}

	peerIdentity, err := DiameterEntityFromCapabilitiesExchangeMessage(m)
	if err != nil {
		b.Notifier.NotifyThatAnErrorOccurred(err)
		return nil, true
	}

	peer := b.PeerFactory.NewPeerFromDiameterEntity(peerIdentity)

	return peer, false
}

type PeerStateConnected struct {
	notifier  *PeerStateNotifier
	transport net.Conn
	peer      *Peer
}

func NewPeerStateConnected(notifier *PeerStateNotifier, transport net.Conn, peer *Peer) *PeerStateConnected {
	return &PeerStateConnected{
		notifier:  notifier,
		transport: transport,
		peer:      peer,
	}
}

func (s *PeerStateConnected) DiameterConnectionIsClosedInThisState() bool {
	return false
}

func (s *PeerStateConnected) CanInitiateDisconnectInThisState() bool {
	return true
}

func (s *PeerStateConnected) ProcessIncomingCER(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, err *PeerStateError) {
	return NewPeerStateDisconnected(s.notifier, s.transport, s.peer), nil, &PeerStateError{fmt.Errorf("received Capabilities-Exchange Request on peer that is already connected"), true}
}
func (s *PeerStateConnected) ProcessIncomingCEA(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, err *PeerStateError) {
	return NewPeerStateDisconnected(s.notifier, s.transport, s.peer), nil, &PeerStateError{fmt.Errorf("received Capabilities-Exchange Answer on peer that is already connected"), true}
}
func (s *PeerStateConnected) ProcessIncomingDWR(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, err *PeerStateError) {
	return s, b.DWA(m), nil
}
func (s *PeerStateConnected) ProcessIncomingDWA(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, err *PeerStateError) {
	return s, nil, nil
}
func (s *PeerStateConnected) ProcessIncomingDPR(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, err *PeerStateError) {
	return NewPeerStateDisconnected(s.notifier, s.transport, s.peer), b.DPA(m), nil
}
func (s *PeerStateConnected) ProcessIncomingDPA(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, err *PeerStateError) {
	return NewPeerStateDisconnected(s.notifier, s.transport, s.peer), nil, &PeerStateError{fmt.Errorf("received unsolicited Disconnect-Peer Answer"), true}
}

func (s *PeerStateConnected) ProcessIncomingNonStateMachineMessage(m *diameter.Message) (nextState PeerState, err *PeerStateError) {
	return s, nil
}

type PeerStateHalfClosed struct {
	notifier  *PeerStateNotifier
	transport net.Conn
	peer      *Peer
}

func NewPeerStateHalfClosed(notifier *PeerStateNotifier, transport net.Conn, peer *Peer) *PeerStateHalfClosed {
	return &PeerStateHalfClosed{
		notifier:  notifier,
		transport: transport,
		peer:      peer,
	}
}

func (s *PeerStateHalfClosed) DiameterConnectionIsClosedInThisState() bool {
	return false
}

func (s *PeerStateHalfClosed) CanInitiateDisconnectInThisState() bool {
	return false
}

func (s *PeerStateHalfClosed) ProcessIncomingCER(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, err *PeerStateError) {
	return NewPeerStateDisconnected(s.notifier, s.transport, s.peer), nil, &PeerStateError{fmt.Errorf("received Capabilities-Exchange Request on peer connection that is half-closed"), false}
}
func (s *PeerStateHalfClosed) ProcessIncomingCEA(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, err *PeerStateError) {
	return NewPeerStateDisconnected(s.notifier, s.transport, s.peer), nil, &PeerStateError{fmt.Errorf("received Capabilities-Exchange Answer on peer connection that is half-closed"), false}
}
func (s *PeerStateHalfClosed) ProcessIncomingDWR(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, err *PeerStateError) {
	return s, nil, nil
}
func (s *PeerStateHalfClosed) ProcessIncomingDWA(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, err *PeerStateError) {
	return s, nil, nil
}
func (s *PeerStateHalfClosed) ProcessIncomingDPR(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, err *PeerStateError) {
	return NewPeerStateDisconnected(s.notifier, s.transport, s.peer), nil, &PeerStateError{fmt.Errorf("received Disconnect-Peer Request on peer connection that is half-closed"), false}
}
func (s *PeerStateHalfClosed) ProcessIncomingDPA(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, err *PeerStateError) {
	return NewPeerStateDisconnected(s.notifier, s.transport, s.peer), nil, nil
}

func (s *PeerStateHalfClosed) ProcessIncomingNonStateMachineMessage(m *diameter.Message) (nextState PeerState, err *PeerStateError) {
	return s, nil
}

type PeerStateDisconnected struct {
	notifier  *PeerStateNotifier
	transport net.Conn
	peer      *Peer
}

func NewPeerStateDisconnected(notifier *PeerStateNotifier, transport net.Conn, peer *Peer) *PeerStateDisconnected {
	notifier.NotifyThatDiameterConnectionHasBeenClosed()
	return &PeerStateDisconnected{notifier, transport, peer}
}

func (s *PeerStateDisconnected) DiameterConnectionIsClosedInThisState() bool {
	return true
}

func (s *PeerStateDisconnected) CanInitiateDisconnectInThisState() bool {
	return false
}

func (s *PeerStateDisconnected) ProcessIncomingCER(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, error *PeerStateError) {
	return NewPeerStateDisconnected(s.notifier, s.transport, s.peer), nil, &PeerStateError{fmt.Errorf("received message from a peer that is disconnected"), false}
}
func (s *PeerStateDisconnected) ProcessIncomingCEA(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, error *PeerStateError) {
	return NewPeerStateDisconnected(s.notifier, s.transport, s.peer), nil, &PeerStateError{fmt.Errorf("received message from a peer that is disconnected"), false}
}
func (s *PeerStateDisconnected) ProcessIncomingDWR(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, error *PeerStateError) {
	return NewPeerStateDisconnected(s.notifier, s.transport, s.peer), nil, &PeerStateError{fmt.Errorf("received message from a peer that is disconnected"), false}
}
func (s *PeerStateDisconnected) ProcessIncomingDWA(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, error *PeerStateError) {
	return NewPeerStateDisconnected(s.notifier, s.transport, s.peer), nil, &PeerStateError{fmt.Errorf("received message from a peer that is disconnected"), false}
}
func (s *PeerStateDisconnected) ProcessIncomingDPR(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, error *PeerStateError) {
	return NewPeerStateDisconnected(s.notifier, s.transport, s.peer), nil, &PeerStateError{fmt.Errorf("received message from a peer that is disconnected"), false}
}
func (s *PeerStateDisconnected) ProcessIncomingDPA(m *diameter.Message, b *MessageBuilder) (nextState PeerState, messageToSend *diameter.Message, error *PeerStateError) {
	return NewPeerStateDisconnected(s.notifier, s.transport, s.peer), nil, &PeerStateError{fmt.Errorf("received message from a peer that is disconnected"), false}
}

func (s *PeerStateDisconnected) ProcessIncomingNonStateMachineMessage(m *diameter.Message) (nextState PeerState, err *PeerStateError) {
	return NewPeerStateDisconnected(s.notifier, s.transport, s.peer), &PeerStateError{fmt.Errorf("received message from a peer that is disconnected"), false}
}

func (s *PeerStateDisconnected) ProcessIncomingMessage(m *diameter.Message) (nextState PeerState, closePeerTransport bool) {
	s.notifier.NotifyThatAnErrorOccurred(fmt.Errorf("received message from a peer that is in a disconnected state"))
	return s, true
}

// WatchdogIntervalTimer wraps a time.Timer object.  It exposes the channel of the
// underlying Timer object.  Each time the timer is started (or restarted), the
// interval is set to some base duration with a jitter.  The jittered value is
// randomly selected from the range [base - 2 second .. base + 2 seconds].
// See RFC 3539 section 3.4.1 for an explanation of this.  As with time.Timer,
// WatchdogIntervalTime has a channel -- C -- which this will write to at the
// jittered time for the current interval.  If C is read and the timer should be
// restarted, the method Restart() must be called.  On the other hand if the timer
// should be (re)started but C was no read since the last (re)start, then
// StopAndRestart() must be called.
type WatchdogIntervalTimer struct {
	C                   <-chan time.Time
	timer               *time.Timer
	twFloorBeforeJitter time.Duration
}

// StartNewWatchdogIntervalTimer creates a new watchdog timer, providing an initial
// jittered interval centered on twInitInSeconds.  twInit must not be less than
// 6 seconds (see RFC 3539).
func StartNewWatchdogIntervalTimer(twInitInSeconds uint) *WatchdogIntervalTimer {
	if twInitInSeconds < 6 {
		panic("twInit must be at least 6 seconds")
	}

	twFloorBeforeJitter := time.Duration(twInitInSeconds) * time.Second
	timer := time.NewTimer(newWatchdogIntervalWithJitter(twFloorBeforeJitter))

	return &WatchdogIntervalTimer{
		C:                   timer.C,
		timer:               timer,
		twFloorBeforeJitter: twFloorBeforeJitter,
	}
}

// Restart restarts the time using the twInit with a random jitter.  This method
// may only be called after reading from the channel C.  This means that the
// underlying timer has stopped.  If it has not, this method will panic.
func (t *WatchdogIntervalTimer) Restart() {
	if t.timer.Stop() {
		panic("Restart() cannot be called on a timer that is still active")
	}

	t.timer.Reset(newWatchdogIntervalWithJitter(t.twFloorBeforeJitter))
}

// StopAndRestart does the same as Restart() but may only be called if the channel
// C has not been read since the last restart.  This drains that channel and restarts
// the underlying timer.  If C was read since the last restart, this will deadlock.
func (t *WatchdogIntervalTimer) StopAndRestart() {
	if !t.timer.Stop() {
		<-t.timer.C
	}

	t.timer.Reset(newWatchdogIntervalWithJitter(t.twFloorBeforeJitter))
}

func newWatchdogIntervalWithJitter(twFloorBeforeJitter time.Duration) time.Duration {
	return twFloorBeforeJitter + time.Duration(rand.Intn(4000))*time.Millisecond
}

type messageReaderEvent struct {
	IncomingMessage *diameter.Message
	Error           error
}
