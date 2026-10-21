// Package autostart starts TelTun at sign-in through Task Scheduler.
// A scheduled task can run with admin rights without a UAC prompt, which
// the virtual network card needs. A Run registry key cannot do that.
package autostart

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unicode/utf16"
)

const taskName = "TelTun"

func schtasks(args ...string) error {
	cmd := exec.Command("schtasks.exe", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%v: %s", err, bytes.TrimSpace(out.Bytes()))
	}
	return nil
}

func Enabled() bool { return schtasks("/Query", "/TN", taskName) == nil }

func Disable() error {
	if !Enabled() {
		return nil
	}
	return schtasks("/Delete", "/TN", taskName, "/F")
}

// Enable registers the current exe to start hidden in the tray at sign-in.
func Enable() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	xml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-16"?>
<Task version="1.2" xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
  <RegistrationInfo><Description>Start TelTun at sign-in</Description></RegistrationInfo>
  <Triggers><LogonTrigger><Enabled>true</Enabled><Delay>PT10S</Delay></LogonTrigger></Triggers>
  <Principals><Principal id="Author"><LogonType>InteractiveToken</LogonType><RunLevel>HighestAvailable</RunLevel></Principal></Principals>
  <Settings>
    <MultipleInstancesPolicy>IgnoreNew</MultipleInstancesPolicy>
    <DisallowStartIfOnBatteries>false</DisallowStartIfOnBatteries>
    <StopIfGoingOnBatteries>false</StopIfGoingOnBatteries>
    <ExecutionTimeLimit>PT0S</ExecutionTimeLimit>
    <Priority>7</Priority>
  </Settings>
  <Actions Context="Author"><Exec><Command>%s</Command><Arguments>--tray</Arguments></Exec></Actions>
</Task>`, xmlEscape(exe))

	f := filepath.Join(os.TempDir(), "teltun-task.xml")
	u := utf16.Encode([]rune(xml))
	b := []byte{0xFF, 0xFE}
	for _, c := range u {
		b = append(b, byte(c), byte(c>>8))
	}
	if err := os.WriteFile(f, b, 0o644); err != nil {
		return err
	}
	defer os.Remove(f)
	return schtasks("/Create", "/TN", taskName, "/XML", f, "/F")
}

func xmlEscape(s string) string {
	var b bytes.Buffer
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
