package strings

import (
	"bytes"
	"unsafe"
)

type Buffer struct {
	bytes.Buffer
}

// String returns the accumulated string.
// Like [strings.Builder.String], it does not allocate a new string.
func (b *Buffer) String() string {
	return unsafe.String(unsafe.SliceData(b.Bytes()), b.Len())
}
