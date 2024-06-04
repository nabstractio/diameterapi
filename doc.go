// Package diameter implements Diameter (RFC 6733) Message and AVP encoders and decoders.  It also provides a method for creating, reading and using
// Diameter dictionaries.  A dictionary provides human-readable names for Message type and AVP types.  It also provides type information for AVPs,
// making AVPs more convenient to create, read and manipulate.  A sample dictionary (describing all Message and AVP types in RFC6733) can be found
// in the examples/ directory.
//
// This package also includes an implementation of a Diameter Agent, which manages the Diameter base protocol state-machine -- and corresponding
// messaging -- for one diameter connections to one or more peers.
//
// 
package diameter
