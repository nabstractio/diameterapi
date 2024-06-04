package agent

import (
	"fmt"
	"net"

	"github.com/blorticus-go/diameter"
)

type diameterEntityCache struct {
	OriginHost      *diameter.AVP
	OriginRealm     *diameter.AVP
	ResultCode      *diameter.AVP
	HostIPAddresses []*diameter.AVP
	VendorId        *diameter.AVP
	ProductName     *diameter.AVP
}

const (
	CapabilitiesExchangeCode = 257
	DeviceWatchdogCode       = 280
	DisconnectPeerCode       = 282
)

// A DiameterEntity provides identifying information about a diameter entity.  The first time an *Avp()
// method is invoked, the AVP it returns is first cached.  Subsequent calls are returned from this cached
// value.  This mechanism assumes the values of the AVPs in a DiameterEntity instance are not changed
// after an instance is created.
type DiameterEntity struct {
	OriginHost      string
	OriginRealm     string
	HostIPAddresses []*net.IP
	VendorID        uint32
	ProductName     string

	cache diameterEntityCache
}

// OriginHostAvp returns the OriginHost as an AVP.
func (e *DiameterEntity) OriginHostAvp() *diameter.AVP {
	if e.cache.OriginHost == nil {
		e.cache.OriginHost = diameter.NewTypedAVP(264, 0, true, diameter.DiamIdent, e.OriginHost)
	}

	return e.cache.OriginHost
}

// OriginRealmAvp returns the OriginRealm as an AVP.
func (e *DiameterEntity) OriginRealmAvp() *diameter.AVP {
	if e.cache.OriginRealm == nil {
		e.cache.OriginRealm = diameter.NewTypedAVP(296, 0, true, diameter.DiamIdent, e.OriginHost)
	}

	return e.cache.OriginRealm
}

// VendorIdAVP returns the VendorId as an AVP.
func (e *DiameterEntity) VendorIdAVP() *diameter.AVP {
	if e.cache.VendorId == nil {
		e.cache.VendorId = diameter.NewTypedAVP(266, 0, true, diameter.Unsigned32, e.VendorID)
	}

	return e.cache.VendorId
}

// ProductNameAvp returns the ProductName as an AVP.
func (e *DiameterEntity) ProductNameAvp() *diameter.AVP {
	if e.cache.ProductName == nil {
		e.cache.ProductName = diameter.NewTypedAVP(269, 0, true, diameter.UTF8String, e.ProductName)
	}

	return e.cache.ProductName
}

// HostIpAddressAvps returns the HostIPAddresses set as a set of AVPs.
func (e *DiameterEntity) HostIpAddressAvps() []*diameter.AVP {
	if len(e.cache.HostIPAddresses) == 0 {
		avps := make([]*diameter.AVP, len(e.HostIPAddresses))
		for i, avp := range e.HostIPAddresses {
			avps[i] = diameter.NewTypedAVP(257, 0, true, diameter.Address, avp)
		}
		e.cache.HostIPAddresses = avps
	}

	return e.cache.HostIPAddresses
}

// CapabilitiesExchangeMandatoryAvps generates the mandatory attributes required for
// a Capabilities-Exchange request or answer based on the DiameterEntity values.
func (e *DiameterEntity) CapabilitiesExchangeMandatoryAvps() []*diameter.AVP {
	avps := make([]*diameter.AVP, 0, 4+len(e.HostIPAddresses))

	avps = append(avps,
		e.OriginHostAvp(),
		e.OriginRealmAvp(),
	)

	avps = append(avps, e.HostIpAddressAvps()...)

	return append(avps,
		e.VendorIdAVP(),
		e.ProductNameAvp(),
	)
}

// CapabilitiesExchangeMandatoryAvps generates the mandatory attributes required for
// a Capabilities-Exchange request or answer based on the DiameterEntity values.
func (e *DiameterEntity) CapabilitiesExchangeMandatoryAvpsWithResultCode(resultCodeAvp *diameter.AVP) []*diameter.AVP {
	avps := make([]*diameter.AVP, 0, 5+len(e.HostIPAddresses))

	avps = append(avps,
		resultCodeAvp,
		e.OriginHostAvp(),
		e.OriginRealmAvp(),
	)

	avps = append(avps, e.HostIpAddressAvps()...)

	return append(avps,
		e.VendorIdAVP(),
		e.ProductNameAvp(),
	)
}

// DiameterEntityFromCapabilitiesExchangeMessage reads a Capabilities-Exchange request or
// answer and extracts the AVPs providing the DiameterEntity information.  Returns an
// error if the message does not contain mandatory AVPs or if the AVPs are malformed.
func DiameterEntityFromCapabilitiesExchangeMessage(m *diameter.Message) (*DiameterEntity, error) {
	for _, avpCode := range []diameter.Uint24{264, 296, 266, 269} {
		if m.NumberOfTopLevelAvpsMatching(0, avpCode) != 1 {
			return nil, fmt.Errorf("missing mandatory AVP with code (%d)", avpCode)
		}
	}

	if m.NumberOfTopLevelAvpsMatching(0, diameter.Uint24(257)) == 0 {
		return nil, fmt.Errorf("missing mandatory AVP with code (257)")
	}

	hostIpAvps := m.TopLevelAvpsMatching(0, 257)

	e := &DiameterEntity{
		HostIPAddresses: make([]*net.IP, len(hostIpAvps)),
	}

	if originHost, err := diameter.ConvertAVPDataToTypedData(m.FirstAvpMatching(0, 264).Data, diameter.DiamIdent); err != nil {
		return nil, fmt.Errorf("Origin-Host AVP cannot be properly decoded: %s", err)
	} else {
		e.OriginHost = originHost.(string)
	}
	if originRealm, err := diameter.ConvertAVPDataToTypedData(m.FirstAvpMatching(0, 296).Data, diameter.DiamIdent); err != nil {
		return nil, fmt.Errorf("Origin-Realm AVP cannot be properly decoded: %s", err)
	} else {
		e.OriginRealm = originRealm.(string)
	}
	if vendorId, err := diameter.ConvertAVPDataToTypedData(m.FirstAvpMatching(0, 266).Data, diameter.Unsigned32); err != nil {
		return nil, fmt.Errorf("Vendor-Id AVP cannot be properly decoded: %s", err)
	} else {
		e.VendorID = vendorId.(uint32)
	}
	if productName, err := diameter.ConvertAVPDataToTypedData(m.FirstAvpMatching(0, 269).Data, diameter.UTF8String); err != nil {
		return nil, fmt.Errorf("Product-Name AVP cannot be properly decoded: %s", err)
	} else {
		e.ProductName = productName.(string)
	}

	for i, ipAddressAvp := range hostIpAvps {
		if ipAddr, err := diameter.ConvertAVPDataToTypedData(ipAddressAvp.Data, diameter.Address); err != nil {
			return nil, fmt.Errorf("Host-IP-Address AVP cannot be properly decoded: %s", err)
		} else {
			ipAddr := ipAddr.(net.IP)
			e.HostIPAddresses[i] = &ipAddr
		}
	}

	return e, nil
}

// Peer represents a diameter peer.  It provides peer identity information and methods
// for sending messages to the peer.
type Peer struct {
	Identity                     DiameterEntity
	sendMessageMethod            func(m *diameter.Message) error
	initiatePeerDisconnectMethod func() error
}

func NewPeer(entityInformation *DiameterEntity, sendMessageMethod func(m *diameter.Message) error, initiatePeerDisconnectMethod func() error) *Peer {
	return &Peer{
		Identity:                     *entityInformation,
		sendMessageMethod:            sendMessageMethod,
		initiatePeerDisconnectMethod: initiatePeerDisconnectMethod,
	}
}

// SendMessage attempts to deliver a Diameter message to the peer.  Returns an error
// if the delivery fails either because the peer is no longer connected or because of
// a transport failure.
func (peer *Peer) SendMessage(m *diameter.Message) error {
	return peer.sendMessageMethod(m)
}

// InitiateDisconnect start the Disconnect Peer procedure by sending a Disconnect-Peer
// request to the peer.
func (peer *Peer) InitiateDisconnect() error {
	return peer.initiatePeerDisconnectMethod()
}

// IsInAConnectedState indicates whether the peer is in a connected state.  This means
// that the transport is active, a Capabilities-Exchange has succesfully completed,
// and a Disconnect Peer procedure is neither pending nor has been completed.
func (peer *Peer) IsInAConnectedState() bool {
	return false
}

// IsDisconnected is the inverse of IsInAConnectedState() and is provided to improve
// readability of conditionals that want to check for a disconnected state.
func (peer *Peer) IsDisconnected() bool {
	return !peer.IsInAConnectedState()
}

// PeerFactory provides a constructor for Peer objects without the caller having to know
// the details of the callback methods.
type PeerFactory struct {
	sendMessageMethod            func(m *diameter.Message) error
	initiatePeerDisconnectMethod func() error
}

// NewPeerFactory creates a new PeerFactory
func NewPeerFactory(sendMessageMethod func(m *diameter.Message) error, initiatePeerDisconnectMethod func() error) *PeerFactory {
	return &PeerFactory{
		sendMessageMethod:            sendMessageMethod,
		initiatePeerDisconnectMethod: initiatePeerDisconnectMethod,
	}
}

// NewPeerFromDiameterEntity returns a new Peer using the supplied DiameterEntity
func (f *PeerFactory) NewPeerFromDiameterEntity(entity *DiameterEntity) *Peer {
	return NewPeer(entity, f.sendMessageMethod, f.initiatePeerDisconnectMethod)
}
