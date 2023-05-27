package mc

import (
	"bytes"
	"fmt"
	"net"
	"time"
)

const maxPacketSize = 1024
const timeoutSeconds = 30

type Conn struct {
	byteBuf []byte
	NetConn net.Conn
}

func Dial(address string) (Conn, error) {
	netConn, err := net.Dial("tcp", address)
	if err != nil {
		return Conn{}, err
	}
	return NewConn(netConn), nil
}

func NewConn(netConn net.Conn) Conn {
	return Conn{
		byteBuf: make([]byte, 1),
		NetConn: netConn,
	}
}

func (conn *Conn) RemoteAddr() net.Addr {
	return conn.NetConn.RemoteAddr()
}

func (conn *Conn) LocalPort() uint16 {
	return uint16(conn.NetConn.LocalAddr().(*net.TCPAddr).Port)
}

func (conn *Conn) Close() {
	_ = conn.NetConn.Close()
}

func (conn *Conn) ReadByte() (uint8, error) {
	err := conn.ReadData(conn.byteBuf)
	if err != nil {
		return 0, err
	}

	return conn.byteBuf[0], nil
}

func (conn *Conn) WriteByte(data uint8) error {
	return conn.WriteData([]byte{data})
}

func (conn *Conn) WriteData(buffer []byte) error {
	if len(buffer) == 0 {
		return nil
	}

	err := conn.NetConn.SetWriteDeadline(time.Now().Add(timeoutSeconds * time.Second))
	if err != nil {
		return err
	}

	written := 0
	for written < len(buffer) {
		n, err := conn.NetConn.Write(buffer[written:])
		if err != nil {
			return err
		}

		written += n
	}
	return nil
}

func (conn *Conn) ReadData(buffer []byte) error {
	if len(buffer) == 0 {
		return nil
	}

	err := conn.NetConn.SetReadDeadline(time.Now().Add(timeoutSeconds * time.Second))
	if err != nil {
		return err
	}

	read := 0
	for read < len(buffer) {
		n, err := conn.NetConn.Read(buffer[read:])
		if err != nil {
			return err
		}

		read += n
	}
	return nil
}

func (conn *Conn) ReadPacket() (Packet, error) {
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

func (conn *Conn) WritePacket(packet Packet) error {
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

func (conn *Conn) DisableDeadlines() {
	_ = conn.NetConn.SetDeadline(time.Time{})
}
