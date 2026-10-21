// Package settings keeps user choices in %APPDATA%\TelTun\settings.json.
package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Settings struct {
	Language       string `json:"language"` // "en" or "tr"
	AutoConnect    bool   `json:"autoConnect"`
	AutoReconnect  bool   `json:"autoReconnect"`
	CloseToTray    bool   `json:"closeToTray"`
	BlockQUIC      bool   `json:"blockQuic"`
	Device         string `json:"device"` // preferred adb serial, empty = any
	HotspotWithTun bool   `json:"hotspotWithTun"`
	HotspotSSID    string `json:"hotspotSsid"`
	HotspotPass    string `json:"hotspotPass"`
}

func Default() Settings {
	return Settings{Language: "", AutoReconnect: true, CloseToTray: true, BlockQUIC: true}
}

func path() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "TelTun", "settings.json")
}

func Load() Settings {
	s := Default()
	if b, err := os.ReadFile(path()); err == nil {
		json.Unmarshal(b, &s)
	}
	return s
}

func Save(s Settings) error {
	p := path()
	os.MkdirAll(filepath.Dir(p), 0o755)
	b, _ := json.MarshalIndent(s, "", "  ")
	return os.WriteFile(p, b, 0o644)
}
