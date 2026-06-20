package devices

import (
	"bytes"
	"context"
	"errors"
	"io"
)

var ErrChangingReaderWriter = errors.New("changing reader/writer")

type SerialCable interface {
	io.ByteReader
	io.ByteWriter
}

type NullSerialCable struct{}

func (sc *NullSerialCable) ReadByte() (byte, error) {
	return 0xFF, nil
}

func (sc *NullSerialCable) WriteByte(value byte) error {
	return nil
}

type HostSerialCable struct {
	rxChan   chan byte
	rxCancel context.CancelCauseFunc

	txChan   chan byte
	txCancel context.CancelCauseFunc
}

func NewHostSerialCable() *HostSerialCable {
	sc := &HostSerialCable{
		rxChan: make(chan byte),
		txChan: make(chan byte),
	}

	ctx := context.Background()
	sc.SetReader(ctx, bytes.NewReader([]byte{}))
	sc.SetWriter(ctx, io.Discard)

	return sc
}

func (sc *HostSerialCable) ReadByte() (byte, error) {
	select {
	case value := <-sc.rxChan:
		return value, nil
	default:
		return 0xFF, errors.New("read channel empty")
	}
}

func (sc *HostSerialCable) WriteByte(value byte) error {
	select {
	case sc.txChan <- value:
	default:
		return errors.New("write channel full")
	}

	return nil
}

func (sc *HostSerialCable) SetReader(ctx context.Context, reader io.Reader) {
	if sc.rxCancel != nil {
		sc.rxCancel(ErrChangingReaderWriter)
	}
	ctx, cancel := context.WithCancelCause(ctx)
	sc.rxCancel = cancel
	go sc.receivePoll(ctx, reader)
}

func (sc *HostSerialCable) SetWriter(ctx context.Context, writer io.Writer) {
	if sc.txCancel != nil {
		sc.txCancel(ErrChangingReaderWriter)
	}
	ctx, cancel := context.WithCancelCause(ctx)
	sc.txCancel = cancel
	go sc.writePoll(ctx, writer)
}

func (sc *HostSerialCable) receivePoll(ctx context.Context, reader io.Reader) {
	var readBuf [1]byte

	for {
		if ctx.Err() != nil {
			return
		}
		if n, err := reader.Read(readBuf[:]); err != nil && err != io.EOF {
			return
		} else if n > 0 {
			select {
			case sc.rxChan <- readBuf[0]:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (sc *HostSerialCable) writePoll(ctx context.Context, writer io.Writer) {
	for {
		if ctx.Err() != nil {
			return
		}

		select {
		case value := <-sc.txChan:
			_, _ = writer.Write([]byte{value})
		case <-ctx.Done():
			return
		}
	}
}
