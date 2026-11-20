package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestUploaded340AllModes(t *testing.T) {
	raw, err := os.ReadFile("/mnt/data/app(8).asar")
	if os.IsNotExist(err) {
		t.Skip("3.7.40 fixture missing in CI")
	}
	if err != nil {
		t.Fatal(err)
	}
	a, err := parseAsar(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got := parseVersion(a); got != "3.7.40-ow" {
		t.Fatal(got)
	}
	oldMain, _ := a.entry("dist/assets/main.js")
	oldAd, _ := a.entry("dist/assets/use-page-search-keybind.js")
	for _, mode := range []string{"ads", "layout", "sakura", "color"} {
		for _, roblox := range []bool{false, true} {
			t.Run(mode+map[bool]string{true: "+roblox", false: ""}[roblox], func(t *testing.T) {
				patched, report, err := patchSource(raw, mode, "#EBA2C7", roblox)
				if err != nil {
					t.Fatal(err)
				}
				if report.Version != "3.7.40-ow" {
					t.Fatal(report.Version)
				}
				b, err := parseAsar(patched)
				if err != nil {
					t.Fatal(err)
				}
				ad, _ := b.entry("dist/assets/use-page-search-keybind.js")
				if bytes.Equal(ad, oldAd) || !bytes.Contains(ad, []byte("if(true)return null")) {
					t.Error("ad guard unchanged")
				}
				html, _ := b.entry("dist/html/main.html")
				if !bytes.Contains(html, []byte(marker)) {
					t.Error("marker absent")
				}
				main, _ := b.entry("dist/assets/main.js")
				if mode != "ads" {
					for _, tok := range []string{"lunarPatcherHomeTools=", "lunarPatcherUpdatesPage=", "className:`lunar-patcher-friends`", "className:`lunar-patcher-hero`", "(0,Z.jsx)(jm,{})", "(0,Z.jsx)(yp,{full:!0,mode:`list`,rootRef:b})"} {
						if !bytes.Contains(main, []byte(tok)) {
							t.Errorf("missing %s", tok)
						}
					}
					if bytes.Contains(main, []byte("(0,Z.jsx)(Ha,{..._.ad.main")) {
						t.Error("ads remain")
					}
				} else if !roblox && !bytes.Equal(oldMain, main) {
					t.Error("ads mode changed homepage")
				}
				if roblox && !bytes.Contains(main, []byte("lunarPatcherRobloxPage=")) {
					t.Error("roblox absent")
				}
				if mode == "sakura" && !bytes.Contains(html, []byte("data:image/jpeg;base64")) {
					t.Error("sakura absent")
				}
				if mode == "color" && bytes.Contains(html, []byte("data:image/jpeg;base64")) {
					t.Error("color copied sakura")
				}
			})
		}
	}
}
func Test340RejectedChanges(t *testing.T) {
	raw, err := os.ReadFile("/mnt/data/app(8).asar")
	if os.IsNotExist(err) {
		t.Skip("fixture unavailable")
	}
	if err != nil {
		t.Fatal(err)
	}
	a, err := parseAsar(raw)
	if err != nil {
		t.Fatal(err)
	}
	main, _ := a.entry("dist/assets/main.js")
	for _, tok := range []string{"(0,Z.jsx)(yp,{full:t===!1,mode:x,rootRef:b})", "(0,Z.jsx)(Ha,{..._.ad.main", "(0,Z.jsx)(jm,{})", "(0,Z.jsx)(Zp,{})"} {
		if strings.Count(string(main), tok) != 1 {
			t.Fatalf("fixture signature unexpected: %s", tok)
		}
		changed := strings.Replace(string(main), tok, "DAMAGE", 1)
		if b, e := patchHome340(changed); e == nil || len(b) != 0 {
			t.Fatalf("patch should reject %s", tok)
		}
	}
}

func Test340SyntheticHomeAndGuard(t *testing.T) {
	src := "nav=[{type:`external`,id:`news`,icon:yt,tooltip:e(`layout.sidebar.news.tooltip`),url:`https://www.lunarclient.com/news`}]," +
		"routes=[(0,Z.jsx)(D,{element:(0,Z.jsx)(Fm,{}),path:`/`})]," +
		"Fm=()=>{return(0,Z.jsx)(I,{display:`flex`,flexDirection:`column`,flexGrow:1,gap:`4`,minHeight:`0`,children:(0,Z.jsxs)(Oo,{children:[" +
		"(0,Z.jsxs)(Lp,{background:l.src,onLoad:()=>{},children:[]})," +
		"(0,Z.jsx)(Zp,{})," +
		"(0,Z.jsxs)(I,{css:{\"&&::-webkit-scrollbar\":{display:`none`}},display:`flex`,flexDirection:`column`,flexShrink:0,gridColumn:`span 2`,height:`100%`,children:[" +
		"(0,Z.jsx)(jm,{}),(0,Z.jsx)(yp,{full:t===!1,mode:x,rootRef:b})," +
		"(0,Z.jsx)(Ha,{..._.ad.secondary}),(0,Z.jsx)(Ha,{..._.ad.main})]})]})})}"
	patched, err := patchHome340(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, tok := range []string{"lunarPatcherHomeTools=", "lunarPatcherUpdatesPage=", "className:`lunar-patcher-home`", "className:`lunar-patcher-friends`", "(0,Z.jsx)(Zp,{})"} {
		if !bytes.Contains(patched, []byte(tok)) {
			t.Errorf("synthetic patch missing %s", tok)
		}
	}
	for _, tok := range []string{"(0,Z.jsx)(yp,{full:t===!1,mode:x,rootRef:b})", "(0,Z.jsx)(jm,{})", "(0,Z.jsx)(Zp,{})"} {
		damaged := strings.Replace(src, tok, "unexpected", 1)
		got, e := patchHome340(damaged)
		if e == nil || len(got) != 0 {
			t.Errorf("synthetic patch accepted corrupted signature %s", tok)
		}
	}
}

func Test340UnknownAdGuardFailsClosed(t *testing.T) {
	raw, err := os.ReadFile("/mnt/data/app(8).asar")
	if os.IsNotExist(err) {
		t.Skip("3.7.40 fixture unavailable")
	}
	if err != nil {
		t.Fatal(err)
	}
	a, err := parseAsar(raw)
	if err != nil {
		t.Fatal(err)
	}
	const path = "dist/assets/use-page-search-keybind.js"
	guard, _ := a.entry(path)
	original := []byte("if(!a||!e.willShowAd)return null;")
	if bytes.Count(guard, original) != 1 {
		t.Fatal("fixture ad guard changed")
	}
	damaged := bytes.Replace(guard, original, []byte("if(mysteryGate())return null;"), 1)
	rebuilt, err := a.rebuild(map[string][]byte{path: damaged})
	if err != nil {
		t.Fatal(err)
	}
	got, _, err := patchSource(rebuilt, "layout", "#EBA2C7", false)
	if err == nil || len(got) != 0 {
		t.Fatal("unknown ad renderer should refuse entire patch")
	}
}
