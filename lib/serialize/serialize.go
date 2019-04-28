package serialize

import (
	"io"
	"math"
)

func String(w io.Writer, s string) {
	write(w, []byte(s))
}

func Bool(w io.Writer, b bool) {
	if b {
		write(w, []byte{1})
	} else {
		write(w, []byte{0})
	}
}

func UInt32(w io.Writer, i uint32) {
	write(w, []byte{byte(i), byte(i >> 8), byte(i >> 16), byte(i >> 24)})
}

func UInt64(w io.Writer, i uint64) {
	write(w, []byte{byte(i), byte(i >> 8), byte(i >> 16), byte(i >> 24), byte(i >> 32), byte(i >> 40), byte(i >> 48), byte(i >> 56)})
}

func Float64(w io.Writer, f float64) {
	UInt64(w, math.Float64bits(f))
}

func StringSlice(w io.Writer, slice []string) {
	for _, s := range slice {
		String(w, s)
	}
}

func write(w io.Writer, bytes []byte) {
	if _, err := w.Write(bytes); err != nil {
		panic(err)
	}
}
