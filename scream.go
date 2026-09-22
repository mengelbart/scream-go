// Copyright (c) Mathis Engelbart. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scream

/*
#cgo CXXFLAGS: -Wno-overflow -Wno-write-strings
*/
import "C"
import "time"

const (
	// ntpShortRange is the period of the NTP short format in seconds.
	ntpShortRange = 1 << 16

	ntpShortHalfRange = ntpShortRange / 2
)

func toNTP32(t time.Time) uint32 {
	return uint32(toNTP(t) >> 16)
}

func toNTP(t time.Time) uint64 {
	s := (float64(t.UnixNano()) / 1000000000) + 2208988800
	integerPart := uint32(s)
	fractionalPart := uint32((s - float64(integerPart)) * 0xFFFFFFFF)
	return uint64(integerPart)<<32 | uint64(fractionalPart)
}
