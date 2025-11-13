package devices

import (
	"bytes"
	"io"
)

type SerialCable interface {
	ReadByte() (byte, error)
	WriteByte(value byte) error
}

type NullSerialCable struct{}

func (sc *NullSerialCable) ReadByte() (byte, error) {
	return 0xFF, nil
}

func (sc *NullSerialCable) WriteByte(value byte) error {
	return nil
}

type HostSerialCable struct {
	reader io.Reader
	writer io.Writer
}

func NewHostSerialCable() *HostSerialCable {
	return &HostSerialCable{
		reader: bytes.NewReader([]byte{}),
		writer: io.Discard,
	}
}

func (sc *HostSerialCable) ReadByte() (byte, error) {
	readBuf := []byte{0x00}

	if _, err := sc.reader.Read(readBuf); err != nil {
		return 0xFF, err
	}

	return readBuf[0], nil
}

func (sc *HostSerialCable) WriteByte(value byte) error {
	_, err := sc.writer.Write([]byte{value})

	return err
}

func (sc *HostSerialCable) SetReader(reader io.Reader) {
	sc.reader = reader
}

func (sc *HostSerialCable) SetWriter(writer io.Writer) {
	sc.writer = writer
}
