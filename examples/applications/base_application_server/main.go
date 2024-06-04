package main

import (
	"fmt"
	"net"
	"os"

	"github.com/blorticus-go/diameter"
	"github.com/blorticus-go/diameter/agent"
)

// server [-bind [<ip>]:<port>] [-originHost <originHost>] [-originRealm <originRealm>] -dictionary /path/to/dictionary.yaml
func main() {
	cliArgs, err := ProcessCommandLineArguments()
	dieOnError(err)

	dictionary, err := diameter.DictionaryFromYamlFile(cliArgs.PathToDictionary)
	dieOnError(err)

	listener, err := net.Listen("tcp", cliArgs.Bind)
	dieOnError(err)

	diameterAgent := agent.New()
	agentEventChannel := diameterAgent.EventChannel()

	go diameterAgent.Run([]*agent.AgentReceiver{
		{
			Listener: listener,
			IdentityToAssert: &agent.DiameterEntity{
				OriginHost:      cliArgs.OriginHost,
				OriginRealm:     cliArgs.OriginRealm,
				HostIPAddresses: nil,
				VendorID:        0,
				ProductName:     "diameter-go-server",
			},
		},
	})

	for {
		event := <-agentEventChannel

		switch event.Type {
		case agent.ListenerAcceptedTransportEvent:
			logGeneralEvent("accepted incoming transport", event.Connection, event.Peer)

		case agent.ClosedTransportToPeerEvent:
			logGeneralEvent("closed transport to peer", event.Connection, event.Peer)

		case agent.PeerClosedTransportEvent:
			logGeneralEvent("peer closed transport", event.Connection, event.Peer)

		case agent.StateMachineMessageReceivedFromPeerEvent:
			logDiameterMessage(event.Message, dictionary, "received", event.Peer)

		case agent.StateMachineMessageSentToPeerEvent:
			logDiameterMessage(event.Message, dictionary, "sent", event.Peer)

		case agent.DiameterConnectionEstablishedEvent:
			logGeneralEvent("diameter connection established", event.Connection, event.Peer)

		case agent.DiameterConnectionClosedEvent:
			logGeneralEvent("diameter connection closed", event.Connection, event.Peer)

		case agent.MessageReceivedFromPeerEvent:
			logDiameterMessage(event.Message, dictionary, "received", event.Peer)

			if event.Message.AppID == 0 && event.Message.Code == 272 {
				if cca, err := generateCCAFromCCR(event.Message, cliArgs.OriginHost, cliArgs.OriginRealm, dictionary); err != nil {
					logError(err, event.Connection, event.Peer)
				} else {
					if err := event.Peer.SendMessage(cca); err != nil {
						logError(err, event.Connection, event.Peer)
						event.Peer.InitiateDisconnect()
					} else {
						logDiameterMessage(cca, dictionary, "sent", event.Peer)
					}
				}
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

func generateCCAFromCCR(ccr *diameter.Message, localOriginHost string, localOriginRealm string, dictionary *diameter.Dictionary) (*diameter.Message, error) {
	if ccr.IsAnswer() {
		return nil, fmt.Errorf("expected a CCR from the peer")
	}

	for _, avpCode := range []diameter.Uint24{263, 258, 416, 415} {
		if ccr.DoesNotHaveATopLevelAvpMatching(0, avpCode) {
			return nil, fmt.Errorf("the CCR is missing AVP with code (%d)", avpCode)
		}
	}

	cca := dictionary.Message("CCA", diameter.MessageFlags{}, []*diameter.AVP{
		ccr.FirstAvpMatching(0, 263),
		dictionary.AVP("Result-Code", uint32(2000)),
		dictionary.AVP("Origin-Host", localOriginHost),
		dictionary.AVP("Origin-Realm", localOriginRealm),
		ccr.FirstAvpMatching(0, 258),
		ccr.FirstAvpMatching(0, 416),
	}, nil)

	cca.BecomeAnAnswerBasedOnTheRequestMessage(ccr)

	return cca, nil
}

func dieOnError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
