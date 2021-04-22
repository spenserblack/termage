package helpers

import "testing"

func TestBitShift(t *testing.T) {
	shifted := BitshiftTo8Bit(1 << 9)
	if shifted != 1<<1 {
		t.Errorf(`BitshiftTo8Bit(1 << 9) = %v, want %v`, shifted, 1<<1)
	}
	shifted = BitshiftTo8Bit(0xABCD)
	if shifted != 0xAB {
		t.Errorf(`BitshiftTo8Bit(0xABCD) = %v, want %v`, shifted, 0xAB)
	}
}
