package proxy

import (
	"flint/config"
	"flint/mc"
	"fmt"
	"log"
)

type Server struct {
	config config.Config
}

func NewServer(config config.Config) Server {
	return Server{
		config: config,
	}
}

func (server *Server) UpdateConfig(config config.Config) {
	server.config = config
}

func (server *Server) HandleConn(conn *mc.Conn) {
	defer conn.Close()

	handshake, err := receiveHandshake(conn)
	if err != nil {
		logPacketReadError(conn, err)
		return
	}

	log.Printf("client handshaking with %s\n", handshake.ServerAddress)

	/*err = server.handlePacket(conn, packet)
	if err != nil {
		log.Printf("error: failed to handle packet: %v\n", err)
		return
	}*/

}

// tries to receive a valid handshake packet from the connection
func receiveHandshake(conn *mc.Conn) (mc.HandshakePacket, error) {
	packet, err := conn.ReadPacket()
	if err != nil {
		return mc.HandshakePacket{}, err
	}
	if packet.Id != 0x00 {
		return mc.HandshakePacket{}, fmt.Errorf("expected handshake, received packet 0x%02x", packet.Id)
	}

	handshakePacket, err := mc.DecodeHandshakePacket(packet)
	if err != nil {
		return mc.HandshakePacket{}, fmt.Errorf("failed to decode handshake: %s", err.Error())
	}
	if handshakePacket.ServerPort != conn.LocalPort() {
		return mc.HandshakePacket{}, fmt.Errorf("client connected to :%d, but specified :%d", conn.LocalPort(), handshakePacket.ServerPort)
	}

	return handshakePacket, nil
}
