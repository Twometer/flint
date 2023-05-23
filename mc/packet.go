package mc

import (
	"bytes"
)

type Packet struct {
	Id   int32
	Data *bytes.Buffer
}

func wrapPacket(id int32, data *bytes.Buffer) Packet {
	return Packet{Id: id, Data: data}
}

func createPacket(id int32) Packet {
	data := make([]byte, maxPacketSize)
	buffer := bytes.NewBuffer(data)
	buffer.Reset()
	return wrapPacket(id, buffer)
}

type HandshakePacket struct {
	ProtocolVersion int32
	ServerAddress   string
	ServerPort      uint16
	NextState       int32
}

func DecodeHandshakePacket(packet Packet) (HandshakePacket, error) {
	var err error
	handshakePacket := HandshakePacket{}

	handshakePacket.ProtocolVersion, err = ReadVarInt(packet.Data)
	if err != nil {
		return handshakePacket, err
	}

	handshakePacket.ServerAddress, err = ReadString(packet.Data)
	if err != nil {
		return handshakePacket, err
	}

	handshakePacket.ServerPort, err = ReadBigEndian[uint16](packet.Data)
	if err != nil {
		return handshakePacket, err
	}

	handshakePacket.NextState, err = ReadVarInt(packet.Data)
	if err != nil {
		return handshakePacket, err
	}

	return handshakePacket, nil
}

type StatusVersion struct {
	Name     string `json:"name"`
	Protocol int    `json:"protocol"`
}

type StatusPlayers struct {
	Max    int                  `json:"max"`
	Online int                  `json:"online"`
	Sample []StatusPlayerSample `json:"sample"`
}

type StatusPlayerSample struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type ChatComponent struct {
	Text string `json:"text"`
}

type StatusResponsePacket struct {
	Version            StatusVersion `json:"version"`
	Players            StatusPlayers `json:"players"`
	Description        ChatComponent `json:"description"`
	Favicon            string        `json:"favicon"`
	EnforcesSecureChat bool          `json:"enforcesSecureChat"`
}

func EncodeStatusResponsePacket(packet StatusResponsePacket) (Packet, error) {
	statusJson, err := jsonEncode(packet)
	if err != nil {
		return Packet{}, err
	}

	outPacket := createPacket(0x00)

	err = WriteString(outPacket.Data, statusJson)
	if err != nil {
		return Packet{}, err
	}

	return outPacket, nil
}

type PingPacket struct {
	Payload uint64
}

func DecodePingPacket(packet Packet) (PingPacket, error) {
	var err error
	pingPacket := PingPacket{}

	pingPacket.Payload, err = ReadBigEndian[uint64](packet.Data)
	if err != nil {
		return pingPacket, err
	}

	return pingPacket, nil
}

func EncodePingPacket(pingPacket PingPacket) (Packet, error) {
	packet := createPacket(0x01)

	err := WriteBigEndian[uint64](packet.Data, pingPacket.Payload)
	if err != nil {
		return Packet{}, err
	}

	return packet, nil
}

type DisconnectPacket struct {
	Reason ChatComponent
}

func EncodeDisconnectPacket(disconnectPacket DisconnectPacket) (Packet, error) {
	packet := createPacket(0x00)

	json, err := jsonEncode(disconnectPacket.Reason)
	if err != nil {
		return Packet{}, err
	}

	err = WriteString(packet.Data, json)
	if err != nil {
		return Packet{}, err
	}

	return packet, nil
}
