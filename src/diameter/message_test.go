package diameter

import (
	"net"
	"testing"
)

type FlagsTestSet struct {
	value       uint8
	isRequest   bool
	isProxiable bool
	isError     bool
	isPr        bool
}

var flagstest = []FlagsTestSet{
	{0x10, false, false, false, true}, {0x20, false, false, true, false}, {0x30, false, false, true, true}, {0x40, false, true, false, false},
	{0x50, false, true, false, true}, {0x60, false, true, true, false}, {0x70, false, true, true, true}, {0x80, true, false, false, false},
	{0x90, true, false, false, true}, {0xa0, true, false, true, false}, {0xb0, true, false, true, true}, {0xc0, true, true, false, false},
	{0xd0, true, true, false, true}, {0xe0, true, true, true, false}, {0xf0, true, true, true, true},
}

func TestFlags(t *testing.T) {
	var i uint8

	for i = 0; i <= 0x0f; i++ {
		m := Message{Flags: i}

		if m.IsRequest() {
			t.Error("For value ", i, " should not be request")
		}
		if m.IsProxiable() {
			t.Error("For value ", i, " should not be proxiable")
		}
		if m.IsError() {
			t.Error("For value ", i, " should not be error")
		}
		if m.IsPotentiallyRetransmitted() {
			t.Error("For value ", i, " should not be PR")
		}
	}

	for _, set := range flagstest {
		m := Message{Flags: set.value}

		if m.IsRequest() != set.isRequest {
			t.Error("For value ", i, " expect isRequest == ", set.isRequest, " but does not match")
		}
		if m.IsProxiable() != set.isProxiable {
			t.Error("For value ", i, " expect isProxiable == ", set.isProxiable, " but does not match")
		}
		if m.IsError() != set.isError {
			t.Error("For value ", i, " expect isError == ", set.isError, " but does not match")
		}
		if m.IsPotentiallyRetransmitted() != set.isPr {
			t.Error("For value ", i, " expect isPotentiallyRetransmittable == ", set.isPr, " but does not match")
		}

	}
}

type EncodeTestSet struct {
	flags         uint8
	code          Uint24
	appID         uint32
	hopByHopID    uint32
	endToEndID    uint32
	mandatoryAvps []*AVP
	optionalAvps  []*AVP
	encoded       []byte
}

// encodetests is a list of structs.  The structs are EncodeTestSet.  Everything but 'encoded' is passed to diameter.NewMessage.
// The message is encoded, and the encoded stream is compared to 'encoded'
var encodetests = []EncodeTestSet{
	// -- TEST 1: just header
	{0x00 | MsgFlagRequest | MsgFlagProxiable, 203, 0, 0x10101010, 0xabcd0000, []*AVP{}, []*AVP{},
		[]byte{0x01, 0x00, 0x00, 0x14, 0xc0, 0x00, 0x00, 0xcb, 0x00, 0x00, 0x00, 0x00, 0x10, 0x10, 0x10, 0x10, 0xab, 0xcd, 0x00, 0x00}},
	// -- TEST 2: CER with only mandatory AVPs
	{0x00 | MsgFlagRequest | MsgFlagProxiable, 257, 0, 0x10101010, 0xabcd0000,
		[]*AVP{
			NewAVP(NewAVPAttribute("Origin-Host", 264, 0, DiamIdent), true, false, nil, "host.example.com"),
			NewAVP(NewAVPAttribute("Origin-Realm", 296, 0, DiamIdent), true, false, nil, "example.com"),
			NewAVP(NewAVPAttribute("Host-IP-Address", 257, 0, Address), true, false, nil, net.ParseIP("10.20.30.1")),
			NewAVP(NewAVPAttribute("Vendor-Id", 266, 0, Unsigned32), true, false, nil, uint32(0)),
			NewAVP(NewAVPAttribute("Product-Name", 269, 0, UTF8String), true, false, nil, "GoDiameter"),
		},
		[]*AVP{},
		[]byte{
			// header
			0x01, 0x00, 0x00, 0x70,
			0xc0, 0x00, 0x01, 0x01,
			0x00, 0x00, 0x00, 0x00,
			0x10, 0x10, 0x10, 0x10,
			0xab, 0xcd, 0x00, 0x00,
			// Origin-Host
			0x00, 0x00, 0x01, 0x08, 0x40, 0x00, 0x00, 0x18, 0x68, 0x6f, 0x73, 0x74, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d,
			// Origin-Realm
			0x00, 0x00, 0x01, 0x28, 0x40, 0x00, 0x00, 0x13, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x00,
			//Host-IP-Address
			0x00, 0x00, 0x01, 0x01, 0x40, 0x00, 0x00, 0x0e, 0x00, 0x01, 0x0a, 0x14, 0x1e, 0x01, 0x00, 0x00,
			// Vendor-Id
			0x00, 0x00, 0x01, 0x0a, 0x40, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x00,
			// Product-Name
			0x00, 0x00, 0x01, 0x0d, 0x40, 0x00, 0x00, 0x12, 0x47, 0x6f, 0x44, 0x69, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x00, 0x00}},
	// -- TEST 3: Same CER, but set flag of some mandatory AVPs to not mandatory; they should be flipped to mandatory
	{0x00 | MsgFlagRequest | MsgFlagProxiable, 257, 0, 0x10101010, 0xabcd0000,
		[]*AVP{
			NewAVP(NewAVPAttribute("Origin-Host", 264, 0, DiamIdent), false, false, nil, "host.example.com"),
			NewAVP(NewAVPAttribute("Origin-Realm", 296, 0, DiamIdent), true, false, nil, "example.com"),
			NewAVP(NewAVPAttribute("Host-IP-Address", 257, 0, Address), false, false, nil, net.ParseIP("10.20.30.1")),
			NewAVP(NewAVPAttribute("Vendor-Id", 266, 0, Unsigned32), false, false, nil, uint32(0)),
			NewAVP(NewAVPAttribute("Product-Name", 269, 0, UTF8String), true, false, nil, "GoDiameter"),
		},
		[]*AVP{},
		[]byte{
			// header
			0x01, 0x00, 0x00, 0x70,
			0xc0, 0x00, 0x01, 0x01,
			0x00, 0x00, 0x00, 0x00,
			0x10, 0x10, 0x10, 0x10,
			0xab, 0xcd, 0x00, 0x00,
			// Origin-Host
			0x00, 0x00, 0x01, 0x08, 0x40, 0x00, 0x00, 0x18, 0x68, 0x6f, 0x73, 0x74, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d,
			// Origin-Realm
			0x00, 0x00, 0x01, 0x28, 0x40, 0x00, 0x00, 0x13, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x00,
			//Host-IP-Address
			0x00, 0x00, 0x01, 0x01, 0x40, 0x00, 0x00, 0x0e, 0x00, 0x01, 0x0a, 0x14, 0x1e, 0x01, 0x00, 0x00,
			// Vendor-Id
			0x00, 0x00, 0x01, 0x0a, 0x40, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x00,
			// Product-Name
			0x00, 0x00, 0x01, 0x0d, 0x40, 0x00, 0x00, 0x12, 0x47, 0x6f, 0x44, 0x69, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x00, 0x00,
		}},
	// -- TEST 4: Same as previous test, but add some optional AVPs, also with mandatory flags set and unset
	{0x00 | MsgFlagRequest | MsgFlagProxiable, 257, 0, 0x10101010, 0xabcd0000,
		[]*AVP{
			NewAVP(NewAVPAttribute("Origin-Host", 264, 0, DiamIdent), false, false, nil, "host.example.com"),
			NewAVP(NewAVPAttribute("Origin-Realm", 296, 0, DiamIdent), true, false, nil, "example.com"),
			NewAVP(NewAVPAttribute("Host-IP-Address", 257, 0, Address), false, false, nil, net.ParseIP("10.20.30.1")),
			NewAVP(NewAVPAttribute("Vendor-Id", 266, 0, Unsigned32), false, false, nil, uint32(0)),
			NewAVP(NewAVPAttribute("Product-Name", 269, 0, UTF8String), true, false, nil, "GoDiameter"),
		},
		[]*AVP{
			NewAVP(NewAVPAttribute("Supported-Vendor-Id", 265, 0, Unsigned32), false, false, nil, uint32(18)),
			NewAVP(NewAVPAttribute("Auth-Application-Id", 258, 0, Unsigned32), true, false, nil, uint32(65536)),
		},
		[]byte{
			// header -- length = 20
			0x01, 0x00, 0x00, 0x88,
			0xc0, 0x00, 0x01, 0x01,
			0x00, 0x00, 0x00, 0x00,
			0x10, 0x10, 0x10, 0x10,
			0xab, 0xcd, 0x00, 0x00,
			// Origin-Host -- length = 24
			0x00, 0x00, 0x01, 0x08, 0x40, 0x00, 0x00, 0x18, 0x68, 0x6f, 0x73, 0x74, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d,
			// Origin-Realm -- length = 20
			0x00, 0x00, 0x01, 0x28, 0x40, 0x00, 0x00, 0x13, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x00,
			//Host-IP-Address -- length = 16
			0x00, 0x00, 0x01, 0x01, 0x40, 0x00, 0x00, 0x0e, 0x00, 0x01, 0x0a, 0x14, 0x1e, 0x01, 0x00, 0x00,
			// Vendor-Id -- length = 12
			0x00, 0x00, 0x01, 0x0a, 0x40, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x00,
			// Product-Name -- length = 20
			0x00, 0x00, 0x01, 0x0d, 0x40, 0x00, 0x00, 0x12, 0x47, 0x6f, 0x44, 0x69, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x00, 0x00,
			// Supported-Vendor-Id -- length = 12
			0x00, 0x00, 0x01, 0x09, 0x00, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x12,
			// Auth-Application-Id -- length = 12
			0x00, 0x00, 0x01, 0x02, 0x40, 0x00, 0x00, 0x0c, 0x00, 0x01, 0x00, 0x00,
		}},
}

func TestEncode(t *testing.T) {
	for testnum, set := range encodetests {
		m := NewMessage(set.flags, set.code, set.appID, set.hopByHopID, set.endToEndID, set.mandatoryAvps, set.optionalAvps)

		if m == nil {
			t.Error("Message is nil")
		}

		eb := m.Encode()

		if len(eb) != len(set.encoded) {
			t.Error("For test [", testnum+1, "] Encode length [", len(eb), "] does not match expected length [", len(set.encoded), "]")
		}

		for i, v := range set.encoded {
			if v != eb[i] {
				t.Error("For test [", testnum+1, "] byte [", i+1, "] with value [", eb[i], "] does not match expected value [", v, "]")
			}
		}
	}
}

type DecodeTestSet struct {
	encoded    []byte
	flags      uint8
	code       Uint24
	appID      uint32
	hopByHopID uint32
	endToEndID uint32
	avpCount   int
}

// decodetests is a list of DecodeTestSet structs, one per test.  'encoded' is the encoded byte stream;
// the remaining elements correspond to the Message fields
var decodetests = []DecodeTestSet{
	{encoded: []byte{
		// -- TEST 1: Basic CER
		0x01, 0x00, 0x00, 0x64, // diameter header word 1
		0x80, 0x00, 0x01, 0x01, // diameter header word 2
		0x00, 0x00, 0x00, 0x00, // diameter header word 3
		0x52, 0xf7, 0x04, 0x2a, // diameter header word 4
		0xc7, 0xf8, 0xd6, 0x02, // diameter header word 5
		// --- AVPS --- //
		0x00, 0x00, 0x01, 0x28, 0x40, 0x00, 0x00, 0x14, 0x64, 0x69, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x2e, 0x6f, 0x72, 0x67,
		0x00, 0x00, 0x01, 0x08, 0x40, 0x00, 0x00, 0x0e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x01, 0x40, 0x00, 0x00, 0x0e, 0x00, 0x02, 0x7f, 0x00, 0x00, 0x01, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x0a, 0x40, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x0d, 0x00, 0x00, 0x00, 0x0e, 0x6a, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x00, 0x00},
		flags: 0x80, code: 257, appID: 0, hopByHopID: 0x52f7042a, endToEndID: 0xc7f8d602, avpCount: 5},
	{encoded: []byte{
		// -- TEST 2: Same as encoding test 4
		0x01, 0x00, 0x00, 0x88,
		0xc0, 0x00, 0x01, 0x01,
		0x00, 0x00, 0x00, 0x00,
		0x10, 0x10, 0x10, 0x10,
		0xab, 0xcd, 0x00, 0x00,
		// -- AVPS -- //
		0x00, 0x00, 0x01, 0x08, 0x40, 0x00, 0x00, 0x18, 0x68, 0x6f, 0x73, 0x74, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d,
		0x00, 0x00, 0x01, 0x28, 0x40, 0x00, 0x00, 0x13, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x00,
		0x00, 0x00, 0x01, 0x01, 0x40, 0x00, 0x00, 0x0e, 0x00, 0x01, 0x0a, 0x14, 0x1e, 0x01, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x0a, 0x40, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x0d, 0x40, 0x00, 0x00, 0x12, 0x47, 0x6f, 0x44, 0x69, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x09, 0x00, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x12,
		0x00, 0x00, 0x01, 0x02, 0x00, 0x00, 0x00, 0x0c, 0x00, 0x01, 0x00, 0x00},
		flags: 0xc0, code: 257, appID: 0, hopByHopID: 0x10101010, endToEndID: 0xabcd0000, avpCount: 7},
}

func TestDecode(t *testing.T) {
	for testnum, set := range decodetests {
		m, err := DecodeMessage(set.encoded)

		if err != nil {
			t.Error("Failed to decode message stream: ", err.Error())
		}

		if m.Flags != set.flags {
			t.Error("For encoded set ", testnum, " flags do not match")
		}
		if m.Code != set.code {
			t.Error("For encoded set ", testnum, " codes do not match")
		}
		if m.AppID != set.appID {
			t.Error("For encoded set ", testnum, " application IDs do not match")
		}
		if m.HopByHopID != set.hopByHopID {
			t.Error("For encoded set ", testnum, " hop-by-hop-ids do not match")
		}
		if m.EndToEndID != set.endToEndID {
			t.Error("For encoded set ", testnum, " end-to-end-ids do not match")
		}
		if len(m.Avps) != set.avpCount {
			t.Error("For encoded set ", testnum, " AVP counts do not match")
		}

	}
}

//func TestContext(t *testing.T) {
//	ln, err := net.Listen("tcp", ":8080")
//	if err != nil {
//		t.Error(err)
//	}
//	instance, err := NewInstance(Identity{"peer1.diameter.org", "diameter.org"}, ln, make(chan *InstanceMessage))
//	if err != nil {
//		t.Error(err)
//	}
//	var cer *Message = NewMessage(MsgFlagRequest, 257, 0, 0x1, 0x2,
//		[]*AVP{
//			NewAVP(AvpOriginHost, true, false, nil, "peer1.diameter.org"),
//			NewAVP(AvpOriginRealm, true, false, nil, "diameter.org"),
//			NewAVP(AvpHostIPAddress, true, false, nil, "\x00\x02\x0f\x00\x00\x01"),
//			NewAVP(AvpVendorID, true, false, nil, uint32(0)),
//			NewAVP(AvpProductName, true, false, nil, "Test"),
//		},
//		[]*AVP{},
//	)
//	context, err := NewMessageContext(instance, cer)
//
//	if context == nil {
//		t.Error("context is nil")
//	}
//}
