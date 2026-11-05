package main

import (
	"bytes"
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
	if r.Announcements != "Removed from the homepage" {
		t.Fatalf("homepage announcements not removed: %s", r.Announcements)
	}
	mainBefore, _ := a.entry("dist/assets/main.js")
	archiveAfter, err := parseAsar(patched)
	if err != nil {
		t.Fatal(err)
	}
	mainAfter, _ := archiveAfter.entry("dist/assets/main.js")
	if len(mainBefore)-len(mainAfter) < 10 {
		t.Fatal("expected actual news-feed JSX removal")
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

func TestAnnouncementGuard(t *testing.T) {
	src := []byte("irrelevant component with no supported launch/news structure")
	if _, err := removeHomeAnnouncements(src); err == nil {
		t.Fatal("must refuse unrecognized homepage structure")
	}
}
