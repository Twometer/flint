package proxy

import (
	"flint/mc"
	"io"
)

type proxyConnectionHandler struct {
	destination string
	handshake   mc.HandshakePacket
}

func newProxyHandler(destination string, handshake mc.HandshakePacket) *proxyConnectionHandler {
	return &proxyConnectionHandler{destination: destination, handshake: handshake}
}

func (p *proxyConnectionHandler) handle(clientConn *mc.Conn) {
	defer clientConn.Close()

	serverConn, err := mc.Dial(p.destination)
	if err != nil {
		return
	}

	packet, err := mc.EncodeHandshakePacket(p.handshake)
	if err != nil {
		return
	}

	err = serverConn.WritePacket(packet)
	if err != nil {
		return
	}

	serverConn.DisableDeadlines()
	clientConn.DisableDeadlines()

	go linkConnections(clientConn.NetConn, serverConn.NetConn)
	linkConnections(serverConn.NetConn, clientConn.NetConn)
}

func linkConnections(dst io.WriteCloser, src io.ReadCloser) {
	_, _ = io.Copy(dst, src)
	_ = dst.Close()
	_ = src.Close()
}
