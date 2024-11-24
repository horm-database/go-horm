package codec

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/horm-database/common/codec"
)

var (
	readerSize = 4 * 1024 // reader buffer size
)

// NewReader returns reader with the default buffer size.
func NewReader(r io.Reader) io.Reader {
	return bufio.NewReaderSize(r, readerSize)
}

// Framer frame reader
type Framer struct {
	reader io.Reader
	header [codec.FrameHeadLen]byte
}

// NewFramer create framer
func NewFramer(reader io.Reader) *Framer {
	return &Framer{
		reader: reader,
	}
}

// ReadFrame read response buffer from frame
func (f *Framer) ReadFrame() (respBuf []byte, err error) {
	num, err := io.ReadFull(f.reader, f.header[:])
	if err != nil {
		return nil, err
	}
	if num != codec.FrameHeadLen {
		return nil, fmt.Errorf("framer: read frame header num %d != %d, invalid", num, codec.FrameHeadLen)
	}
	totalLen := binary.BigEndian.Uint32(f.header[4:8])
	if totalLen < uint32(codec.FrameHeadLen) {
		return nil, fmt.Errorf(
			"framer: read frame header total len %d < %d, invalid", totalLen, uint32(codec.FrameHeadLen))
	}

	if totalLen > uint32(codec.MaxFrameSize) {
		return nil, fmt.Errorf(
			"framer: read frame header total len %d > %d, too large", totalLen, uint32(codec.MaxFrameSize))
	}

	respBuf = make([]byte, totalLen)
	copy(respBuf, f.header[:])

	num, err = io.ReadFull(f.reader, respBuf[codec.FrameHeadLen:totalLen])
	if err != nil {
		return nil, err
	}

	if num != int(totalLen-uint32(codec.FrameHeadLen)) {
		return nil, fmt.Errorf("framer: read frame total num %d != %d, invalid",
			num, int(totalLen-uint32(codec.FrameHeadLen)))
	}

	return respBuf, nil
}
