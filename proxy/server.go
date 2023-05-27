package proxy

import (
	"flint/config"
	"flint/mc"
	"fmt"
)

type Server struct {
	config    config.Config
	upstreams *upstreamTracker
}

func NewServer() Server {
	return Server{
		upstreams: newUpstreamTracker(),
	}
}

func (server *Server) UpdateConfig(config config.Config) {
	server.config = config
	server.upstreams.setUpstreams(config.Upstreams)
}

func (server *Server) HandleConn(conn *mc.Conn) {
	defer conn.Close()
	handshake, err := receiveHandshake(conn)
	if err != nil {
		logPacketReadError(conn, err)
		return
	}

	server.createConnectionHandler(handshake).handle(conn)
}

func (server *Server) createConnectionHandler(handshake mc.HandshakePacket) connectionHandler {
	upstream, found := server.upstreams.findUpstream(handshake.ServerAddress)
	msgParams := interpolationParams{
		connectHost: handshake.ServerAddress,
	}

	if !found {
		message := interpolateMessage(server.config.Messages.ServerNotFound, msgParams)
		return newStatusHandler(message, false, handshake)
	}

	msgParams.serverAddress = upstream.config.Address
	msgParams.serverName = upstream.config.Name

	if upstream.config.Maintenance {
		message := interpolateMessage(server.config.Messages.Maintenance, msgParams)
		return newStatusHandler(message, true, handshake)
	}

	if !upstream.status.Online {
		message := interpolateMessage(server.config.Messages.ServerDown, msgParams)
		return newStatusHandler(message, false, handshake)
	}

	return newProxyHandler(upstream.config.Address, handshake)
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
