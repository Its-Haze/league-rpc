// Command league-rpc-gui runs the daemon and the desktop GUI in one process.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/its-haze/league-rpc/frontend"
	"github.com/its-haze/league-rpc/internal/app"
	"github.com/its-haze/league-rpc/internal/config"
	"github.com/its-haze/league-rpc/internal/daemon"
	"github.com/its-haze/league-rpc/internal/discordapp"
	"github.com/its-haze/league-rpc/internal/logging"
	"github.com/its-haze/league-rpc/internal/startup"
	"github.com/its-haze/league-rpc/internal/updates"
	"github.com/its-haze/league-rpc/internal/version"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/icons"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"
)

// singleInstanceID keeps a second launch from starting its own daemon; it
// signals the running instance to surface its window instead.
const singleInstanceID = "com.its-haze.league-rpc"

func main() {
	// A run launched by the Run entry carries the hidden marker. Drop any
	// console Windows attached to it before anything can write there.
	startHidden := startup.StartedHidden(os.Args[1:])
	if startHidden {
		startup.DetachConsole()
	}

	cfg, err := config.LoadOrCreate()
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}

	sink, err := logging.New(logging.Options{Debug: cfg.Advanced.DebugMode})
	if err != nil {
		log.Fatalf("init logging: %v", err)
	}
	defer sink.Close()

	store := config.NewStore(cfg)
	d := daemon.Wire(store, sink.Logger)

	// Keep the "start with Windows" registry entry matching the setting, both
	// now and whenever the GUI toggles it.
	reconciler := startup.New(startup.SystemRunKey())
	if err := reconciler.Reconcile(cfg.Behavior.LaunchAtStartup); err != nil {
		sink.Logger.Warn().Err(err).Msg("could not reconcile start-with-Windows entry")
	}

	ctx, cancel := context.WithCancel(context.Background())
	daemonDone := make(chan struct{})

	// Assigned once the window exists; the single-instance callback reads it.
	var mainWindow *application.WebviewWindow

	wailsApp := application.New(application.Options{
		Name:        "League RPC",
		Description: "League of Legends Discord Rich Presence",
		Icon:        appIcon,
		Assets:      application.AssetOptions{Handler: application.AssetFileServerFS(frontend.Assets())},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: singleInstanceID,
			OnSecondInstanceLaunch: func(application.SecondInstanceData) {
				if mainWindow != nil {
					mainWindow.Show()
					mainWindow.Focus()
				}
			},
		},
		OnShutdown: func() {
			cancel()
			<-daemonDone
		},
	})

	// wailsApp.Updater exists only once application.New has run, hence wiring
	// App Update here rather than earlier alongside the rest of guiApp.
	updateCoord := updates.New(wailsApp.Updater, updates.NewProductionHTTPDoer(), version.IsDev(), sink.Logger)
	if updCfg, err := updates.BuildConfig(version.Version()); err != nil {
		sink.Logger.Warn().Err(err).Msg("could not configure the app updater")
	} else if err := wailsApp.Updater.Init(updCfg); err != nil {
		sink.Logger.Warn().Err(err).Msg("could not initialize the app updater")
	}

	logDir, err := logging.LogDir()
	if err != nil {
		sink.Logger.Warn().Err(err).Msg("could not resolve the logs directory")
	}

	guiApp := app.New(store, d,
		app.WithStatus(d, d, d.SubscribeState()),
		app.WithUpdater(updateAdapter{updateCoord}),
		app.WithLogs(sink.Ring, logDir),
		app.WithAppNameLookup(discordapp.New(discordapp.NewProductionHTTPDoer())),
	)
	svc := newGUIService(guiApp)
	wailsApp.RegisterService(application.NewService(svc))

	// Registered so an update-ready toast can be pushed from OnUpdateChange
	// below; clicking it brings the window forward to the About screen.
	notifier := notifications.New()
	wailsApp.RegisterService(application.NewService(notifier))
	notifier.OnNotificationResponse(func(result notifications.NotificationResult) {
		if result.Response.ID != updateReadyNotificationID {
			return
		}
		application.InvokeAsync(func() {
			if mainWindow != nil {
				mainWindow.Show()
				mainWindow.Focus()
			}
			wailsApp.Event.Emit(navigateAboutEvent)
		})
	})

	// Start the daemon only after application.New has run the single-instance
	// guard; a second launch exits inside New and never touches Discord.
	go func() {
		defer close(daemonDone)
		d.Run(ctx)
	}()

	// Bridge config.Store changes to a frontend event so screens can react
	// without polling. Runs until the app shuts down.
	go svc.publishConfigChanges(ctx, wailsApp)

	// Live-tail the log ring to the frontend so the Help screen's viewer
	// doesn't have to poll. Runs until the app shuts down.
	go svc.publishLogLines(ctx, wailsApp)

	// Push status snapshots to the frontend on change, and drive the bridge
	// that assembles them. Both run until the app shuts down.
	guiApp.OnStatusChange(func(s app.StatusSnapshot) {
		wailsApp.Event.Emit(statusChangedEvent, s)
	})
	go guiApp.RunStatus(ctx)

	// Apply a start-with-Windows toggle to the registry the moment it changes,
	// not just on the next launch.
	go watchConfigField(ctx, store,
		func(c *config.Config) bool { return c.Behavior.LaunchAtStartup },
		func(want bool) {
			if err := reconciler.Reconcile(want); err != nil {
				sink.Logger.Warn().Err(err).Msg("could not update start-with-Windows entry")
			}
		})

	// Follow a debug-logging toggle immediately, not just on the next launch.
	go watchConfigField(ctx, store, func(c *config.Config) bool { return c.Advanced.DebugMode }, logging.SetDebug)

	// A hidden launch opens straight to the tray; a manual run shows the
	// window.
	windowWidth, windowHeight := defaultWindowSize()
	mainWindow = wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "League RPC",
		Width:            windowWidth,
		Height:           windowHeight,
		MinWidth:         minWindowWidth,
		MinHeight:        minWindowHeight,
		Hidden:           startHidden,
		BackgroundColour: application.NewRGB(15, 17, 23),
		URL:              "/",
		// Custom chrome (frontend/src/components/shell/TitleBar.tsx) replaces
		// the native titlebar, so the window carries no OS decorations.
		Frameless: true,
	})

	tray := newTrayController(windowAdapter{mainWindow}, d)
	tray.closeAction = func() string { return store.Load().Behavior.CloseAction }
	tray.quit = wailsApp.Quit
	tray.askClose = func() {
		mainWindow.Show() // a close from the taskbar can arrive minimized
		wailsApp.Event.Emit(closeRequestedEvent)
	}

	// Every close is cancelled and re-decided by the tray controller, so the
	// window only ever goes away by hiding or by a real quit.
	mainWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		e.Cancel()
		tray.handleClose()
	})

	systemTray := wailsApp.SystemTray.New()
	systemTray.SetIcon(trayIcon)
	systemTray.SetDarkModeIcon(trayIcon)
	systemTray.SetTooltip("League RPC")
	if runtime.GOOS == "darwin" {
		systemTray.SetTemplateIcon(icons.SystrayMacTemplate)
	}
	systemTray.OnClick(tray.showWindow)
	systemTray.SetMenu(buildTrayMenu(wailsApp, tray, guiApp))

	// Push App Update status to the frontend and the tray tooltip, and fire a
	// one-time toast once a new version is first discovered (see updates.Coordinator).
	var notifyMu sync.Mutex
	var notifiedVersion string
	guiApp.OnUpdateChange(func(s app.UpdateStatus) {
		wailsApp.Event.Emit(updateChangedEvent, s)
		if s.Available {
			systemTray.SetTooltip("League RPC (update available)")
		} else {
			systemTray.SetTooltip("League RPC")
		}

		if !s.Available {
			return
		}
		notifyMu.Lock()
		alreadyNotified := notifiedVersion == s.Version
		notifiedVersion = s.Version
		notifyMu.Unlock()
		// Still tracked above even when suppressed, so re-enabling the
		// setting later doesn't fire a toast for a version already seen.
		if alreadyNotified || !store.Load().Behavior.NotifyUpdates {
			return
		}
		if err := notifier.SendNotification(notifications.NotificationOptions{
			ID:    updateReadyNotificationID,
			Title: "League RPC update available",
			Body:  fmt.Sprintf("Version %s is available. Click to review and install.", s.Version),
		}); err != nil {
			sink.Logger.Warn().Err(err).Msg("could not show update-available notification")
		}
	})
	go guiApp.RunUpdates(ctx)

	// Frontend pause toggles flow through the same path as tray toggles, so
	// the daemon flag and the tray checkbox stay in agreement.
	svc.pauseHook = tray.setPaused
	svc.closeHook = tray.resolveClose

	if err := wailsApp.Run(); err != nil {
		log.Fatal(err)
	}
}

// watchConfigField calls fn whenever extract's live-config reading changes,
// until ctx is canceled. Shared by every apply-immediately watcher below.
func watchConfigField[T comparable](ctx context.Context, store *config.Store, extract func(*config.Config) T, fn func(T)) {
	changes := store.Subscribe()
	last := extract(store.Load())
	for {
		select {
		case <-ctx.Done():
			return
		case cfg, ok := <-changes:
			if !ok {
				return
			}
			if next := extract(cfg); next != last {
				last = next
				fn(next)
			}
		}
	}
}

// buildTrayMenu assembles the right-click menu: Open, Pause presence, Check
// for updates, Quit.
func buildTrayMenu(wailsApp *application.App, tray *trayController, guiApp *app.App) *application.Menu {
	menu := wailsApp.NewMenu()
	menu.Add("Open").OnClick(func(*application.Context) { tray.showWindow() })

	pauseItem := menu.AddCheckbox("Pause presence", tray.pause.IsPaused())
	// A frontend toggle runs off the main thread; marshal the native menu update.
	tray.reflectChecked = func(paused bool) {
		application.InvokeAsync(func() { pauseItem.SetChecked(paused) })
	}
	pauseItem.OnClick(func(*application.Context) { tray.togglePause() })

	menu.AddSeparator()
	// Fire-and-forget: the result reaches the window (and this tooltip) the
	// same way the launch/periodic checks already do, via OnUpdateChange.
	menu.Add("Check for updates").OnClick(func(*application.Context) {
		go func() { _, _ = guiApp.CheckForUpdates(context.Background()) }()
	})

	menu.AddSeparator()
	menu.Add("Quit").OnClick(func(*application.Context) { wailsApp.Quit() })
	return menu
}

// windowAdapter fits *application.WebviewWindow to the tray's windowController;
// the window's Show/Hide return a value the interface does not.
type windowAdapter struct{ w *application.WebviewWindow }

func (a windowAdapter) Show()  { a.w.Show() }
func (a windowAdapter) Hide()  { a.w.Hide() }
func (a windowAdapter) Focus() { a.w.Focus() }
