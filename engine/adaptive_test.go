package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionNeutralAndFailClosed(t *testing.T) {
	raw, e := os.ReadFile("/mnt/data/app.asar")
	if os.IsNotExist(e) {
		t.Skip("local test fixture not mounted")
	}
	if e != nil {
		t.Fatal(e)
	}
	a, e := parseAsar(raw)
	if e != nil {
		t.Fatal(e)
	}
	pkg, ok := a.entry("package.json")
	if !ok {
		t.Fatal("no package.json")
	}
	newPkg := bytes.Replace(pkg, []byte(`3.7.31-ow`), []byte(`3.7.32-master.15672`), 1)
	newer, e := a.rebuild(map[string][]byte{"package.json": newPkg})
	if e != nil {
		t.Fatal(e)
	}
	patched, r, e := patchSource(newer, "color", "#F19CC1")
	if e != nil {
		t.Fatal(e)
	}
	if r.Announcements != "News moved to Updates; Quick Play, shortcuts, and full friends list enabled" {
		t.Fatalf("homepage announcements not removed: %s", r.Announcements)
	}
	mainBefore, _ := a.entry("dist/assets/main.js")
	archiveAfter, err := parseAsar(patched)
	if err != nil {
		t.Fatal(err)
	}
	mainAfter, _ := archiveAfter.entry("dist/assets/main.js")
	if bytes.Equal(mainBefore, mainAfter) || !bytes.Contains(mainAfter, []byte("lunarPatcherHomeTools=")) || !bytes.Contains(mainAfter, []byte("lunarPatcherUpdatesPage=")) {
		t.Fatal("expected home tools, expanded friends and updates route")
	}
	if !bytes.Contains(mainAfter, []byte("onOptionsClick:")) {
		t.Fatal("must keep launcher settings controls")
	}
	if r.Version != "3.7.32-master.15672" {
		t.Fatalf("wrong new version: %s", r.Version)
	}
	b, e := parseAsar(patched)
	if e != nil {
		t.Fatal(e)
	}
	html, ok := b.entry("dist/html/main.html")
	if !ok || !bytes.Contains(bytes.ToLower(html), []byte("#f19cc1")) {
		t.Fatalf("custom color absent, first snippet: %s", string(html[:min(len(html), 400)]))
	}
	ad, ok := b.entry("dist/assets/use-show-ad.js")
	if !ok || !bytes.Contains(ad, []byte("if(true)return null;")) {
		t.Fatal("ad guard not disabled")
	}
	// Sakura mode is independent of the color picker and retains its wallpaper.
	sakuraA, _, e := patchSource(newer, "sakura", "#F19CC1")
	if e != nil {
		t.Fatal(e)
	}
	sakuraB, _, e := patchSource(newer, "sakura", "#0020FF")
	if e != nil {
		t.Fatal(e)
	}
	pa, er := parseAsar(sakuraA)
	if er != nil {
		t.Fatal(er)
	}
	pb, er := parseAsar(sakuraB)
	if er != nil {
		t.Fatal(er)
	}
	ah, _ := pa.entry("dist/html/main.html")
	bh, _ := pb.entry("dist/html/main.html")
	if !bytes.Equal(ah, bh) {
		t.Fatal("Sakura's rendered HTML must not depend on custom color")
	}
	if bytes.Contains(ah, []byte("%!")) {
		t.Fatal("theme CSS contains formatting error")
	}
	if bytes.Contains(html, []byte("data:image/jpeg;base64")) {
		t.Fatal("Custom Color unexpectedly embeds Sakura artwork")
	}
	fixed, e := parseAsar(sakuraA)
	if e != nil {
		t.Fatal(e)
	}
	fixedHTML, ok := fixed.entry("dist/html/main.html")
	if !ok || !bytes.Contains(fixedHTML, []byte("data:image/jpeg;base64")) {
		t.Fatal("Sakura wallpaper missing")
	}
	// If the ad guard changes completely, refuse rather than patch stale selectors.
	adOriginal, ok := a.entry("dist/assets/use-show-ad.js")
	if !ok {
		t.Fatal("missing ad asset")
	}
	removed := guardPattern.ReplaceAll(adOriginal, []byte("if(x) return null;"))
	unknown, e := a.rebuild(map[string][]byte{"dist/assets/use-show-ad.js": removed})
	if e != nil {
		t.Fatal(e)
	}
	_, _, e = patchSource(unknown, "ads", "#D676A5")
	if e == nil || !strings.Contains(e.Error(), "ad-renderer signature") {
		t.Fatalf("unknown version should fail safely: %v", e)
	}
}
func TestFullLifecycle(t *testing.T) {
	exe := filepath.Join("/mnt/data/lunar-adaptive-fixture", "Lunar Client.exe")
	asar := filepath.Join("/mnt/data/lunar-adaptive-fixture", "resources", "app.asar")
	e1, e := os.ReadFile(exe)
	if os.IsNotExist(e) {
		t.Skip("local EXE fixture not mounted")
	}
	if e != nil {
		t.Fatal(e)
	}
	a1, e := os.ReadFile(asar)
	if e != nil {
		t.Fatal(e)
	}
	d := t.TempDir()
	if e = os.MkdirAll(filepath.Join(d, "resources"), 0700); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(d, "Lunar Client.exe"), e1, 0600); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(d, "resources", "app.asar"), a1, 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = Install(d, "sakura", "#F19CC1", false); e != nil {
		t.Fatal(e)
	}
	if _, e = Install(d, "color", "#8EBACA", false); e != nil {
		t.Fatal(e)
	}
	state, e := readState(d)
	if e != nil || state == nil || state.Mode != "color" || state.Color != "#8EBACA" {
		t.Fatalf("custom color state not saved: %v %v", state, e)
	}
	if _, e = Install(d, "layout", "#D676A5", false); e != nil {
		t.Fatal(e)
	}
	if e = Restore(d); e != nil {
		t.Fatal(e)
	}
	e2, _ := os.ReadFile(filepath.Join(d, "Lunar Client.exe"))
	a2, _ := os.ReadFile(filepath.Join(d, "resources", "app.asar"))
	if !bytes.Equal(e1, e2) || !bytes.Equal(a1, a2) {
		t.Fatal("restore not byte-for-byte")
	}
}

func TestUnknownLayoutFailsClosed(t *testing.T) {
	if _, err := patchHomeLayout([]byte("unrecognized launcher home")); err == nil {
		t.Fatal("must refuse unrecognized homepage composition")
	}
}
func TestThemeModesRemainIndependent(t *testing.T) {
	s := makeCSS("sakura", "#0088FF")
	if !strings.Contains(s, "data:image/jpeg;base64,") || !strings.Contains(s, ".lunar-patcher-home") {
		t.Fatal("Sakura wallpaper or scoped home styling missing")
	}
	c := makeCSS("color", "#0088FF")
	if strings.Contains(c, "data:image/jpeg;base64,") || !strings.Contains(c, "#0088ff") {
		t.Fatal("custom color accidentally contains Sakura or ignores selected color")
	}
	if strings.Contains(s, "%!") || strings.Contains(c, "%!") {
		t.Fatal("invalid fmt CSS placeholder")
	}
}

// The user supplied a v8.1-patched 3.7.32 archive. Reconstruct the three
// exact renderer changes only inside this test fixture to exercise current
// version detection without treating a modified installation as pristine.
func originalForTest(raw []byte) ([]byte, error) {
	a, err := parseAsar(raw)
	if err != nil {
		return nil, err
	}
	html, ok := a.entry("dist/html/main.html")
	if !ok {
		return nil, fmt.Errorf("missing html")
	}
	ad, ok := a.entry("dist/assets/use-show-ad.js")
	if !ok {
		return nil, fmt.Errorf("missing ad")
	}
	start := bytes.Index(html, []byte(`<style id="lunar-adaptive-css">/* LUNAR_ADAPTIVE_V8_1 */`))
	if start < 0 {
		return nil, fmt.Errorf("not recognized v8.1 fixture")
	}
	endrel := bytes.Index(html[start:], []byte(`</style>`))
	if endrel < 0 {
		return nil, fmt.Errorf("css end missing")
	}
	end := start + endrel + len(`</style>`)
	originalHTML := append(append([]byte{}, html[:start]...), html[end:]...)
	guard := []byte(`if(true)return null;/* Lunar Adaptive: no ad slot */`)
	if bytes.Count(ad, guard) != 1 {
		return nil, fmt.Errorf("expected single ad guard")
	}
	originalAd := bytes.Replace(ad, guard, []byte(`if(!(a&&e.willShowAd))return null;`), 1)
	return a.rebuild(map[string][]byte{"dist/html/main.html": originalHTML, "dist/assets/use-show-ad.js": originalAd})
}
func TestCurrentUploadedLauncher3_7_32(t *testing.T) {
	patchedRaw, err := os.ReadFile("/mnt/data/app(3).asar")
	if os.IsNotExist(err) {
		t.Skip("uploaded latest ASAR unavailable")
	}
	if err != nil {
		t.Fatal(err)
	}
	raw, err := originalForTest(patchedRaw)
	if err != nil {
		t.Fatal(err)
	}
	a, err := parseAsar(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got := parseVersion(a); got != "3.7.32-ow" {
		t.Fatalf("unexpected uploaded version: %s", got)
	}
	main, _ := a.entry("dist/assets/main.js")
	for _, mode := range []string{"ads", "layout", "sakura", "color"} {
		updated, report, err := patchSource(raw, mode, "#E8A3C0")
		if err != nil {
			t.Fatalf("%s: %v", mode, err)
		}
		if report.Version != "3.7.32-ow" {
			t.Fatal(report.Version)
		}
		b, err := parseAsar(updated)
		if err != nil {
			t.Fatal(err)
		}
		if mode != "ads" {
			patched, _ := b.entry("dist/assets/main.js")
			if bytes.Equal(main, patched) || !bytes.Contains(patched, []byte("lunarPatcherHomeTools=")) || !bytes.Contains(patched, []byte("lunarPatcherUpdatesPage=")) {
				t.Fatalf("%s: homepage tools / updates route missing", mode)
			}
			if !bytes.Contains(patched, []byte("(0,V.jsx)(iU,{full:!0,mode:`list`,rootRef:D})")) {
				t.Fatal("friends panel not forced to full list")
			}
			if !bytes.Contains(patched, []byte("kU=()=>{")) {
				t.Fatal("original announcements component lost; Updates would be empty")
			}
			if !bytes.Contains(patched, []byte("(0,V.jsx)(aU,{})")) || !bytes.Contains(patched, []byte("(0,V.jsx)(gU,{})")) {
				t.Fatal("toolbar controls must remain")
			}
			if err := os.WriteFile("/mnt/data/lunarpatcher_current_patched_main.mjs", patched, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
}
