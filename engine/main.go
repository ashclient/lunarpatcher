package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var appVersion = "1.4.1"

func defaultDir() string { return filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Lunar Client") }
func main() {
	dir := flag.String("dir", defaultDir(), "Lunar installation directory")
	mode := flag.String("mode", "ads", "ads, layout, sakura or color")
	color := flag.String("color", "#D676A5", "Custom background color, #RRGGBB")
	roblox := flag.Bool("roblox", false, "Enable Roblox launcher in Lunar sidebar")
	restore := flag.Bool("restore", false, "restore backed-up original Lunar files")
	check := flag.Bool("check", false, "validate compatibility without modifying files")
	checkUpdate := flag.Bool("check-update", false, "check for new LunarPatcher release")
	downloadUpdate := flag.Bool("download-update", false, "download and verify latest installer")
	version := flag.Bool("version", false, "print installed application version")
	status := flag.Bool("status", false, "read launcher version and current patch status")
	flag.Parse()
	if *status {
		v, err := readInstallStatus(*dir)
		if err != nil {
			fail(err)
		}
		if err := json.NewEncoder(os.Stdout).Encode(v); err != nil {
			fail(err)
		}
		return
	}
	if *version {
		fmt.Println(appVersion)
		return
	}
	if *checkUpdate || *downloadUpdate {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		release, err := latestRelease(ctx)
		if err != nil {
			fail(err)
		}
		if *checkUpdate {
			_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"current": appVersion, "latest": release.TagName, "available": newerVersion(release.TagName, appVersion), "url": release.HTMLURL})
			return
		}
		if !newerVersion(release.TagName, appVersion) {
			fmt.Println("NO_UPDATE")
			return
		}
		path, err := downloadInstaller(ctx, release)
		if err != nil {
			fail(err)
		}
		fmt.Println(path)
		return
	}
	if *restore {
		if err := Restore(*dir); err != nil {
			fail(err)
		}
		fmt.Println("Original launcher restored")
		return
	}
	report, err := Install(*dir, *mode, *color, *check, *roblox)
	if err != nil {
		fail(err)
	}
	_ = json.NewEncoder(os.Stdout).Encode(report)
}
func fail(e error) { fmt.Fprintln(os.Stderr, "ERROR:", e); os.Exit(1) }
