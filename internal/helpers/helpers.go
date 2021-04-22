package helpers

// BitshiftTo8Bit shifts a value over until it is in the range 0-255.
//
// It does not convert to an 8-bit type.
func BitshiftTo8Bit(n uint32) uint32 {
	for ; n > 0xFF; n = n >> 8 {
	}
	return n
}
