package engine

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"teltun/assets"
)

const createNoWindow = 0x08000000

// hidden makes a child process run with no console window.
func hidden(cmd *exec.Cmd) *exec.Cmd {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
	return cmd
}

// DataDir is %LOCALAPPDATA%\TelTun.
func DataDir() string {
	base, err := os.UserCacheDir()
	if err != nil {
		base = os.TempDir()
	}
	return filepath.Join(base, "TelTun")
}

// unpack writes the embedded helper files to disk, skipping unchanged ones.
func unpack() (string, error) {
	dir := filepath.Join(DataDir(), "runtime")
	err := fs.WalkDir(assets.FS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		want, err := assets.FS.ReadFile(p)
		if err != nil {
			return err
		}
		dst := filepath.Join(dir, filepath.FromSlash(p))
		if have, err := os.ReadFile(dst); err == nil && sha256.Sum256(have) == sha256.Sum256(want) {
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dst, want, 0o755)
	})
	return dir, err
}

// adb runs one adb command on the current phone.
func (e *Engine) adb(timeout time.Duration, args ...string) (string, error) {
	e.mu.Lock()
	serial := e.serial
	e.mu.Unlock()
	return e.runAdb(serial, timeout, args...)
}

// runAdb runs adb and returns its trimmed output. Empty serial = no -s.
func (e *Engine) runAdb(serial string, timeout time.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if serial != "" {
		args = append([]string{"-s", serial}, args...)
	}
	var out bytes.Buffer
	cmd := hidden(exec.CommandContext(ctx, e.adbPath, args...))
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}
