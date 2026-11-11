package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRealUploaded3_7_34(t *testing.T) {
	raw, err := os.ReadFile("/mnt/data/app(4).asar")
	if os.IsNotExist(err) {
		t.Skip("new Lunar ASAR fixture unavailable")
	}
	if err != nil {
		t.Fatal(err)
	}
	a, err := parseAsar(raw)
	if err != nil {
		t.Fatal(err)
	}
	if version := parseVersion(a); version != "3.7.34-ow" {
		t.Fatalf("unexpected fixture %s", version)
	}
	originalMain, _ := a.entry("dist/assets/main.js")
	originalAd, _ := a.entry("dist/assets/use-show-ad.js")
	for _, mode := range []string{"ads", "layout", "sakura", "color"} {
		for _, roblox := range []bool{false, true} {
			name := mode
			if roblox {
				name += "+roblox"
			}
			t.Run(name, func(t *testing.T) {
				patched, report, err := patchSource(raw, mode, "#ECA3C4", roblox)
				if err != nil {
					t.Fatal(err)
				}
				if report.Version != "3.7.34-ow" {
					t.Fatal(report.Version)
				}
				p, err := parseAsar(patched)
				if err != nil {
					t.Fatal(err)
				}
				ad, _ := p.entry("dist/assets/use-show-ad.js")
				if bytes.Equal(ad, originalAd) || bytes.Contains(ad, []byte("if(!a||!e.willShowAd)return null;")) || !bytes.Contains(ad, []byte("if(true)return null;")) {
					t.Fatal("ad guard still enabled")
				}
				html, _ := p.entry("dist/html/main.html")
				if !bytes.Contains(html, []byte(marker)) {
					t.Fatal("patch marker absent")
				}
				main, _ := p.entry("dist/assets/main.js")
				if mode != "ads" {
					for _, token := range []string{"lunarPatcherHomeTools=", "lunarPatcherUpdatesPage=", "className:`lunar-patcher-friends`", "(0,V.jsx)(qG,{full:!0,mode:`list`,rootRef:E})", "(0,V.jsx)(WK,{})", "path:`/updates`"} {
						if !bytes.Contains(main, []byte(token)) {
							t.Errorf("missing %s", token)
						}
					}
					if bytes.Contains(main, []byte("(0,V.jsx)(xf,{...C.ad.main")) {
						t.Error("home ad container still present")
					}
					if strings.Count(string(main), "(0,V.jsx)(vK,{})") != 1 {
						t.Error("news should render only in Updates")
					}
				} else if !bytes.Equal(main, originalMain) && !roblox {
					t.Error("Ads Only should retain original home renderer")
				}
				if roblox {
					if !bytes.Contains(main, []byte("lunarPatcherRobloxPage=")) {
						t.Fatal("Roblox route missing")
					}
					for _, file := range []string{"dist-electron/electron/main.js", "dist-electron/preload/launcher/index.js"} {
						b, _ := p.entry(file)
						if !bytes.Contains(b, []byte("lunarpatcher:roblox:show")) && !bytes.Contains(b, []byte("lunarPatcherRobloxBrowser")) {
							t.Error("Roblox bridge missing in", file)
						}
					}
				}
				if mode == "sakura" && !bytes.Contains(html, []byte("data:image/jpeg;base64")) {
					t.Error("Sakura wallpaper missing")
				}
				if mode == "color" && bytes.Contains(html, []byte("data:image/jpeg;base64")) {
					t.Error("color mode must not embed Sakura art")
				}
				if mode == "sakura" && roblox { // dump representative patched build for external syntax checks
					dir := "/mnt/data/lunar_hotfix_180/validation"
					if err = os.MkdirAll(dir, 0755); err != nil {
						t.Fatal(err)
					}
					for _, f := range []string{"dist/assets/main.js", "dist/assets/use-show-ad.js", "dist-electron/electron/main.js", "dist-electron/preload/launcher/index.js"} {
						out := filepath.Join(dir, filepath.Base(f)+"."+strings.ReplaceAll(f, "/", "_")+".mjs")
						b, _ := p.entry(f)
						if err = os.WriteFile(out, b, 0644); err != nil {
							t.Fatal(err)
						}
					}
				}
			})
		}
	}
}

func Test334ChangedSignaturesFailClosed(t *testing.T) {
	raw, err := os.ReadFile("/mnt/data/app(4).asar")
	if os.IsNotExist(err) {
		t.Skip("3.7.34 fixture unavailable")
	}
	if err != nil {
		t.Fatal(err)
	}
	a, err := parseAsar(raw)
	if err != nil {
		t.Fatal(err)
	}
	ad, _ := a.entry("dist/assets/use-show-ad.js")
	if bytes.Count(ad, []byte("if(!a||!e.willShowAd)return null;")) != 1 {
		t.Fatal("unexpected test fixture ad guard")
	}
	badAd := bytes.Replace(ad, []byte("if(!a||!e.willShowAd)return null;"), []byte("if(!a||e.willShowAd)return null;"), 1)
	badArchive, err := a.rebuild(map[string][]byte{"dist/assets/use-show-ad.js": badAd})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = patchSource(badArchive, "ads", "#ECA3C4"); err == nil || !strings.Contains(err.Error(), "ad-renderer signature") {
		t.Fatalf("unknown ad guard must fail closed: %v", err)
	}
	main, _ := a.entry("dist/assets/main.js")
	if bytes.Count(main, []byte("(0,V.jsx)(qG,{full:n===!1,mode:D,rootRef:E})")) != 1 {
		t.Fatal("unexpected friends fixture")
	}
	badMain := bytes.Replace(main, []byte("(0,V.jsx)(qG,{full:n===!1,mode:D,rootRef:E})"), []byte("(0,V.jsx)(qG,{full:n===!1,mode:X,rootRef:E})"), 1)
	badArchive, err = a.rebuild(map[string][]byte{"dist/assets/main.js": badMain})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = patchSource(badArchive, "layout", "#ECA3C4"); err == nil || !strings.Contains(err.Error(), "signature") {
		t.Fatalf("unknown home layout must fail closed: %v", err)
	}
}
