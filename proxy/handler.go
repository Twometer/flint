package proxy

import (
	"flint/config"
	"flint/mc"
	"fmt"
	"log"
)

type Handler struct {
	config config.Config
}

func NewHandler(config config.Config) Handler {
	return Handler{
		config: config,
	}
}

func (handler *Handler) UpdateConfig(config config.Config) {
	handler.config = config
}

func (handler *Handler) HandleConn(conn mc.Conn) {
	defer conn.Close()
	for {
		packet, err := conn.ReadPacket()
		if err != nil {
			logPacketReadError(conn, err)
			return
		}

		err = handler.handlePacket(&conn, packet)
		if err != nil {
			log.Printf("error: failed to handle packet: %v\n", err)
			return
		}
	}
}

func (handler *Handler) handlePacket(conn *mc.Conn, packet mc.Packet) error {
	//log.Printf("Received packet %d with len %d\n", packet.Id, packet.Data.Len())

	if conn.State == mc.StateInitial && packet.Id == 0x00 { // handshake
		handshakePacket, err := mc.DecodeHandshakePacket(packet)
		if err != nil {
			return err
		}

		expectedPort := conn.LocalPort()
		if handshakePacket.ServerPort != expectedPort {
			return fmt.Errorf("client connected to port %d but specified port %d", expectedPort, handshakePacket.ServerPort)
		}

		conn.State = handshakePacket.NextState
	} else if conn.State == mc.StateStatus && packet.Id == 0x00 {
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
				Text: handler.config.Messages.ServerNotFound,
			},
		})
		if err != nil {
			return err
		}

		err = conn.WritePacket(responsePacket)
		if err != nil {
			return err
		}
	} else if conn.State == mc.StateStatus && packet.Id == 0x01 {
		pingPacket, err := mc.DecodePingPacket(packet)
		if err != nil {
			return err
		}

		pongPacket, err := mc.EncodePingPacket(pingPacket)
		if err != nil {
			return err
		}

		err = conn.WritePacket(pongPacket)
		if err != nil {
			return err
		}
	} else if conn.State == mc.StateLogin && packet.Id == 0x00 {
		disconnectPacket, err := mc.EncodeDisconnectPacket(mc.DisconnectPacket{
			Reason: mc.ChatComponent{Text: handler.config.Messages.ServerNotFound},
		})
		if err != nil {
			return err
		}

		err = conn.WritePacket(disconnectPacket)
		if err != nil {
			return err
		}
	}

	return nil
}
