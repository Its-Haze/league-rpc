package main

import (
	"slices"
	"testing"

	"github.com/its-haze/league-rpc/internal/config"
)

type fakeWindow struct {
	shown   int
	hidden  int
	focused int
}

func (f *fakeWindow) Show()  { f.shown++ }
func (f *fakeWindow) Hide()  { f.hidden++ }
func (f *fakeWindow) Focus() { f.focused++ }

type fakePause struct{ paused bool }

func (f *fakePause) SetPaused(p bool) { f.paused = p }
func (f *fakePause) IsPaused() bool   { return f.paused }

// newTestTray builds a controller wired to counters, with the close action
// fixed to action.
func newTestTray(win windowController, action string) (*trayController, *int, *int) {
	asks, quits := 0, 0
	c := newTrayController(win, &fakePause{})
	c.closeAction = func() string { return action }
	c.askClose = func() { asks++ }
	c.quit = func() { quits++ }
	return c, &asks, &quits
}

func TestTrayController_CloseAsksWithoutHiding(t *testing.T) {
	win := &fakeWindow{}
	c, asks, quits := newTestTray(win, config.CloseAsk)

	c.handleClose()
	c.handleClose()

	if *asks != 2 {
		t.Fatalf("every close under %q should ask, asked %d times", config.CloseAsk, *asks)
	}
	if win.hidden != 0 {
		t.Fatal("the window must stay up while the dialog is asking")
	}
	if *quits != 0 {
		t.Fatal("asking must never quit on its own")
	}
}

func TestTrayController_CloseHidesWhenRememberedAsTray(t *testing.T) {
	win := &fakeWindow{}
	c, asks, quits := newTestTray(win, config.CloseTray)

	c.handleClose()

	if win.hidden != 1 || *asks != 0 || *quits != 0 {
		t.Fatalf("expected a silent hide: hidden=%d asks=%d quits=%d", win.hidden, *asks, *quits)
	}
}

func TestTrayController_CloseQuitsWhenRememberedAsQuit(t *testing.T) {
	win := &fakeWindow{}
	c, asks, quits := newTestTray(win, config.CloseQuit)

	c.handleClose()

	if *quits != 1 || win.hidden != 0 || *asks != 0 {
		t.Fatalf("expected a silent quit: quits=%d hidden=%d asks=%d", *quits, win.hidden, *asks)
	}
}

// A close arriving before the frontend is wired must not strand the user with
// a window that refuses to go away.
func TestTrayController_CloseHidesWhenNothingCanAsk(t *testing.T) {
	win := &fakeWindow{}
	c := newTrayController(win, &fakePause{})

	c.handleClose()

	if win.hidden != 1 {
		t.Fatalf("expected a fallback hide, got %d hides", win.hidden)
	}
}

func TestTrayController_ResolveCloseAppliesTheAnswer(t *testing.T) {
	win := &fakeWindow{}
	c, _, quits := newTestTray(win, config.CloseAsk)

	c.resolveClose(config.CloseTray)
	if win.hidden != 1 || *quits != 0 {
		t.Fatalf("tray answer should hide only: hidden=%d quits=%d", win.hidden, *quits)
	}

	c.resolveClose(config.CloseQuit)
	if *quits != 1 || win.hidden != 1 {
		t.Fatalf("quit answer should quit only: quits=%d hidden=%d", *quits, win.hidden)
	}
}

func TestTrayController_ShowWindowShowsAndFocuses(t *testing.T) {
	win := &fakeWindow{}
	c := newTrayController(win, &fakePause{})

	c.showWindow()

	if win.shown != 1 || win.focused != 1 {
		t.Fatalf("showWindow should show and focus once: shown=%d focused=%d", win.shown, win.focused)
	}
}

func TestTrayController_TogglePauseFlipsFlag(t *testing.T) {
	pause := &fakePause{}
	c := newTrayController(&fakeWindow{}, pause)

	if got := c.togglePause(); !got || !pause.paused {
		t.Fatalf("first toggle should pause: returned=%v flag=%v", got, pause.paused)
	}
	if got := c.togglePause(); got || pause.paused {
		t.Fatalf("second toggle should unpause: returned=%v flag=%v", got, pause.paused)
	}
}

func TestTrayController_SetPausedMirrorsToCheckbox(t *testing.T) {
	pause := &fakePause{}
	c := newTrayController(&fakeWindow{}, pause)

	var checked []bool
	c.reflectChecked = func(p bool) { checked = append(checked, p) }

	c.setPaused(true) // frontend-style call, not via the menu
	c.togglePause()   // menu-style call, now unpauses

	if !slices.Equal(checked, []bool{true, false}) {
		t.Fatalf("checkbox did not mirror every pause change: %v", checked)
	}
	if pause.paused {
		t.Fatal("daemon flag should be unpaused after the toggle")
	}
}
