package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test338SyntheticHome(t *testing.T) {
	const src = "nav=[{type:`external`,id:`news`,icon:Cee,tooltip:e(`layout.sidebar.news.tooltip`),url:`https://www.lunarclient.com/news`}]," +
		"routes=[(0,H.jsx)(O,{element:(0,H.jsx)(mq,{}),path:`/`})]," +
		"mq=()=>{return(0,H.jsx)(j,{display:`flex`,flexDirection:`column`,flexGrow:1,gap:`4`,minHeight:`0`,children:(0,H.jsxs)($c,{children:[" +
		"(0,H.jsxs)(gK,{background:d.src,onLoad:()=>{},children:[]})," +
		"(0,H.jsx)(MK,{})," +
		"(0,H.jsxs)(j,{css:{\"&&::-webkit-scrollbar\":{display:`none`}},display:`flex`,flexDirection:`column`,flexShrink:0,gridColumn:`span 2`,height:`100%`,children:[" +
		"(0,H.jsx)(iq,{}),(0,H.jsx)(sK,{full:n===!1,mode:O,rootRef:D})," +
		"(0,H.jsx)(kf,{...w.ad.secondary}),(0,H.jsx)(kf,{...w.ad.main})]})]})})}"
	b, err := patchHome338(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, tok := range []string{"lunarPatcherHomeTools=", "lunarPatcherUpdatesPage=", "className:`lunar-patcher-friends`", "(0,H.jsx)(sK,{full:!0,mode:`list`,rootRef:D})"} {
		if !bytes.Contains(b, []byte(tok)) {
			t.Errorf("missing %s", tok)
		}
	}
	for _, damage := range []string{"(0,H.jsx)(sK,{full:n===!1,mode:O,rootRef:D})", "(0,H.jsx)(iq,{})", "(0,H.jsx)(kf,{...w.ad.main", "(0,H.jsx)(MK,{})"} {
		damaged := strings.Replace(src, damage, "unexpected", 1)
		result, err := patchHome338(damaged)
		if err == nil || len(result) != 0 {
			t.Fatalf("invalid change to %q accepted", damage)
		}
	}
}

func TestUploaded338AllModes(t *testing.T) {
	raw, err := os.ReadFile("/mnt/data/app(7).asar")
	if os.IsNotExist(err) {
		t.Skip("uploaded 3.7.38 archive unavailable in CI")
	}
	if err != nil {
		t.Fatal(err)
	}
	a, err := parseAsar(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got := parseVersion(a); got != "3.7.38-ow" {
		t.Fatalf("unexpected fixture %s", got)
	}
	oldMain, _ := a.entry("dist/assets/main.js")
	oldAd, _ := a.entry("dist/assets/use-show-ad.js")
	for _, mode := range []string{"ads", "layout", "sakura", "color"} {
		for _, roblox := range []bool{false, true} {
			t.Run(mode+map[bool]string{false: "", true: "+roblox"}[roblox], func(t *testing.T) {
				patched, report, err := patchSource(raw, mode, "#EBA2C7", roblox)
				if err != nil {
					t.Fatal(err)
				}
				if report.Version != "3.7.38-ow" {
					t.Fatal(report.Version)
				}
				u, err := parseAsar(patched)
				if err != nil {
					t.Fatal(err)
				}
				ad, _ := u.entry("dist/assets/use-show-ad.js")
				if bytes.Equal(ad, oldAd) || !bytes.Contains(ad, []byte("if(true)return null;")) {
					t.Error("ad renderer unchanged")
				}
				html, _ := u.entry("dist/html/main.html")
				if !bytes.Contains(html, []byte(marker)) {
					t.Error("missing patch marker")
				}
				main, _ := u.entry("dist/assets/main.js")
				if mode != "ads" {
					for _, token := range []string{"lunarPatcherHomeTools=", "lunarPatcherUpdatesPage=", "className:`lunar-patcher-home`", "className:`lunar-patcher-hero`", "className:`lunar-patcher-friends`", "(0,H.jsx)(iq,{})", "(0,H.jsx)(sK,{full:!0,mode:`list`,rootRef:D})"} {
						if !bytes.Contains(main, []byte(token)) {
							t.Errorf("missing %s", token)
						}
					}
					if bytes.Contains(main, []byte("(0,H.jsx)(kf,{...w.ad.main")) {
						t.Error("homepage ads retained")
					}
					if bytes.Count(main, []byte("(0,H.jsx)(MK,{})")) != 1 {
						t.Error("news should only remain in Updates")
					}
				} else if !roblox && !bytes.Equal(main, oldMain) {
					t.Error("Ads Only modified homepage")
				}
				if roblox && !bytes.Contains(main, []byte("lunarPatcherRobloxPage=")) {
					t.Error("Roblox missing")
				}
				if mode == "sakura" && !bytes.Contains(html, []byte("data:image/jpeg;base64")) {
					t.Error("Sakura wallpaper absent")
				}
				if mode == "color" && bytes.Contains(html, []byte("data:image/jpeg;base64")) {
					t.Error("Custom Color contains Sakura wallpaper")
				}
				// Save only one mode's parsed bundles for JS syntax verification.
				if mode == "sakura" && roblox {
					out := filepath.Join(os.TempDir(), "lunar338_verify")
					if err := os.MkdirAll(out, 0700); err != nil {
						t.Fatal(err)
					}
					for _, name := range []string{"dist/assets/main.js", "dist/assets/use-show-ad.js", "dist-electron/electron/main.js", "dist-electron/preload/launcher/index.js"} {
						content, _ := u.entry(name)
						dest := filepath.Join(out, strings.ReplaceAll(name, "/", "_")+".js")
						if err := os.WriteFile(dest, content, 0600); err != nil {
							t.Fatal(err)
						}
						t.Log("bundle: " + dest)
					}
				}
			})
		}
	}
}
