//go:build windows

package startup

import (
	"log"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32                  = windows.NewLazySystemDLL("kernel32.dll")
	procFreeConsole           = kernel32.NewProc("FreeConsole")
	procGetConsoleProcessList = kernel32.NewProc("GetConsoleProcessList")
)

// DetachConsole closes the console window Windows attached to this process,
// but only when nothing else shares it, so a logon start flashes none.
func DetachConsole() {
	if !ownsConsoleAlone() {
		return
	}
	silenceStdio()
	_, _, _ = procFreeConsole.Call()
}

// ownsConsoleAlone reports whether a console is attached and nothing else is
// sharing it, which means Windows created it for this launch alone.
func ownsConsoleAlone() bool {
	var pids [2]uint32
	n, _, _ := procGetConsoleProcessList.Call(uintptr(unsafe.Pointer(&pids[0])), uintptr(len(pids)))
	return n == 1
}

// silenceStdio points the standard handles at NUL, so writes issued after
// FreeConsole go nowhere instead of at a closed handle.
func silenceStdio() {
	nul, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return
	}
	h := windows.Handle(nul.Fd())
	for _, std := range []uint32{windows.STD_INPUT_HANDLE, windows.STD_OUTPUT_HANDLE, windows.STD_ERROR_HANDLE} {
		_ = windows.SetStdHandle(std, h)
	}
	os.Stdin, os.Stdout, os.Stderr = nul, nul, nul
	log.SetOutput(nul)
}
