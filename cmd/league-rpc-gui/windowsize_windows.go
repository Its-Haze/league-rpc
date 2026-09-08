//go:build windows

package main

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                    = windows.NewLazySystemDLL("user32.dll")
	procSystemParametersInfoW = user32.NewProc("SystemParametersInfoW")
	procGetDpiForSystem       = user32.NewProc("GetDpiForSystem")
	procGetDC                 = user32.NewProc("GetDC")
	procReleaseDC             = user32.NewProc("ReleaseDC")
	gdi32                     = windows.NewLazySystemDLL("gdi32.dll")
	procGetDeviceCaps         = gdi32.NewProc("GetDeviceCaps")
)

const (
	spiGetWorkArea = 0x0030
	logPixelsX     = 88
	defaultDPI     = 96
)

type rect struct {
	left, top, right, bottom int32
}

// workAreaDIP returns the primary monitor's usable area (the desktop minus
// the taskbar) in DIPs, the same unit Wails window sizes are given in.
func workAreaDIP() (int, int, bool) {
	var area rect
	ok, _, _ := procSystemParametersInfoW.Call(spiGetWorkArea, 0, uintptr(unsafe.Pointer(&area)), 0)
	if ok == 0 {
		return 0, 0, false
	}
	dpi := systemDPI()
	width := int((area.right - area.left)) * defaultDPI / dpi
	height := int((area.bottom - area.top)) * defaultDPI / dpi
	if width <= 0 || height <= 0 {
		return 0, 0, false
	}
	return width, height, true
}

// systemDPI reads the system-wide scaling, falling back to the device caps
// on Windows builds older than 1607, where GetDpiForSystem is missing.
func systemDPI() int {
	if err := procGetDpiForSystem.Find(); err == nil {
		if dpi, _, _ := procGetDpiForSystem.Call(); dpi > 0 {
			return int(dpi)
		}
	}
	screen, _, _ := procGetDC.Call(0)
	if screen == 0 {
		return defaultDPI
	}
	defer procReleaseDC.Call(0, screen)
	dpi, _, _ := procGetDeviceCaps.Call(screen, logPixelsX)
	if dpi == 0 {
		return defaultDPI
	}
	return int(dpi)
}
