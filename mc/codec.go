package mc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const maxVarIntBits = 32
const maxStringSize = 255

func ReadVarInt(reader io.ByteReader) (int32, error) {
	result := int32(0)
	position := 0

	for {
		data, err := reader.ReadByte()
		if err != nil {
			return 0, err
		}

		result |= int32(data&0x7f) << position
		position += 7

		if data&0x80 == 0 {
			break
		}
		if position >= maxVarIntBits {
			return 0, fmt.Errorf("varint is too big")
		}
	}
	return result, nil
}

func WriteVarInt(writer io.ByteWriter, value int32) error {
	for {
		if value & ^0x7f == 0 {
			return writer.WriteByte(byte(value))
		}

		err := writer.WriteByte(byte((value & 0x7f) | 0x80))
		if err != nil {
			return err
		}

		value >>= 7
	}
}

func GetVarIntSize(value int32) int {
	size := 0
	for {
		size += 1
		if value & ^0x7f == 0 {
			break
		}
		value >>= 7
	}
	return size
}

func ReadString(reader *bytes.Buffer) (string, error) {
	stringLen, err := ReadVarInt(reader)
	if err != nil {
		return "", err
	}
	if stringLen > maxStringSize {
		return "", fmt.Errorf("string is too big (%d > %d)", stringLen, maxStringSize)
	}

	stringData := make([]byte, stringLen)
	n, err := reader.Read(stringData)
	if err != nil {
		return "", err
	}
	if int32(n) != stringLen {
		return "", io.EOF
	}

	return string(stringData), nil
}

func WriteString(writer *bytes.Buffer, value string) error {
	err := WriteVarInt(writer, int32(len(value)))
	if err != nil {
		return err
	}

	stringData := []byte(value)
	n, err := writer.Write(stringData)
	if err != nil {
		return err
	}
	if n != len(stringData) {
		return fmt.Errorf("failed to write string")
	}

	return nil
}

func ReadBigEndian[T any](reader *bytes.Buffer) (T, error) {
	var data T
	err := binary.Read(reader, binary.BigEndian, &data)
	if err != nil {
		return data, err
	}
	return data, nil
}

func WriteBigEndian[T any](writer io.Writer, value T) error {
	return binary.Write(writer, binary.BigEndian, value)
}
