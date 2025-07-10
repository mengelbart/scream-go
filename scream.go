package scream

/*
#cgo CPPFLAGS: -Wno-overflow -Wno-write-strings -I${SRCDIR}/include
#cgo LDFLAGS: ${SRCDIR}/libscream.a
*/
import "C"
import "time"

func toNTP32(t time.Time) uint32 {
	return uint32(toNTP(t) >> 16)
}

func toNTP(t time.Time) uint64 {
	s := (float64(t.UnixNano()) / 1000000000) + 2208988800
	integerPart := uint32(s)
	fractionalPart := uint32((s - float64(integerPart)) * 0xFFFFFFFF)
	return uint64(integerPart)<<32 | uint64(fractionalPart)
}
