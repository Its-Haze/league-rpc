package main

// Sized to the tallest, widest screen (Display, with its eight template tabs
// and the preview under them) so nothing needs a manual resize to be read.
const (
	preferredWindowWidth  = 1280
	preferredWindowHeight = 950
	minWindowWidth        = 940
	minWindowHeight       = 620
)

// Fraction of the work area the window may occupy, leaving the desktop
// visible around it rather than filling the screen edge to edge.
const workAreaFraction = 0.92

// defaultWindowSize returns the startup size in DIPs, shrunk to fit the
// primary monitor when the preferred size would run off a smaller screen.
func defaultWindowSize() (int, int) {
	availableWidth, availableHeight, ok := workAreaDIP()
	if !ok {
		return preferredWindowWidth, preferredWindowHeight
	}
	width := clampDimension(preferredWindowWidth, minWindowWidth, availableWidth)
	height := clampDimension(preferredWindowHeight, minWindowHeight, availableHeight)
	return width, height
}

// clampDimension caps preferred at a share of available, but never shrinks
// below min: a screen too small for the minimum gets a window it can scroll.
func clampDimension(preferred, min, available int) int {
	limit := int(float64(available) * workAreaFraction)
	if limit < min {
		return min
	}
	if preferred > limit {
		return limit
	}
	return preferred
}
