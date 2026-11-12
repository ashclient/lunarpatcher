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
	if strings.Contains(s, "(0,V.jsx)(A,{element:(0,V.jsx)(tq,{}),path:`/`})") {
		jsx, react, componentAnchor = "V", "B", "tq=()=>{"
		routeAnchor = "(0,V.jsx)(A,{element:(0,V.jsx)(tq,{}),path:`/`})"
		routeAdd = "(0,V.jsx)(A,{element:(0,V.jsx)(lunarPatcherRobloxPage,{}),path:`/roblox`})"
		internal = "{type:`route`,id:`news`,inactiveIcon:Kee,activeIcon:Kee,tooltip:`Updates`,path:`/updates`}"
		external = "{type:`external`,id:`news`,icon:Kee,tooltip:e(`layout.sidebar.news.tooltip`),url:`https://www.lunarclient.com/news`}"
	} else if strings.Contains(s, "(0,V.jsx)(j,{element:(0,V.jsx)(eq,{}),path:`/`})") {

		jsx, react, componentAnchor = "V", "B", "eq=()=>{"
		routeAnchor = "(0,V.jsx)(j,{element:(0,V.jsx)(eq,{}),path:`/`})"
		routeAdd = "(0,V.jsx)(j,{element:(0,V.jsx)(lunarPatcherRobloxPage,{}),path:`/roblox`})"
		internal = "{type:`route`,id:`news`,inactiveIcon:Kee,activeIcon:Kee,tooltip:`Updates`,path:`/updates`}"
		external = "{type:`external`,id:`news`,icon:Kee,tooltip:e(`layout.sidebar.news.tooltip`),url:`https://www.lunarclient.com/news`}"
	} else if strings.Contains(s, "(0,V.jsx)(y,{element:(0,V.jsx)(dW,{}),path:`/`})") {
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
/* The native Roblox website fills Lunar's existing main panel. No extra sidebars. */
#root .lunar-patcher-roblox-browser{height:100%;min-width:0;min-height:0;display:flex;flex-direction:column;position:relative;overflow:hidden;background:#15171d;box-sizing:border-box}
#root .lpr-browser-viewport{flex:1;min-width:0;min-height:150px;width:100%;background:#15171d}
#root .lpr-browser-error{position:absolute;bottom:18px;left:18px;right:18px;z-index:2;display:flex;align-items:center;gap:16px;padding:13px 16px;background:#272b34;color:#f9fafc;border:1px solid #e36d84;border-radius:9px;font-size:12px;line-height:1.45;font-family:inherit;box-shadow:0 8px 24px #0008}
#root .lpr-browser-error button{margin-left:auto;flex-shrink:0;background:#444b59;border:1px solid #6a7080;border-radius:6px;padding:7px 10px;color:#fff;cursor:pointer}
#root .lpr-browser-error button:hover{background:#576073}
`
