package mc

import (
	"bytes"
	"fmt"
	"net"
	"time"
)

const maxPacketSize = 1024
const timeoutSeconds = 30

const (
	StateInitial = iota
	StateStatus
	StateLogin
)

type Conn struct {
	State   int32
	byteBuf []byte
	netConn net.Conn
}

func NewConn(netConn net.Conn) Conn {
	return Conn{
		State:   StateInitial,
		byteBuf: make([]byte, 1),
		netConn: netConn,
	}
}

func (conn Conn) RemoteAddr() net.Addr {
	return conn.netConn.RemoteAddr()
}

func (conn Conn) LocalPort() uint16 {
	return uint16(conn.netConn.LocalAddr().(*net.TCPAddr).Port)
}

func (conn Conn) Close() {
	_ = conn.netConn.Close()
}

func (conn Conn) ReadByte() (uint8, error) {
	err := conn.ReadData(conn.byteBuf)
	if err != nil {
		return 0, err
	}

	return conn.byteBuf[0], nil
}

func (conn Conn) WriteByte(data uint8) error {
	return conn.WriteData([]byte{data})
}

func (conn Conn) WriteData(buffer []byte) error {
	if len(buffer) == 0 {
		return nil
	}

	err := conn.netConn.SetWriteDeadline(time.Now().Add(timeoutSeconds * time.Second))
	if err != nil {
		return err
	}

	written := 0
	for written < len(buffer) {
		n, err := conn.netConn.Write(buffer[written:])
		if err != nil {
			return err
		}

		written += n
	}
	return nil
}

func (conn Conn) ReadData(buffer []byte) error {
	if len(buffer) == 0 {
		return nil
	}

	err := conn.netConn.SetReadDeadline(time.Now().Add(timeoutSeconds * time.Second))
	if err != nil {
		return err
	}

	read := 0
	for read < len(buffer) {
		n, err := conn.netConn.Read(buffer[read:])
		if err != nil {
			return err
		}

		read += n
	}
	return nil
}

func (conn Conn) ReadPacket() (Packet, error) {
	packetSize, err := ReadVarInt(conn)
	if err != nil {
		return Packet{}, err
	}
	if packetSize > maxPacketSize {
		return Packet{}, fmt.Errorf("refusing to decode packet of size %d, maximum is %d", packetSize, maxPacketSize)
	}

	packetData := make([]byte, packetSize)
	err = conn.ReadData(packetData)
	if err != nil {
		return Packet{}, err
	}

	packetBuffer := bytes.NewBuffer(packetData)

	packetId, err := ReadVarInt(packetBuffer)
	if err != nil {
		return Packet{}, err
	}

	return wrapPacket(packetId, packetBuffer), nil
}

func (conn Conn) WritePacket(packet Packet) error {
	packetData := packet.Data.Bytes()
	packetLen := int32(len(packetData) + GetVarIntSize(packet.Id))

	err := WriteVarInt(conn, packetLen)
	if err != nil {
		return err
	}

	err = WriteVarInt(conn, packet.Id)
	if err != nil {
		return err
	}

	err = conn.WriteData(packetData)
	if err != nil {
		return err
	}

	return nil
}
