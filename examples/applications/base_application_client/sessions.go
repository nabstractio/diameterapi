package main

import (
	"github.com/blorticus-go/diameter"
	"github.com/blorticus-go/diameter/agent"
)

var cachedResponseCode2001 = diameter.NewTypedAVP(268, 0, true, diameter.Unsigned32, 2001)

type phase int

const (
	initial phase = iota
	updates
	terminate
	terminating
)

type DiameterSession struct {
	SessionId             string
	diameterEntity        *agent.DiameterEntity
	dictionary            *diameter.Dictionary
	phase                 phase
	numberOfUpdatesToSend uint
	updateSequenceNumber  uint
}

func NewDiameterSession(diameterEntity *agent.DiameterEntity, dictionary *diameter.Dictionary, numberOfUpdatesToSend uint) *DiameterSession {
	return &DiameterSession{
		SessionId:             diameter.GenerateSessionId(diameterEntity.OriginHost),
		diameterEntity:        diameterEntity,
		dictionary:            dictionary,
		phase:                 initial,
		numberOfUpdatesToSend: numberOfUpdatesToSend,
		updateSequenceNumber:  0,
	}
}

func (s *DiameterSession) NextMessageForSession() *diameter.Message {
	switch s.phase {
	case initial:
		s.phase = updates
		return generateCCRi(s.SessionId, s.dictionary, s.diameterEntity.OriginHost, s.diameterEntity.OriginRealm)

	case updates:
		s.updateSequenceNumber++
		if s.updateSequenceNumber >= s.numberOfUpdatesToSend {
			s.phase = terminate
		}
		return generateCCRu(s.SessionId, s.updateSequenceNumber, s.dictionary, s.diameterEntity.OriginHost, s.diameterEntity.OriginRealm)

	case terminate:
		s.phase = terminating
		return generateCCRt(s.SessionId, s.updateSequenceNumber+1, s.dictionary, s.diameterEntity.OriginHost, s.diameterEntity.OriginRealm)
	}

	return nil
}

func (s *DiameterSession) WasTerminating() bool {
	return s.phase == terminating
}

func generateCCRi(sessionId string, dictionary *diameter.Dictionary, originHost string, originRealm string) *diameter.Message {
	return dictionary.Message("CCR", diameter.MessageFlags{}, []*diameter.AVP{
		cachedResponseCode2001,
		dictionary.AVP("Session-Id", sessionId),
		dictionary.AVP("Origin-Host", originHost),
		dictionary.AVP("Origin-Realm", originRealm),
		dictionary.AVP("Destination-Realm", originRealm),
		dictionary.AVP("Auth-Application-Id", uint32(4)),
		diameter.NewTypedAVP(461, 0, true, diameter.UTF8String, "service@example.com"),
		diameter.NewTypedAVP(416, 0, true, diameter.Enumerated, int32(1)),
		diameter.NewTypedAVP(415, 0, true, diameter.Unsigned32, uint32(0)),
	}, nil)
}

func generateCCRu(sessionId string, requestNumber uint, dictionary *diameter.Dictionary, originHost string, originRealm string) *diameter.Message {
	return dictionary.Message("CCR", diameter.MessageFlags{}, []*diameter.AVP{
		cachedResponseCode2001,
		dictionary.AVP("Session-Id", sessionId),
		dictionary.AVP("Origin-Host", originHost),
		dictionary.AVP("Origin-Realm", originRealm),
		dictionary.AVP("Destination-Realm", originRealm),
		dictionary.AVP("Auth-Application-Id", uint32(4)),
		diameter.NewTypedAVP(461, 0, true, diameter.UTF8String, "service@example.com"),
		diameter.NewTypedAVP(416, 0, true, diameter.Enumerated, int32(2)),
		diameter.NewTypedAVP(415, 0, true, diameter.Unsigned32, uint32(requestNumber)),
	}, nil)
}

func generateCCRt(sessionId string, requestNumber uint, dictionary *diameter.Dictionary, originHost string, originRealm string) *diameter.Message {
	return dictionary.Message("CCR", diameter.MessageFlags{}, []*diameter.AVP{
		cachedResponseCode2001,
		dictionary.AVP("Session-Id", sessionId),
		dictionary.AVP("Origin-Host", originHost),
		dictionary.AVP("Origin-Realm", originRealm),
		dictionary.AVP("Destination-Realm", originRealm),
		dictionary.AVP("Auth-Application-Id", uint32(4)),
		diameter.NewTypedAVP(461, 0, true, diameter.UTF8String, "service@example.com"),
		diameter.NewTypedAVP(416, 0, true, diameter.Enumerated, int32(3)),
		diameter.NewTypedAVP(415, 0, true, diameter.Unsigned32, uint32(requestNumber)),
	}, nil)
}
