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
var newsFunctionStart = regexp.MustCompile(`[A-Za-z_$][\w$]*=\(\)=>\{let\{t:[A-Za-z_$][\w$]*\}=[A-Za-z_$][\w$]*\(\),`)
var newsFunctionEnd = regexp.MustCompile(`\},[A-Za-z_$][\w$]*=[A-Za-z_$][\w$]*\(\{base:\{backgroundOrigin:\x60border-box\x60`)
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

// removeHomeAnnouncements disables exactly the blog/news component with
// metadata.blogPosts and home.latest-news. It must never touch the settings,
// cancel-launch or friends controls adjacent to the launch button.
func removeHomeAnnouncements(src []byte) ([]byte, error) {
	key := []byte("e=>e.metadata.blogPosts")
	if bytes.Count(src, key) != 1 {
		return nil, errors.New("news component signature missing or ambiguous; no patch applied")
	}
	p := bytes.Index(src, key)
	lo := max(0, p-260)
	candidates := newsFunctionStart.FindAllIndex(src[lo:p], -1)
	if len(candidates) != 1 {
		return nil, fmt.Errorf("news component start changed (%d candidates)", len(candidates))
	}
	begin := lo + candidates[0][0]
	hi := min(len(src), p+8000)
	// The following status-icon component is a stable, unique boundary in
	// both tested bundles. Anchor against it instead of guessing JSX delimiters.
	ends := newsFunctionEnd.FindAllIndex(src[p:hi], -1)
	if len(ends) != 1 {
		return nil, fmt.Errorf("news component boundary changed (%d candidates)", len(ends))
	}
	// Keep the following comma and next variable assignment intact.
	end := p + ends[0][0] + 1
	segment := src[begin:end]
	if !bytes.Contains(segment, []byte("home.latest-news")) ||
		!bytes.Contains(segment, []byte("home_blog_post")) ||
		!bytes.Contains(segment, []byte("home_news")) ||
		len(segment) < 1000 || len(segment) > 7500 {
		return nil, errors.New("news component safeguards did not match; refusing patch")
	}
	lhs := bytes.Index(segment, []byte("=()=>"))
	if lhs < 1 || lhs > 80 {
		return nil, errors.New("unexpected news component declaration")
	}
	replacement := append(append([]byte{}, segment[:lhs]...), []byte("=()=>null")...)
	out := make([]byte, 0, len(src)-len(segment)+len(replacement))
	out = append(out, src[:begin]...)
	out = append(out, replacement...)
	out = append(out, src[end:]...)
	return out, nil
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
		revised, err := removeHomeAnnouncements(main)
		if err != nil {
			return nil, report, err
		}
		changes[mainName] = revised
		report.Announcements = "News-feed component disabled"
	} else {
		report.Announcements = "Unchanged in Ads Only mode"
	}
	if mode == "layout" {
		report.Theme = "Clean homepage and layout styling"
		report.Notes = append(report.Notes, "Full friends panel and Updates route still require a build-specific React patch")
	}
	if mode == "sakura" {
		report.Theme = "Sakura pink landscape"
		report.Notes = append(report.Notes, "Fixed pink cherry-blossom wallpaper and matching panels. Choose Custom Color for a plain color-based theme")
	}
	if mode == "color" {
		report.Theme = "Custom Color"
		report.Notes = append(report.Notes, "The chosen color controls the launcher background and compatible panels, without Sakura artwork")
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
func makeCSS(mode, color string) string {
	basic := `#root .lunar-adaptive-ads,[data-ad-placement="Home"],[data-ad-placement="HomeTop"],iframe[src*="playwire"],iframe[src*="overwolf"],iframe[src*="doubleclick.net"]{display:none!important}`
	switch mode {
	case "ads":
		return basic
	case "layout":
		return basic + ` #root{--lunar-adaptive-radius:10px} #root button{transition:background-color .15s ease} #root button:focus-visible{outline:2px solid #cf88b4;outline-offset:2px}`
	case "sakura":
		// Sakura remains a fixed rose-pink landscape. The custom color is NOT used.
		return basic + fmt.Sprintf(`
html.dark,html.dark body{background:#b96892!important}
#root{--colors-space-1:#8c4670!important;--colors-space-2:#99537b!important;--colors-space-3:#ab628a!important;--colors-space-4:#b97298!important;--colors-midnight-1:#713254!important;--colors-space-a4:#d783b0!important;
 background:linear-gradient(135deg,rgba(221,105,164,.28),rgba(106,43,85,.31)),url("data:image/jpeg;base64,%s") center center/cover no-repeat fixed!important;color:#fff4f9}
#root>div{background-color:transparent!important}
#root [class*="sidebar"],#root [class*="Sidebar"],#root [class*="friends"],#root [class*="Friends"],#root .lunar-clean-friends,#root .lunar-clean-quick-play,#root .lunar-clean-shortcuts{background-color:rgba(151,68,110,.78)!important;border-color:rgba(255,191,224,.55)!important}
#root [class*="Card"],#root [class*="panel"],#root [class*="Panel"]{border-color:rgba(255,188,220,.43)}
#root button:focus-visible{outline:2px solid #ffc9e2!important;outline-offset:2px}
#root ::-webkit-scrollbar-thumb{background:rgba(255,181,221,.55)!important}
`, base64.StdEncoding.EncodeToString(sakuraWallpaper))
	case "color":
		// Color mode has NO sakura landscape: a pure palette-based gradient.
		base, dark, deep, light := colorPalette(color)
		return basic + fmt.Sprintf(`
html.dark,html.dark body{background:%s!important}
#root{--colors-space-1:%s!important;--colors-space-2:%s!important;--colors-space-3:%s!important;--colors-space-4:%s!important;--colors-midnight-1:%s!important;--colors-space-a4:%s!important;
 background:linear-gradient(145deg,%s 0%%,%s 50%%,%s 100%%) center center/cover no-repeat fixed!important;color:#fff4f9}
#root>div{background-color:transparent!important}
#root [class*="sidebar"],#root [class*="Sidebar"],#root [class*="friends"],#root [class*="Friends"],#root .lunar-clean-friends,#root .lunar-clean-quick-play,#root .lunar-clean-shortcuts{background-color:%sE6!important;border-color:%s88!important}
#root [class*="Card"],#root [class*="panel"],#root [class*="Panel"]{border-color:%s88}
#root button:focus-visible{outline:2px solid %s!important;outline-offset:2px}
#root ::-webkit-scrollbar-thumb{background:%s!important}
`, base, deep, dark, base, deep, base, light, base, dark, deep, dark, light, light, light, light)
	default:
		return basic
	}
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
