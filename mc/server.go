package mc

import (
	"fmt"
	"net"
)

type Server struct {
	listener net.Listener
}

func Listen(ip string, port uint16) (Server, error) {
	address := fmt.Sprintf("%s:%d", ip, port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return Server{}, err
	}

	return Server{
		listener: listener,
	}, nil
}

func (server Server) Accept() (Conn, error) {
	conn, err := server.listener.Accept()
	if err != nil {
		return Conn{}, err
	}

	return NewConn(conn), nil
}

func (server Server) Addr() net.Addr {
	return server.listener.Addr()
}

func (server Server) Close() {
	_ = server.listener.Close()
}
