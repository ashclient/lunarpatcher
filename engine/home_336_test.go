package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// Integration fixture: supplied 3.7.36-ow ASAR (not redistributed).
func TestUploaded3_7_36(t *testing.T) {
	raw, err := os.ReadFile("/mnt/data/app(5).asar")
	if os.IsNotExist(err) {
		t.Skip("uploaded 3.7.36 fixture unavailable")
	}
	if err != nil {
		t.Fatal(err)
	}
	a, err := parseAsar(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got := parseVersion(a); got != "3.7.36-ow" {
		t.Fatalf("fixture version %s", got)
	}
	oldMain, _ := a.entry("dist/assets/main.js")
	oldAd, _ := a.entry("dist/assets/use-show-ad.js")
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
				if report.Version != "3.7.36-ow" {
					t.Fatal(report.Version)
				}
				p, err := parseAsar(patched)
				if err != nil {
					t.Fatal(err)
				}
				ad, _ := p.entry("dist/assets/use-show-ad.js")
				if bytes.Equal(ad, oldAd) || !bytes.Contains(ad, []byte("if(true)return null;")) {
					t.Fatal("ad guard still enabled")
				}
				html, _ := p.entry("dist/html/main.html")
				if !bytes.Contains(html, []byte(marker)) {
					t.Fatal("patch marker missing")
				}
				main, _ := p.entry("dist/assets/main.js")
				if mode != "ads" {
					for _, token := range []string{"lunarPatcherHomeTools=", "lunarPatcherUpdatesPage=", "className:`lunar-patcher-friends`", "className:`lunar-patcher-hero`", "(0,V.jsx)(GK,{})", "(0,V.jsx)(JG,{full:!0,mode:`list`,rootRef:E})", "path:`/updates`"} {
						if !bytes.Contains(main, []byte(token)) {
							t.Errorf("missing %s", token)
						}
					}
					if bytes.Contains(main, []byte("(0,V.jsx)(wf,{...C.ad.main")) {
						t.Error("home ad container retained")
					}
					if strings.Count(string(main), "(0,V.jsx)(yK,{})") != 1 {
						t.Error("news should appear once on Updates")
					}
				} else if !roblox && !bytes.Equal(main, oldMain) {
					t.Error("Ads Only changed home renderer")
				}
				if roblox {
					if !bytes.Contains(main, []byte("lunarPatcherRobloxPage=")) {
						t.Error("Roblox page missing")
					}
					for _, file := range []string{"dist-electron/electron/main.js", "dist-electron/preload/launcher/index.js"} {
						b, _ := p.entry(file)
						if !bytes.Contains(b, []byte("lunarpatcher:roblox:show")) && !bytes.Contains(b, []byte("lunarPatcherRobloxBrowser")) {
							t.Error("Roblox bridge missing", file)
						}
					}
				}
				if mode == "sakura" && !bytes.Contains(html, []byte("data:image/jpeg;base64")) {
					t.Error("Sakura wallpaper missing")
				}
				if mode == "color" && bytes.Contains(html, []byte("data:image/jpeg;base64")) {
					t.Error("Custom Color has Sakura wallpaper")
				}
				if mode == "sakura" && roblox {
					for _, file := range []string{"dist/assets/main.js", "dist/assets/use-show-ad.js", "dist-electron/electron/main.js", "dist-electron/preload/launcher/index.js"} {
						b, _ := p.entry(file)
						name := strings.ReplaceAll(file, "/", "_")
						if err := os.WriteFile("/mnt/data/lunarpatcher_v173/"+name+"_patched.js", b, 0600); err != nil {
							t.Fatal(err)
						}
					}
				}
			})
		}
	}
}
func Test336ChangedSignaturesFailClosed(t *testing.T) {
	raw, err := os.ReadFile("/mnt/data/app(5).asar")
	if os.IsNotExist(err) {
		t.Skip("3.7.36 fixture unavailable")
	}
	if err != nil {
		t.Fatal(err)
	}
	a, err := parseAsar(raw)
	if err != nil {
		t.Fatal(err)
	}
	oldMain, _ := a.entry("dist/assets/main.js")
	before := []byte("(0,V.jsx)(JG,{full:n===!1,mode:D,rootRef:E})")
	if bytes.Count(oldMain, before) != 1 {
		t.Fatal("expected one friends signature")
	}
	broken := bytes.Replace(oldMain, before, []byte("(0,V.jsx)(JG,{full:n===!1,mode:Z,rootRef:E})"), 1)
	asar, err := a.rebuild(map[string][]byte{"dist/assets/main.js": broken})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := patchSource(asar, "layout", "#ECA3C4"); err == nil || !strings.Contains(err.Error(), "signature") {
		t.Fatalf("unexpected friends change accepted: %v", err)
	}
	oldAd, _ := a.entry("dist/assets/use-show-ad.js")
	guard := []byte("if(!a||!e.willShowAd)return null;")
	if bytes.Count(oldAd, guard) != 1 {
		t.Fatal("expected one ad signature")
	}
	ad := bytes.Replace(oldAd, guard, []byte("if(!a||e.willShowAd)return null;"), 1)
	asar, err = a.rebuild(map[string][]byte{"dist/assets/use-show-ad.js": ad})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := patchSource(asar, "ads", "#ECA3C4"); err == nil || !strings.Contains(err.Error(), "ad-renderer signature") {
		t.Fatalf("unexpected ad guard accepted: %v", err)
	}
}

// A small independently written skeleton makes sure CI exercises the 3.7.36
// adapter even without the user-supplied proprietary ASAR fixture.
func Test336SyntheticHome(t *testing.T) {
	src := "const sidebar=[{type:`external`,id:`news`,icon:Kee,tooltip:e(`layout.sidebar.news.tooltip`),url:`https://www.lunarclient.com/news`}]," +
		"routes=[(0,V.jsx)(A,{element:(0,V.jsx)(tq,{}),path:`/`})]," +
		"tq=()=>{return(0,V.jsx)(g,{display:`flex`,flexDirection:`column`,flexGrow:1,gap:`4`,minHeight:`0`,children:(0,V.jsxs)(Js,{children:[" +
		"(0,V.jsxs)(rK,{background:u.src,onLoad:()=>{},children:[]})," +
		"(0,V.jsx)(yK,{})," +
		"(0,V.jsxs)(g,{css:{\"&&::-webkit-scrollbar\":{display:`none`}},display:`flex`,flexDirection:`column`,flexShrink:0,gridColumn:`span 2`,height:`100%`,children:[" +
		"(0,V.jsx)(GK,{}),(0,V.jsx)(JG,{full:n===!1,mode:D,rootRef:E})," +
		"(0,V.jsx)(wf,{...C.ad.secondary}),(0,V.jsx)(wf,{...C.ad.main})]})]})})}"
	output, err := patchHome336(src)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(output), "(0,V.jsx)(wf,{...C.ad.main") {
		t.Fatal("ad still present")
	}
	for _, token := range []string{"lunarPatcherHomeTools=", "lunarPatcherUpdatesPage=", "(0,V.jsx)(GK,{})", "(0,V.jsx)(JG,{full:!0,mode:`list`,rootRef:E})"} {
		if !strings.Contains(string(output), token) {
			t.Fatalf("missing %s", token)
		}
	}
}
