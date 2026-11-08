package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRobloxEmbeddedWebsitePatchingBothRenderers(t *testing.T) {
	for _, fixture := range []struct {
		path        string
		reconstruct bool
	}{{"/mnt/data/app.asar", false}, {"/mnt/data/app(3).asar", true}} {
		t.Run(filepath.Base(fixture.path), func(t *testing.T) {
			raw, err := os.ReadFile(fixture.path)
			if os.IsNotExist(err) {
				t.Skip("fixture unavailable")
			}
			if err != nil {
				t.Fatal(err)
			}
			if fixture.reconstruct {
				raw, err = originalForTest(raw)
				if err != nil {
					t.Fatal(err)
				}
			}
			original, err := parseAsar(raw)
			if err != nil {
				t.Fatal(err)
			}
			origMain, _ := original.entry("dist-electron/electron/main.js")
			origPreload, _ := original.entry("dist-electron/preload/launcher/index.js")
			for _, mode := range []string{"ads", "layout", "sakura", "color"} {
				patched, report, err := patchSource(raw, mode, "#D676A5", true)
				if err != nil {
					t.Fatalf("%s: %v", mode, err)
				}
				if !strings.Contains(report.Roblox, "Roblox.com") {
					t.Fatalf("missing browser report %s", report.Roblox)
				}
				a, err := parseAsar(patched)
				if err != nil {
					t.Fatal(err)
				}
				renderer, _ := a.entry("dist/assets/main.js")
				main, _ := a.entry("dist-electron/electron/main.js")
				preload, _ := a.entry("dist-electron/preload/launcher/index.js")
				html, _ := a.entry("dist/html/main.html")
				for _, tok := range []string{"lunarPatcherRobloxPage=", "lunarPatcherRobloxIcon=", "path:`/roblox`", "lunarPatcherRobloxBrowser", "lpr-browser-viewport"} {
					if !bytes.Contains(renderer, []byte(tok)) {
						t.Fatalf("%s renderer lacks %s", mode, tok)
					}
				}
				if !bytes.Contains(main, []byte("lunarpatcher:roblox:show")) || !bytes.Contains(main, []byte("persist:lunarpatcher-roblox")) {
					t.Fatal("browser backend missing")
				}
				if !bytes.Contains(preload, []byte("lunarPatcherRobloxBrowser")) {
					t.Fatal("preload bridge missing")
				}
				if bytes.Equal(main, origMain) || bytes.Equal(preload, origPreload) {
					t.Fatal("browser bridge not installed")
				}
				if !bytes.Contains(html, []byte("lunar-patcher-roblox-css")) {
					t.Fatal("browser styles missing")
				}
				if bytes.Contains(renderer, []byte("__ROBLOX_ART__")) || bytes.Contains(renderer, []byte("roblox://experiences/start?placeId=")) {
					t.Fatal("old faux launcher code remains")
				}
				if mode != "ads" && !bytes.Contains(renderer, []byte("lunarPatcherHomeTools=")) {
					t.Fatal("clean home layout regressed")
				}
			}
			patched, report, err := patchSource(raw, "layout", "#D676A5", false)
			if err != nil {
				t.Fatal(err)
			}
			if report.Roblox != "Disabled" {
				t.Fatalf("unexpected toggle report %s", report.Roblox)
			}
			a, _ := parseAsar(patched)
			main, _ := a.entry("dist-electron/electron/main.js")
			preload, _ := a.entry("dist-electron/preload/launcher/index.js")
			if !bytes.Equal(main, origMain) || !bytes.Equal(preload, origPreload) {
				t.Fatal("disabled Roblox changed Electron backend")
			}
		})
	}
}
func TestRobloxBridgeSafety(t *testing.T) {
	for _, token := range []string{"nodeIntegration:false", "contextIsolation:true", "sandbox:true", "webSecurity:true", "persist:lunarpatcher-roblox", "setPermissionRequestHandler", "will-navigate", "setWindowOpenHandler", "isOfficial(url)", "BrowserWindow.fromWebContents", "lunarpatcher:roblox:hide"} {
		if !strings.Contains(robloxMainBridge, token) {
			t.Fatalf("browser missing safety check %s", token)
		}
	}
	if strings.Contains(robloxMainBridge, "webSecurity:false") || strings.Contains(robloxMainBridge, "executeJavaScript(") {
		t.Fatal("untrusted-site security regression")
	}
	for _, token := range []string{"useEffect", "bridge.show(b)", "bridge.hide()", "lpr-browser-address", "https://www.roblox.com/home"} {
		if !strings.Contains(robloxWidgetJS, token) {
			t.Fatalf("renderer missing %s", token)
		}
	}
}
func TestRobloxBridgeRejectsUnrecognizedLunarMain(t *testing.T) {
	_, _, err := patchRobloxElectronBridge([]byte("unknown"), []byte("unknown"))
	if err == nil {
		t.Fatal("unrecognized Electron should fail closed")
	}
}
