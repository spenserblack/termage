package gif

// Looper is a helper to determine if looping should continue.
type looper interface {
	// NextLoop should be called at the end of each loop,
	// to determine if it should continue.
	nextLoop() error
}

// InfiniteLoop will never end animation.
type infiniteLoop struct{}

// NoLoop will always end animation.
type noLoop struct{}

// countLoop will end animation once the set number of loops completes.
type countLoop struct {
	count int
}

func (l infiniteLoop) nextLoop() error {
	return nil
}

func (l noLoop) nextLoop() error {
	return ErrAnimationComplete
}

func (l *countLoop) nextLoop() error {
	l.count--
	if l.count < 0 {
		return ErrAnimationComplete
	}
	return nil
}
