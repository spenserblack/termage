package gif

import "testing"

// TestInfiniteLoop checks that an infiniteLoop looper will never return
// an error.
func TestInfiniteLoop(t *testing.T) {
	var l looper = infiniteLoop{}

	// Let's just run it a bunch of times
	for i := 0; i < 1_000_000; i++ {
		if err := l.nextLoop(); err != nil {
			t.Fatalf(`err = %v, want nil`, err)
		}
	}
}

// TestNoLoop checks that that the very first loop returns an
// animation completion error.
func TestNoLoop(t *testing.T) {
	var l looper = noLoop{}

	if err := l.nextLoop(); err != ErrAnimationComplete {
		t.Fatalf(`err = %v, want %v`, err, ErrAnimationComplete)
	}
}

// TestCountLoop checks that that it loops a certain number of times
// before returning an animation completion error.
func TestCountLoop(t *testing.T) {
	var l looper = &countLoop{5}

	for i := 0; i <= 4; i++ {
		if err := l.nextLoop(); err != nil {
			t.Fatalf(`err = %v, want nil`, err)
		}
	}

	if err := l.nextLoop(); err != ErrAnimationComplete {
		t.Fatalf(`err = %v, want %v`, err, ErrAnimationComplete)
	}
}
