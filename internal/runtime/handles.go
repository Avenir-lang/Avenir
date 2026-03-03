package runtime

import (
	"encoding/binary"
	"fmt"
)

func encodeHandle(id uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, id)
	return buf
}

func decodeHandle(b []byte) (uint64, error) {
	if len(b) != 8 {
		return 0, fmt.Errorf("invalid handle")
	}
	return binary.LittleEndian.Uint64(b), nil
}
