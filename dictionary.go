package diameter

import (
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v2"
)

// DictionaryYamlMetadataSpecificationType is the type for a dictionary yaml field Metadata section
type DictionaryYamlMetadataSpecificationType struct {
	Type       string `yaml:"Type"`
	Identifier string `yaml:"Identifier"`
	URL        string `yaml:"URL"`
}

// DictionaryYamlMetadataType is the type for a dictionary yaml Metadata section Specification subsection
type DictionaryYamlMetadataType struct {
	Name           string                                    `yaml:"Name"`
	Specifications []DictionaryYamlMetadataSpecificationType `yaml:"Specifications"`
}

// DictionaryYamlAvpEnumerationType is the type for Avp Enumerations
type DictionaryYamlAvpEnumerationType struct {
	Name  string `yaml:"Name"`
	Value uint32 `yaml:"Value"`
}

// DictionaryYamlAvpType is the type for AvpTypes in a Diameter YAML Dictionary
type DictionaryYamlAvpType struct {
	Name        string                             `yaml:"Name"`
	Code        uint32                             `yaml:"Code"`
	Type        string                             `yaml:"Type"`
	VendorID    uint32                             `yaml:"VendorId"`
	Enumeration []DictionaryYamlAvpEnumerationType `yaml:"Enumeration"`
}

// DictionaryYamlMessageAbbreviation is the type for MessageTypes.Abbreviations in a Diameter YAML Dictionary
type DictionaryYamlMessageAbbreviation struct {
	Request string `yaml:"Request"`
	Answer  string `yaml:"Answer"`
}

// DictionaryYamlMessageType is the type for MessageTypes in a Diameter YAML Dictionary
type DictionaryYamlMessageType struct {
	Basename      string                            `yaml:"Basename"`
	Code          uint32                            `yaml:"Code"`
	ApplicationID uint32                            `yaml:"ApplicationId"`
	Abbreviations DictionaryYamlMessageAbbreviation `yaml:"Abbreviations"`
}

// DictionaryYaml represents a YAML dictionary containing Diameter message type and AVP definitions
type DictionaryYaml struct {
	AvpTypes     []DictionaryYamlAvpType     `yaml:"AvpTypes"`
	MessageTypes []DictionaryYamlMessageType `yaml:"MessageTypes"`
}

type dictionaryMessageDescriptor struct {
	name          string
	abbreviation  string
	code          uint32
	appID         uint32
	isRequestType bool
}

type dictionaryAvpDescriptor struct {
	name             string
	code             uint32
	isVendorSpecific bool
	vendorID         uint32
	dataType         AVPDataType
}

type avpFullyQualifiedCodeType struct {
	vendorID uint32
	code     uint32
}

type messageFullyQualifiedCodeType struct {
	applicationID uint32
	code          uint32
}

// Dictionary is a Diameter dictionary, mapping AVP and message type data to names
type Dictionary struct {
	messageDescriptorByNameOrAbbreviation map[string]*dictionaryMessageDescriptor
	requestMessageDescriptorByCode        map[messageFullyQualifiedCodeType]*dictionaryMessageDescriptor
	answerMessageDescriptorByCode         map[messageFullyQualifiedCodeType]*dictionaryMessageDescriptor
	avpDescriptorByName                   map[string]*dictionaryAvpDescriptor
	avpDescriptorByFullyQualifiedCode     map[avpFullyQualifiedCodeType]*dictionaryAvpDescriptor
}

var mapOfYamlAvpTypeStringToAVPDataType = map[string]AVPDataType{
	"Unsigned32":  Unsigned32,
	"Unsigned64":  Unsigned64,
	"Integer32":   Integer32,
	"Integer64":   Integer64,
	"Enumerated":  Enumerated,
	"OctetString": OctetString,
	"UTF8String":  UTF8String,
	"Grouped":     Grouped,
	"Address":     Address,
	"Time":        Time,
	"DiamIdent":   DiamIdent,
	"DiamURI":     DiamURI,
}

func convertYamlAvpToDictionaryAvpDescriptor(yamlAvp *DictionaryYamlAvpType) (*dictionaryAvpDescriptor, error) {
	avpDescriptor := &dictionaryAvpDescriptor{
		code:     yamlAvp.Code,
		name:     yamlAvp.Name,
		vendorID: yamlAvp.VendorID,
	}

	if avpDataType, typeStringIsRecognized := mapOfYamlAvpTypeStringToAVPDataType[yamlAvp.Type]; typeStringIsRecognized {
		avpDescriptor.dataType = avpDataType
	} else {
		return nil, fmt.Errorf("provided Type (%s) invalid", yamlAvp.Type)
	}

	if yamlAvp.VendorID != 0 {
		avpDescriptor.isVendorSpecific = true
	}

	return avpDescriptor, nil
}

// fromYamlForm converts a DictionaryYaml to a Dictionary.  Returns error if a failure occurs
// or the values in the DictionaryYaml are malformed.
func fromYamlForm(yamlForm *DictionaryYaml) (*Dictionary, error) {
	dictionary := Dictionary{
		messageDescriptorByNameOrAbbreviation: make(map[string]*dictionaryMessageDescriptor),
		requestMessageDescriptorByCode:        make(map[messageFullyQualifiedCodeType]*dictionaryMessageDescriptor),
		answerMessageDescriptorByCode:         make(map[messageFullyQualifiedCodeType]*dictionaryMessageDescriptor),
		avpDescriptorByName:                   make(map[string]*dictionaryAvpDescriptor),
		avpDescriptorByFullyQualifiedCode:     make(map[avpFullyQualifiedCodeType]*dictionaryAvpDescriptor),
	}

	for _, yamlAvpType := range yamlForm.AvpTypes {
		avpDescriptor, err := convertYamlAvpToDictionaryAvpDescriptor(&yamlAvpType)

		if err != nil {
			return nil, err
		}

		dictionary.avpDescriptorByName[yamlAvpType.Name] = avpDescriptor
		dictionary.avpDescriptorByFullyQualifiedCode[avpFullyQualifiedCodeType{code: yamlAvpType.Code, vendorID: yamlAvpType.VendorID}] = avpDescriptor
	}

	for _, yamlMessageType := range yamlForm.MessageTypes {
		messageDescriptor := &dictionaryMessageDescriptor{
			code:          yamlMessageType.Code,
			abbreviation:  yamlMessageType.Abbreviations.Request,
			name:          yamlMessageType.Basename + "-Request",
			appID:         yamlMessageType.ApplicationID,
			isRequestType: true,
		}

		dictionary.messageDescriptorByNameOrAbbreviation[yamlMessageType.Basename+"-Request"] = messageDescriptor
		dictionary.messageDescriptorByNameOrAbbreviation[yamlMessageType.Abbreviations.Request] = messageDescriptor
		dictionary.requestMessageDescriptorByCode[messageFullyQualifiedCodeType{yamlMessageType.ApplicationID, yamlMessageType.Code}] = messageDescriptor

		messageDescriptor = &dictionaryMessageDescriptor{
			code:          yamlMessageType.Code,
			abbreviation:  yamlMessageType.Abbreviations.Answer,
			name:          yamlMessageType.Basename + "-Answer",
			appID:         yamlMessageType.ApplicationID,
			isRequestType: false,
		}

		dictionary.messageDescriptorByNameOrAbbreviation[yamlMessageType.Basename+"-Answer"] = messageDescriptor
		dictionary.messageDescriptorByNameOrAbbreviation[yamlMessageType.Abbreviations.Answer] = messageDescriptor
		dictionary.answerMessageDescriptorByCode[messageFullyQualifiedCodeType{yamlMessageType.ApplicationID, yamlMessageType.Code}] = messageDescriptor
	}

	return &dictionary, nil
}

// DictionaryFromYamlFile processes a file that should be a YAML formatted Diameter dictionary
func DictionaryFromYamlFile(filepath string) (*Dictionary, error) {
	contentsOfFileAsString, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file (%s): %s", filepath, err.Error())
	}

	return DictionaryFromYamlString(string(contentsOfFileAsString))
}

// DictionaryFromYamlString reads a string containing a Diameter dictionary in YAML format
func DictionaryFromYamlString(yamlString string) (*Dictionary, error) {
	dictionaryYaml := new(DictionaryYaml)
	err := yaml.Unmarshal([]byte(yamlString), &dictionaryYaml)

	if err != nil {
		return nil, err
	}

	dictionary, err := fromYamlForm(dictionaryYaml)

	if err != nil {
		return nil, err
	}

	return dictionary, nil
}

func (dictionary *Dictionary) MessageCodeAsAString(m *Message) string {
	if m.IsRequest() {
		if name := dictionary.requestMessageDescriptorByCode[messageFullyQualifiedCodeType{m.AppID, uint32(m.Code)}]; name != nil {
			return name.name
		}
	} else {
		if name := dictionary.answerMessageDescriptorByCode[messageFullyQualifiedCodeType{m.AppID, uint32(m.Code)}]; name != nil {
			return name.name
		}
	}

	return ""
}

// DataTypeForAVPNamed looks up the data type for the specific AVP
func (dictionary *Dictionary) DataTypeForAVPNamed(name string) (AVPDataType, error) {
	descriptor, isInMap := dictionary.avpDescriptorByName[name]

	if !isInMap {
		return TypeOrAvpUnknown, fmt.Errorf("no AVP named (%s) in the dictionary", name)
	}

	return descriptor.dataType, nil
}

// DataTypeForAvp returns the AVPDataType for the AVP based on its vendor-id and code.  If the type is not in the dictionary, returns TypeOrAvpUnknown.
func (dictionary *Dictionary) DataTypeForAvp(avp *AVP) AVPDataType {
	if diameterType, isInMap := dictionary.avpDescriptorByFullyQualifiedCode[avpFullyQualifiedCodeType{avp.VendorID, avp.Code}]; isInMap {
		return diameterType.dataType
	}

	return TypeOrAvpUnknown
}

// AVPErrorable returns an AVP based on the dictionary definition.  If the name is not in
// the dictionary, or the value type is incorrect based on the dictionary definition,
// return an error.  This is Errorable because it may throw an error.  It is assumed
// that this will be the uncommon case, because ordinarily, the value will be known in
// advance by the application creating it.
func (dictionary *Dictionary) AVPErrorable(name string, value interface{}) (*AVP, error) {
	descriptor, isInMap := dictionary.avpDescriptorByName[name]

	if !isInMap {
		return nil, fmt.Errorf("no AVP named (%s) in the dictionary", name)
	}

	return NewTypedAVPErrorable(descriptor.code, descriptor.vendorID, false, descriptor.dataType, value)
}

// AVP is the same as AVPErrorable, except that, if an error occurs, panic() is invoked
// with the error string
func (dictionary *Dictionary) AVP(name string, value interface{}) *AVP {
	avp, err := dictionary.AVPErrorable(name, value)

	if err != nil {
		panic(err)
	}

	return avp
}

// TypeAnAvp attempts to provide the ExtendedAttributes for the provided AVP.  If the AVP type
// is not found in the dictionary, the ExtendedAttributes for untypedAvp is set to nil and the
// untypedAvp is returned.  If an error occurs when attempting to conver the AVP's data to the
// type in the dictionary, return (nil, err).  Otherwise, return untypedAvp with its
// ExtendedAttributes set.
func (dictionary *Dictionary) TypeAnAvp(untypedAvp *AVP) (*AVP, error) {
	avpInfo, isInMap := dictionary.avpDescriptorByFullyQualifiedCode[avpFullyQualifiedCodeType{untypedAvp.VendorID, untypedAvp.Code}]

	if !isInMap || avpInfo.dataType == TypeOrAvpUnknown {
		untypedAvp.ExtendedAttributes = nil
		return untypedAvp, nil
	}

	typedData, err := untypedAvp.ConvertDataToTypedData(avpInfo.dataType)
	if err != nil {
		return nil, err
	}

	untypedAvp.ExtendedAttributes = &AVPExtendedAttributes{
		Name:       avpInfo.name,
		DataType:   avpInfo.dataType,
		TypedValue: typedData,
	}

	return untypedAvp, nil
}

// MessageFlags provides the Diameter Message flag types
type MessageFlags struct {
	Proxiable           bool
	Error               bool
	PotentialRetransmit bool
}

// MessageErrorable returns a Message based on the dictionary definition.  If the name is
// not present in the dictionary, an error is returned.
func (dictionary *Dictionary) MessageErrorable(name string, flags MessageFlags, mandatoryAVPs []*AVP, additionalAVPs []*AVP) (*Message, error) {
	messageDescriptor, messageTypeIsDefined := dictionary.messageDescriptorByNameOrAbbreviation[name]
	if !messageTypeIsDefined {
		return nil, fmt.Errorf("message of type (%s) is not known", name)
	}

	flagsEncoded := uint8(0)
	if flags.PotentialRetransmit {
		flagsEncoded |= MsgFlagPotentialRetransmit
	}
	if flags.Error {
		flagsEncoded |= MsgFlagError
	}
	if flags.Proxiable {
		flagsEncoded |= MsgFlagProxiable
	}

	if messageDescriptor.isRequestType {
		flagsEncoded |= MsgFlagRequest
	}

	return NewMessage(flagsEncoded, Uint24(messageDescriptor.code), messageDescriptor.appID, 0, 0, mandatoryAVPs, additionalAVPs), nil
}

// Message is the same as MessageErrorable, except that, if an error occurs, panic() is
// invoked with the error string
func (dictionary *Dictionary) Message(name string, flags MessageFlags, mandatoryAVPs []*AVP, additionalAVPs []*AVP) *Message {
	m, err := dictionary.MessageErrorable(name, flags, mandatoryAVPs, additionalAVPs)
	if err != nil {
		panic(err)
	}

	return m
}

// TypeAMessage attempts to provide ExendedAttribute information for the provided message based on a message
// definition in the dictionary.  If no definition exists for the message type, the ExtendedAttributes is set to nil.
// This method then iterates through the message AVP set, attempting to convert each AVP to its typed value (see TypeAnAvp).
// If no error occurs, returns the original message with (possibly) typed AVPs.  Otherwise, returns nil and the error.
func (dictionary *Dictionary) TypeAMessage(m *Message) (*Message, error) {
	var descriptor *dictionaryMessageDescriptor
	var descriptorIsInMap bool

	if m.IsRequest() {
		descriptor, descriptorIsInMap = dictionary.requestMessageDescriptorByCode[messageFullyQualifiedCodeType{m.AppID, uint32(m.Code)}]
	} else {
		descriptor, descriptorIsInMap = dictionary.answerMessageDescriptorByCode[messageFullyQualifiedCodeType{m.AppID, uint32(m.Code)}]
	}

	if descriptorIsInMap {
		m.ExtendedAttributes = &MessageExtendedAttributes{
			Name:            descriptor.name,
			AbbreviatedName: descriptor.abbreviation,
		}
	} else {
		m.ExtendedAttributes = nil
	}

	for _, avp := range m.Avps {
		_, err := dictionary.TypeAnAvp(avp)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}
