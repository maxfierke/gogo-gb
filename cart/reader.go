package cart

import (
	"bufio"
	"errors"
	"io"
)

var (
	// ErrChecksum is returned when reading cartridge data that has an invalid checksum.
	ErrChecksum = errors.New("cart: invalid checksum")

	// ErrHeader is returned when reading cartridge data that has an invalid header.
	ErrHeader = errors.New("cart: invalid header")
)

type byteReader interface {
	io.Reader
	io.ByteReader
}

// An io.Reader that mostly cribs from compress/gzip's gunzip.Reader
type Reader struct {
	Header    Header
	r         byteReader
	err       error
	headerBuf [HEADER_END + 1]byte
}

// NewReader creates a new Reader reading the given reader.
// If r does not also implement io.ByteReader,
// the decoder may read more data than necessary from r.
//
// It is the caller's responsibility to call Close on the Reader when done.
//
// The Reader.Metadata fields will be valid in the Reader returned.
func NewReader(r io.Reader) (*Reader, error) {
	cartReader := new(Reader)
	if err := cartReader.Reset(r); err == ErrHeader {
		// Pass header checksum errors onto caller and let them handle appropriately
		return cartReader, err
	} else if err != nil {
		return nil, err
	}
	return cartReader, nil
}

// Reset discards the Reader cr's state and makes it equivalent to the
// result of its original state from NewReader, but reading from r instead.
// This permits reusing a Reader rather than allocating a new one.
func (cr *Reader) Reset(r io.Reader) error {
	*cr = Reader{}
	if rr, ok := r.(byteReader); ok {
		cr.r = rr
	} else {
		cr.r = bufio.NewReader(r)
	}
	cr.Header, cr.err = cr.readHeader()
	return cr.err
}

func (cr *Reader) readHeader() (hdr Header, err error) {
	if _, err = io.ReadFull(cr.r, cr.headerBuf[:]); err != nil {
		return hdr, err
	}

	hdr = NewHeader(cr.headerBuf[:])

	// Check actual checksum against expected. Computed according to
	// https://gbdev.io/pandocs/The_Cartridge_Header.html#014d--header-checksum
	// The BootROM does this, but so can we. Earlier.
	var hdrChksum byte
	for addr := titleOffset; addr <= maskRomVerOffset; addr++ {
		hdrChksum = hdrChksum - cr.headerBuf[addr] - 1
	}

	if hdrChksum != hdr.HeaderChecksum {
		return hdr, ErrHeader
	}

	return hdr, nil
}

// Read implements io.Reader, reading cartridge ROM bytes from its underlying Reader.
func (cr *Reader) Read(p []byte) (n int, err error) {
	return cr.r.Read(p)
}
