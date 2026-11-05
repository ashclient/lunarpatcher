package main

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type PatchReport struct {
	Version       string   `json:"version"`
	Mode          string   `json:"mode"`
	Color         string   `json:"color"`
	Ads           string   `json:"ads"`
	Announcements string   `json:"announcements"`
	Theme         string   `json:"theme"`
	Notes         []string `json:"notes"`
}
type State struct {
	OriginalExe  string `json:"original_exe"`
	OriginalAsar string `json:"original_asar"`
	PatchedExe   string `json:"patched_exe"`
	PatchedAsar  string `json:"patched_asar"`
	Mode         string `json:"mode"`
	Color        string `json:"color"`
}

const marker = "LUNAR_ADAPTIVE_V8_1"

//go:embed assets/sakura_wallpaper.jpg
var sakuraWallpaper []byte

var integrityRecord = regexp.MustCompile(`\[\{"file":"resources\\\\app\.asar","alg":"SHA256","value":"([0-9a-f]{64})"\}\]`)
var guardPattern = regexp.MustCompile(`if\(!\([\w$]+&&[\w$]+\.willShowAd\)\)return null;`)

// A signature for the standalone news component, not the adjacent Launch
// toolbar. The previous implementation accidentally matched the Settings icon.
var headTag = regexp.MustCompile(`(?i)</head\s*>`)
var bodyTag = regexp.MustCompile(`(?i)</body\s*>`)
var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
var versionPattern = regexp.MustCompile(`"version"\s*:\s*"([^"]+)"`)

func getPaths(base string) (string, string, error) {
	exe := filepath.Join(base, "Lunar Client.exe")
	asar := filepath.Join(base, "resources", "app.asar")
	for _, p := range []string{exe, asar} {
		f, e := os.Stat(p)
		if e != nil {
			return "", "", fmt.Errorf("required file %s: %w", p, e)
		}
		if f.IsDir() {
			return "", "", fmt.Errorf("expected file: %s", p)
		}
	}
	return exe, asar, nil
}
func copyFileSecure(src, dst string) error {
	b, e := os.ReadFile(src)
	if e != nil {
		return e
	}
	return os.WriteFile(dst, b, 0600)
}
func writeAtomic(path string, data []byte) error {
	tmp := path + ".adaptive.tmp"
	if e := os.WriteFile(tmp, data, 0600); e != nil {
		return e
	}
	if e := os.Rename(tmp, path); e != nil {
		os.Remove(tmp)
		return e
	}
	return nil
}
func writeVerifiedBackup(path string, data []byte) error {
	if old, err := os.ReadFile(path); err == nil {
		if hashBytes(old) == hashBytes(data) {
			return nil
		}
		// Preserve a backup from the previous launcher version if Lunar updated.
		archive := path + ".prior." + hashBytes(old)[:12]
		if _, err := os.Stat(archive); os.IsNotExist(err) {
			if err = os.Rename(path, archive); err != nil {
				return err
			}
		} else {
			if err = os.Remove(path); err != nil {
				return err
			}
		}
	}
	return writeAtomic(path, data)
}

func savePair(exe, asar string, exeBytes, asarBytes []byte) error {
	// Two-phase file replacement with rollback, never touch unrelated files.
	nextExe := exe + ".adaptive.new"
	nextAsar := asar + ".adaptive.new"
	if e := os.WriteFile(nextExe, exeBytes, 0600); e != nil {
		return e
	}
	defer os.Remove(nextExe)
	if e := os.WriteFile(nextAsar, asarBytes, 0600); e != nil {
		return e
	}
	defer os.Remove(nextAsar)
	for _, p := range []string{exe + ".adaptive.old", asar + ".adaptive.old"} {
		if _, e := os.Stat(p); e == nil {
			return fmt.Errorf("interrupted prior install detected at %s; do not overwrite", p)
		}
	}
	if e := os.Rename(exe, exe+".adaptive.old"); e != nil {
		return fmt.Errorf("close Lunar Client before patching: %w", e)
	}
	if e := os.Rename(asar, asar+".adaptive.old"); e != nil {
		os.Rename(exe+".adaptive.old", exe)
		return e
	}
	rollback := func() {
		os.Remove(exe)
		os.Remove(asar)
		os.Rename(exe+".adaptive.old", exe)
		os.Rename(asar+".adaptive.old", asar)
	}
	if e := os.Rename(nextAsar, asar); e != nil {
		rollback()
		return e
	}
	if e := os.Rename(nextExe, exe); e != nil {
		rollback()
		return e
	}
	os.Remove(exe + ".adaptive.old")
	os.Remove(asar + ".adaptive.old")
	return nil
}
func statePaths(base string) (string, string, string, string) {
	d := filepath.Join(base, ".lunar-adaptive")
	return d, filepath.Join(d, "original.exe"), filepath.Join(d, "original.asar"), filepath.Join(d, "state.json")
}
func readState(base string) (*State, error) {
	_, _, _, p := statePaths(base)
	b, e := os.ReadFile(p)
	if os.IsNotExist(e) {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	s := new(State)
	if e = json.Unmarshal(b, s); e != nil {
		return nil, e
	}
	return s, nil
}
func integrityHash(exeBytes []byte) (string, error) {
	all := integrityRecord.FindAllSubmatch(exeBytes, -1)
	if len(all) != 1 {
		return "", fmt.Errorf("expected one Electron ASAR integrity record; found %d", len(all))
	}
	return string(all[0][1]), nil
}
func patchEXE(exeBytes []byte, old, new string) ([]byte, error) {
	oldToken := []byte(`[{"file":"resources\\app.asar","alg":"SHA256","value":"` + old + `"}]`)
	newToken := []byte(`[{"file":"resources\\app.asar","alg":"SHA256","value":"` + new + `"}]`)
	if bytes.Count(exeBytes, oldToken) != 1 {
		return nil, errors.New("integrity record no longer matches source archive")
	}
	return bytes.Replace(exeBytes, oldToken, newToken, 1), nil
}
func validatePair(exeBytes, asarBytes []byte) error {
	archive, e := parseAsar(asarBytes)
	if e != nil {
		return e
	}
	integrity, e := integrityHash(exeBytes)
	if e != nil {
		return e
	}
	if integrity != archive.headerHash {
		return fmt.Errorf("EXE expects ASAR header %s but archive has %s", integrity, archive.headerHash)
	}
	return nil
}
func parseVersion(a *asarArchive) string {
	if p, ok := a.entry("package.json"); ok {
		m := versionPattern.FindSubmatch(p)
		if len(m) > 1 {
			return string(m[1])
		}
	}
	return "unknown"
}

func patchSource(asarBytes []byte, mode, color string) ([]byte, PatchReport, error) {
	report := PatchReport{Mode: mode, Color: color}
	if mode != "color" {
		report.Color = ""
	}
	a, e := parseAsar(asarBytes)
	if e != nil {
		return nil, report, e
	}
	report.Version = parseVersion(a)
	htmlName := "dist/html/main.html"
	html, ok := a.entry(htmlName)
	if !ok {
		return nil, report, errors.New("renderer HTML has moved; unable to patch this build safely")
	}
	if !headTag.Match(html) || !bodyTag.Match(html) {
		return nil, report, errors.New("unexpected HTML structure; refusing patch")
	}
	if bytes.Contains(html, []byte(marker)) {
		return nil, report, errors.New("already patched: restore the original before trying again")
	}
	changes := map[string][]byte{}
	// The slot-renderer guard may be minified differently across versions. Patch only if unique.
	adFiles := []string{"dist/assets/use-show-ad.js"}
	slotPatched := false
	for _, name := range adFiles {
		b, exists := a.entry(name)
		if !exists {
			continue
		}
		matches := guardPattern.FindAllIndex(b, -1)
		if len(matches) == 1 {
			revised := guardPattern.ReplaceAll(b, []byte("if(true)return null;/* Lunar Adaptive: no ad slot */"))
			changes[name] = revised
			slotPatched = true
			report.Ads = "Ad renderer disabled"
			break
		}
	}
	// Intentionally fail closed. A generic HTML tweak by itself is NOT proof ads were removed.
	if !slotPatched {
		return nil, report, errors.New("ad-renderer signature changed in this build; no files were modified. Supply the new app.asar to update detection")
	}
	htmlOut := string(html)
	// Add scoped CSS; avoid altering Electron main or authentication/game-launch modules.
	css := makeCSS(mode, color)
	htmlOut = headTag.ReplaceAllString(htmlOut, "<style id=\"lunar-adaptive-css\">/* "+marker+" */"+css+"</style></head>")
	changes[htmlName] = []byte(htmlOut)
	// In Clean Layout and themed modes, the announcements below Launch are
	// removed in the renderer itself, rather than hidden with fragile CSS.
	if mode == "layout" || mode == "sakura" || mode == "color" {
		mainName := "dist/assets/main.js"
		main, exists := a.entry(mainName)
		if !exists {
			return nil, report, errors.New("main launcher renderer moved; homepage announcements cannot be removed safely")
		}
		revised, err := patchHomeLayout(main)
		if err != nil {
			return nil, report, err
		}
		changes[mainName] = revised
		report.Announcements = "News moved to Updates; Quick Play, shortcuts, and full friends list enabled"
	} else {
		report.Announcements = "Unchanged in Ads Only mode"
	}
	if mode == "layout" {
		report.Theme = "Full friends list, Quick Play, shortcuts and Updates"
	}
	if mode == "sakura" {
		report.Theme = "Sakura pink landscape"
		report.Notes = append(report.Notes, "Full friends list, Quick Play and separate Updates page with Sakura wallpaper and rosy surfaces")
	}
	if mode == "color" {
		report.Theme = "Custom Color"
		report.Notes = append(report.Notes, "The chosen color controls the launcher background and compatible panels, without Sakura artwork; full friends and shortcuts are included")
	}
	if mode == "ads" {
		report.Theme = "Original layout"
	}
	patched, e := a.rebuild(changes)
	return patched, report, e
}
func colorPalette(hex string) (string, string, string, string) {
	hex = strings.TrimPrefix(hex, "#")
	val := func(start int) int { v := 0; fmt.Sscanf(hex[start:start+2], "%02x", &v); return v }
	r, g, b := val(0), val(2), val(4)
	mix := func(amt float64, target int) string {
		rr := int(float64(r)*(1-amt) + float64(target)*amt)
		gg := int(float64(g)*(1-amt) + float64(target)*amt)
		bb := int(float64(b)*(1-amt) + float64(target)*amt)
		return fmt.Sprintf("#%02x%02x%02x", rr, gg, bb)
	}
	return fmt.Sprintf("#%02x%02x%02x", r, g, b), mix(.32, 0), mix(.56, 0), mix(.40, 255)
}

// Theme selectors are anchored to JSX classes inserted by patchHomeLayout.
// Styling a stable, explicitly marked component is more reliable than trying
// to theme every hashed class that Lunar uses internally.
func makeCSS(mode, color string) string {
	const adsCSS = `
#root .lunar-adaptive-ads,[data-ad-placement="Home"],[data-ad-placement="HomeTop"],iframe[src*="playwire"],iframe[src*="overwolf"],iframe[src*="doubleclick.net"]{display:none!important}
`
	const layoutCSS = `
#root .lunar-patcher-home{padding:0 12px 16px;box-sizing:border-box;border-radius:12px;min-width:0;min-height:0}
#root .lunar-patcher-home-content{min-height:0;min-width:0}
#root .lunar-patcher-home-tools{display:flex;flex-direction:column;gap:13px;margin:14px 0 8px;min-width:0}
#root .lunar-patcher-section{padding:14px;border-radius:11px;border:1px solid rgba(180,190,210,.16);background:rgba(29,32,41,.92);overflow:hidden;min-width:0}
#root .lunar-patcher-section>h3{color:#e9e9ee;font-size:13px;font-weight:650;line-height:1.3;margin:0 0 12px}
#root .lunar-patcher-friends{box-sizing:border-box!important;overflow-y:auto!important;overflow-x:hidden!important;padding:15px 13px!important;background:rgba(29,32,41,.93)!important;border-radius:12px!important;border:1px solid rgba(180,190,210,.16)!important}
#root .lunar-patcher-shortcuts{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:10px}
#root .lunar-patcher-shortcut{font-family:inherit;font-size:13px;font-weight:600;line-height:1.2;min-height:44px;padding:10px 8px;border-radius:8px;color:#ebebf0;border:1px solid rgba(190,192,211,.18);background:rgba(110,121,146,.14);cursor:pointer}
#root .lunar-patcher-shortcut:hover{background:rgba(152,161,191,.25);border-color:rgba(192,206,228,.4)}
#root .lunar-patcher-updates{height:100%;overflow-y:auto;display:flex;flex-direction:column;gap:14px;padding:20px 24px}
#root .lunar-patcher-updates>h1{font-size:22px;font-weight:700;margin:0;color:inherit}
#root button:focus-visible{outline:2px solid #e5a3c5;outline-offset:2px}
@media(max-width:1000px){#root .lunar-patcher-friends{width:310px!important}#root .lunar-patcher-shortcuts{grid-template-columns:repeat(3,minmax(0,1fr))}}
/* Use a real two-column grid, not a 12-track grid with an oversized friends column. */
#root .lunar-patcher-home{height:100%;overflow:hidden;padding:15px 17px 17px!important;box-sizing:border-box}
#root .lunar-patcher-home-content{display:grid!important;grid-template-columns:minmax(0,1fr) minmax(285px,350px)!important;grid-template-rows:minmax(0,1fr)!important;gap:20px!important;width:100% !important;height:100% !important;min-height:0!important;overflow:hidden!important;align-items:stretch!important}
#root .lunar-patcher-home-content>*:first-child{grid-column:1!important;grid-row:1!important;min-width:0!important;min-height:0!important;max-height:100% !important;width:100% !important;margin-right:0!important;overflow-y:auto!important;overflow-x:hidden!important;padding-right:3px!important;scrollbar-gutter:stable}
#root .lunar-patcher-home-content>.lunar-patcher-friends{grid-column:2!important;grid-row:1!important;min-width:0!important;min-height:0!important;max-height:100% !important;width:100% !important;height:100% !important;padding:15px!important;overflow-y:auto!important;overscroll-behavior:contain}
#root .lunar-patcher-home-tools{gap:12px!important;margin:0 0 8px!important;min-width:0!important;flex-shrink:0!important}
#root .lunar-patcher-hero{flex-shrink:0!important;min-width:0!important}
#root .lunar-patcher-section{width:100%;box-sizing:border-box!important;flex-shrink:0!important}
#root .lunar-patcher-shortcuts{gap:10px!important}
#root .lunar-patcher-shortcut{min-width:0!important;white-space:normal!important}
@media(max-width:1070px){#root .lunar-patcher-home-content{grid-template-columns:minmax(0,1fr) minmax(260px,315px)!important;gap:13px!important}#root .lunar-patcher-home{padding:10px!important}}
@media(max-width:820px){#root .lunar-patcher-home-content{grid-template-columns:minmax(0,1fr) 245px!important;gap:10px!important}#root .lunar-patcher-shortcuts{gap:6px!important}#root .lunar-patcher-shortcut{font-size:11px!important;padding:8px 5px!important}}
@media(prefers-reduced-motion:reduce){#root .lunar-patcher-shortcut{transition:none!important}}
`
	if mode == "ads" {
		return adsCSS
	}
	if mode == "layout" {
		return adsCSS + layoutCSS
	}
	if mode == "sakura" {
		artwork := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(sakuraWallpaper)
		return adsCSS + layoutCSS + fmt.Sprintf(`
html.dark,html.dark body{background:#8b4c74!important}
#root{--colors-space-1:#783d60!important;--colors-space-2:#85476b!important;--colors-space-3:#985377!important;--colors-space-4:#ad6489!important;--colors-midnight-1:#683151!important;--colors-space-a4:rgba(246,154,195,.25)!important;--lunar-sakura-wallpaper:url("%s");background:linear-gradient(120deg,rgba(103,42,80,.17),rgba(206,115,167,.17)),var(--lunar-sakura-wallpaper) center/cover no-repeat fixed!important}
#root>div{background-color:transparent!important}
#root .lunar-patcher-home{background:linear-gradient(145deg,rgba(79,27,62,.35),rgba(79,31,65,.27)),var(--lunar-sakura-wallpaper) center center/cover no-repeat!important;box-shadow:inset 0 0 0 1px rgba(254,188,221,.25)!important}
#root .lunar-patcher-home-content{background:transparent!important}
#root .lunar-patcher-hero{position:relative!important;overflow:hidden!important;border:1px solid rgba(252,186,220,.55)!important;isolation:isolate}
#root .lunar-patcher-hero::after{content:"";position:absolute;inset:0;pointer-events:none;background:var(--lunar-sakura-wallpaper) center/cover no-repeat;opacity:.33;mix-blend-mode:screen;z-index:1}
#root .lunar-patcher-friends,#root .lunar-patcher-section{background:linear-gradient(150deg,rgba(111,49,84,.9),rgba(137,66,104,.87))!important;border-color:rgba(255,184,218,.40)!important;backdrop-filter:blur(10px)!important}
#root .lunar-patcher-section>h3{color:#ffe5f2}
#root .lunar-patcher-shortcut{color:#ffeef7;background:rgba(235,147,188,.18);border-color:rgba(254,202,226,.30)}
#root .lunar-patcher-shortcut:hover{background:rgba(248,171,211,.3);border-color:#fbcce3}
#root .lunar-patcher-updates{background:linear-gradient(145deg,rgba(116,54,94,.85),rgba(75,30,64,.87))!important}
#root ::-webkit-scrollbar-thumb{background:rgba(255,177,216,.54)!important}

/* Sakura refinement: layered rose panels and accessible contrast. */
html.dark,html.dark body{background:#261727!important}
#root{--colors-space-1:#302034!important;--colors-space-2:#3e2942!important;--colors-space-3:#503150!important;--colors-space-4:#67395f!important;--colors-midnight-1:#211521!important;--colors-space-a4:rgba(242,150,192,.19)!important;background:linear-gradient(135deg,rgba(24,15,32,.56),rgba(78,27,60,.42)),var(--lunar-sakura-wallpaper) center/cover no-repeat fixed!important}
#root .lunar-patcher-home{position:relative!important;isolation:isolate;background:linear-gradient(120deg,rgba(45,24,47,.89),rgba(51,23,48,.79)),var(--lunar-sakura-wallpaper) center/cover no-repeat!important;border:1px solid rgba(255,181,212,.28)!important;box-shadow:inset 0 1px 0 rgba(255,210,231,.12),0 12px 42px rgba(15,5,26,.24)!important}
#root .lunar-patcher-home-content{position:relative!important;z-index:1!important;background:transparent!important}
#root .lunar-patcher-hero{position:relative!important;overflow:hidden!important;border:1px solid rgba(255,164,207,.52)!important;isolation:isolate!important;box-shadow:0 8px 26px rgba(15,5,30,.3)!important}
#root .lunar-patcher-hero::after{content:"";position:absolute;inset:0;pointer-events:none;background:linear-gradient(120deg,rgba(129,42,101,.18),rgba(250,139,188,.06)),var(--lunar-sakura-wallpaper) center/cover no-repeat;opacity:.28;mix-blend-mode:screen;z-index:1}
#root .lunar-patcher-friends,#root .lunar-patcher-section{background:linear-gradient(145deg,rgba(65,34,62,.95),rgba(81,42,75,.93))!important;border-color:rgba(255,174,214,.38)!important;box-shadow:inset 0 1px 0 rgba(255,204,226,.07),0 8px 20px rgba(20,7,21,.14)!important;backdrop-filter:blur(10px)!important}
#root .lunar-patcher-section>h3{color:#ffe1f0!important;font-weight:700!important}
#root .lunar-patcher-shortcut{color:#fff0f8!important;background:linear-gradient(155deg,rgba(228,122,180,.25),rgba(146,70,120,.23))!important;border-color:rgba(255,185,219,.35)!important}
#root .lunar-patcher-shortcut:hover{background:rgba(243,154,199,.39)!important;border-color:#ffd3e8!important}
#root .lunar-patcher-updates{background:linear-gradient(130deg,rgba(77,35,69,.96),rgba(36,22,43,.98))!important}
#root .lunar-patcher-updates>h1{color:#ffdbea!important}
#root ::-webkit-scrollbar-thumb{background:rgba(241,152,198,.51)!important}
#root button:focus-visible{outline-color:#fbb5d4!important}
`, artwork)
	}
	if mode == "color" {
		base, dark, deep, light := colorPalette(color)
		return adsCSS + layoutCSS + fmt.Sprintf(`
html.dark,html.dark body{background:%s!important}
#root{--colors-space-1:%s!important;--colors-space-2:%s!important;--colors-space-3:%s!important;--colors-space-4:%s!important;--colors-midnight-1:%s!important;--colors-space-a4:%s!important;background:linear-gradient(145deg,%s,%s)!important}
#root>div{background-color:transparent!important}
#root .lunar-patcher-home{background:linear-gradient(145deg,%s,%s)!important;box-shadow:inset 0 0 0 1px %s66!important}
#root .lunar-patcher-friends,#root .lunar-patcher-section{background:%sDD!important;border-color:%s88!important}
#root .lunar-patcher-shortcut{background:%s55!important;border-color:%s88!important;color:#ffffff}
#root .lunar-patcher-shortcut:hover{background:%s99!important}
#root .lunar-patcher-updates{background:%s!important}
#root button:focus-visible{outline:2px solid %s!important}
`, base, deep, dark, base, deep, deep, base, base, dark, base, dark, light, deep, light, base, light, deep, light, light)
	}
	return adsCSS
}

func Install(base, mode, color string, checkOnly bool) (PatchReport, error) {
	report := PatchReport{Mode: mode, Color: color}
	if mode != "color" {
		report.Color = ""
	}
	if mode != "ads" && mode != "layout" && mode != "sakura" && mode != "color" {
		return report, errors.New("invalid patch mode")
	}
	if !hexColor.MatchString(color) {
		return report, errors.New("color must be #RRGGBB")
	}
	exe, asar, e := getPaths(base)
	if e != nil {
		return report, e
	}
	curExe, e := os.ReadFile(exe)
	if e != nil {
		return report, e
	}
	curAsar, e := os.ReadFile(asar)
	if e != nil {
		return report, e
	}
	st, e := readState(base)
	if e != nil {
		return report, e
	}
	var originalExe, originalAsar []byte
	if st != nil && hashBytes(curExe) == st.PatchedExe && hashBytes(curAsar) == st.PatchedAsar {
		_, backupExe, backupAsar, _ := statePaths(base)
		originalExe, e = os.ReadFile(backupExe)
		if e != nil {
			return report, fmt.Errorf("backed-up executable missing: %w", e)
		}
		originalAsar, e = os.ReadFile(backupAsar)
		if e != nil {
			return report, e
		}
		if hashBytes(originalExe) != st.OriginalExe || hashBytes(originalAsar) != st.OriginalAsar {
			return report, errors.New("backups do not match state: refusing to continue")
		}
	} else {
		// Older v2-v7 patchers left backups next to the executable and archive.
		// Prefer a verified original from those backups only if the current EXE is
		// byte-identical to it after reversing the SINGLE integrity hash change.
		// Otherwise accept a new stock version after the official updater runs.
		legacyExe := exe + ".lunarclean.bak"
		legacyAsar := asar + ".lunarclean.bak"
		candidateExe, ee := os.ReadFile(legacyExe)
		candidateAsar, ea := os.ReadFile(legacyAsar)
		legacyMatch := false
		if ee == nil && ea == nil && validatePair(candidateExe, candidateAsar) == nil {
			originalHdr, _ := integrityHash(candidateExe)
			currentHdr, er := integrityHash(curExe)
			if er == nil {
				reverted, er := patchEXE(curExe, currentHdr, originalHdr)
				legacyMatch = er == nil && bytes.Equal(reverted, candidateExe) &&
					(hashBytes(curAsar) != hashBytes(candidateAsar))
			}
		}
		if legacyMatch {
			originalExe, originalAsar = candidateExe, candidateAsar
		} else if e := validatePair(curExe, curAsar); e != nil {
			return report, fmt.Errorf("installed files have unexpected integrity records and no compatible original backup (%w). Restore with the previous patcher first", e)
		} else {
			originalExe, originalAsar = curExe, curAsar
		}
	}
	if e := validatePair(originalExe, originalAsar); e != nil {
		return report, e
	}
	newAsar, report, e := patchSource(originalAsar, mode, color)
	if e != nil {
		return report, e
	}
	oldAsarParsed, _ := parseAsar(originalAsar)
	newAsarParsed, _ := parseAsar(newAsar)
	newExe, e := patchEXE(originalExe, oldAsarParsed.headerHash, newAsarParsed.headerHash)
	if e != nil {
		return report, e
	}
	if e := validatePair(newExe, newAsar); e != nil {
		return report, e
	}
	if checkOnly {
		return report, nil
	}
	stateDir, backupExe, backupAsar, statePath := statePaths(base)
	if e := os.MkdirAll(stateDir, 0700); e != nil {
		return report, e
	}
	// Backup data from the source version, not from the current modified installation.
	if e := writeVerifiedBackup(backupExe, originalExe); e != nil {
		return report, e
	}
	if e := writeVerifiedBackup(backupAsar, originalAsar); e != nil {
		return report, e
	}
	if e := savePair(exe, asar, newExe, newAsar); e != nil {
		return report, e
	}
	s := State{OriginalExe: hashBytes(originalExe), OriginalAsar: hashBytes(originalAsar), PatchedExe: hashBytes(newExe), PatchedAsar: hashBytes(newAsar), Mode: mode, Color: color}
	sb, _ := json.MarshalIndent(s, "", "  ")
	if e := writeAtomic(statePath, sb); e != nil {
		return report, fmt.Errorf("patch applied but state could not be saved: %w", e)
	}
	return report, nil
}
func Restore(base string) error {
	exe, asar, e := getPaths(base)
	if e != nil {
		return e
	}
	st, e := readState(base)
	if e != nil {
		return e
	}
	if st == nil {
		return errors.New("no adaptive-patcher backup found; use your previous installer's Restore Original function")
	}
	originalExePath, originalAsarPath := "", ""
	_, originalExePath, originalAsarPath, _ = statePaths(base)
	originalExe, e := os.ReadFile(originalExePath)
	if e != nil {
		return e
	}
	originalAsar, e := os.ReadFile(originalAsarPath)
	if e != nil {
		return e
	}
	if hashBytes(originalExe) != st.OriginalExe || hashBytes(originalAsar) != st.OriginalAsar {
		return errors.New("backup failed verification")
	}
	if e := validatePair(originalExe, originalAsar); e != nil {
		return e
	}
	currentExe, e := os.ReadFile(exe)
	if e != nil {
		return e
	}
	currentAsar, e := os.ReadFile(asar)
	if e != nil {
		return e
	}
	if hashBytes(currentExe) == st.OriginalExe && hashBytes(currentAsar) == st.OriginalAsar {
		return nil
	}
	if hashBytes(currentExe) != st.PatchedExe || hashBytes(currentAsar) != st.PatchedAsar {
		return errors.New("Lunar has updated or changed since patching; refusing to replace newer installed files")
	}
	if e := savePair(exe, asar, originalExe, originalAsar); e != nil {
		return e
	}
	// Preserve verified backup for safety; mark restored state by deleting state.json only.
	_, _, _, statePath := statePaths(base)
	return os.Remove(statePath)
}
