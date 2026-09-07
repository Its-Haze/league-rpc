package main

import "testing"

func TestClampDimension(t *testing.T) {
	tests := []struct {
		name                      string
		preferred, min, available int
		want                      int
	}{
		{"fits on a large screen", 1200, 940, 1920, 1200},
		{"shrinks to the work area", 1200, 940, 1280, 1177},
		{"never goes below the minimum", 1200, 940, 800, 940},
		{"preferred just under the limit", 940, 940, 1024, 940},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clampDimension(tt.preferred, tt.min, tt.available); got != tt.want {
				t.Errorf("clampDimension(%d, %d, %d) = %d, want %d", tt.preferred, tt.min, tt.available, got, tt.want)
			}
		})
	}
}

func TestDefaultWindowSizeIsAtLeastTheMinimum(t *testing.T) {
	width, height := defaultWindowSize()
	if width < minWindowWidth || height < minWindowHeight {
		t.Errorf("defaultWindowSize() = %dx%d, smaller than the %dx%d minimum", width, height, minWindowWidth, minWindowHeight)
	}
}
