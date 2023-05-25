package proxy

import "flint/mc"

type connectionHandler interface {
	handle(conn *mc.Conn)
}
