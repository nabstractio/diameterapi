package diameter

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

// Uint24 is a documentation reference type.  There is no enforcement of boundaries;
// it is simply a visual reminder of the type
type Uint24 uint32

// Possible flag values for a Diameter message
const (
	MsgFlagNone                = 0x00
	MsgFlagRequest             = 0x80
	MsgFlagProxiable           = 0x40
	MsgFlagError               = 0x20
	MsgFlagPotentialRetransmit = 0x10
	MsgHeaderSize              = Uint24(20)
)

// MessageExtendedAttributes includes extended Message attributes that can be
// provided by, for example, a dictionary.  It includes a human-friendly name
// and an abbreviated name.
type MessageExtendedAttributes struct {
	Name            string
	AbbreviatedName string
}

// Message represents a single Diameter message
type Message struct {
	Version            uint8
	Length             Uint24
	Flags              uint8
	Code               Uint24
	AppID              uint32
	HopByHopID         uint32
	EndToEndID         uint32
	Avps               []*AVP
	ExtendedAttributes *MessageExtendedAttributes

	mapOfAvpsByVendorAndCode map[AvpVendorIdAndCode][]*AVP
}

// FirstAvpMatching returns the first instance of the identified AVP associated
// with the current Message, or nil if the Message has no instances of the AVP
func (m *Message) FirstAvpMatching(vendorId uint32, code Uint24) *AVP {
	if m.mapOfAvpsByVendorAndCode == nil {
		m.mapOfAvpsByVendorAndCode = GenerateMapOfAvpsByVendorAndCode(m.Avps)
	}

	if avpSet := m.mapOfAvpsByVendorAndCode[AvpVendorIdAndCode{vendorId, uint32(code)}]; len(avpSet) == 0 {
		return nil
	} else {
		return avpSet[0]
	}
}

// MapOfAvpsByCode creates a map of the AVPs by AVP code, providing a list of
// all AVPs matching that code.  Each call to this method regenerates the map
// (i.e., the conversion is not cached).
func (m *Message) MapOfAvpsByCode() map[AvpVendorIdAndCode][]*AVP {
	// don't use internal mapOfAvpsByVendorAndCode so that caller doesn't modify that
	// internal structure
	return GenerateMapOfAvpsByVendorAndCode(m.Avps)
}

// TopLevelAvpsMatching returns the set of top-level AVPs in the message that match
// the provided vendorId and code.  "top-level" here means AVPs that are not part of
// a Grouped AVP contained within the message.
func (m *Message) TopLevelAvpsMatching(vendorId uint32, code Uint24) []*AVP {
	if m.mapOfAvpsByVendorAndCode == nil {
		m.mapOfAvpsByVendorAndCode = GenerateMapOfAvpsByVendorAndCode(m.Avps)
	}

	return m.mapOfAvpsByVendorAndCode[AvpVendorIdAndCode{vendorId, uint32(code)}]
}

// HasATopLevelAvpMatching returns true if there is at least one top-level AVP in the message
// that has matching vendorId and code.
func (m *Message) HasATopLevelAvpMatching(vendorId uint32, code Uint24) bool {
	return len(m.TopLevelAvpsMatching(vendorId, code)) > 0
}

// DoesNotHaveATopLevelAvpMatching is the opposite of HasATopLevelAvpMatching(), provided
// to enhance conditional statement natural readability.
func (m *Message) DoesNotHaveATopLevelAvpMatching(vendorId uint32, code Uint24) bool {
	return len(m.TopLevelAvpsMatching(vendorId, code)) == 0
}

// NumberOfTopLevelAvpsMatching returns the count of top-level AVPs in the message that
// have the matching vendorId and code.
func (m *Message) NumberOfTopLevelAvpsMatching(vendorId uint32, code Uint24) int {
	return len(m.TopLevelAvpsMatching(vendorId, code))
}

// IsRequest returns true if the message is a Diameter Request message (that
// is, the request flag in the Diameter message header is set)
func (m *Message) IsRequest() bool {
	return (m.Flags & MsgFlagRequest) != 0
}

// IsAnswer returns true if the message is a Diameter Answer message (that is,
// the request flag in the Diameter message header is not set)
func (m *Message) IsAnswer() bool {
	return !m.IsRequest()
}

// IsProxiable returns true if the proxiable flag in the Diameter message header is set
func (m *Message) IsProxiable() bool {
	return (m.Flags & MsgFlagProxiable) != 0
}

// IsError returns true if the message is a Diameter erro9r message (that
// is, the error flag in the Diameter message header is set)
func (m *Message) IsError() bool {
	return (m.Flags & MsgFlagError) != 0
}

// IsPotentiallyRetransmitted returns true if the potentially retransmit
// flag in the Diameter message header is set
func (m *Message) IsPotentiallyRetransmitted() bool {
	return (m.Flags & MsgFlagPotentialRetransmit) != 0
}

// Encode transforms the current message into an octet stream appropriate
// for network transmission
func (m *Message) Encode() []byte {
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.BigEndian, uint32(m.Version)<<24|uint32(m.Length)&0x00ffffff)
	binary.Write(buf, binary.BigEndian, uint32(m.Flags)<<24|uint32(m.Code)&0x00ffffff)
	binary.Write(buf, binary.BigEndian, m.AppID)
	binary.Write(buf, binary.BigEndian, m.HopByHopID)
	binary.Write(buf, binary.BigEndian, m.EndToEndID)
	for _, avp := range m.Avps {
		buf.Write(avp.Encode())
	}
	return buf.Bytes()
}

// DecodeMessage accepts an octet stream and attempts to interpret it as a Diameter
// message.  The stream must contain at least a single Diameter
// message.  To decode incoming streams, use a MessageStreamReader.  If the input
// stream is at least one Diameter message, or an error occurs in the reading of
// the stream or creation of the message, return nil and an error; otherwise
// return a Message object and nil for the error.
func DecodeMessage(input []byte) (*Message, error) {
	m := new(Message)
	buf := bytes.NewReader(input)
	var flagsAndLength uint32
	err := binary.Read(buf, binary.BigEndian, &flagsAndLength)
	if err != nil {
		return nil, err
	}

	m.Version = byte((flagsAndLength & 0xFF000000) >> 24)
	m.Length = Uint24(flagsAndLength & 0x00FFFFFF)

	if Uint24(len(input)) < m.Length {
		return nil, errors.New("header length does not match stream length")
	}

	err = binary.Read(buf, binary.BigEndian, &flagsAndLength)
	if err != nil {
		return nil, err
	}

	m.Flags = byte((flagsAndLength & 0xFF000000) >> 24)
	m.Code = Uint24(flagsAndLength & 0x00FFFFFF)

	err = binary.Read(buf, binary.BigEndian, &m.AppID)
	if err != nil {
		return nil, err
	}

	err = binary.Read(buf, binary.BigEndian, &m.HopByHopID)
	if err != nil {
		return nil, err
	}

	err = binary.Read(buf, binary.BigEndian, &m.EndToEndID)
	if err != nil {
		return nil, err
	}

	m.Avps = make([]*AVP, 0)
	b := input[MsgHeaderSize:int(m.Length)]
	for len(b) > 0 {
		var avp *AVP
		avp, err = DecodeAVP(b)

		if err != nil {
			return nil, err
		}

		b = b[avp.PaddedLength:]
		m.Avps = append(m.Avps, avp)
	}

	if err != nil {
		return nil, err
	}
	return m, err
}

// NewMessage creates a new diameter.Message instance.  'mandatoryAvps' will all
// have their Mandatory flag set to true.  The Mandatory flag for 'additionalAvps'
// will be left untouched.
func NewMessage(flags uint8, code Uint24, appID uint32, hopByHopID uint32, endToEndID uint32, mandatoryAvps []*AVP, additionalAvps []*AVP) (m *Message) {
	m = new(Message)

	m.Version = 1
	m.Flags = flags & 0xf0
	m.Code = code & 0x00ffffff
	m.AppID = appID
	m.HopByHopID = hopByHopID
	m.EndToEndID = endToEndID
	m.Avps = make([]*AVP, len(mandatoryAvps)+len(additionalAvps))

	m.Length = MsgHeaderSize
	for i := 0; i < len(mandatoryAvps); i++ {
		m.Length += Uint24(mandatoryAvps[i].PaddedLength)
		m.Avps[i] = mandatoryAvps[i]
		m.Avps[i].Mandatory = true
	}

	t := len(mandatoryAvps)

	for i := 0; i < len(additionalAvps); i++ {
		m.Length += Uint24(additionalAvps[i].PaddedLength)
		m.Avps[t+i] = additionalAvps[i]
	}

	return m
}

// Clone makes a copy of the current message.  No effort is made to be thread-safe
// against changes to the message being cloned.  All AVPs in this message are also
// cloned.
func (m *Message) Clone() *Message {
	clonedAvps := make([]*AVP, len(m.Avps))
	for _, srcAvp := range m.Avps {
		clonedAvps = append(clonedAvps, srcAvp.Clone())
	}

	clonedMessage := *m
	clonedMessage.Avps = clonedAvps

	return &clonedMessage
}

// Equals compares the current Message object to a different message object.  If
// they have equivalent values for all fields and AVPs, return true; otherwise
// return false.  AVPs are compared exactly in order.
func (m *Message) Equals(c *Message) bool {
	// XXX: This can almost certainly be made into just a straight memory
	// value comparison between the two objects.
	if m == nil {
		return false
	}

	if m.Version != c.Version || m.Flags != c.Flags || m.Code != c.Code || m.AppID != c.AppID || m.HopByHopID != c.HopByHopID || m.EndToEndID != c.EndToEndID {
		return false
	}

	if len(m.Avps) != len(c.Avps) {
		return false
	}

	for i := 0; i < len(m.Avps); i++ {
		if !m.Avps[i].Equal(c.Avps[i]) {
			return false
		}
	}

	return true
}

// BecomeAnAnswerBasedOnTheRequestMessage extracts the end-to-end-id and hop-by-hop-id
// from the request message and applies them to this message.  It also clears
// the request flag if it is set and sets this message's code to the request
// message's code.  Return this message, so that this call may be chained, if
// desired.
func (m *Message) BecomeAnAnswerBasedOnTheRequestMessage(request *Message) *Message {
	m.EndToEndID = request.EndToEndID
	m.HopByHopID = request.HopByHopID
	m.AppID = request.AppID
	m.Code = request.Code
	m.Flags &^= MsgFlagRequest

	return m
}

// GenerateMatchingResponseWithAvps duplicates the end-to-end-id, hop-by-hop-id, code
// and flags from the message, but clearing the request flag.  The newly generated message
// will contain the provided AVPs.
func (m *Message) GenerateMatchingResponseWithAvps(mandatoryAvps []*AVP, optionalAvps []*AVP) *Message {
	return NewMessage(m.Flags&^MsgFlagRequest, m.Code, m.AppID, m.HopByHopID, m.EndToEndID, mandatoryAvps, optionalAvps)
}

const (
	streamReaderBaseBufferSizeInBytes int = 16384
)

// MessageByteReader simplifies the reading of an octet stream which must be
// converted to one or more diameter.Message objects.  Generally, a new
// MessageByteReader is created, then ReceiveBytes() is repeatedly called on
// an input stream (which must be in network byte order) as bytes arrive.
// This method will return diameter.Message objects as they can be extracted, and
// store any bytes that are left over after message conversion
type MessageByteReader struct {
	incomingBuffer []byte
}

// NewMessageByteReader creates a new MessageStreamReader object
func NewMessageByteReader() *MessageByteReader {
	return &MessageByteReader{
		incomingBuffer: make([]byte, 0, streamReaderBaseBufferSizeInBytes),
	}
}

// ReceiveBytes returns one or more diameter.Message objects read from the incoming
// byte stream.  Return nil if no Message is yet found.  Return error on malformed
// byte stream.  If an error is returned, subsequent calls are no longer reliable.
func (reader *MessageByteReader) ReceiveBytes(incoming []byte) ([]*Message, error) {
	reader.incomingBuffer = append(reader.incomingBuffer, incoming...)

	setOfExtractedMessages := make([]*Message, 0, 3)

	for {
		nextMessageInStream, err := reader.ReceiveBytesButReturnAtMostOneMessage([]byte{})

		if err != nil {
			return nil, err
		}

		if nextMessageInStream != nil {
			setOfExtractedMessages = append(setOfExtractedMessages, nextMessageInStream)
		} else {
			return setOfExtractedMessages, nil
		}
	}
}

// ReceiveBytesButReturnAtMostOneMessage is the same as ReceiveBytes(), but it will return no more
// than one message.  If more than one message is available in the internal buffer plus the incoming bytes,
// all messages after the first are saved in the internal buffer, which means they'll be returned on the
// next call to ReceiveBytes().
func (reader *MessageByteReader) ReceiveBytesButReturnAtMostOneMessage(incoming []byte) (*Message, error) {
	reader.incomingBuffer = append(reader.incomingBuffer, incoming...)

	nextMessageInStream, incomingBytesLeftToProcess, err := extractNextMessageInByteBufferIfThereIsOne(reader.incomingBuffer)

	if err != nil {
		return nil, err
	}

	if nextMessageInStream != nil {
		reader.incomingBuffer = incomingBytesLeftToProcess
		return nextMessageInStream, nil
	}

	return nil, nil
}

// Read a stream buffer and attempt to extract a Message, if there are enough
// bytes in the stream.  If not, return (nil, incoming, nil).  If the stream is malformed for
// a message, return (nil, incoming, error). If there is at least enough bytes for a message
// and the stream is well-formed, return (m, leftOverBytes, nil), where m is a Message and
// remainder is a slice of incoming, starting one byte after the extracted message.
func extractNextMessageInByteBufferIfThereIsOne(incoming []byte) (*Message, []byte, error) {
	if len(incoming) == 0 {
		return nil, incoming, nil
	}

	buf := bytes.NewReader(incoming)

	// 20 is the diameter header length
	if len(incoming) < 20 {
		var version uint8
		err := binary.Read(buf, binary.BigEndian, &version)

		if err != nil {
			return nil, incoming, err
		} else if version != 1 {
			return nil, incoming, errors.New("unknown Diameter version")
		} else {
			return nil, incoming, nil
		}
	} else {
		var flagsAndLength uint32
		err := binary.Read(buf, binary.BigEndian, &flagsAndLength)

		if err != nil {
			return nil, incoming, err
		}

		version := byte((flagsAndLength & 0xFF000000) >> 24)
		length := Uint24(flagsAndLength & 0x00FFFFFF)

		if version != 1 {
			return nil, incoming, errors.New("invalid Diameter message version")
		}

		if len(incoming) < int(length) {
			return nil, incoming, nil
		}

		m, err := DecodeMessage(incoming)

		if err != nil {
			return nil, incoming, err
		}

		return m, incoming[m.Length:], nil
	}
}

// MessageStreamReader is the same as MessageByteReader, but instead of being passed
// bytes repeatedly, it is supplied an io.Reader, and reads from that, blocking until
// messages are found on each call to ReadNextMessage().
type MessageStreamReader struct {
	underlyingReader   io.Reader
	internalByteBuffer []byte
	readBuffer         []byte
}

// NewMessageStreamReader creates an empty reader which will use the provided io.Reader
// for each call to ReadNextMessage().
func NewMessageStreamReader(usingReader io.Reader) *MessageStreamReader {
	return &MessageStreamReader{
		underlyingReader:   usingReader,
		internalByteBuffer: make([]byte, 0, 16384),
		readBuffer:         make([]byte, 9100),
	}
}

// ReadNextMessage will repeatedly perform a Read() on the underlying Reader until
// a message is found.  It will then queue any additional bytes after the returned
// message.  If that internal byte buffer contains a complete message, a subsequent
// call will return that message and buffer again any left over bytes.  This will
// continue until the internal buffer no longer contains a complete message, at which
// point, another Read() will occur.  The returned error may be io.EOF.  In this case,
// the returned message will still be nil.
func (reader *MessageStreamReader) ReadNextMessage() (*Message, error) {
	for {
		message, err := reader.ReadOnce()
		if err != nil {
			return nil, err
		}

		if message != nil {
			return message, nil
		}
	}
}

// ReadOnce does the same as ReadNextMessage(), but it will perform no more than a
// single Read() on the underlying Reader.  If the Read() (plus any internal buffer)
// does not yield a complete message, this will return.  In that case, the returned
// Message and error will both be nil.
func (reader *MessageStreamReader) ReadOnce() (*Message, error) {
	message, leftOverBytes, err := extractNextMessageInByteBufferIfThereIsOne(reader.internalByteBuffer)
	if err != nil {
		return nil, err
	}

	if message != nil {
		reader.internalByteBuffer = leftOverBytes
		return message, nil
	}

	bytesRead, err := reader.underlyingReader.Read(reader.readBuffer)
	if err != nil {
		return nil, err
	}

	reader.internalByteBuffer = append(reader.internalByteBuffer, reader.readBuffer[:bytesRead]...)

	return nil, nil
}
