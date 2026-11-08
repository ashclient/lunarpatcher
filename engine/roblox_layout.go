package main

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"
)

//go:embed assets/roblox_widget.js
var robloxWidgetJS string

//go:embed assets/roblox_bridge_main.js
var robloxMainBridge string

//go:embed assets/roblox_bridge_preload.js
var robloxPreloadBridge string

// patchRobloxSidebar adds the official Roblox website route to Lunar's
// own sidebar, instead of installing a separate Roblox desktop launcher.
func patchRobloxSidebar(source []byte) ([]byte, error) {
	return patchRobloxSidebarWithArtwork(source, nil)
}

// Kept for old caller compatibility. Roblox.com renders its real images now;
// there is no game-thumbnail download or static game catalog.
func patchRobloxSidebarWithArtwork(source []byte, _ map[string]string) ([]byte, error) {
	s := string(source)
	var jsx, react, componentAnchor, routeAnchor, routeAdd, internal, external string
	entry := "{type:`route`,id:`roblox`,inactiveIcon:lunarPatcherRobloxIcon,activeIcon:lunarPatcherRobloxIcon,tooltip:`Roblox`,path:`/roblox`}"
	if strings.Contains(s, "(0,V.jsx)(y,{element:(0,V.jsx)(dW,{}),path:`/`})") {
		jsx, react, componentAnchor = "V", "B", "dW=()=>{"
		routeAnchor = "(0,V.jsx)(y,{element:(0,V.jsx)(dW,{}),path:`/`})"
		routeAdd = "(0,V.jsx)(y,{element:(0,V.jsx)(lunarPatcherRobloxPage,{}),path:`/roblox`})"
		internal = "{type:`route`,id:`news`,inactiveIcon:ts,activeIcon:ts,tooltip:`Updates`,path:`/updates`}"
		external = "{type:`external`,id:`news`,icon:ts,tooltip:e(`layout.sidebar.news.tooltip`),url:`https://www.lunarclient.com/news`}"
	} else if strings.Contains(s, "(0,H.jsx)(b,{element:(0,H.jsx)(IK,{}),path:`/`})") {
		jsx, react, componentAnchor = "H", "V", "IK=()=>{"
		routeAnchor = "(0,H.jsx)(b,{element:(0,H.jsx)(IK,{}),path:`/`})"
		routeAdd = "(0,H.jsx)(b,{element:(0,H.jsx)(lunarPatcherRobloxPage,{}),path:`/roblox`})"
		internal = "{type:`route`,id:`news`,inactiveIcon:rs,activeIcon:rs,tooltip:`Updates`,path:`/updates`}"
		external = "{type:`external`,id:`news`,icon:rs,tooltip:e(`layout.sidebar.news.tooltip`),url:`https://www.lunarclient.com/news`}"
	} else {
		return nil, errors.New("unsupported Roblox sidebar layout; no patch applied")
	}
	navAnchor := internal
	if strings.Count(s, navAnchor) != 1 {
		navAnchor = external
	}
	if strings.Count(s, navAnchor) != 1 {
		return nil, errors.New("Roblox sidebar navigation changed; no patch applied")
	}
	if strings.Count(s, routeAnchor) != 1 || strings.Count(s, componentAnchor) != 1 {
		return nil, errors.New("Roblox route or component placement changed; no patch applied")
	}
	s = strings.Replace(s, navAnchor, navAnchor+","+entry, 1)
	s = strings.Replace(s, routeAnchor, routeAnchor+","+routeAdd, 1)
	widget := strings.ReplaceAll(strings.ReplaceAll(robloxWidgetJS, "__JSX__", jsx), "__REACT__", react)
	s = strings.Replace(s, componentAnchor, widget+componentAnchor, 1)
	if strings.Count(s, "lunarPatcherRobloxPage=") != 1 || strings.Count(s, "path:`/roblox`") != 2 {
		return nil, fmt.Errorf("Roblox page injection count mismatch; no patch applied")
	}
	return []byte(s), nil
}

// Installation changes are limited to three packaged JavaScript files. Both
// the preload and main-process bridges are required for actual website browsing.
func patchRobloxElectronBridge(main, preload []byte) ([]byte, []byte, error) {
	if len(main) < 1000 || len(preload) < 200 {
		return nil, nil, errors.New("Lunar Electron main/preload module missing")
	}
	if !strings.Contains(string(main), `require("electron")`) || !strings.Contains(string(main), `n.BrowserWindow`) {
		return nil, nil, errors.New("unsupported Lunar Electron main module; refusing browser patch")
	}
	if !strings.Contains(string(preload), `contextBridge.exposeInMainWorld`) {
		return nil, nil, errors.New("unsupported Lunar Electron preload module; refusing browser patch")
	}
	for _, marker := range []string{"lunarpatcher:roblox:show", "lunarPatcherRobloxBrowser"} {
		if strings.Contains(string(main), marker) || strings.Contains(string(preload), marker) {
			return nil, nil, errors.New("Roblox bridge already present; restore first")
		}
	}
	return append(append([]byte{}, main...), []byte("\n"+robloxMainBridge+"\n")...), append(append([]byte{}, preload...), []byte("\n"+robloxPreloadBridge+"\n")...), nil
}

const robloxCSS = `
#root .lunar-patcher-roblox-browser{height:100%;min-height:0;min-width:0;display:flex;flex-direction:column;overflow:hidden;color:#f6f7fb;background:#15171e;font-family:inherit;box-sizing:border-box}
#root .lunar-patcher-roblox-browser *{box-sizing:border-box}
#root .lpr-browser-toolbar{flex:0 0 auto;min-height:52px;display:flex;align-items:center;gap:8px;padding:9px 14px;border-bottom:1px solid rgba(255,255,255,.1);background:#1b1d25}
#root .lpr-browser-heading{font-weight:700;min-width:70px;font-size:14px;letter-spacing:-.01em}
#root .lpr-browser-toolbar button{border:1px solid rgba(255,255,255,.1);background:#292c35;color:#fff;border-radius:7px;min-width:34px;height:32px;padding:0 9px;font-size:17px;cursor:pointer}
#root .lpr-browser-toolbar button:hover:not(:disabled){background:#353944}
#root .lpr-browser-toolbar button:disabled{opacity:.35;cursor:default}
#root .lpr-browser-address{min-width:120px;flex:1;display:flex;align-items:center}
#root .lpr-browser-address input{font:inherit;font-size:12px;color:#e5e8ed;min-width:0;flex:1;height:32px;padding:0 11px;border-radius:7px;border:1px solid rgba(255,255,255,.12);background:#14161c;outline:none}
#root .lpr-browser-address input:focus{border-color:#aab3d1}
#root .lpr-browser-viewport{flex:1;min-height:180px;width:100%;background:#111318;position:relative}
#root .lpr-browser-footer{display:flex;gap:12px;align-items:center;justify-content:space-between;padding:6px 13px;min-height:35px;font-size:11px;color:#a6abb6;background:#1b1d25;border-top:1px solid rgba(255,255,255,.1)}
#root .lpr-browser-footer button{font:inherit;background:transparent;color:inherit;cursor:pointer;border:0;padding:3px 6px}
#root .lpr-browser-footer button:hover{color:#fff}
#root .lunar-patcher-roblox-browser button:focus-visible{outline:2px solid #f6a2c5;outline-offset:2px}
@media(max-width:760px){#root .lpr-browser-heading{min-width:0}#root .lpr-browser-toolbar{gap:4px;padding:5px}#root .lpr-browser-toolbar button{min-width:28px;padding:0 6px}}
`
