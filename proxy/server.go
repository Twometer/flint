package proxy

import (
	"flint/config"
	"flint/mc"
	"fmt"
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

	createConnectionHandler(handshake).handle(conn)
}

func createConnectionHandler(handshake mc.HandshakePacket) connectionHandler {
	// TODO: do actual handler
	message := fmt.Sprintf("§4No Minecraft server at §6%s", handshake.ServerAddress)
	return &statusConnectionHandler{message: message, state: handshake.NextState}
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
