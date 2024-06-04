# golang Diameter

A golang library for the Diameter protocol.

## Diameter Library

This module provides methods for encoding and decoding Diameter messages and AVPs.  Since Diameter is typically carried over TCP or SCTP, it also provides a stream reader, which extracts Diameter messages from a stream protocol.  The following example illustrates reading and writing Diameter messages from a stream:

```go
package main

import (
  "fmt"
  "os"
  "github.com/blorticus-go/diameter"
)

func main() {
  c, err := net.Dial("tcp", "10.10.10.1:5060")
  if err != nil {
    panic(err)
  }
  defer c.Close()

  sr := diameter.NewMessageStreamReader(c)
  g := diameter.NewSequenceGeneratorSet()

  cer := diameter.NewMessage(diameter.MsgFlagRequest, 257, 0, g.NextHopByHopId(), g.NextEndToEndId(), []*AVP{
    diameter.NewTypedAVP(264, 0, true, diameter.DiamIdent, "client01.example.com"),
    diameter.NewTypedAVP(296, 0, true, diameter.DiamIdent, "example.com"),
    diameter.NewTypedAVP(257, 0, true, diameter.Address, c.LocalAddr().(net.TCPAddr).IP),
    diameter.NewTypedAVP(266, 0, true, diameter.Unsigned32, 0),
    diameter.NewTypedAVP(269, 0, true, diameter.UTF8String, "go-diameter"),
  }, nil)

  if _, err := c.Write(cer.Encode()); err != nil {
    panic(err)
  }

  incomingMessage, err := sr.ReadNextMessage()
  if err != nil {
    panic(err)
  }

  if incomingMessage.Code != 257 {
    fmt.Fprintf(os.Stderr, "expected CEA, got message with Code (%d)\n", incomingMessage.Code)
    os.Exit(1)
  }

  if incomingMessage.IsRequest() {
    fmt.Fprintf(os.Stderr, "expected CEA, CER\n")
    os.Exit(2)
  }

  peerOriginHostAvp := FirstAvpMatching(264, 0)
  if peerOriginHostAvp == nil {
    fmt.Fprintf(os.Stderr, "peer failed to send Origin-Host in CEA\n")
    os.Exit(3)
  }

  peerOriginHostValue := string(peerOriginHostAvp.Data)
  fmt.Printf("received CEA from peer with Origin-Host (%s)\n", peerOriginHostValue)

  // ...
}
```

## Diameter Dictionaries

A Diameter dictionary allows one to provide human-readable names for Diameter message and AVP codes, and to provide type definitions for AVPs so that, when they are read, they can automatically be typed.  The `diameter.Dictionary` type can be generated from a YAML file.  The `examples` directory contains a sample dictionary for the base Diameter application.  Here is an example using a dictionary.

```go
package main

import (
  "fmt"
  "os"
  "github.com/blorticus-go/diameter"
)

func main() {
  dictionary, err := diameter.DictionaryFromYamlFile("./dictionary.yaml")
  if err != nil {
    panic(err)
  }

  c, err := net.Dial("tcp", "10.10.10.1:5060")
  if err != nil {
    panic(err)
  }
  defer c.Close()

  sr := diameter.NewMessageStreamReader(c)
  g := diameter.NewSequenceGeneratorSet()

  cer := dictionary.NewMessage("CER", MessageFlags(0), 0, []*AVP{
    dictionary.AVP("Origin-Host", "client01.example.com"),
    dictionary.AVP("Origin-Realm", "example.com"),
    dictionary.AVP("Host-IP-Address", c.LocalAddr().(net.TCPAddr).IP),
    dictionary.AVP("Vendor-Id", 0),
    dictionary.AVP("Product-Name", "go-diameter"),
  }, nil)

  if _, err := c.Write(cer.Encode()); err != nil {
    panic(err)
  }

  incomingMessage, err := sr.ReadNextMessage()
  if err != nil {
    panic(err)
  }

  answer, err := dictionary.TypeAMessage(incomingMessage)
  if err != nil {
    panic(err)
  }

  if answer.ExtendedAttributes.Abbreviated != "CEA" {*
    fmt.Fprintf(os.Stderr, "expected CEA, CER\n")
    os.Exit(2)
  }

  peerOriginHostAvp := FirstAvpMatching(264, 0)
  if peerOriginHostAvp == nil {
    fmt.Fprintf(os.Stderr, "peer failed to send Origin-Host in CEA\n")
    os.Exit(3)
  }

  peerOriginHostValue := string(peerOriginHostAvp.Data)
  fmt.Printf("received CEA from peer with Origin-Host (%s)\n", peerOriginHostValue)

  // ...
}
```