package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func Test337SyntheticHome(t *testing.T) {
	const stub = "qT=(e,t)=>[{type:`external`,id:`news`,icon:Bo,tooltip:e(`layout.sidebar.news.tooltip`),url:`https://www.lunarclient.com/news`}]," +
		"routes=[(0,V.jsx)(S,{element:(0,V.jsx)($K,{}),path:`/`})]," +
		"$K=()=>{return(0,V.jsx)(h,{display:`flex`,flexDirection:`column`,flexGrow:1,gap:`4`,minHeight:`0`,children:(0,V.jsxs)(Kc,{children:[" +
		"(0,V.jsxs)(tK,{background:u.src,onLoad:()=>{},children:[]})," +
		"(0,V.jsx)(_K,{})," +
		"(0,V.jsxs)(h,{css:{\"&&::-webkit-scrollbar\":{display:`none`}},display:`flex`,flexDirection:`column`,flexShrink:0,gridColumn:`span 2`,height:`100%`,children:[" +
		"(0,V.jsx)(UK,{}),(0,V.jsx)(KG,{full:n===!1,mode:ee,rootRef:O})," +
		"(0,V.jsx)(wf,{...T.ad.secondary}),(0,V.jsx)(wf,{...T.ad.main})]})]})})}"
	patched, err := patchHome337(stub)
	if err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{"lunarPatcherHomeTools=", "lunarPatcherUpdatesPage=", "className:`lunar-patcher-friends`", "className:`lunar-patcher-home`", "(0,V.jsx)(KG,{full:!0,mode:`list`,rootRef:O})"} {
		if !bytes.Contains(patched, []byte(token)) {
			t.Fatalf("missing %s", token)
		}
	}
	if bytes.Contains(patched, []byte("(0,V.jsx)(wf,{...T.ad.main")) {
		t.Fatal("ad retained")
	}
	for _, damage := range []string{"(0,V.jsx)(KG,{full:n===!1,mode:ee,rootRef:O})", "(0,V.jsx)(wf,{...T.ad.main", "(0,V.jsx)(UK,{})"} {
		altered := strings.Replace(stub, damage, "break", 1)
		result, err := patchHome337(altered)
		if err == nil || len(result) != 0 {
			t.Fatalf("unsafe partial output after changing %s", damage)
		}
	}
}

func TestUploaded3_7_37(t *testing.T) {
	raw, err := os.ReadFile("/mnt/data/app(6).asar")
	if os.IsNotExist(err) {
		t.Skip("3.7.37 user-provided ASAR not available in CI")
	}
	if err != nil {
		t.Fatal(err)
	}
	a, err := parseAsar(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got := parseVersion(a); got != "3.7.37-ow" {
		t.Fatalf("fixture version %s", got)
	}
	oldMain, _ := a.entry("dist/assets/main.js")
	oldAds, _ := a.entry("dist/assets/use-show-ad.js")
	for _, mode := range []string{"ads", "layout", "sakura", "color"} {
		for _, roblox := range []bool{false, true} {
			name := mode
			if roblox {
				name += "+roblox"
			}
			t.Run(name, func(t *testing.T) {
				updated, report, err := patchSource(raw, mode, "#EFADC9", roblox)
				if err != nil {
					t.Fatal(err)
				}
				if report.Version != "3.7.37-ow" {
					t.Fatal(report.Version)
				}
				u, err := parseAsar(updated)
				if err != nil {
					t.Fatal(err)
				}
				ad, _ := u.entry("dist/assets/use-show-ad.js")
				if bytes.Equal(ad, oldAds) || !bytes.Contains(ad, []byte("if(true)return null;")) {
					t.Fatal("ad guard not removed")
				}
				html, _ := u.entry("dist/html/main.html")
				if !bytes.Contains(html, []byte(marker)) {
					t.Fatal("missing HTML marker")
				}
				main, _ := u.entry("dist/assets/main.js")
				if mode != "ads" {
					for _, token := range []string{"lunarPatcherHomeTools=", "lunarPatcherUpdatesPage=", "className:`lunar-patcher-home`", "className:`lunar-patcher-hero`", "className:`lunar-patcher-friends`", "(0,V.jsx)(UK,{})", "(0,V.jsx)(KG,{full:!0,mode:`list`,rootRef:O})"} {
						if !bytes.Contains(main, []byte(token)) {
							t.Errorf("missing %s", token)
						}
					}
					if bytes.Contains(main, []byte("(0,V.jsx)(wf,{...T.ad.main")) {
						t.Error("home ad still present")
					}
					if bytes.Count(main, []byte("(0,V.jsx)(_K,{})")) != 1 {
						t.Error("news should only render on Updates")
					}
				} else if !roblox && !bytes.Equal(main, oldMain) {
					t.Error("ads only changed home renderer")
				}
				if roblox {
					if !bytes.Contains(main, []byte("lunarPatcherRobloxPage=")) {
						t.Error("Roblox page not injected")
					}
					for _, file := range []string{"dist-electron/electron/main.js", "dist-electron/preload/launcher/index.js"} {
						b, ok := u.entry(file)
						if !ok {
							t.Error("missing " + file)
							continue
						}
						if !bytes.Contains(b, []byte("lunarpatcher:roblox:show")) && !bytes.Contains(b, []byte("lunarPatcherRobloxBrowser")) {
							t.Error("Roblox bridge not installed in " + file)
						}
					}
				}
				if mode == "sakura" && !bytes.Contains(html, []byte("data:image/jpeg;base64")) {
					t.Error("wallpaper missing")
				}
				if mode == "color" && bytes.Contains(html, []byte("data:image/jpeg;base64")) {
					t.Error("Sakura erroneously applied to Custom Color")
				}
				if mode == "sakura" && roblox {
					// Checked by node --check outside Go. Files are not bundled with the app.
					os.MkdirAll("/mnt/data/lunarpatcher_v174/validation", 0700)
					for _, file := range []string{"dist/assets/main.js", "dist/assets/use-show-ad.js", "dist-electron/electron/main.js", "dist-electron/preload/launcher/index.js"} {
						b, _ := u.entry(file)
						out := strings.ReplaceAll(file, "/", "_")
						if err := os.WriteFile("/mnt/data/lunarpatcher_v174/validation/"+out+".mjs", b, 0600); err != nil {
							t.Fatal(err)
						}
					}
				}
			})
		}
	}
}

func Test337UnexpectedRendererFailsClosed(t *testing.T) {
	raw, err := os.ReadFile("/mnt/data/app(6).asar")
	if os.IsNotExist(err) {
		t.Skip("fixture unavailable")
	}
	if err != nil {
		t.Fatal(err)
	}
	archive, err := parseAsar(raw)
	if err != nil {
		t.Fatal(err)
	}
	main, _ := archive.entry("dist/assets/main.js")
	damaged := bytes.Replace(main, []byte("(0,V.jsx)(KG,{full:n===!1,mode:ee,rootRef:O})"), []byte("(0,V.jsx)(KG,{full:n===!1,mode:unknown,rootRef:O})"), 1)
	altered, err := archive.rebuild(map[string][]byte{"dist/assets/main.js": damaged})
	if err != nil {
		t.Fatal(err)
	}
	patched, _, err := patchSource(altered, "layout", "#EFADC9")
	if err == nil || len(patched) != 0 {
		t.Fatalf("unsupported friend signature accepted: %v", err)
	}
	ad, _ := archive.entry("dist/assets/use-show-ad.js")
	alteredAd := bytes.Replace(ad, []byte("if(!a||!e.willShowAd)return null;"), []byte("if(!a||e.willShowAd)return null;"), 1)
	badAd, err := archive.rebuild(map[string][]byte{"dist/assets/use-show-ad.js": alteredAd})
	if err != nil {
		t.Fatal(err)
	}
	patched, _, err = patchSource(badAd, "ads", "#EFADC9")
	if err == nil || len(patched) != 0 {
		t.Fatalf("unsupported ad guard accepted: %v", err)
	}
}
