package main

import (
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/blorticus-go/diameter"
	"github.com/blorticus-go/diameter/agent"
)

func main() {
	cliArgs, err := ProcessCommandLineArguments()
	dieOnError(err)

	dictionary, err := diameter.DictionaryFromYamlFile(cliArgs.PathToDictionary)
	dieOnError(err)

	conn, err := net.Dial("tcp", cliArgs.Connect)
	dieOnError(err)

	diameterAgent := agent.New()
	agentEventChannel := diameterAgent.EventChannel()

	go diameterAgent.Run(nil)

	clientEntity := &agent.DiameterEntity{
		OriginHost:      cliArgs.OriginHost,
		OriginRealm:     cliArgs.OriginRealm,
		HostIPAddresses: []*net.IP{&conn.LocalAddr().(*net.TCPAddr).IP},
		VendorID:        0,
		ProductName:     "diameter-go",
	}

	diameterAgent.EstablishDiameterConnectionTo(conn, clientEntity)

	sessionBySessionId := make(map[string]*DiameterSession)

	for i := uint(0); i < cliArgs.NumberOfSessionsToGenerate; i++ {
		s := NewDiameterSession(clientEntity, dictionary, 3)
		if sessionBySessionId[s.SessionId] != nil {
			die("generated two SessionIds with the same value: %s\n", s.SessionId)
		}
		sessionBySessionId[s.SessionId] = s
	}

	for {
		event := <-agentEventChannel

		switch event.Type {
		case agent.ClosedTransportToPeerEvent:
			logGeneralEvent("closed transport to peer", event.Connection, event.Peer)
			return

		case agent.PeerClosedTransportEvent:
			logGeneralEvent("peer closed transport", event.Connection, event.Peer)

		case agent.StateMachineMessageReceivedFromPeerEvent:
			logDiameterMessage(event.Message, dictionary, "received", event.Peer)

		case agent.StateMachineMessageSentToPeerEvent:
			logDiameterMessage(event.Message, dictionary, "sent", event.Peer)

		case agent.DiameterConnectionEstablishedEvent:
			logGeneralEvent("diameter connection established", event.Connection, event.Peer)

			for _, s := range sessionBySessionId {
				ccr := s.NextMessageForSession()
				if failedToSend := tryToSendMessageToPeer(ccr, event.Peer, event.Connection); failedToSend {
					os.Exit(2)
				}
				logDiameterMessage(ccr, dictionary, "sent", event.Peer)
			}

		case agent.DiameterConnectionClosedEvent:
			logGeneralEvent("diameter connection closed", event.Connection, event.Peer)

		case agent.MessageReceivedFromPeerEvent:
			logDiameterMessage(event.Message, dictionary, "received", event.Peer)

			if event.Message.AppID == 0 && event.Message.Code == 272 && event.Message.IsAnswer() {
				sessionIdAvp := event.Message.FirstAvpMatching(0, 263)
				if sessionIdAvp == nil {
					logError(errors.New("received CCA without a Session-Id"), event.Connection, event.Peer)
					continue
				}

				sessionId := string(sessionIdAvp.Data)
				session := sessionBySessionId[sessionId]
				if session == nil {
					logError(fmt.Errorf("peer sent CCA with Session-Id (%s) that was not locally generated", sessionId), event.Connection, event.Peer)
					continue
				}

				if session.WasTerminating() {
					delete(sessionBySessionId, sessionId)
					if len(sessionBySessionId) == 0 {
						if err := event.Peer.InitiateDisconnect(); err != nil {
							logError(fmt.Errorf("failed to deliver Peer-Disconnect Request: %s", err), event.Connection, event.Peer)
							os.Exit(3)
						}
					}
					continue
				}

				ccr := session.NextMessageForSession()

				if ccr == nil {
					logError(errors.New("received unexpected CCA from peer after session is already terminated"), event.Connection, event.Peer)
					continue
				}

				if failedToSend := tryToSendMessageToPeer(ccr, event.Peer, event.Connection); failedToSend {
					os.Exit(2)
				}

				logDiameterMessage(ccr, dictionary, "sent", event.Peer)
			}

		case agent.ErrorEvent:
			logError(event.Error, event.Connection, event.Peer)
		}
	}
}

func logGeneralEvent(eventDetail string, conn net.Conn, peer *agent.Peer) {
	fmt.Printf(`event msg="%s",localAddress=%s,remoteAddress=%s`, eventDetail, conn.LocalAddr().String(), conn.RemoteAddr().String())
	if peer != nil {
		fmt.Printf(`,peer="%s"`, peer.Identity.OriginHost)
	}
	fmt.Println()
}

func logDiameterMessage(m *diameter.Message, dictionary *diameter.Dictionary, direction string, peer *agent.Peer) {
	fmt.Printf(`message direction=%s,type=%s`, direction, dictionary.MessageCodeAsAString(m))
	if peer != nil {
		fmt.Printf(`,peer="%s"`, peer.Identity.OriginHost)
	}
	fmt.Println()
}

func logError(err error, conn net.Conn, peer *agent.Peer) {
	fmt.Printf(`error msg="%s"`, err)
	if conn != nil {
		fmt.Printf(",localAddress=%s,remoteAddress=%s", conn.LocalAddr().String(), conn.RemoteAddr().String())
	}
	if peer != nil {
		fmt.Printf(`,peer="%s"`, peer.Identity.OriginHost)
	}
	fmt.Println()
}

func tryToSendMessageToPeer(message *diameter.Message, peer *agent.Peer, transport net.Conn) (failedToSend bool) {
	if err := peer.SendMessage(message); err != nil {
		logError(err, transport, peer)

		if err := peer.InitiateDisconnect(); err != nil {
			logError(fmt.Errorf("failed to deliver Peer-Disconnect Request: %s", err), transport, peer)
			os.Exit(4)
		}

		return true
	}

	return false
}

func dieOnError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func die(f string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, f, a...)
	os.Exit(1)
}
