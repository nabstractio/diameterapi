package diameter

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
	"unicode/utf8"
)

const (
	avpProtectedFlag                 = 0x20
	avpMandatoryFlag                 = 0x40
	avpFlagVendorSpecific            = 0x80
	nonVendorSpecificAvpHeaderLength = 8
	vendorSpecificAvpHeaderLength    = 12
)

// AVPDataType is an enumeration of Diameter AVP types.  For each AVPDataType, the
// "typed" type is articulated (e.g., the "typed" value for an Unsigned32 is uint32).
// The allowed types for input to NewTypedAVP() are provided. Generally, a type is allowed
// only if it doesn't force a conversion of sign or magnitude.  Sometimes, an 'int
// is allowed for otherwise unsigned types.  This is for convenience, since uncast literal
// integers are 'int' values. However, in these cases, a negative
// number will result in an error.  No effort, on the other hand, is made to ensure that
// an 'int' does not overflow, particularly for the Unsigned32, Integer32 and Float32 types.
type AVPDataType int

const (
	// Unsigned32 indicates AVP type for unsigned 32-bit integer.  The typed value is uint32.
	// Allowed source types: uint32, int.
	Unsigned32 AVPDataType = 1 + iota
	// Unsigned64 indicates AVP type for unsigned 64-bit integer.  The typed value is uint64.
	// Allowed source types: uint64, uint32, uint, int.
	Unsigned64
	// Integer32 indicates AVP type for signed 32-bit integer.  The typed value is int32.
	// Allowed source types: int32, int.
	Integer32
	// Integer64 indicates AVP type for signed 64-bit integer.  The typed value is int64.
	// Allowed source types: int64, int, int32.
	Integer64
	// Float32 indicates AVP type for signed 32-bit floating point.  The typed value is float32.
	// Allows source types: float32, int.
	Float32
	// Float64 indicates AVP type for signed 64-bit floating point.  The typed value is float64.
	// Allowed source types: float32, float64, int.
	Float64
	// Enumerated indicates AVP type for Enumerated.  The typed value is int32.
	// Allowed source types: int32, int.
	Enumerated
	// UTF8String indicates AVP type for UTF8String (a UTF8 encoded octet stream).  The typed
	// value string.
	// Allowed source types: string, []byte and []rune.  All three types are subject to
	// validation for a properly encoded utf8 byte stream.
	UTF8String
	// OctetString indicates AVP type for OctetString (an arbitrary octet stream).  The typed
	// value is []byte.
	// Allowed source types: []byte, string.
	OctetString
	// Time indicates AVP type for Time (number of seconds since Jan 1, 1900 as unsigned 32).  The typed value is
	// *time.Time.  If a time.Time is supplied that exceeds the maximum or is less than the minimum that
	// the Diameter Time type can represent, an error is returned.
	// Allowed source types: time.Time, *time.Time, [4]byte (network byte order), uint32, int.
	Time
	// Address indicates AVP type for Address.  The typed value is *diameter.AddressType.
	// Allowed source types: AddressType, *AddressType, net.IP, *net.IP, net.IPAddr, *net.IPAddr.
	Address
	// DiamIdent indicates AVP type for diameter identity (an octet stream).  The typed value is
	// String.
	DiamIdent
	// DiamURI indicates AVP type for a diameter URI (an octet stream).  The typed value is String.
	DiamURI
	// Grouped indicates AVP type for grouped (a set of AVPs).  The typed value is []*AVP.
	Grouped
	// IPFilterRule indicates AVP type for IP Filter Rule.  The typed value is []byte.
	IPFilterRule
	// TypeOrAvpUnknown is used when a query is made for an unknown AVP or the dictionary
	// contains an unknown type.  The typed value is []byte.
	TypeOrAvpUnknown
)

type AddressFamilyNumber uint16

const (
	AddressFamilyNumberInvalid         AddressFamilyNumber = 0
	IP4                                AddressFamilyNumber = 1
	IP6                                AddressFamilyNumber = 2
	NSAP                               AddressFamilyNumber = 3
	HDLC                               AddressFamilyNumber = 4
	BBN1822                            AddressFamilyNumber = 5
	Ethernet                           AddressFamilyNumber = 6
	E163                               AddressFamilyNumber = 7
	E164                               AddressFamilyNumber = 8
	F69                                AddressFamilyNumber = 9
	X121                               AddressFamilyNumber = 10
	IPX                                AddressFamilyNumber = 11
	Appletalk                          AddressFamilyNumber = 12
	DecnetIV                           AddressFamilyNumber = 13
	BanyanVines                        AddressFamilyNumber = 14
	E164withNSAP                       AddressFamilyNumber = 15
	DNS                                AddressFamilyNumber = 16
	DistinguishedName                  AddressFamilyNumber = 17
	ASNumber                           AddressFamilyNumber = 18
	XTPoverIP4                         AddressFamilyNumber = 19
	XTPoverIP6                         AddressFamilyNumber = 20
	XTPNativeMode                      AddressFamilyNumber = 21
	FibreChannelPortName               AddressFamilyNumber = 22
	FibreChannelNodeName               AddressFamilyNumber = 23
	GWID                               AddressFamilyNumber = 24
	AFIforL2VPN                        AddressFamilyNumber = 25
	MPLSTPSectionEndpointIdentifier    AddressFamilyNumber = 26
	MPLSTPLSPEndpointIdentifier        AddressFamilyNumber = 27
	MPLSTPPseudowireEndpointIdentifier AddressFamilyNumber = 28
	MTIP4                              AddressFamilyNumber = 29
	MTIP6                              AddressFamilyNumber = 30
	BGPSFC                             AddressFamilyNumber = 31
	EIGRPCommonServiceFamily           AddressFamilyNumber = 16384
	EIGRPIP4ServiceFamily              AddressFamilyNumber = 16385
	EIGRPIPv6ServiceFamily             AddressFamilyNumber = 16386
	LISPCanonicalAddressFormat         AddressFamilyNumber = 16387
	BGPLS                              AddressFamilyNumber = 16388
	MAC48Bit                           AddressFamilyNumber = 16389
	MAC64Bit                           AddressFamilyNumber = 16390
	OUI                                AddressFamilyNumber = 16391
	MAC24                              AddressFamilyNumber = 16392
	MAC40                              AddressFamilyNumber = 16393
	IPv6_64                            AddressFamilyNumber = 16394
	RBridgePortID                      AddressFamilyNumber = 16395
	TRILLNickname                      AddressFamilyNumber = 16396
	UniversallyUniqueIdentifier        AddressFamilyNumber = 16397
	RoutingPolicyAFI                   AddressFamilyNumber = 16398
	MPLSNamespaces                     AddressFamilyNumber = 16399
)

type AddressType []byte

// NewAddressType is the same as NewAddressTypeErrorable but panics if an error occurs.
func NewAddressType(addressFamilyNumber AddressFamilyNumber, value []byte) AddressType {
	a, err := NewAddressTypeErrorable(addressFamilyNumber, value)
	if err != nil {
		panic(err)
	}
	return a
}

// NewAddressTypeErrorable creates a new AddressType object from the address family number and
// the appropriate byte sequence (usually in Network Byte Order).  If the addressFamilyNumber
// is IP4 or IP6 but the value is not correspondingly 4 or 16 bytes, return an error.
func NewAddressTypeErrorable(addressFamilyNumber AddressFamilyNumber, value []byte) (AddressType, error) {
	switch addressFamilyNumber {
	case IP4:
		if len(value) != 4 {
			return nil, fmt.Errorf("an IP4 address must have exactly 4 bytes")
		}

	case IP6:
		if len(value) != 16 {
			return nil, fmt.Errorf("an IP6 address must have exactly 16 bytes")
		}
	}

	a := make([]byte, 2+len(value))
	binary.BigEndian.PutUint16(a, uint16(addressFamilyNumber))
	copy(a[2:], value)

	return a, nil
}

// NewAddressTypeFromIP creates an AddressType object from a net.IP.  Panics if 'ip'
// is not the correct number of bytes for a net.IP object.
func NewAddressTypeFromIP(ip net.IP) AddressType {
	asIpV4 := ip.To4()
	if asIpV4 != nil {
		a := make([]byte, 6)
		binary.BigEndian.PutUint16(a, uint16(IP4))
		copy(a[2:], []byte(asIpV4))
		return a
	}

	if len(ip) == 16 {
		a := make([]byte, 18)
		binary.BigEndian.PutUint16(a, uint16(IP6))
		copy(a[2:], []byte(ip))
		return a
	}

	panic("provided value is not an IP address")
}

// Address returns the address part of the AddressType, or nil if there
// are not enough bytes for that.
func (a *AddressType) Address() []byte {
	if len([]byte(*a)) < 2 {
		return nil
	}

	b := []byte(*a)
	return b[2:]
}

// Type returns the AddressFamilyNumber part of the AddressType.  If the
// AddressType does not contain at least two bytes, it returns AddressFamilyNumberInvalid.
func (a *AddressType) Type() AddressFamilyNumber {
	b := []byte(*a)
	if len(b) < 2 {
		return AddressFamilyNumberInvalid
	}
	return AddressFamilyNumber(binary.BigEndian.Uint16(b[:2]))
}

// IsAnIP returns true if the AddressFamilyNumber if IP4 or IP6.
func (a *AddressType) IsAnIP() bool {
	t := a.Type()
	return t == IP4 || t == IP6
}

// IsNotAnIP is the opposite of IsAnIP(), provided to help with readability
// when checking the negative case.
func (a *AddressType) IsNotAnIP() bool {
	return !a.IsAnIP()
}

// ToIP returns a net.IP if the AddressType is IP4 or IP6.  Returns nil otherwise.
func (a *AddressType) ToIP() *net.IP {
	switch a.Type() {
	case IP4:
		b := []byte(*a)
		if len(b) != 6 {
			return nil
		}
		n := net.IP(b[2:])
		return &n

	case IP6:
		b := []byte(*a)
		if len(b) != 18 {
			return nil
		}
		n := net.IP(b[2:])
		return &n
	}

	return nil
}

var diameterBaseTime time.Time = time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)

// AVPExtendedAttributes includes extended AVP attributes that can be
// provided by, for example, a dictionary.  It includes a human-friendly name
// and a typed value (e.g., a uint32 for AVPs of Unsigned32 type).
type AVPExtendedAttributes struct {
	Name       string
	DataType   AVPDataType
	TypedValue interface{}
}

// AVP represents a Diameter Message AVP
type AVP struct {
	// The AVP Code value.
	Code uint32
	// Whether the Vendor-specific (V) flag is set.
	VendorSpecific bool
	// Whether the Mandatory (M) flag is set.
	Mandatory bool
	// Whether the Protected (P) flag is set.
	Protected bool
	// The VendorID field value.  This value is irrelevant for encoding if
	// VendorSpecific is false.
	VendorID uint32
	// The unpadded data.
	Data []byte
	// The value of the Length field.  This is the length of the header
	// in bytes plus len(Data).  It does not include the PaddedLength.
	Length int
	// The total length including pad bytes; thus it will be Length + numberOfPadBytes.
	PaddedLength int
	// The AVPExtendedAttributes, if they are includes.  If they are not included,
	// this will be nil.
	ExtendedAttributes *AVPExtendedAttributes
}

// NewAVP is an AVP constructor.  This will set the Vendor-Specific (V) flag if the
// VendorID is not 0.  It will set the Mandatory (M) flag according to the value
// of 'mandatory'.  It will also set the Length and PaddedLength values appropriately.
// ExtendedAttributes will be nil.
func NewAVP(code uint32, VendorID uint32, mandatory bool, data []byte) *AVP {
	avp := new(AVP)
	avp.Code = code
	avp.VendorID = VendorID
	if VendorID != 0 {
		avp.Length = vendorSpecificAvpHeaderLength
		avp.VendorSpecific = true
	} else {
		avp.Length = nonVendorSpecificAvpHeaderLength
		avp.VendorSpecific = false
	}
	avp.Mandatory = mandatory
	avp.Protected = false
	avp.Data = data

	avp.Length += len(data)
	avp.updatePaddedLength()

	avp.ExtendedAttributes = nil

	return avp
}

// NewTypedAVPErrorable is an AVP constructor provided typed data rather than the raw data.  Returns an
// error if the value is not convertible from the avpType.  The ExtendedAttributes will be set, but the
// Name will be the empty string.  For each avpType, more than one value type that may be accepted.
// For example, for all number types (e.g., Unsigned32 and Float32), a value of type 'int' is acceptable.
// However, for unsigned types -- like Unsigned32 -- if a typed value -- like 'int' -- is provided, then
// the value is blindly coerced into the proper type.  Thus, if an 'int' is supplied for an Unsigned32
// AVP, the value is cast as a uint32.  If the 'int' value is negative, the value will be accepted and
// still cast without any alteration.  AVPDataType cannot be TypeOrAvpUnknown.
func NewTypedAVPErrorable(code uint32, vendorID uint32, mandatory bool, avpType AVPDataType, value interface{}) (*AVP, error) {
	var data []byte
	var coercedValue any

	switch avpType {
	case Unsigned32:
		data = make([]byte, 4)

		switch v := value.(type) {
		case uint32:
			coercedValue = v
			binary.BigEndian.PutUint32(data, v)
		case int:
			coercedValue = uint32(v)
			binary.BigEndian.PutUint32(data, uint32(v))
		default:
			return nil, fmt.Errorf("supplied type cannot be converted to Unsighed32")
		}

	case Unsigned64:
		data = make([]byte, 8)

		switch v := value.(type) {
		case uint64:
			coercedValue = v
			binary.BigEndian.PutUint64(data, v)
		case int:
			coercedValue = uint64(v)
			binary.BigEndian.PutUint64(data, uint64(v))
		case uint:
			coercedValue = uint64(v)
			binary.BigEndian.PutUint64(data, uint64(v))
		case uint32:
			coercedValue = uint64(v)
			binary.BigEndian.PutUint64(data, uint64(v))
		default:
			return nil, fmt.Errorf("supplied type cannot be converted to Unsighed6")
		}

	case Integer32:
		buf := new(bytes.Buffer)

		switch v := value.(type) {
		case int32:
			coercedValue = v
			binary.Write(buf, binary.BigEndian, v)
		case int:
			coercedValue = int32(v)
			binary.Write(buf, binary.BigEndian, coercedValue)
		default:
			return nil, fmt.Errorf("supplied type cannot be converted to Integer32")
		}

		data = buf.Bytes()

	case Integer64:
		buf := new(bytes.Buffer)

		switch v := value.(type) {
		case int64:
			coercedValue = v
			binary.Write(buf, binary.BigEndian, v)
		case int:
			coercedValue = int64(v)
			binary.Write(buf, binary.BigEndian, coercedValue)
		case int32:
			coercedValue = int64(v)
			binary.Write(buf, binary.BigEndian, coercedValue)
		default:
			return nil, fmt.Errorf("supplied type cannot be converted to Integer64")
		}

		data = buf.Bytes()

	case Float32:
		buf := new(bytes.Buffer)

		switch v := value.(type) {
		case float32:
			coercedValue = v
			binary.Write(buf, binary.BigEndian, v)
		case int:
			coercedValue = float32(v)
			binary.Write(buf, binary.BigEndian, coercedValue)
		default:
			return nil, fmt.Errorf("supplied type cannot be converted to Float32")
		}

		data = buf.Bytes()

	case Float64:
		buf := new(bytes.Buffer)

		switch v := value.(type) {
		case float32:
			coercedValue = v
			binary.Write(buf, binary.BigEndian, v)
		case float64:
			coercedValue = float64(v)
			binary.Write(buf, binary.BigEndian, coercedValue)
		case int:
			coercedValue = float64(v)
			binary.Write(buf, binary.BigEndian, coercedValue)
		default:
			return nil, fmt.Errorf("supplied type cannot be converted to Float64")
		}

		data = buf.Bytes()

	case UTF8String:
		switch v := value.(type) {
		case string:
			data = []byte(v)
			coercedValue = v
		case []byte:
			data = v
			coercedValue = string(v)
		case []rune:
			data = []byte(string(v))
			coercedValue = string(v)
		default:
			return nil, fmt.Errorf("supplied type cannot be converted to UTF8String")
		}

		if !utf8.Valid(data) {
			return nil, fmt.Errorf("supplied value is not encoded utf8")
		}

	case OctetString:
		switch v := value.(type) {
		case []byte:
			data = v
			coercedValue = v
		case string:
			data = []byte(v)
			coercedValue = []byte(v)
		default:
			return nil, fmt.Errorf("supplied type cannot be converted to OctetString")
		}

	case Enumerated:
		buf := new(bytes.Buffer)

		switch v := value.(type) {
		case int32:
			coercedValue = v
			binary.Write(buf, binary.BigEndian, v)
		case int:
			coercedValue = int32(v)
			binary.Write(buf, binary.BigEndian, coercedValue)
		default:
			return nil, fmt.Errorf("supplied type cannot be converted to Enumerated")
		}

		data = buf.Bytes()

	case Time:
		switch v := value.(type) {
		case time.Time:
			return NewTypedAVPErrorable(code, vendorID, mandatory, avpType, &v)

		case *time.Time:
			durationSinceDiameterBaseTime := v.Sub(diameterBaseTime) / time.Second

			if durationSinceDiameterBaseTime < 0 {
				return nil, fmt.Errorf("provided Time is earlier than the Diameter Epoch (Jan 01, 1900 UTC)")
			}

			if durationSinceDiameterBaseTime > 4294967295 {
				return nil, fmt.Errorf("provided Time is later than Diameter time can represent")
			}

			data = make([]byte, 4)
			binary.BigEndian.PutUint32(data, uint32(durationSinceDiameterBaseTime))

			coercedValue = v

		case []byte:
			if len(v) != 4 {
				return nil, fmt.Errorf("byte slice for Time must have a length of exactly 4")
			}

			byteSliceToUint32 := binary.BigEndian.Uint32(v)

			c := diameterBaseTime.Add(time.Second * time.Duration(byteSliceToUint32))
			coercedValue = &c
			data = v

		case uint32:
			c := diameterBaseTime.Add(time.Second * time.Duration(v))
			coercedValue = &c

			data = make([]byte, 4)
			binary.BigEndian.PutUint32(data, v)

		case int:
			if v < 0 {
				return nil, fmt.Errorf("value for Time cannot be negative")
			}
			c := diameterBaseTime.Add(time.Second * time.Duration(v))
			coercedValue = &c

			data = make([]byte, 4)
			binary.BigEndian.PutUint32(data, uint32(v))

		default:
			return nil, fmt.Errorf("supplied type cannot be converted to Time")
		}

	case Address:
		switch v := value.(type) {
		case AddressType:
			data = []byte(v)
			coercedValue = AddressType(data)

		case *AddressType:
			data = []byte(*v)
			coercedValue = AddressType(data)

		case *net.IP:
			a := NewAddressTypeFromIP(*v)
			data = []byte(a)
			coercedValue = AddressType(data)

		case net.IP:
			a := NewAddressTypeFromIP(v)
			data = []byte(a)
			coercedValue = AddressType(data)

		case net.IPAddr:
			a := NewAddressTypeFromIP(v.IP)
			data = []byte(a)
			coercedValue = AddressType(data)

		case *net.IPAddr:
			a := NewAddressTypeFromIP(v.IP)
			data = []byte(a)
			coercedValue = AddressType(data)

		default:
			return nil, fmt.Errorf("supplied type cannot be converted to Address")
		}

	case DiamIdent:
		v, isString := value.(string)

		if !isString {
			return nil, fmt.Errorf("supplied type cannot be converted to DiamIdent")
		}

		data = []byte(v)
		coercedValue = v

	case DiamURI:
		v, isByteSlice := value.(string)

		if !isByteSlice {
			return nil, fmt.Errorf("supplied type cannot be converted to DiamURI")
		}

		data = []byte(v)
		coercedValue = v

	case Grouped:
		v, isAvpSlice := value.([]*AVP)

		if !isAvpSlice {
			return nil, fmt.Errorf("supplied type cannot be converted to Grouped")
		}

		avpDataLen := 0
		for _, avp := range v {
			avpDataLen += avp.PaddedLength
		}

		data = make([]byte, 0, avpDataLen)

		for _, avp := range v {
			data = append(data, avp.Encode()...)
		}

		coercedValue = v

	case IPFilterRule:
		switch v := value.(type) {
		case string:
			coercedValue = v
			data = []byte(v)
		case []byte:
			coercedValue = string(v)
			data = v
		default:
			return nil, fmt.Errorf("supplied type cannot be converted to IPFilterRule")
		}

	default:
		return nil, fmt.Errorf("type not valid for an AVP")
	}

	isVendorSpecific := false
	avpLength := nonVendorSpecificAvpHeaderLength
	if vendorID != 0 {
		isVendorSpecific = true
		avpLength = vendorSpecificAvpHeaderLength
	}

	avpLength += len(data)

	paddedLength := avpLength
	carry := avpLength % 4
	if carry > 0 {
		paddedLength += (4 - carry)
	}

	return &AVP{
		Code:           code,
		VendorID:       vendorID,
		VendorSpecific: isVendorSpecific,
		Mandatory:      mandatory,
		Protected:      false,
		Data:           data,
		Length:         avpLength,
		PaddedLength:   paddedLength,
		ExtendedAttributes: &AVPExtendedAttributes{
			DataType:   avpType,
			TypedValue: coercedValue,
			Name:       "",
		},
	}, nil
}

// NewTypedAVP is the same as NewTypedAVPErrorable, except that it raises panic() on an error.
func NewTypedAVP(code uint32, vendorID uint32, mandatory bool, avpType AVPDataType, value interface{}) *AVP {
	avp, err := NewTypedAVPErrorable(code, vendorID, mandatory, avpType, value)

	if err != nil {
		panic(err)
	}

	return avp
}

// ConvertAVPDataToTypedData attempts to convert the provided AVP data into a typed value,
// according to the data type provided.  dataType cannot be TypeOrAvpUnknown.
func ConvertAVPDataToTypedData(avpData []byte, dataType AVPDataType) (interface{}, error) {
	switch dataType {
	case Unsigned32:
		if len(avpData) != 4 {
			return nil, fmt.Errorf("type Unsigned32 requires exactly four bytes")
		}

		return binary.BigEndian.Uint32(avpData), nil

	case Unsigned64:
		if len(avpData) != 8 {
			return nil, fmt.Errorf("type Unsigned64 requires exactly eight bytes")
		}

		return binary.BigEndian.Uint64(avpData), nil

	case Integer32:
		if len(avpData) != 4 {
			return nil, fmt.Errorf("type Integer32 requires exactly four bytes")
		}

		return int32(binary.BigEndian.Uint32(avpData)), nil

	case Integer64:
		if len(avpData) != 8 {
			return nil, fmt.Errorf("type Integer64 requires exactly eight bytes")
		}

		return int64(binary.BigEndian.Uint64(avpData)), nil

	case Float32:
		if len(avpData) != 4 {
			return nil, fmt.Errorf("type Float32 requires exactly four bytes")
		}

		return float32(binary.BigEndian.Uint32(avpData)), nil

	case Float64:
		if len(avpData) != 8 {
			return nil, fmt.Errorf("type Float64 requires exactly eight bytes")
		}

		return float64(binary.BigEndian.Uint64(avpData)), nil

	case UTF8String:
		return string(avpData), nil

	case OctetString:
		return avpData[:], nil

	case Enumerated:
		if len(avpData) != 4 {
			return nil, fmt.Errorf("type Enumerated requires exactly four bytes")
		}

		return int32(binary.BigEndian.Uint32(avpData)), nil

	case Time:
		if len(avpData) != 4 {
			return nil, fmt.Errorf("type time requires exactly four bytes")
		}

		return binary.BigEndian.Uint32(avpData), nil

	case Address:
		switch len(avpData) {
		case 6:
			if binary.BigEndian.Uint16(avpData[:2]) != 1 {
				return nil, fmt.Errorf("type Address must be for IPv4 or IPv6 address only")
			}
			return net.IPv4(avpData[2], avpData[3], avpData[4], avpData[5]), nil

		case 10:
			if binary.BigEndian.Uint16(avpData[:2]) != 2 {
				return nil, fmt.Errorf("type Address must be for IPv4 or IPv6 address only")
			}
			ipAddr := net.IP(avpData[2:])
			return &ipAddr, nil

		default:
			return nil, fmt.Errorf("type Address requires exactly 6 bytes or 10 bytes")
		}

	case DiamIdent:
		return string(avpData), nil

	case Grouped:
		groupedBytes := avpData
		avpsInGroup := make([]*AVP, 10)

		for len(groupedBytes) > 0 {
			nextAvp, err := DecodeAVP(groupedBytes)
			if err != nil {
				return nil, fmt.Errorf("unable to decode AVP inside group: %s", err.Error())
			}
			avpsInGroup = append(avpsInGroup, nextAvp)
			groupedBytes = groupedBytes[nextAvp.PaddedLength+1:]
		}

		return avpsInGroup, nil

	case IPFilterRule:
		return avpData[:], nil

	default:
		return nil, fmt.Errorf("type not valid for an AVP")
	}
}

// MustConvertAVPDataToTypedData does the same thing as ConvertAVPDataToTypedData but panics
// if there is an error.
func MustConvertAVPDataToTypedData(avpData []byte, dataType AVPDataType) interface{} {
	v, err := ConvertAVPDataToTypedData(avpData, dataType)
	if err != nil {
		panic(err)
	}
	return v
}

// MakeProtected sets avp.Protected to true and returns the AVP reference.  It is so rare for
// this flag to be set, this provides a convenient method to set the value inline after
// AVP creation
func (avp *AVP) MakeProtected() *AVP {
	avp.Protected = true
	return avp
}

// ConvertDataToTypedData overrides any internally stored typed data representation for
// the AVP and attempts to convert the raw data into the named type.
func (avp *AVP) ConvertDataToTypedData(dataType AVPDataType) (interface{}, error) {
	return ConvertAVPDataToTypedData(avp.Data, dataType)
}

func appendUint32(avp *bytes.Buffer, dataUint32 uint32) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, dataUint32)
	err := binary.Write(avp, binary.LittleEndian, data)
	if err != nil {
		panic(fmt.Sprintf("binary.Write failed: %s", err))
	}
}

func appendByteArray(avp *bytes.Buffer, dataBytes []byte) {
	err := binary.Write(avp, binary.LittleEndian, dataBytes)
	if err != nil {
		panic(fmt.Sprintf("binary.Write failed: %s", err))
	}
}

// Encode produces an octet stream in network byte order from this AVP.
func (avp *AVP) Encode() []byte {
	buf := new(bytes.Buffer)
	padded := make([]byte, (avp.PaddedLength - avp.Length))
	appendUint32(buf, avp.Code)
	flags := 0

	if avp.VendorSpecific {
		flags = 0x80
	}
	if avp.Mandatory {
		flags |= 0x40
	}
	if avp.Protected {
		flags |= 0x20
	}

	appendUint32(buf, ((uint32(flags) << 24) | (uint32(avp.Length) & 0x00ffffff)))

	if avp.VendorSpecific {
		appendUint32(buf, avp.VendorID)
	}

	appendByteArray(buf, avp.Data)
	appendByteArray(buf, padded)

	return buf.Bytes()
}

func (avp *AVP) updatePaddedLength() {
	plen := (avp.Length) & 0x00000003
	if plen > 0 {
		avp.PaddedLength = avp.Length + (4 - plen)
	} else {
		avp.PaddedLength = avp.Length
	}
}

// Clone makes a copy of this AVP and returns it.  The source AVP
// object must not be updated during the cloning process.
func (avp *AVP) Clone() *AVP {
	clone := *avp
	clone.Data = make([]byte, len(avp.Data))
	copy(clone.Data, avp.Data)
	return &clone
}

// Equal compares the current AVP to another AVP to determine if they are byte-wise
// identical (that is, if they would map identically as a byte stream using Encode).
func (avp *AVP) Equal(a *AVP) bool {
	if a == nil {
		return false
	}

	if avp.Code != a.Code || avp.VendorSpecific != a.VendorSpecific || avp.Mandatory != a.Mandatory || avp.VendorID != a.VendorID || avp.Length != a.Length || avp.PaddedLength != a.PaddedLength {
		return false
	}

	if len(avp.Data) != len(a.Data) {
		return false
	}

	for i, leftAvpByteValue := range avp.Data {
		if leftAvpByteValue != a.Data[i] {
			return false
		}
	}

	return true
}

// DecodeAVP accepts a byte stream in network byte order and produces an AVP
// object from it.
func DecodeAVP(input []byte) (*AVP, error) {
	avp := new(AVP)
	buf := bytes.NewReader(input)
	var code uint32
	err := binary.Read(buf, binary.BigEndian, &code)
	if err != nil {
		return nil, fmt.Errorf("stream read failure: %s", err)
	}

	avp.Code = code

	var flagsAndLength uint32
	err = binary.Read(buf, binary.BigEndian, &flagsAndLength)
	if err != nil {
		return nil, fmt.Errorf("stream read failure: %s", err)
	}
	flags := byte((flagsAndLength & 0xFF000000) >> 24)
	avp.Length = int(flagsAndLength & 0x00FFFFFF)

	avp.Mandatory = bool((avpMandatoryFlag & flags) == avpMandatoryFlag)
	avp.Protected = bool((avpProtectedFlag & flags) == avpProtectedFlag)
	avp.VendorSpecific = bool((avpFlagVendorSpecific & flags) == avpFlagVendorSpecific)

	if avp.Length > len(input) {
		return nil, fmt.Errorf("length field in AVP header greater than encoded length")
	}

	headerLength := nonVendorSpecificAvpHeaderLength

	if avp.VendorSpecific {
		err = binary.Read(buf, binary.BigEndian, &avp.VendorID)
		if err != nil {
			return nil, fmt.Errorf("stream read failure: %s", err)
		}
		headerLength = vendorSpecificAvpHeaderLength
	}

	avp.Data = make([]byte, avp.Length-headerLength)

	err = binary.Read(buf, binary.BigEndian, avp.Data)

	if err != nil {
		return nil, err
	}

	avp.updatePaddedLength()

	return avp, nil
}

// AvpVendorIdAndCode is a union representing the vendor-id for an AVP and the code for an AVP.
type AvpVendorIdAndCode struct {
	VendorId uint32
	Code     uint32
}

// GenerateMapOfAvpsByVendorAndCode walks through the set of supplied AVPs, and creates a map
// indexed by vendor-id and code, pointing to the subset of AVPs with that vendor-id/code.
func GenerateMapOfAvpsByVendorAndCode(avps []*AVP) map[AvpVendorIdAndCode][]*AVP {
	m := make(map[AvpVendorIdAndCode][]*AVP)
	for _, avp := range avps {
		v := AvpVendorIdAndCode{avp.VendorID, avp.Code}
		existingAvpsWithThisVendorAndCode := m[v]
		if existingAvpsWithThisVendorAndCode == nil {
			m[v] = make([]*AVP, 1)
			m[v][0] = avp
		} else {
			m[v] = append(existingAvpsWithThisVendorAndCode, avp)
		}
	}

	return m
}
