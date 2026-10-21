// Package hotspot turns Windows "Mobile hotspot" on and off.
//
// The tethering API is WinRT only, so a short PowerShell script calls it.
// The script runs hidden and prints one JSON line as its answer.
package hotspot

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unicode/utf16"
)

//go:embed hotspot.ps1
var script string

type Info struct {
	OK      bool   `json:"ok"`
	State   string `json:"state"` // On, Off, InTransition, Unknown
	SSID    string `json:"ssid"`
	Pass    string `json:"pass"`
	Clients int    `json:"clients"`
	Max     int    `json:"max"`
	Profile string `json:"profile"`
	Msg     string `json:"msg"`
}

func Status() (Info, error) { return run("status", "", "") }
func Stop() (Info, error)   { return run("stop", "", "") }

// Start turns the hotspot on. Empty ssid/pass keep the current Windows values.
func Start(ssid, pass string) (Info, error) { return run("start", ssid, pass) }

func run(action, ssid, pass string) (Info, error) {
	q := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
	full := "$Action=" + q(action) + ";$Ssid=" + q(ssid) + ";$Pass=" + q(pass) + "\n" + script

	u := utf16.Encode([]rune(full))
	b := make([]byte, len(u)*2)
	for i, c := range u {
		b[2*i], b[2*i+1] = byte(c), byte(c>>8)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive",
		"-ExecutionPolicy", "Bypass", "-EncodedCommand", base64.StdEncoding.EncodeToString(b))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()

	var info Info
	for _, ln := range strings.Split(out.String(), "\n") {
		ln = strings.TrimSpace(ln)
		if strings.HasPrefix(ln, "{") && json.Unmarshal([]byte(ln), &info) == nil {
			if !info.OK {
				return info, errors.New(info.Msg)
			}
			return info, nil
		}
	}
	if err == nil {
		err = errors.New(strings.TrimSpace(out.String()))
	}
	return info, err
}
