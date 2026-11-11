package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestPatchedLunarJSParses(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node unavailable")
	}
	for _, tt := range []struct {
		fixture     string
		reconstruct bool
	}{{"/mnt/data/app.asar", false}, {"/mnt/data/app(3).asar", true}, {"/mnt/data/app(4).asar", false}} {
		t.Run(filepath.Base(tt.fixture), func(t *testing.T) {
			raw, err := os.ReadFile(tt.fixture)
			if os.IsNotExist(err) {
				t.Skip("Lunar fixture unavailable")
			}
			if err != nil {
				t.Fatal(err)
			}
			if tt.reconstruct {
				raw, err = originalForTest(raw)
				if err != nil {
					t.Fatal(err)
				}
			}
			modified, _, err := patchSource(raw, "sakura", "#D676A5", true)
			if err != nil {
				t.Fatal(err)
			}
			archive, err := parseAsar(modified)
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range []string{"dist/assets/main.js", "dist-electron/electron/main.js", "dist-electron/preload/launcher/index.js"} {
				code, ok := archive.entry(file)
				if !ok {
					t.Fatal("missing", file)
				}
				target := filepath.Join(t.TempDir(), filepath.Base(filepath.Dir(file))+"_"+filepath.Base(file))
				if err := os.WriteFile(target, code, 0600); err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command("node", "--check", target)
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("%s JavaScript syntax invalid: %s (%v)", file, output, err)
				}
			}
		})
	}
}
