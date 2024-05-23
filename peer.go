package diameter

import (
	"errors"
	"fmt"
	"io"
	"net"
)

type DiameterEntity struct {
	OriginHost      string
	OriginRealm     string
	HostIPAddresses []*net.IP
	VendorID        uint32
	ProductName     string
}

func (e *DiameterEntity) CapabilitiesExchangeMandatoryAvps() []*AVP {
	avps := make([]*AVP, 0, 4+len(e.HostIPAddresses))

	avps = append(avps,
		NewTypedAVP(264, 0, true, DiamIdent, e.OriginHost),
		NewTypedAVP(296, 0, true, DiamIdent, e.OriginRealm),
	)

	for _, ip := range e.HostIPAddresses {
		avps = append(avps, NewTypedAVP(257, 0, true, Address, ip))
	}

	return append(avps,
		NewTypedAVP(266, 0, true, Unsigned32, e.VendorID),
		NewTypedAVP(269, 0, true, UTF8String, e.ProductName),
	)
}

func DiameterEntityFromCapabilitiesExchangeMessage(m *Message) (*DiameterEntity, error) {
	avpsByCode := m.MapOfAvpsByCode()

	for _, avpCode := range []uint32{264, 296, 266, 269} {
		if doesNotHaveExactlyOne(avpsByCode, avpCode) {
			return nil, fmt.Errorf("missing mandatory AVP with code (%d)", avpCode)
		}
	}

	if doesNotHaveAtLeastOne(avpsByCode, 257) {
		return nil, fmt.Errorf("missing mandatory AVP with code (257)")
	}

	e := &DiameterEntity{
		HostIPAddresses: make([]*net.IP, len(avpsByCode[257])),
	}

	if originHost, err := ConvertAVPDataToTypedData(avpsByCode[264][0].Data, DiamIdent); err != nil {
		return nil, fmt.Errorf("Origin-Host AVP cannot be properly decoded: %s", err)
	} else {
		e.OriginHost = originHost.(string)
	}
	if originRealm, err := ConvertAVPDataToTypedData(avpsByCode[296][0].Data, DiamIdent); err != nil {
		return nil, fmt.Errorf("Origin-Realm AVP cannot be properly decoded: %s", err)
	} else {
		e.OriginHost = originRealm.(string)
	}
	if vendorId, err := ConvertAVPDataToTypedData(avpsByCode[266][0].Data, Unsigned32); err != nil {
		return nil, fmt.Errorf("Vendor-Id AVP cannot be properly decoded: %s", err)
	} else {
		e.VendorID = vendorId.(uint32)
	}
	if productName, err := ConvertAVPDataToTypedData(avpsByCode[269][0].Data, UTF8String); err != nil {
		return nil, fmt.Errorf("Product-Name AVP cannot be properly decoded: %s", err)
	} else {
		e.ProductName = productName.(string)
	}

	for i, ipAddressAvp := range avpsByCode[257] {
		if ipAddr, err := ConvertAVPDataToTypedData(ipAddressAvp.Data, Address); err != nil {
			return nil, fmt.Errorf("Host-IP-Address AVP cannot be properly decoded: %s", err)
		} else {
			e.HostIPAddresses[i] = ipAddr.(*net.IP)
		}
	}

	return e, nil
}

func doesNotHaveExactlyOne(avpMap map[uint32][]*AVP, avpWithCode uint32) bool {
	avpSet := avpMap[avpWithCode]
	return avpSet == nil || len(avpSet) != 1
}

func doesNotHaveAtLeastOne(avpMap map[uint32][]*AVP, avpWithCode uint32) bool {
	avpSet := avpMap[avpWithCode]
	return avpSet == nil && len(avpSet) == 0
}

type PeerHandlerEvent struct {
	Type       EventType
	RemotePeer *DiameterEntity
	Conn       net.Conn
	Error      error
}

type PeerHandler struct {
	self                      *DiameterEntity
	peer                      *DiameterEntity
	conn                      net.Conn
	messageStreamReader       *MessageStreamReader
	thePeerInitiatedTransport bool
	eventChannel              chan<- *PeerHandlerEvent
	dwaDpaMandatoryAvps       []*AVP
	peerCloseInitiated        bool
}

func NewInitiatorPeerHandler(self *DiameterEntity, conn net.Conn, eventChannel chan<- *PeerHandlerEvent) *PeerHandler {
	h := newPeerHandler(self, conn, eventChannel)
	h.thePeerInitiatedTransport = true
	return h
}

func NewInitiatedPeerHandler(self *DiameterEntity, conn net.Conn, eventChannel chan<- *PeerHandlerEvent) *PeerHandler {
	return newPeerHandler(self, conn, eventChannel)
}

func newPeerHandler(self *DiameterEntity, conn net.Conn, eventChannel chan<- *PeerHandlerEvent) *PeerHandler {
	if self == nil {
		panic("self must not be null")
	}
	if conn == nil {
		panic("conn must not be nil")
	}

	return &PeerHandler{
		self:                      self,
		peer:                      nil,
		conn:                      conn,
		messageStreamReader:       NewMessageStreamReader(conn),
		thePeerInitiatedTransport: false,
		eventChannel:              eventChannel,
		dwaDpaMandatoryAvps: []*AVP{
			NewTypedAVP(268, 0, true, Unsigned32, 2000),
			NewTypedAVP(264, 0, true, DiamIdent, self.OriginHost),
			NewTypedAVP(296, 0, true, DiamIdent, self.OriginRealm),
		},
		peerCloseInitiated: false,
	}
}

type stateMachineResult struct {
	MessageToSend         *Message
	PeerMayCloseTransport bool
	EventsToSend          []*PeerHandlerEvent
	ReceivedPeerDetails   *DiameterEntity
	AFatalErrorOccurred   bool
}

type initialPeerState func() (nextState successivePeerState, result *stateMachineResult)
type successivePeerState func(msgFromPeer *Message) (nextState successivePeerState, result *stateMachineResult)

func (handler *PeerHandler) Run() {
	defer handler.conn.Close()

	var initialState initialPeerState
	if handler.thePeerInitiatedTransport {
		initialState = handler.initialPeerStateWhenPeerInitiatedTransport
	} else {
		initialState = handler.initialPeerStateWhenTransportWasLocallyInitiated
	}

	nextState, stateResult := initialState()

	for {
		for _, event := range stateResult.EventsToSend {
			handler.eventChannel <- event
		}

		if stateResult.MessageToSend != nil {
			handler.SendMessage(stateResult.MessageToSend)
		}

		if stateResult.ReceivedPeerDetails != nil {
			handler.peer = stateResult.ReceivedPeerDetails
		}

		if stateResult.AFatalErrorOccurred {
			return
		}

		msgFromPeer, err := handler.messageStreamReader.ReadNextMessage()
		if err != nil {
			if err == io.EOF {
				if stateResult.PeerMayCloseTransport {
					handler.notifyOfTrasportClosedEvent()
					return
				}
			}
			handler.notifyOfFatalTransportErrorEvent(err)
			return
		}

		nextState, stateResult = nextState(msgFromPeer)
	}
}

func (handler *PeerHandler) initialPeerStateWhenPeerInitiatedTransport() (nextState successivePeerState, result *stateMachineResult) {
	return handler.peerStateAwaitCER, &stateMachineResult{}
}

func (handler *PeerHandler) initialPeerStateWhenTransportWasLocallyInitiated() (nextState successivePeerState, result *stateMachineResult) {
	cer := NewMessage(MsgFlagRequest, CapabilitiesExchangeCode, 0, 0, 0, handler.self.CapabilitiesExchangeMandatoryAvps(), []*AVP{})

	return handler.peerStateAwaitCEA, &stateMachineResult{MessageToSend: cer}
}

func (handler *PeerHandler) peerStateAwaitCEA(msgFromPeer *Message) (nextState successivePeerState, result *stateMachineResult) {
	if msgFromPeer.Code != CapabilitiesExchangeCode || msgFromPeer.AppID != 0 || msgFromPeer.IsRequest() {
		return nil, &stateMachineResult{
			EventsToSend:        []*PeerHandlerEvent{handler.makeStateMachineErrorEventFromError(errors.New("expected Capabilities-Exchange answer"))},
			AFatalErrorOccurred: true,
		}
	}

	return handler.peerStateSteady, nil
}

func (handler *PeerHandler) peerStateAwaitCER(msgFromPeer *Message) (nextState successivePeerState, result *stateMachineResult) {
	if msgFromPeer.Code != CapabilitiesExchangeCode || msgFromPeer.AppID != 0 || !msgFromPeer.IsRequest() {
		return nil, &stateMachineResult{
			EventsToSend:        []*PeerHandlerEvent{handler.makeStateMachineErrorEventFromError(errors.New("expected Capabilities-Exchange request"))},
			AFatalErrorOccurred: true,
		}
	}

	peerDetails, err := DiameterEntityFromCapabilitiesExchangeMessage(msgFromPeer)
	if err != nil {
		return nil, &stateMachineResult{
			EventsToSend: []*PeerHandlerEvent{handler.makeStateMachineErrorEventFromError(err)},
		}
	}

	cea := msgFromPeer.GenerateMatchingResponseWithAvps(handler.self.CapabilitiesExchangeMandatoryAvps(), []*AVP{})
	return handler.peerStateSteady, &stateMachineResult{
		EventsToSend:        []*PeerHandlerEvent{{Type: PeerConnectionEsablished, RemotePeer: peerDetails, Conn: handler.conn}},
		MessageToSend:       cea,
		ReceivedPeerDetails: peerDetails,
	}
}

func (handler *PeerHandler) peerStateSteady(msgFromPeer *Message) (nextState successivePeerState, result *stateMachineResult) {
	if msgFromPeer.AppID == 0 {
		switch msgFromPeer.Code {
		case CapabilitiesExchangeCode:
			return nil, &stateMachineResult{
				EventsToSend:        []*PeerHandlerEvent{handler.makeStateMachineErrorEventFromError(errors.New("received unexpected Capabilities-Exchange message"))},
				AFatalErrorOccurred: true,
			}

		case DeviceWatchdogCode:
			if msgFromPeer.IsRequest() {
				dwa := msgFromPeer.GenerateMatchingResponseWithAvps(handler.dwaDpaMandatoryAvps, []*AVP{})

				return handler.peerStateSteady, &stateMachineResult{
					MessageToSend: dwa,
				}
			}

		case DisconnectPeerCode:
			if !msgFromPeer.IsRequest() {
				return nil, &stateMachineResult{
					EventsToSend:        []*PeerHandlerEvent{handler.makeStateMachineErrorEventFromError(errors.New("received unsolicited Disconnect-Peer answer"))},
					AFatalErrorOccurred: true,
				}
			}

			return nil, nil
		}
	}

	return handler.peerStateSteady, nil
}

func (handler *PeerHandler) InitiateDisconnect() {
}

func (handler *PeerHandler) SendMessage(msg *Message) {
}

func (handler *PeerHandler) notifyOfFatalTransportErrorEvent(err error) error {
	if err == io.EOF {
		handler.eventChannel <- &PeerHandlerEvent{
			Type:       PeerConnectionTerminated,
			Conn:       handler.conn,
			RemotePeer: handler.peer,
		}
	} else {
		handler.eventChannel <- &PeerHandlerEvent{
			Type:       TransportError,
			Conn:       handler.conn,
			RemotePeer: handler.peer,
			Error:      err,
		}
	}

	return err
}

func (handler *PeerHandler) makeStateMachineErrorEventFromError(err error) *PeerHandlerEvent {
	return &PeerHandlerEvent{
		Type:       StateMachineError,
		Conn:       handler.conn,
		RemotePeer: handler.peer,
		Error:      err,
	}
}

func (handler *PeerHandler) notifyOfReceivedMessage(msg *Message) {
	handler.eventChannel <- &PeerHandlerEvent{
		Type:       MessageReceivedFromPeer,
		Conn:       handler.conn,
		RemotePeer: handler.peer,
	}
}

func (handler *PeerHandler) notifyOfPeerConnection() {
	handler.eventChannel <- &PeerHandlerEvent{
		Type:       PeerConnectionEsablished,
		Conn:       handler.conn,
		RemotePeer: handler.peer,
	}
}

func (handler *PeerHandler) notifyOfTrasportClosedEvent() {
	handler.eventChannel <- &PeerHandlerEvent{
		Type:       TransportClosed,
		Conn:       handler.conn,
		RemotePeer: handler.peer,
	}
}

// type PeerType int

// const (
// 	Initator  PeerType = iota
// 	Responder PeerType = iota
// )

// type PeerConnectionInformation struct {
// 	RemoteAddress     *net.IP
// 	RemotePort        uint16
// 	TransportProtocol string
// 	LocalAddress      *net.IP
// 	LocalPort         uint16
// }

// type Peer struct {
// 	TypeOfPeer                  PeerType
// 	PeerCapabilitiesInformation *CapabiltiesExchangeInformation
// 	ConnectionInformation       PeerConnectionInformation
// }

// type PeerHandler struct {
// 	peer                   *Peer
// 	connection             net.Conn
// 	diameterByteReader     *MessageByteReader
// 	eventChannel           chan<- *NodeEvent
// 	flowReadBuffer         []byte
// 	myCapabilities         *CapabiltiesExchangeInformation
// 	nextHopByHopIdentifier uint32
// 	nextEndToEndIdentifier uint32
// }

// func NewHandlerForInitiatorPeer(flowConnection net.Conn, eventChannel chan<- *NodeEvent) *PeerHandler {
// 	return newPeerHandler(flowConnection, Initator, eventChannel)
// }

// func NewHandlerForResponderPeer(flowConnection net.Conn, eventChannel chan<- *NodeEvent) *PeerHandler {
// 	return newPeerHandler(flowConnection, Responder, eventChannel)
// }

// func (handler *PeerHandler) WithCapabilities(capabilities *CapabiltiesExchangeInformation) *PeerHandler {
// 	handler.myCapabilities = capabilities
// 	return handler
// }

// func (handler *PeerHandler) SeedIdentifiers() (*PeerHandler, error) {
// 	initialEndToEndIdentifier, err := generateIdentifierSeedValue()
// 	if err != nil {
// 		return handler, fmt.Errorf("failed to generate cryptographic seed for end-to-end identifier: %s", err.Error())
// 	}

// 	initialHopByHopIdentifier, err := generateIdentifierSeedValue()
// 	if err != nil {
// 		return handler, fmt.Errorf("failed to generate cryptographic seed for hop-by-hop identifier: %s", err.Error())
// 	}

// 	handler.nextEndToEndIdentifier = initialEndToEndIdentifier
// 	handler.nextHopByHopIdentifier = initialHopByHopIdentifier

// 	return handler, nil
// }

// func (handler *PeerHandler) WithSeededIdentifiers() (*PeerHandler, error) {
// 	return handler.SeedIdentifiers()
// }

// func newPeerHandler(flowConnection net.Conn, typeOfPeer PeerType, eventChannel chan<- *NodeEvent) *PeerHandler {
// 	var localIPAddr, remoteIPAddr *net.IP
// 	var localPort, remotePort uint16

// 	if flowConnection.LocalAddr().Network() == "tcp" {
// 		localTCPAddr := flowConnection.LocalAddr().(*net.TCPAddr)
// 		remoteTCPAddr := flowConnection.RemoteAddr().(*net.TCPAddr)

// 		localIPAddr = &localTCPAddr.IP
// 		localPort = uint16(localTCPAddr.Port)
// 		remoteIPAddr = &remoteTCPAddr.IP
// 		remotePort = uint16(remoteTCPAddr.Port)
// 	} else {
// 		localIPAddr, localPort = extractIPAddressAndPortFromAddrNetworkString(flowConnection.LocalAddr().String())
// 		remoteIPAddr, remotePort = extractIPAddressAndPortFromAddrNetworkString(flowConnection.RemoteAddr().String())
// 	}

// 	return &PeerHandler{
// 		peer: &Peer{
// 			TypeOfPeer:                  typeOfPeer,
// 			PeerCapabilitiesInformation: nil,
// 			ConnectionInformation: PeerConnectionInformation{
// 				RemoteAddress:     remoteIPAddr,
// 				RemotePort:        remotePort,
// 				LocalAddress:      localIPAddr,
// 				LocalPort:         localPort,
// 				TransportProtocol: flowConnection.LocalAddr().Network(),
// 			},
// 		},
// 		connection:         flowConnection,
// 		eventChannel:       eventChannel,
// 		diameterByteReader: NewMessageByteReader(),
// 		flowReadBuffer:     make([]byte, 9000),
// 	}
// }

// func generateIdentifierSeedValue() (uint32, error) {
// 	randBytes := make([]byte, 3)
// 	if _, err := rand.Read(randBytes); err != nil {
// 		return 0, err
// 	}

// 	var seedLower20 uint32 = (uint32(randBytes[0]) << 12) | (uint32(randBytes[1]) << 4) | (uint32(randBytes[2] >> 4))
// 	var seed uint32 = (uint32(time.Now().Unix()) << 20) | seedLower20

// 	return seed, nil
// }

// func extractIPAddressAndPortFromAddrNetworkString(networkAddress string) (*net.IP, uint16) {
// 	parts := strings.Split(networkAddress, ":")
// 	if len(parts) < 2 {
// 		panic(fmt.Sprintf("provided invalid IP transport address: %s", networkAddress))
// 	}

// 	portAsString := parts[len(parts)-1]
// 	portAsUint64, err := strconv.ParseUint(portAsString, 10, 16)
// 	if err != nil {
// 		panic(fmt.Sprintf("provided invalid IP transport address (port xlat failed): %s", portAsString))
// 	}

// 	ipAsString := strings.Join(parts[:len(parts)-1], ":")
// 	ipAsString = strings.Trim(ipAsString, "[]")

// 	ipAddr := net.ParseIP(ipAsString)
// 	if ipAddr == nil {
// 		panic(fmt.Sprintf("provided invalid IP transport address (IP xlat failed): %s", ipAsString))
// 	}

// 	return &ipAddr, uint16(portAsUint64)
// }

// func (handler *PeerHandler) StartHandling() {
// 	defer handler.connection.Close()

// 	if handler.myCapabilities == nil {
// 		panic("attempt to StartHandling without having set local agent capabilities")
// 	}

// 	if err := handler.completeCapabilitiesExchangeWithPeer(); err != nil {
// 		return
// 	}

// }

// func (handler *PeerHandler) completeCapabilitiesExchangeWithPeer() error {
// 	if handler.peer.TypeOfPeer == Initator {
// 		cer := handler.waitForCER()
// 		if cer == nil {
// 			return fmt.Errorf("failed to receive CER")
// 		}

// 		ceaToSendInResponse := handler.myCapabilities.MakeCEA().BecomeAnAnswerBasedOnTheRequestMessage(cer)

// 		err := handler.SendMessageToPeer(ceaToSendInResponse)
// 		if err != nil {
// 			handler.sendFatalTransportErrorEvent(err)
// 			handler.sendCapabilitiesExchangeFailureEvent(fmt.Errorf("failed to send CEA"))
// 			return err
// 		}

// 		return nil
// 	}

// 	// peer is Responder
// 	cer := handler.myCapabilities.MakeCER()
// 	handler.populateMessageIdentifiers(cer)

// 	err := handler.SendMessageToPeer(cer)
// 	if err != nil {
// 		handler.sendFatalTransportErrorEvent(err)
// 		handler.sendCapabilitiesExchangeFailureEvent(fmt.Errorf("failed to send CER"))
// 	}

// 	cea := handler.waitForCEA()
// 	if cea == nil {
// 		return fmt.Errorf("failed to receive CEA")
// 	}

// 	handler.peer.PeerCapabilitiesInformation, err = extractCapabilitiesForPeerFromCapabilitiesExchangeMessage(cea)
// 	if err != nil {
// 		handler.sendCapabilitiesExchangeFailureEvent(err)
// 		return err
// 	}

// 	return nil
// }

// func extractCapabilitiesForPeerFromCapabilitiesExchangeMessage(message *Message) (*CapabiltiesExchangeInformation, error) {
// 	var originHost, originRealm, productName string
// 	var hostIPAddresses []*net.IPAddr
// 	var vendorID uint32

// 	var foundOriginHost, foundOriginRealm, foundProductName, foundHostIPAddresses, foundVendorID bool

// 	additionalAVPs := make([]*AVP, 10)

// 	for _, avp := range message.Avps {
// 		switch avp.Code {
// 		case 264:
// 			originHost = string(avp.Data)
// 			foundOriginHost = true

// 		case 296:
// 			originRealm = string(avp.Data)
// 			foundOriginRealm = true

// 		case 269:
// 			productName = string(avp.Data)
// 			foundProductName = true

// 		case 266:
// 			wrappedVendorID, err := avp.ConvertDataToTypedData(Unsigned32)
// 			if err != nil {
// 				return nil, fmt.Errorf("Vendor-ID AVP is improperly formatted")
// 			}
// 			vendorID = wrappedVendorID.(uint32)
// 			foundVendorID = true

// 		case 257:
// 			wrappedIP, err := avp.ConvertDataToTypedData(Address)
// 			if err != nil {
// 				return nil, fmt.Errorf("Host-IP-Address AVP contains invalid data: %s", err.Error())
// 			}

// 			ip := wrappedIP.(*net.IP)
// 			hostIPAddresses = append(hostIPAddresses, &net.IPAddr{IP: *ip, Zone: ""})

// 		default:
// 			additionalAVPs = append(additionalAVPs, avp)
// 		}
// 	}

// 	if !foundOriginHost {
// 		return nil, fmt.Errorf("peer asserted no Origin-Host")
// 	}
// 	if !foundOriginRealm {
// 		return nil, fmt.Errorf("peer asserted no Origin-Realm")
// 	}
// 	if !foundHostIPAddresses {
// 		return nil, fmt.Errorf("peer asserted no Host-IP-Addresses")
// 	}
// 	if !foundProductName {
// 		return nil, fmt.Errorf("peer asserted no Product-Name")
// 	}
// 	if !foundVendorID {
// 		return nil, fmt.Errorf("peer asserted no Vendor-ID")
// 	}

// 	return &CapabiltiesExchangeInformation{
// 		OriginHost:                 originHost,
// 		OriginRealm:                originRealm,
// 		HostIPAddresses:            hostIPAddresses,
// 		VendorID:                   vendorID,
// 		ProductName:                productName,
// 		AdditionalAVPsToSendToPeer: additionalAVPs,
// 	}, nil
// }

// func (handler *PeerHandler) populateMessageIdentifiers(message *Message) {
// 	message.HopByHopID = handler.nextHopByHopIdentifier
// 	message.EndToEndID = handler.nextEndToEndIdentifier

// 	handler.nextEndToEndIdentifier++
// 	handler.nextEndToEndIdentifier++
// }

// func (handler *PeerHandler) waitForCER() (cerMessage *Message) {
// 	var messages []*Message

// 	for {
// 		bytesRead, err := handler.connection.Read(handler.flowReadBuffer)
// 		if err != nil {
// 			handler.sendTransportReadErrorEvent(err)
// 			return nil
// 		}

// 		messages, err = handler.diameterByteReader.ReceiveBytes(handler.flowReadBuffer[:bytesRead])
// 		if err != nil {
// 			handler.sendUnableToParseIncomingMessageStreamEvent(err)
// 			return nil
// 		}

// 		if len(messages) > 0 {
// 			break
// 		}
// 	}

// 	if messageIsNotACER(messages[0]) {
// 		handler.sendCapabilitiesExchangeFailureEvent(fmt.Errorf("first message from peer is not a CER"))
// 		return nil
// 	}

// 	if len(messages) > 1 {
// 		handler.sendCapabilitiesExchangeFailureEvent(fmt.Errorf("peer sent messages before completing capabilities-exchange"))
// 		return nil
// 	}

// 	return messages[0]
// }

// func (handler *PeerHandler) waitForCEA() (ceaMessage *Message) {
// 	var message *Message

// 	for {
// 		bytesRead, err := handler.connection.Read(handler.flowReadBuffer)
// 		if err != nil {
// 			handler.sendTransportReadErrorEvent(err)
// 			return nil
// 		}

// 		message, err = handler.diameterByteReader.ReceiveBytesButReturnAtMostOneMessage(handler.flowReadBuffer[:bytesRead])
// 		if err != nil {
// 			handler.sendUnableToParseIncomingMessageStreamEvent(err)
// 			return nil
// 		}

// 		if message != nil {
// 			break
// 		}
// 	}

// 	if messageIsNotACEA(message) {
// 		handler.sendCapabilitiesExchangeFailureEvent(fmt.Errorf("first message from peer is not a CEA"))
// 		return nil
// 	}

// 	return message
// }

// func (handler *PeerHandler) sendCapabilitiesExchangeAnswerBasedOnCER(cer *Message) error {
// 	cea := handler.generateCEA().BecomeAnAnswerBasedOnTheRequestMessage(cer)

// 	err := handler.SendMessageToPeer(cea)
// 	if err != nil {
// 		handler.sendFatalTransportErrorEvent(err)
// 		return fmt.Errorf("transport error occured after sending CEA")
// 	}

// 	return nil
// }

// func (handler *PeerHandler) sendFatalTransportErrorEvent(err error) {
// 	handler.eventChannel <- &NodeEvent{
// 		Type:       FatalTransportError,
// 		Peer:       handler.peer,
// 		Connection: handler.connection,
// 		Error:      err,
// 	}
// }

// func (handler *PeerHandler) generateCEA() *Message {
// 	return NewMessage(0, 257, 0, 0, 0, nil, nil)
// }

// func (handler *PeerHandler) sendCapabilitiesExchangeFailureEvent(err error) {
// 	handler.eventChannel <- &NodeEvent{
// 		Type:       CapabilitiesExchangeFailed,
// 		Peer:       handler.peer,
// 		Connection: handler.connection,
// 		Error:      err,
// 	}
// }

// func messageIsNotACER(message *Message) bool {
// 	return message.AppID != 0 || message.Code != 257 || !message.IsRequest()
// }

// func messageIsNotACEA(message *Message) bool {
// 	return message.AppID != 0 || message.Code != 257 || message.IsRequest()
// }

// func (handler *PeerHandler) sendUnableToParseIncomingMessageStreamEvent(readError error) {
// 	handler.eventChannel <- &NodeEvent{
// 		Type:       UnableToParseIncomingMessageStream,
// 		Peer:       handler.peer,
// 		Connection: handler.connection,
// 		Error:      readError,
// 	}
// }

// func (handler *PeerHandler) sendTransportReadErrorEvent(err error) {
// 	if err == io.EOF {
// 		handler.eventChannel <- &NodeEvent{
// 			Type:       TransportClosed,
// 			Peer:       handler.peer,
// 			Connection: handler.connection,
// 		}
// 	} else {
// 		handler.eventChannel <- &NodeEvent{
// 			Type:       FatalTransportError,
// 			Peer:       handler.peer,
// 			Connection: handler.connection,
// 			Error:      err,
// 		}
// 	}
// }

// func (handler *PeerHandler) SendMessageToPeer(message *Message) error {
// 	_, err := handler.connection.Write(message.Encode())
// 	return err
// }

// func (handler *PeerHandler) CloseDiameterFlow() error {
// 	return nil
// }

// func (handler *PeerHandler) Terminate() error {
// 	return nil
// }

// type IncomingPeerListener struct {
// 	listener                net.Listener
// 	capabilitiesInformation *CapabiltiesExchangeInformation
// 	cloneableCEA            *Message
// }

// func NewIncomingPeerListener(usingUnderlyingListener net.Listener, usingCapabilitiesInformation CapabiltiesExchangeInformation) *IncomingPeerListener {
// 	copyOfCapabilitiesInformation := usingCapabilitiesInformation
// 	return &IncomingPeerListener{
// 		listener:                usingUnderlyingListener,
// 		capabilitiesInformation: &copyOfCapabilitiesInformation,
// 		cloneableCEA:            (&copyOfCapabilitiesInformation).MakeCEAUsingResultCode(2002),
// 	}
// }

// func (peerListener *IncomingPeerListener) StartListening(eventChannel chan<- *NodeEvent) {
// 	for {
// 		flowConnection, err := peerListener.listener.Accept()
// 		if err != nil {
// 			eventChannel <- &NodeEvent{
// 				Type:       RecoverableTransportError,
// 				Connection: flowConnection,
// 				Error:      err,
// 			}
// 		}

// 		peerHandler, err := NewHandlerForInitiatorPeer(flowConnection, eventChannel).WithCapabilities(peerListener.capabilitiesInformation).WithSeededIdentifiers()
// 		if err != nil {
// 			eventChannel <- &NodeEvent{
// 				Type:  InternalFailure,
// 				Error: fmt.Errorf("failed to create PeerHandler: %s", err.Error()),
// 			}

// 			flowConnection.Close()
// 			continue
// 		}

// 		go peerHandler.StartHandling()
// 	}
// }

// func (peerListener *IncomingPeerListener) StopListening() error {
// 	return nil
// }
