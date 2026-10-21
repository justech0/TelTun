package main

import (
	_ "embed"
	goruntime "runtime"
	"sync"

	"fyne.io/systray"

	"teltun/internal/engine"
)

//go:embed build/tray-on.ico
var trayOn []byte

//go:embed build/tray-off.ico
var trayOff []byte

var trayText = map[string]map[string]string{
	"en": {"show": "Open TelTun", "connect": "Connect", "disconnect": "Disconnect",
		"hotspot": "Mobile hotspot", "quit": "Quit", "on": "TelTun - connected", "off": "TelTun - not connected"},
	"tr": {"show": "TelTun'u aç", "connect": "Bağlan", "disconnect": "Bağlantıyı kes",
		"hotspot": "Mobil etkin nokta", "quit": "Çıkış", "on": "TelTun - bağlı", "off": "TelTun - bağlı değil"},
}

// tray is the notification-area icon. It runs its own message loop on a
// locked OS thread, next to the Wails window loop.
type tray struct {
	mu                          sync.Mutex
	ready                       bool
	state                       engine.State
	lang                        string
	show, toggle, hotspot, quit *systray.MenuItem
}

func startTray(a *App) *tray {
	t := &tray{state: engine.Idle, lang: a.Settings().Language}
	go func() {
		goruntime.LockOSThread()
		systray.Run(func() { t.onReady(a) }, nil)
	}()
	return t
}

func (t *tray) onReady(a *App) {
	systray.SetIcon(trayOff)
	systray.SetOnTapped(a.Show)
	t.mu.Lock()
	t.show = systray.AddMenuItem("", "")
	systray.AddSeparator()
	t.toggle = systray.AddMenuItem("", "")
	t.hotspot = systray.AddMenuItem("", "")
	systray.AddSeparator()
	t.quit = systray.AddMenuItem("", "")
	t.ready = true
	t.mu.Unlock()
	t.update(t.state, t.lang)

	go func() {
		for {
			select {
			case <-t.show.ClickedCh:
				a.Show()
			case <-t.toggle.ClickedCh:
				a.Toggle()
			case <-t.hotspot.ClickedCh:
				go func() {
					if i, err := a.HotspotStatus(); err == nil && i.State == "On" {
						a.HotspotStop()
					} else {
						a.HotspotStart()
					}
				}()
			case <-t.quit.ClickedCh:
				a.Quit()
				return
			}
		}
	}()
}

func (t *tray) update(state engine.State, lang string) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.state = state
	if lang != "" {
		t.lang = lang
	}
	if !t.ready {
		return
	}
	tx, ok := trayText[t.lang]
	if !ok {
		tx = trayText["en"]
	}
	t.show.SetTitle(tx["show"])
	t.hotspot.SetTitle(tx["hotspot"])
	t.quit.SetTitle(tx["quit"])
	switch state {
	case engine.Connected:
		systray.SetIcon(trayOn)
		systray.SetTooltip(tx["on"])
		t.toggle.SetTitle(tx["disconnect"])
		t.toggle.Enable()
	case engine.Idle:
		systray.SetIcon(trayOff)
		systray.SetTooltip(tx["off"])
		t.toggle.SetTitle(tx["connect"])
		t.toggle.Enable()
	default:
		t.toggle.Disable()
	}
}
