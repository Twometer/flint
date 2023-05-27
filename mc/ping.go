package mc

import (
	"golang.org/x/text/encoding/unicode"
	"net"
	"strconv"
	"strings"
	"time"
)

type ServerStatus struct {
	Online          bool
	ProtocolVersion int
	ServerVersion   string
	Motd            string
	CurPlayers      int
	MaxPlayers      int
}

var ServerOffline = ServerStatus{Online: false}

const pongHeaderSize = 9

func PingServer(address string) ServerStatus {
	dialer := net.Dialer{Timeout: 5 * time.Second}
	client, err := dialer.Dial("tcp", address)
	if err != nil {
		return ServerOffline
	}

	err = client.SetDeadline(time.Now().Add(time.Second * 2))
	if err != nil {
		return ServerOffline
	}

	n, err := client.Write([]byte{0xfe, 0x01})
	if err != nil || n != 2 {
		return ServerOffline
	}

	recvBuf := make([]byte, 1024)
	n, err = client.Read(recvBuf)
	if err != nil || n == 0 {
		return ServerOffline
	}

	pongData := recvBuf[:n]
	if len(pongData) < pongHeaderSize || pongData[0] != 0xff || pongData[4] != 0xa7 || pongData[6] != 0x31 {
		return ServerOffline
	}

	decoder := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder()
	decodedPong, err := decoder.Bytes(pongData[pongHeaderSize:])
	if err != nil {
		return ServerOffline
	}

	pongFields := strings.Split(string(decodedPong), "\x00")
	if len(pongFields) < 5 {
		return ServerOffline
	}

	serverState := ServerStatus{Online: true}
	serverState.ProtocolVersion, err = strconv.Atoi(pongFields[0])
	if err != nil {
		return ServerOffline
	}
	serverState.ServerVersion = pongFields[1]
	serverState.Motd = pongFields[2]
	serverState.CurPlayers, err = strconv.Atoi(pongFields[3])
	if err != nil {
		return ServerOffline
	}
	serverState.MaxPlayers, err = strconv.Atoi(pongFields[4])
	if err != nil {
		return ServerOffline
	}

	return serverState
}
