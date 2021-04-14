package scream

// Convert 64 bit NTP timestamp to 32 bit NTP timestamp where
// the 16 most significant bits are seconds and
// the 16 least significant bits are fraction
func ntpToQ16(ntpTime uint64) uint32 {
	return uint32((ntpTime >> 16) & 0xFFFFFFFF)
}
