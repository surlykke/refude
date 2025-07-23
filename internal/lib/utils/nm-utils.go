// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package utils

import "encoding/binary"

func PrependWithLength(buf []byte) []byte {
	var size = uint32(len(buf))
	var prepended = make([]byte, size+4)
	binary.Encode(prepended[0:4], binary.NativeEndian, &size)
	copy(prepended[4:], buf)
	return prepended
}

func StripLength(buf []byte) []byte {
	if len(buf) < 4 {
		return nil
	}
	var size uint32
	binary.Decode(buf[0:4], binary.NativeEndian, &size)
	return buf[4 : size+4]
}
