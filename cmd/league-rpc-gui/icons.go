package main

import _ "embed"

// trayIcon is the League crest at systray size; used for light and dark alike.
//
//go:embed tray-icon.png
var trayIcon []byte

// appIcon is the League crest; also the dev-mode taskbar/titlebar fallback.
//
//go:embed app-icon.png
var appIcon []byte
