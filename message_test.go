package diameter_test

import (
	"io"
	"net"
	"testing"

	diameter "github.com/blorticus-go/diameter"
	"github.com/go-test/deep"
)

type ControlledReader struct {
	readChunks              [][]byte
	nextChunkProvidedOnRead int
}

func NewControlledReader(readChunks [][]byte) *ControlledReader {
	return &ControlledReader{
		readChunks:              readChunks,
		nextChunkProvidedOnRead: 0,
	}
}

func (r *ControlledReader) Read(into []byte) (n int, err error) {
	if r.nextChunkProvidedOnRead >= len(r.readChunks) {
		return 0, io.EOF
	}
	nextChunk := r.readChunks[r.nextChunkProvidedOnRead]
	r.nextChunkProvidedOnRead++
	return copy(into, nextChunk), nil
}

type EncodedAndDecodedAvps struct {
	EncodedBytes []byte
	Avp          *diameter.AVP
}

var encDecAvpByName = map[string]*EncodedAndDecodedAvps{
	"originHost-host.example.com": {
		EncodedBytes: []byte{0x00, 0x00, 0x01, 0x08, 0x40, 0x00, 0x00, 0x18, 0x68, 0x6f, 0x73, 0x74, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d},
		Avp: &diameter.AVP{
			Code: 264, VendorSpecific: false, Mandatory: true, Protected: false, VendorID: 0, Length: 24, PaddedLength: 24, ExtendedAttributes: nil,
			Data: []byte{0x68, 0x6f, 0x73, 0x74, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d},
		},
	},
	"originRealm-example.com": {
		EncodedBytes: []byte{0x00, 0x00, 0x01, 0x28, 0x40, 0x00, 0x00, 0x13, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x00},
		Avp: &diameter.AVP{
			Code: 296, VendorSpecific: false, Mandatory: true, Protected: false, VendorID: 0, Length: 19, PaddedLength: 20, ExtendedAttributes: nil,
			Data: []byte{0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d},
		},
	},
	"hostIpAddress-10.20.30.1": {
		EncodedBytes: []byte{0x00, 0x00, 0x01, 0x01, 0x40, 0x00, 0x00, 0x0e, 0x00, 0x01, 0x0a, 0x14, 0x1e, 0x01, 0x00, 0x00},
		Avp: &diameter.AVP{
			Code: 257, VendorSpecific: false, Mandatory: true, Protected: false, VendorID: 0, Length: 14, PaddedLength: 16, ExtendedAttributes: nil,
			Data: []byte{0x00, 0x01, 0x0a, 0x14, 0x1e, 0x01},
		},
	},
	"vendorId-0": {
		EncodedBytes: []byte{0x00, 0x00, 0x01, 0x0a, 0x40, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x00},
		Avp: &diameter.AVP{
			Code: 266, VendorSpecific: false, Mandatory: true, Protected: false, VendorID: 0, Length: 12, PaddedLength: 12, ExtendedAttributes: nil,
			Data: []byte{0x00, 0x00, 0x00, 0x00},
		},
	},
	"productName-GoDiameter": {
		EncodedBytes: []byte{0x00, 0x00, 0x01, 0x0d, 0x40, 0x00, 0x00, 0x12, 0x47, 0x6f, 0x44, 0x69, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x00, 0x00},
		Avp: &diameter.AVP{
			Code: 269, VendorSpecific: false, Mandatory: true, Protected: false, VendorID: 0, Length: 18, PaddedLength: 20, ExtendedAttributes: nil,
			Data: []byte{0x47, 0x6f, 0x44, 0x69, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72},
		},
	},
}

func flattedBytes(in ...[]byte) []byte {
	totalLength := 0
	for _, ba := range in {
		totalLength += len(ba)
	}
	b := make([]byte, 0, totalLength)

	for _, ba := range in {
		b = append(b, ba...)
	}

	return b
}

type EncodedAndDecodedMessage struct {
	EncodedBytes []byte
	Message      *diameter.Message
}

var testMessagesByName = map[string]*EncodedAndDecodedMessage{
	"Basic-CER-01": {
		EncodedBytes: flattedBytes(
			[]byte{0x01, 0x00, 0x00, 0x70, 0xc0, 0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x10, 0x10, 0x10, 0x10, 0xab, 0xcd, 0x00, 0x00},
			encDecAvpByName["originHost-host.example.com"].EncodedBytes,
			encDecAvpByName["originRealm-example.com"].EncodedBytes,
			encDecAvpByName["hostIpAddress-10.20.30.1"].EncodedBytes,
			encDecAvpByName["vendorId-0"].EncodedBytes,
			encDecAvpByName["productName-GoDiameter"].EncodedBytes,
		),
		Message: &diameter.Message{
			Version: 1, Length: 112, Flags: 0xc0, Code: 257, AppID: 0, HopByHopID: 0x10101010, EndToEndID: 0xabcd0000,
			Avps: []*diameter.AVP{
				encDecAvpByName["originHost-host.example.com"].Avp,
				encDecAvpByName["originRealm-example.com"].Avp,
				encDecAvpByName["hostIpAddress-10.20.30.1"].Avp,
				encDecAvpByName["vendorId-0"].Avp,
				encDecAvpByName["productName-GoDiameter"].Avp,
			},
		},
	},
}

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
		m := diameter.Message{Flags: i}

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
		m := diameter.Message{Flags: set.value}

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
	code          diameter.Uint24
	appID         uint32
	hopByHopID    uint32
	endToEndID    uint32
	mandatoryAvps []*diameter.AVP
	optionalAvps  []*diameter.AVP
	encoded       []byte
}

// encodetests is a list of structs.  The structs are EncodeTestSet.  Everything but 'encoded' is passed to diameter.NewMessage.
// The message is encoded, and the encoded stream is compared to 'encoded'
var encodetests = []EncodeTestSet{
	// -- TEST 1: just header
	{0x00 | diameter.MsgFlagRequest | diameter.MsgFlagProxiable, 203, 0, 0x10101010, 0xabcd0000, []*diameter.AVP{}, []*diameter.AVP{},
		[]byte{0x01, 0x00, 0x00, 0x14, 0xc0, 0x00, 0x00, 0xcb, 0x00, 0x00, 0x00, 0x00, 0x10, 0x10, 0x10, 0x10, 0xab, 0xcd, 0x00, 0x00}},
	// -- TEST 2: CER with only mandatory AVPs
	{0x00 | diameter.MsgFlagRequest | diameter.MsgFlagProxiable, 257, 0, 0x10101010, 0xabcd0000,
		[]*diameter.AVP{
			diameter.NewTypedAVP(264, 0, true, diameter.DiamIdent, "host.example.com"),
			diameter.NewTypedAVP(296, 0, true, diameter.DiamIdent, "example.com"),
			diameter.NewTypedAVP(257, 0, true, diameter.Address, net.ParseIP("10.20.30.1")),
			diameter.NewTypedAVP(266, 0, true, diameter.Unsigned32, uint32(0)),
			diameter.NewTypedAVP(269, 0, true, diameter.UTF8String, "GoDiameter"),
		},
		[]*diameter.AVP{},
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
	{0x00 | diameter.MsgFlagRequest | diameter.MsgFlagProxiable, 257, 0, 0x10101010, 0xabcd0000,
		[]*diameter.AVP{
			diameter.NewTypedAVP(264, 0, false, diameter.DiamIdent, "host.example.com"),
			diameter.NewTypedAVP(296, 0, true, diameter.DiamIdent, "example.com"),
			diameter.NewTypedAVP(257, 0, false, diameter.Address, net.ParseIP("10.20.30.1")),
			diameter.NewTypedAVP(266, 0, false, diameter.Unsigned32, uint32(0)),
			diameter.NewTypedAVP(269, 0, true, diameter.UTF8String, "GoDiameter"),
		},
		[]*diameter.AVP{},
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
	{0x00 | diameter.MsgFlagRequest | diameter.MsgFlagProxiable, 257, 0, 0x10101010, 0xabcd0000,
		[]*diameter.AVP{
			diameter.NewTypedAVP(264, 0, false, diameter.DiamIdent, "host.example.com"),
			diameter.NewTypedAVP(296, 0, true, diameter.DiamIdent, "example.com"),
			diameter.NewTypedAVP(257, 0, false, diameter.Address, net.ParseIP("10.20.30.1")),
			diameter.NewTypedAVP(266, 0, false, diameter.Unsigned32, uint32(0)),
			diameter.NewTypedAVP(269, 0, true, diameter.UTF8String, "GoDiameter"),
		},
		[]*diameter.AVP{
			diameter.NewTypedAVP(265, 0, false, diameter.Unsigned32, uint32(18)),
			diameter.NewTypedAVP(258, 0, true, diameter.Unsigned32, uint32(65536)),
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
		m := diameter.NewMessage(set.flags, set.code, set.appID, set.hopByHopID, set.endToEndID, set.mandatoryAvps, set.optionalAvps)

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
	code       diameter.Uint24
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
		m, err := diameter.DecodeMessage(set.encoded)

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

func TestMessageStreamWithOneCompleteMessageOnlyInOneRead(t *testing.T) {
	stream := []byte{
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
		0x00, 0x00, 0x01, 0x0d, 0x00, 0x00, 0x00, 0x0e, 0x6a, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x00, 0x00,
	}

	messageReader := diameter.NewMessageByteReader()

	messages, err := messageReader.ReceiveBytes(stream)

	if err != nil {
		t.Errorf("Expected no error on ReceiveBytes, got = (%s)", err)
	}

	if len(messages) != 1 {
		t.Errorf("On ReceiveBytes, expected (1) message, got = (%d)", len(messages))
	}
}

func TestMessageStreamWithOneCompleteMessageInThreeReads(t *testing.T) {
	stream := []byte{
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
		0x00, 0x00, 0x01, 0x0d, 0x00, 0x00, 0x00, 0x0e, 0x6a, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x00, 0x00,
	}

	messageReader := diameter.NewMessageByteReader()

	messages, err := messageReader.ReceiveBytes(stream[0:20])

	if err != nil {
		t.Errorf("Expected no error on first ReceiveBytes, got = (%s)", err)
	}

	if len(messages) != 0 {
		t.Errorf("Expected no messages on first ReceiveBytes, got = (%d)", len(messages))
	}

	messages, err = messageReader.ReceiveBytes(stream[20:58])

	if err != nil {
		t.Errorf("Expected no error on second ReceiveBytes, got = (%s)", err)
	}

	if len(messages) != 0 {
		t.Errorf("Expected no messages on second ReceiveBytes, got = (%d)", len(messages))
	}

	messages, err = messageReader.ReceiveBytes(stream[58:])

	if err != nil {
		t.Errorf("Expected no error on third ReceiveBytes, got = (%s)", err)
	}

	if len(messages) != 1 {
		t.Errorf("Expected one messages on third ReceiveBytes, got = (%d)", len(messages))
	}

}

func TestMessageStreamWithThreeCompleteMessagesInOneRead(t *testing.T) {
	stream := []byte{
		// -- MESSAGE 1 -- //
		0x01, 0x00, 0x00, 0x64,
		0x80, 0x00, 0x01, 0x01,
		0x00, 0x00, 0x00, 0x00,
		0x52, 0xf7, 0x04, 0x2a,
		0xc7, 0xf8, 0xd6, 0x02,
		0x00, 0x00, 0x01, 0x28, 0x40, 0x00, 0x00, 0x14, 0x64, 0x69, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x2e, 0x6f, 0x72, 0x67,
		0x00, 0x00, 0x01, 0x08, 0x40, 0x00, 0x00, 0x0e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x01, 0x40, 0x00, 0x00, 0x0e, 0x00, 0x02, 0x7f, 0x00, 0x00, 0x01, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x0a, 0x40, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x0d, 0x00, 0x00, 0x00, 0x0e, 0x6a, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x00, 0x00,
		// -- MESSAGE 2 -- //
		0x01, 0x00, 0x00, 0x88,
		0xc0, 0x00, 0x01, 0x01,
		0x00, 0x00, 0x00, 0x00,
		0x10, 0x10, 0x10, 0x10,
		0xab, 0xcd, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x08, 0x40, 0x00, 0x00, 0x18, 0x68, 0x6f, 0x73, 0x74, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d,
		0x00, 0x00, 0x01, 0x28, 0x40, 0x00, 0x00, 0x13, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x00,
		0x00, 0x00, 0x01, 0x01, 0x40, 0x00, 0x00, 0x0e, 0x00, 0x01, 0x0a, 0x14, 0x1e, 0x01, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x0a, 0x40, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x0d, 0x40, 0x00, 0x00, 0x12, 0x47, 0x6f, 0x44, 0x69, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x09, 0x00, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x12,
		0x00, 0x00, 0x01, 0x02, 0x00, 0x00, 0x00, 0x0c, 0x00, 0x01, 0x00, 0x00,
		// -- MESSAGE 3 -- //
		0x01, 0x00, 0x00, 0x64,
		0x80, 0x00, 0x01, 0x01,
		0x00, 0x00, 0x00, 0x00,
		0x52, 0xf7, 0x04, 0x2a,
		0xc7, 0xf8, 0xd6, 0x02,
		0x00, 0x00, 0x01, 0x28, 0x40, 0x00, 0x00, 0x14, 0x64, 0x69, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x2e, 0x6f, 0x72, 0x67,
		0x00, 0x00, 0x01, 0x08, 0x40, 0x00, 0x00, 0x0e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x01, 0x40, 0x00, 0x00, 0x0e, 0x00, 0x02, 0x7f, 0x00, 0x00, 0x01, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x0a, 0x40, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x0d, 0x00, 0x00, 0x00, 0x0e, 0x6a, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x00, 0x00,
	}

	messageReader := diameter.NewMessageByteReader()

	messages, err := messageReader.ReceiveBytes(stream)

	if err != nil {
		t.Errorf("Expected no error on ReceiveBytes, got = (%s)", err)
	}

	if len(messages) != 3 {
		t.Errorf("Expected three messages on ReceiveBytes, got = (%d)", len(messages))
	}
}

func TestMessageStreamWithThreeCompleteMessagesInThreeReads(t *testing.T) {
	stream := []byte{
		// -- MESSAGE 1 -- //
		0x01, 0x00, 0x00, 0x64,
		0x80, 0x00, 0x01, 0x01,
		0x00, 0x00, 0x00, 0x00,
		0x52, 0xf7, 0x04, 0x2a,
		0xc7, 0xf8, 0xd6, 0x02,
		0x00, 0x00, 0x01, 0x28, 0x40, 0x00, 0x00, 0x14, 0x64, 0x69, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x2e, 0x6f, 0x72, 0x67,
		0x00, 0x00, 0x01, 0x08, 0x40, 0x00, 0x00, 0x0e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x01, 0x40, 0x00, 0x00, 0x0e, 0x00, 0x02, 0x7f, 0x00, 0x00, 0x01, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x0a, 0x40, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x0d, 0x00, 0x00, 0x00, 0x0e, 0x6a, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x00, 0x00,
		// -- MESSAGE 2 -- //
		0x01, 0x00, 0x00, 0x88,
		0xc0, 0x00, 0x01, 0x01,
		0x00, 0x00, 0x00, 0x00,
		0x10, 0x10, 0x10, 0x10,
		0xab, 0xcd, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x08, 0x40, 0x00, 0x00, 0x18, 0x68, 0x6f, 0x73, 0x74, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d,
		0x00, 0x00, 0x01, 0x28, 0x40, 0x00, 0x00, 0x13, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x00,
		0x00, 0x00, 0x01, 0x01, 0x40, 0x00, 0x00, 0x0e, 0x00, 0x01, 0x0a, 0x14, 0x1e, 0x01, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x0a, 0x40, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x0d, 0x40, 0x00, 0x00, 0x12, 0x47, 0x6f, 0x44, 0x69, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x09, 0x00, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x12,
		0x00, 0x00, 0x01, 0x02, 0x00, 0x00, 0x00, 0x0c, 0x00, 0x01, 0x00, 0x00,
		// -- MESSAGE 3 -- //
		0x01, 0x00, 0x00, 0x64,
		0x80, 0x00, 0x01, 0x01,
		0x00, 0x00, 0x00, 0x00,
		0x52, 0xf7, 0x04, 0x2a,
		0xc7, 0xf8, 0xd6, 0x02,
		0x00, 0x00, 0x01, 0x28, 0x40, 0x00, 0x00, 0x14, 0x64, 0x69, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x2e, 0x6f, 0x72, 0x67,
		0x00, 0x00, 0x01, 0x08, 0x40, 0x00, 0x00, 0x0e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x01, 0x40, 0x00, 0x00, 0x0e, 0x00, 0x02, 0x7f, 0x00, 0x00, 0x01, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x0a, 0x40, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x01, 0x0d, 0x00, 0x00, 0x00, 0x0e, 0x6a, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x00, 0x00,
	}

	messageReader := diameter.NewMessageByteReader()

	messages, err := messageReader.ReceiveBytes(stream[0:2])

	if err != nil {
		t.Errorf("Expected no error on first ReceiveBytes, got = (%s)", err)
	}

	if len(messages) != 0 {
		t.Errorf("Expected no messages on first ReceiveBytes, got = (%d)", len(messages))
	}

	messages, err = messageReader.ReceiveBytes(stream[2:236])

	if err != nil {
		t.Errorf("Expected no error on second ReceiveBytes, got = (%s)", err)
	}

	if len(messages) != 2 {
		t.Errorf("Expected two messages on second ReceiveBytes, got = (%d)", len(messages))
	}

	messages, err = messageReader.ReceiveBytes(stream[236:])

	if err != nil {
		t.Errorf("Expected no error on third ReceiveBytes, got = (%s)", err)
	}

	if len(messages) != 1 {
		t.Errorf("Expected one messages on third ReceiveBytes, got = (%d)", len(messages))
	}
}

func TestStreamReaderWithExactlyOneMessageInOnePart(t *testing.T) {
	basicCer01 := testMessagesByName["Basic-CER-01"]

	stream := basicCer01.EncodedBytes
	reader := NewControlledReader([][]byte{stream})

	streamReader := diameter.NewMessageStreamReader(reader)

	m, err := streamReader.ReadNextMessage()
	if err != nil {
		t.Fatalf("on first ReadNextMessage(), expected no error, got = (%s)", err)
	}

	if diff := deep.Equal(m, basicCer01.Message); diff != nil {
		t.Fatalf("after first ReadNextMessage(), messages differ: %s", diff)
	}

	m, err = streamReader.ReadNextMessage()
	if err == nil {
		t.Errorf("on second ReadNextMessage(), expected io.EOF, got no error")
	} else if err != io.EOF {
		t.Errorf("on second ReadNextMessage(), expected io.EOF, got error = (%s)", err)
	}
	if m != nil {
		t.Fatalf("on second ReadNextMessage(), expected message to be nil, but it is not")
	}
}

func TestFindFirstAVPByCode(t *testing.T) {
	message := diameter.NewMessage(diameter.MsgFlagRequest|diameter.MsgFlagProxiable, 257, 0, 0x10101010, 0xabcd0000, []*diameter.AVP{
		diameter.NewTypedAVP(264, 0, true, diameter.DiamIdent, "host.example.com"),
		diameter.NewTypedAVP(296, 0, true, diameter.DiamIdent, "example.com"),
		diameter.NewTypedAVP(257, 0, true, diameter.Address, net.ParseIP("10.20.30.1")),
		diameter.NewTypedAVP(266, 0, true, diameter.Unsigned32, uint32(0)),
		diameter.NewTypedAVP(269, 0, true, diameter.UTF8String, "GoDiameter"),
	}, []*diameter.AVP{})

	for _, code := range []diameter.Uint24{264, 296, 257, 266, 269} {
		matchingAvp := message.FirstAvpMatching(0, code)

		if matchingAvp == nil {
			t.Errorf("For First test, expected FindFirstAVPByCode(%d) to return non-nil, returned nil", code)
		}
	}

	for _, code := range []diameter.Uint24{263, 265, 0, 270, 2690} {
		matchingAvp := message.FirstAvpMatching(0, code)

		if matchingAvp != nil {
			t.Errorf("For First test, expected FindFirstAVPByCode(%d) to return nil, returned non-nil", code)
		}
	}

	message = diameter.NewMessage(diameter.MsgFlagRequest|diameter.MsgFlagProxiable, 257, 0, 0x10101010, 0xabcd0000, []*diameter.AVP{
		diameter.NewTypedAVP(264, 0, true, diameter.DiamIdent, "host.example.com"),
		diameter.NewTypedAVP(296, 0, true, diameter.DiamIdent, "example.com"),
		diameter.NewTypedAVP(257, 0, true, diameter.Address, net.ParseIP("10.20.30.1")),
		diameter.NewTypedAVP(266, 0, true, diameter.Unsigned32, uint32(0)),
		diameter.NewTypedAVP(257, 0, true, diameter.Address, net.ParseIP("10.20.30.2")),
		diameter.NewTypedAVP(269, 0, true, diameter.UTF8String, "GoDiameter"),
	}, []*diameter.AVP{
		diameter.NewTypedAVP(265, 0, false, diameter.Unsigned32, uint32(1)),
		diameter.NewTypedAVP(265, 0, false, diameter.Unsigned32, uint32(10)),
		diameter.NewTypedAVP(265, 0, false, diameter.Unsigned32, uint32(100)),
	})

	for _, code := range []diameter.Uint24{264, 296, 266, 269} {
		matchingAvp := message.FirstAvpMatching(0, code)

		if matchingAvp == nil {
			t.Errorf("For Second test, expected FindFirstAVPByCode(%d) to return non-nil, returned nil", code)
		}
	}

	for _, code := range []diameter.Uint24{263, 0, 270, 2690} {
		matchingAvp := message.FirstAvpMatching(0, code)

		if matchingAvp != nil {
			t.Errorf("For Second test, expected FindFirstAVPByCode(%d) to return nil, returned non-nil", code)
		}
	}

	matchingAvp := message.FirstAvpMatching(0, 257)

	if matchingAvp == nil {
		t.Errorf("For Second test, expected FindFirstAVPByCode(257) to return nil, returned non-nil")
	} else {
		if !matchingAvp.Equal(diameter.NewTypedAVP(257, 0, true, diameter.Address, net.ParseIP("10.20.30.1"))) {
			t.Errorf("For Second test, FindFirstAVPByCode(257) does not return the AVP instance expected")
		}
	}

	matchingAvp = message.FirstAvpMatching(0, 265)

	if matchingAvp == nil {
		t.Errorf("For Second test, expected FindFirstAVPByCode(265) to return nil, returned non-nil")
	} else {
		if !matchingAvp.Equal(diameter.NewTypedAVP(265, 0, false, diameter.Unsigned32, uint32(1))) {
			t.Errorf("For Second test, FindFirstAVPByCode(265) does not return the AVP instance expected")
		}
	}
}

func TestMessageEqualsWhenMessagesAreEqual(t *testing.T) {
	leftMessage := diameter.NewMessage(diameter.MsgFlagRequest|diameter.MsgFlagProxiable, 257, 0, 0x10101010, 0xabcd0000, []*diameter.AVP{
		diameter.NewTypedAVP(264, 0, true, diameter.DiamIdent, "host.example.com"),
		diameter.NewTypedAVP(296, 0, true, diameter.DiamIdent, "example.com"),
		diameter.NewTypedAVP(257, 0, true, diameter.Address, net.ParseIP("10.20.30.1")),
		diameter.NewTypedAVP(266, 0, true, diameter.Unsigned32, uint32(0)),
		diameter.NewTypedAVP(269, 0, true, diameter.UTF8String, "GoDiameter"),
	}, []*diameter.AVP{})

	rightMessage := diameter.NewMessage(diameter.MsgFlagRequest|diameter.MsgFlagProxiable, 257, 0, 0x10101010, 0xabcd0000, []*diameter.AVP{
		diameter.NewTypedAVP(264, 0, true, diameter.DiamIdent, "host.example.com"),
		diameter.NewTypedAVP(296, 0, true, diameter.DiamIdent, "example.com"),
		diameter.NewTypedAVP(257, 0, true, diameter.Address, net.ParseIP("10.20.30.1")),
		diameter.NewTypedAVP(266, 0, true, diameter.Unsigned32, uint32(0)),
		diameter.NewTypedAVP(269, 0, true, diameter.UTF8String, "GoDiameter"),
	}, []*diameter.AVP{})

	if !leftMessage.Equals(rightMessage) {
		t.Errorf("Expected leftMessage.Equal(rightMessage) to be true, is false")
	}

	if !rightMessage.Equals(leftMessage) {
		t.Errorf("Expected rightMessage.Equal(leftMessage) to be true, is false")
	}
}

func TestMessageEqualsWhenMessagesAreNotEqual(t *testing.T) {
	leftMessage := diameter.NewMessage(diameter.MsgFlagRequest|diameter.MsgFlagProxiable, 257, 0, 0x10101010, 0xabcd0000, []*diameter.AVP{
		diameter.NewTypedAVP(264, 0, true, diameter.DiamIdent, "host.example.com"),
		diameter.NewTypedAVP(296, 0, true, diameter.DiamIdent, "example.com"),
		diameter.NewTypedAVP(257, 0, true, diameter.Address, net.ParseIP("10.20.30.1")),
		diameter.NewTypedAVP(266, 0, true, diameter.Unsigned32, uint32(0)),
		diameter.NewTypedAVP(269, 0, true, diameter.UTF8String, "GoDiameter"),
	}, []*diameter.AVP{})

	messagesToCompare := []*diameter.Message{
		// Flags differ
		diameter.NewMessage(diameter.MsgFlagRequest, 257, 0, 0x10101010, 0xabcd0000, []*diameter.AVP{
			diameter.NewTypedAVP(264, 0, true, diameter.DiamIdent, "host.example.com"),
			diameter.NewTypedAVP(296, 0, true, diameter.DiamIdent, "example.com"),
			diameter.NewTypedAVP(257, 0, true, diameter.Address, net.ParseIP("10.20.30.1")),
			diameter.NewTypedAVP(266, 0, true, diameter.Unsigned32, uint32(0)),
			diameter.NewTypedAVP(269, 0, true, diameter.UTF8String, "GoDiameter"),
		}, []*diameter.AVP{}),
		// Message codes differ
		diameter.NewMessage(diameter.MsgFlagRequest|diameter.MsgFlagProxiable, 258, 0, 0x10101010, 0xabcd0000, []*diameter.AVP{
			diameter.NewTypedAVP(264, 0, true, diameter.DiamIdent, "host.example.com"),
			diameter.NewTypedAVP(296, 0, true, diameter.DiamIdent, "example.com"),
			diameter.NewTypedAVP(257, 0, true, diameter.Address, net.ParseIP("10.20.30.1")),
			diameter.NewTypedAVP(266, 0, true, diameter.Unsigned32, uint32(0)),
			diameter.NewTypedAVP(269, 0, true, diameter.UTF8String, "GoDiameter"),
		}, []*diameter.AVP{}),
		// Vendor IDs differ
		diameter.NewMessage(diameter.MsgFlagRequest|diameter.MsgFlagProxiable, 257, 1, 0x10101010, 0xabcd0000, []*diameter.AVP{
			diameter.NewTypedAVP(264, 0, true, diameter.DiamIdent, "host.example.com"),
			diameter.NewTypedAVP(296, 0, true, diameter.DiamIdent, "example.com"),
			diameter.NewTypedAVP(257, 0, true, diameter.Address, net.ParseIP("10.20.30.1")),
			diameter.NewTypedAVP(266, 0, true, diameter.Unsigned32, uint32(0)),
			diameter.NewTypedAVP(269, 0, true, diameter.UTF8String, "GoDiameter"),
		}, []*diameter.AVP{}),
		// Hop-By-Hop IDs differ
		diameter.NewMessage(diameter.MsgFlagRequest|diameter.MsgFlagProxiable, 257, 0, 0x10, 0xabcd0000, []*diameter.AVP{
			diameter.NewTypedAVP(264, 0, true, diameter.DiamIdent, "host.example.com"),
			diameter.NewTypedAVP(296, 0, true, diameter.DiamIdent, "example.com"),
			diameter.NewTypedAVP(257, 0, true, diameter.Address, net.ParseIP("10.20.30.1")),
			diameter.NewTypedAVP(266, 0, true, diameter.Unsigned32, uint32(0)),
			diameter.NewTypedAVP(269, 0, true, diameter.UTF8String, "GoDiameter"),
		}, []*diameter.AVP{}),
		// End-To-End IDs differ
		diameter.NewMessage(diameter.MsgFlagRequest|diameter.MsgFlagProxiable, 257, 0, 0x10101010, 0xa, []*diameter.AVP{
			diameter.NewTypedAVP(264, 0, true, diameter.DiamIdent, "host.example.com"),
			diameter.NewTypedAVP(296, 0, true, diameter.DiamIdent, "example.com"),
			diameter.NewTypedAVP(257, 0, true, diameter.Address, net.ParseIP("10.20.30.1")),
			diameter.NewTypedAVP(266, 0, true, diameter.Unsigned32, uint32(0)),
			diameter.NewTypedAVP(269, 0, true, diameter.UTF8String, "GoDiameter"),
		}, []*diameter.AVP{}),
		// AVP Set missing one AVP
		diameter.NewMessage(diameter.MsgFlagRequest|diameter.MsgFlagProxiable, 257, 0, 0x10101010, 0xabcd0000, []*diameter.AVP{
			diameter.NewTypedAVP(296, 0, true, diameter.DiamIdent, "example.com"),
			diameter.NewTypedAVP(257, 0, true, diameter.Address, net.ParseIP("10.20.30.1")),
			diameter.NewTypedAVP(266, 0, true, diameter.Unsigned32, uint32(0)),
			diameter.NewTypedAVP(269, 0, true, diameter.UTF8String, "GoDiameter"),
		}, []*diameter.AVP{}),
		// Second AVP in Set has different value
		diameter.NewMessage(diameter.MsgFlagRequest|diameter.MsgFlagProxiable, 257, 0, 0x10101010, 0xabcd0000, []*diameter.AVP{
			diameter.NewTypedAVP(264, 0, true, diameter.DiamIdent, "host.example.com"),
			diameter.NewTypedAVP(296, 0, true, diameter.DiamIdent, "example.org"),
			diameter.NewTypedAVP(257, 0, true, diameter.Address, net.ParseIP("10.20.30.1")),
			diameter.NewTypedAVP(266, 0, true, diameter.Unsigned32, uint32(0)),
			diameter.NewTypedAVP(269, 0, true, diameter.UTF8String, "GoDiameter"),
		}, []*diameter.AVP{}),
		// AVP Set order differs value
		diameter.NewMessage(diameter.MsgFlagRequest|diameter.MsgFlagProxiable, 257, 0, 0x10101010, 0xabcd0000, []*diameter.AVP{
			diameter.NewTypedAVP(264, 0, true, diameter.DiamIdent, "host.example.com"),
			diameter.NewTypedAVP(296, 0, true, diameter.DiamIdent, "example.com"),
			diameter.NewTypedAVP(266, 0, true, diameter.Unsigned32, uint32(0)),
			diameter.NewTypedAVP(257, 0, true, diameter.Address, net.ParseIP("10.20.30.1")),
			diameter.NewTypedAVP(269, 0, true, diameter.UTF8String, "GoDiameter"),
		}, []*diameter.AVP{}),
	}

	for comparisonMessageIndex, rightMessage := range messagesToCompare {
		if leftMessage.Equals(rightMessage) {
			t.Errorf("On comparison message at index (%d), expected leftMessage.Equal(rightMessage) to be false, but is true", comparisonMessageIndex)
		}

		if rightMessage.Equals(leftMessage) {
			t.Errorf("On comparison message at index (%d), expected rightMessage.Equal(leftMessage) to be false, but is true", comparisonMessageIndex)
		}
	}
}
