//go:build windows

package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

//go:embed app.ps1
var appScript []byte

func showGUI() {
	exe, err := os.Executable()
	if err != nil {
		showError(err.Error())
		return
	}
	temp, err := os.MkdirTemp("", "lunarpatcher-ui-")
	if err != nil {
		showError(err.Error())
		return
	}
	defer os.RemoveAll(temp)
	script := filepath.Join(temp, "app.ps1")
	if err = os.WriteFile(script, appScript, 0600); err != nil {
		showError(err.Error())
		return
	}
	cmd := exec.Command("powershell.exe", "-NoProfile", "-STA", "-ExecutionPolicy", "Bypass", "-WindowStyle", "Hidden", "-File", script, "-AppExe", exe, "-AppVersion", appVersion)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if b, err := cmd.CombinedOutput(); err != nil {
		showError(fmt.Sprintf("Unable to open LunarPatcher: %v\n%s", err, strings.TrimSpace(string(b))))
	}
}
func showError(s string) {
	p := syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW")
	m, _ := syscall.UTF16PtrFromString(s)
	title, _ := syscall.UTF16PtrFromString("LunarPatcher")
	p.Call(0, uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(title)), 0x10)
}
