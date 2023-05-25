package proxy

import (
	"flint/mc"
	"fmt"
)

const (
	stateStatus = iota + 1
	stateLogin
)

const (
	packetStatusRequest = 0x00
	packetStatusPing    = 0x01
	packetLogin         = 0x00
)

type statusConnectionHandler struct {
	message string
	state   int32
}

func (s *statusConnectionHandler) handle(conn *mc.Conn) {
	for {
		packet, err := conn.ReadPacket()
		if err != nil {
			logPacketReadError(conn, err)
			return
		}

		err = s.handlePacket(conn, packet)
		if err != nil {
			logPacketReadError(conn, err)
			return
		}
	}
}

func (s *statusConnectionHandler) handleStatusPacket(conn *mc.Conn, packet mc.Packet) error {
	switch packet.Id {
	case packetStatusRequest:
		responsePacket, err := mc.EncodeStatusResponsePacket(mc.StatusResponsePacket{
			Version: mc.StatusVersion{
				Name:     "1.7.10",
				Protocol: 5,
			},
			Players: mc.StatusPlayers{
				Max:    0,
				Online: 0,
				Sample: []mc.StatusPlayerSample{
					{Name: "powered by §bflint", Id: "00000000-0000-0000-0000-000000000000"},
				},
			},
			Description: mc.ChatComponent{
				Text: s.message,
			},
		})
		if err != nil {
			return err
		}

		return conn.WritePacket(responsePacket)

	case packetStatusPing:
		pingPacket, err := mc.DecodePingPacket(packet)
		if err != nil {
			return err
		}

		pongPacket, err := mc.EncodePingPacket(pingPacket)
		if err != nil {
			return err
		}

		return conn.WritePacket(pongPacket)
	default:
		return fmt.Errorf("received unknown status packet 0x%02x", packet.Id)
	}
}

func (s *statusConnectionHandler) handleLoginPacket(conn *mc.Conn, packet mc.Packet) error {
	switch packet.Id {
	case packetLogin:
		disconnectPacket, err := mc.EncodeDisconnectPacket(mc.DisconnectPacket{
			Reason: mc.ChatComponent{Text: s.message},
		})
		if err != nil {
			return err
		}

		return conn.WritePacket(disconnectPacket)
	default:
		return fmt.Errorf("received unknown login packet 0x%02x", packet.Id)
	}
}

func (s *statusConnectionHandler) handlePacket(conn *mc.Conn, packet mc.Packet) error {
	switch s.state {
	case stateStatus:
		return s.handleStatusPacket(conn, packet)
	case stateLogin:
		return s.handleLoginPacket(conn, packet)
	default:
		return fmt.Errorf("invalid connection state %d", s.state)
	}
}
