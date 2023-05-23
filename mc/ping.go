package mc

import (
	"golang.org/x/text/encoding/unicode"
	"net"
	"strconv"
	"strings"
)

type ServerStatus struct {
	Online          bool
	ProtocolVersion int
	ServerVersion   string
	Motd            string
	CurPlayers      int
	MaxPlayers      int
}

var serverOffline = ServerStatus{Online: false}

const pongHeaderSize = 9

func PingServer(address string) ServerStatus {
	client, err := net.Dial("tcp", address)
	if err != nil {
		return serverOffline
	}

	n, err := client.Write([]byte{0xfe, 0x01})
	if err != nil || n != 2 {
		return serverOffline
	}

	recvBuf := make([]byte, 1024)
	n, err = client.Read(recvBuf)
	if err != nil || n == 0 {
		return serverOffline
	}

	pongData := recvBuf[:n]
	if len(pongData) < pongHeaderSize || pongData[0] != 0xff || pongData[4] != 0xa7 || pongData[6] != 0x31 {
		return serverOffline
	}

	decoder := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder()
	decodedPong, err := decoder.Bytes(pongData[pongHeaderSize:])
	if err != nil {
		return serverOffline
	}

	pongFields := strings.Split(string(decodedPong), "\x00")
	if len(pongFields) < 5 {
		return serverOffline
	}

	serverState := ServerStatus{Online: true}
	serverState.ProtocolVersion, err = strconv.Atoi(pongFields[0])
	if err != nil {
		return serverOffline
	}
	serverState.ServerVersion = pongFields[1]
	serverState.Motd = pongFields[2]
	serverState.CurPlayers, err = strconv.Atoi(pongFields[3])
	if err != nil {
		return serverOffline
	}
	serverState.MaxPlayers, err = strconv.Atoi(pongFields[4])
	if err != nil {
		return serverOffline
	}

	return serverState
}
