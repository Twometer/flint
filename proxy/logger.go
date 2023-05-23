package proxy

import (
	"flint/mc"
	"io"
	"log"
)

func logPacketReadError(conn *mc.Conn, err error) {
	if err == nil || err == io.EOF {
		return
	}

	log.Printf("failed to read packet from %s: %v\n", conn.RemoteAddr().String(), err)
}
