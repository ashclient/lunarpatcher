package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type InstallStatus struct {
	Found             bool   `json:"found"`
	Version           string `json:"version"`
	Mode              string `json:"mode"`
	Color             string `json:"color"`
	CanRestore        bool   `json:"canRestore"`
	ChangedSincePatch bool   `json:"changedSincePatch"`
	Folder            string `json:"folder"`
}

func readInstallStatus(base string) (InstallStatus, error) {
	status := InstallStatus{Folder: base, Mode: "original"}
	exe := filepath.Join(base, "Lunar Client.exe")
	asar := filepath.Join(base, "resources", "app.asar")
	if _, e := os.Stat(exe); e != nil {
		return status, fmt.Errorf("Lunar Client.exe not found at %s", base)
	}
	data, e := os.ReadFile(asar)
	if e != nil {
		return status, e
	}
	archive, e := parseAsar(data)
	if e != nil {
		return status, e
	}
	status.Found = true
	status.Version = parseVersion(archive)
	st, e := readState(base)
	if e != nil {
		return status, e
	}
	if st != nil {
		status.Mode = st.Mode
		status.Color = st.Color
		status.CanRestore = true
		currentExe, err := os.ReadFile(exe)
		if err != nil {
			return status, err
		}
		status.ChangedSincePatch = hashBytes(currentExe) != st.PatchedExe || hashBytes(data) != st.PatchedAsar
	}
	return status, nil
}
