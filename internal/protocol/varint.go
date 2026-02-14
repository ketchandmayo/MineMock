package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

func ReadVarInt(r io.Reader) (int32, error) {
	var numRead int32
	var result int32
	var read [1]byte

	for {
		if _, err := io.ReadFull(r, read[:]); err != nil {
			return 0, err
		}

		value := int32(read[0] & 0x7F)
		result |= value << (7 * numRead)

		numRead++
		if numRead > 5 {
			return 0, fmt.Errorf("varint is too big")
		}

		if read[0]&0x80 == 0 {
			break
		}
	}

	return result, nil
}

func EncodeVarInt(value int32) []byte {
	buf := make([]byte, binary.MaxVarintLen32)
	n := 0

	for {
		if (value & ^0x7F) == 0 {
			buf[n] = byte(value)
			n++
			break
		}

		buf[n] = byte(value&0x7F | 0x80)
		value >>= 7
		n++
	}

	return buf[:n]
}
