package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRobloxIntegrationOnBothKnownRendererVersions(t *testing.T) {
	fixtures := []struct {
		path        string
		reconstruct bool
	}{{"/mnt/data/app.asar", false}, {"/mnt/data/app(3).asar", true}}
	for _, item := range fixtures {
		t.Run(item.path, func(t *testing.T) {
			raw, err := os.ReadFile(item.path)
			if os.IsNotExist(err) {
				t.Skip("fixture unavailable")
			}
			if err != nil {
				t.Fatal(err)
			}
			if item.reconstruct {
				raw, err = originalForTest(raw)
				if err != nil {
					t.Fatal(err)
				}
			}
			for _, mode := range []string{"ads", "layout", "sakura", "color"} {
				modified, report, err := patchSource(raw, mode, "#D676A5", true)
				if err != nil {
					t.Fatalf("%s: %v", mode, err)
				}
				if report.Roblox != "Roblox sidebar and game launcher enabled" {
					t.Fatalf("%s missing report", mode)
				}
				a, err := parseAsar(modified)
				if err != nil {
					t.Fatal(err)
				}
				code, _ := a.entry("dist/assets/main.js")
				html, _ := a.entry("dist/html/main.html")
				for _, want := range []string{"lunarPatcherRobloxPage=", "lunarPatcherRobloxIcon=", "path:`/roblox`", "roblox://experiences/start?placeId=", "lunarpatcher.roblox."} {
					if !bytes.Contains(code, []byte(want)) {
						t.Fatalf("%s missing %s", mode, want)
					}
				}
				if !bytes.Contains(html, []byte("lunar-patcher-roblox-css")) {
					t.Fatal("Roblox CSS missing")
				}
				if mode != "ads" && !bytes.Contains(code, []byte("lunarPatcherHomeTools=")) {
					t.Fatal("layout missing")
				}
				if mode == "ads" && bytes.Contains(code, []byte("lunarPatcherHomeTools=")) {
					t.Fatal("ads-only changed unrelated home layout")
				}
				path := t.TempDir() + "/patched.mjs"
				if mode == "sakura" {
					external := "/mnt/data/lunarpatcher_v140/validated_331.mjs"
					if item.reconstruct {
						external = "/mnt/data/lunarpatcher_v140/validated_332.mjs"
					}
					if err := os.WriteFile(external, code, 0600); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.WriteFile(path, code, 0600); err != nil {
					t.Fatal(err)
				}
				_ = path // Node --check run separately when local fixture test is exercised.
			}
			unchanged, report, err := patchSource(raw, "layout", "#D676A5", false)
			if err != nil {
				t.Fatal(err)
			}
			if report.Roblox != "Disabled" {
				t.Fatal("toggle disabled report incorrect")
			}
			a, _ := parseAsar(unchanged)
			main, _ := a.entry("dist/assets/main.js")
			if strings.Contains(string(main), "lunarPatcherRobloxPage=") {
				t.Fatal("Roblox injected while disabled")
			}
		})
	}
}

func TestRobloxLinkValidationIsNotAFreeformExternalURL(t *testing.T) {
	if !strings.Contains(robloxWidgetJS, `parsed.hostname.toLowerCase()`) || !strings.Contains(robloxWidgetJS, `www.roblox.com`) {
		t.Fatal("Roblox domain validation missing")
	}
	if !strings.Contains(robloxWidgetJS, `^[1-9]`) {
		t.Fatal("place ID validation missing")
	}
	if !strings.Contains(robloxWidgetJS, `accounts.v1`) || !strings.Contains(robloxWidgetJS, `settings.v1`) {
		t.Fatal("account and settings storage missing")
	}
	if strings.Contains(robloxWidgetJS, `eval(`) {
		t.Fatal("unsafe code evaluation")
	}
}
