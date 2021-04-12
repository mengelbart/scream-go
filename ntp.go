package scream

func ntpToQ16(ntpTime uint64) uint32 {
	return uint32((ntpTime >> 16) & 0xFFFFFFFF)
}
