package main

import "github.com/its-haze/league-rpc/internal/config"

// windowController is the subset of the Wails window the tray drives. An
// adapter over *application.WebviewWindow satisfies it; tests use a fake.
type windowController interface {
	Show()
	Hide()
	Focus()
}

// pauseControl is the runtime pause surface the tray toggles.
type pauseControl interface {
	SetPaused(bool)
	IsPaused() bool
}

// trayController holds the window and pause wiring behind the tray menu,
type trayController struct {
	win   windowController
	pause pauseControl
	// closeAction reads the live config.CloseAction setting.
	closeAction func() string
	// askClose tells the frontend to raise the close confirmation dialog.
	askClose func()
	// quit shuts the whole app down, the same as tray Quit.
	quit func()
	// reflectChecked mirrors the pause flag onto the tray menu checkbox.
	// Set by the Wails wiring once the menu item exists.
	reflectChecked func(bool)
}

func newTrayController(win windowController, pause pauseControl) *trayController {
	return &trayController{win: win, pause: pause}
}

// showWindow brings the window back and focuses it. Used by the tray icon
// click, the "Open" menu item, and a second launch of the app.
func (c *trayController) showWindow() {
	c.win.Show()
	c.win.Focus()
}

// handleClose routes a window close through the user's close_action setting.
// Under "ask" the window stays up so the in-app dialog has something to sit on.
func (c *trayController) handleClose() {
	switch c.action() {
	case config.CloseQuit:
		c.doQuit()
	case config.CloseTray:
		c.win.Hide()
	default:
		if c.askClose == nil {
			c.win.Hide() // no frontend to ask; hiding is the safe default
			return
		}
		c.askClose()
	}
}

// resolveClose applies the answer the close dialog came back with. Remembering
// the choice is the frontend's job: it writes close_action before calling in.
func (c *trayController) resolveClose(action string) {
	if action == config.CloseQuit {
		c.doQuit()
		return
	}
	c.win.Hide()
}

// action reads the configured close behavior, defaulting to asking.
func (c *trayController) action() string {
	if c.closeAction == nil {
		return config.CloseAsk
	}
	return c.closeAction()
}

func (c *trayController) doQuit() {
	if c.quit != nil {
		c.quit()
	}
}

// setPaused is the one place pause changes flow through, so the daemon flag
// and the tray checkbox never disagree, whichever surface triggered it.
func (c *trayController) setPaused(paused bool) {
	c.pause.SetPaused(paused)
	if c.reflectChecked != nil {
		c.reflectChecked(paused)
	}
}

// togglePause flips the pause flag and returns the new value so a tray-menu
// click can update the checkbox it was clicked on.
func (c *trayController) togglePause() bool {
	next := !c.pause.IsPaused()
	c.setPaused(next)
	return next
}
