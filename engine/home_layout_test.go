package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestHomeToolsInsertionSurvivesDifferentJSXClosing(t *testing.T) {
	// This structurally valid variant uses a key on the toolbar button and
	// contains brackets in strings, so the old literal suffix cannot match.
	js := "dW=()=>{const home=(0,V.jsxs)(El,{children:[" +
		"(0,V.jsxs)(pU,{className:`lunar-patcher-hero`,background:u.src,onLoad:()=>{},children:[" +
		"(0,V.jsx)(A,{text:`[not an array]`,onOptionsClick:()=>{window.lunar.misc.switchRoute({route:`/versions`})}})," +
		"(0,V.jsxs)(Kc,{children:[(0,V.jsx)(gU,{}),(0,V.jsx)(aU,{key:`changed`})]})" +
		"]})" +
		"]}),(0,V.jsxs)(p,{width:`400px`,children:[(0,V.jsx)(iU,{full:n===!1,mode:O,rootRef:D})]})};return home}"
	if strings.Contains(js, "(0,V.jsx)(aU,{})]})]})]}),") {
		t.Fatal("test variant must defeat the old literal replacement")
	}
	patched, err := injectHomeTools332(js)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(patched, "(0,V.jsx)(lunarPatcherHomeTools,{})") != 1 {
		t.Fatalf("tools were not inserted exactly once: %s", patched)
	}
	if !strings.Contains(patched, "(0,V.jsx)(aU,{key:`changed`})") ||
		!strings.Contains(patched, "(0,V.jsx)(gU,{})") ||
		!strings.Contains(patched, "(0,V.jsxs)(p,{width:`400px`") {
		t.Fatal("home toolbar or friends column was lost")
	}
	bad := strings.Replace(js, "(0,V.jsx)(gU,{})", "(0,V.jsx)(unknownControl,{})", 1)
	if _, err := injectHomeTools332(bad); err == nil {
		t.Fatal("must refuse an unrecognized toolbar")
	}
}

func TestUploaded332HomeToolsInsertionVariation(t *testing.T) {
	raw, err := os.ReadFile("/mnt/data/app(3).asar")
	if os.IsNotExist(err) {
		t.Skip("user-supplied 3.7.32 fixture not mounted")
	}
	if err != nil {
		t.Fatal(err)
	}
	// The supplied archive is v8.1-patched; reverse its known ad/CSS changes
	// only within this test to recover an original renderer fixture.
	pristine, err := originalForTest(raw)
	if err != nil {
		t.Fatal(err)
	}
	archive, err := parseAsar(pristine)
	if err != nil {
		t.Fatal(err)
	}
	original, ok := archive.entry("dist/assets/main.js")
	if !ok {
		t.Fatal("missing main renderer")
	}
	old := []byte("(0,V.jsx)(aU,{})")
	changed := bytes.Replace(original, old, []byte("(0,V.jsx)(aU,{key:`launcher-update`})"), 1)
	if bytes.Equal(changed, original) {
		t.Fatal("test fixture missing old toolbar")
	}
	if bytes.Contains(changed, []byte("(0,V.jsx)(aU,{})]})]})]}),")) {
		t.Fatal("modified fixture must not match the fragile old signature")
	}
	result, err := patchHomeLayout(changed)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"lunarPatcherHomeTools=", "lunarPatcherUpdatesPage=",
		"(0,V.jsx)(iU,{full:!0,mode:`list`,rootRef:D})",
		"(0,V.jsx)(aU,{key:`launcher-update`})",
	} {
		if !bytes.Contains(result, []byte(expected)) {
			t.Errorf("missing %q", expected)
		}
	}
	if err := os.WriteFile("/mnt/data/LunarPatcher_v121_layout_regression.mjs", result, 0600); err != nil {
		t.Fatal(err)
	}
}
