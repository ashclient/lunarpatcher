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
#root .lunar-patcher-roblox-browser{height:100%;min-height:0;min-width:0;display:flex;overflow:hidden;color:#f5f6fb;background:#15171c;font-family:inherit;font-size:13px;box-sizing:border-box}
#root .lunar-patcher-roblox-browser *{box-sizing:border-box}
#root .lpr-navigation{width:186px;min-width:176px;background:#14161c;border-right:1px solid #2c3039;display:flex;flex-direction:column;min-height:0}
#root .lpr-brand{height:66px;display:flex;align-items:center;gap:12px;padding:0 19px;font-size:17px;font-weight:750;letter-spacing:-.025em}
#root .lpr-brand svg{color:#f8f9fb}
#root .lpr-nav-scroll{flex:1;overflow-y:auto;min-height:0;padding:8px}
#root .lpr-navigation-group{padding:4px 0 13px}
#root .lpr-group-heading{font-size:10px!important;color:#737985!important;margin:10px 12px 8px!important;letter-spacing:.09em;font-weight:800!important}
#root .lpr-nav-link{display:flex;align-items:center;gap:12px;width:100%;height:35px;padding:0 12px;border:0;text-align:left;background:transparent;color:#b1b6c2;border-radius:7px;cursor:pointer;font:inherit;font-size:12px;font-weight:570}
#root .lpr-nav-link:hover{background:#252933;color:#f9f9fd}
#root .lpr-nav-link.is-active{background:#333741;color:#fff}
#root .lpr-nav-symbol{font-size:18px;line-height:1;width:17px;text-align:center;font-weight:500;color:#a1a5b1}
#root .lpr-nav-link.is-active .lpr-nav-symbol{color:#fff}
#root .lpr-saved-link .lpr-remove{margin-left:auto;color:#818794;font-size:17px;padding:2px 3px}
#root .lpr-saved-link .lpr-remove:hover{color:#fff}
#root .lpr-empty-saved{padding:0 12px;color:#8c919f;font-size:11px;line-height:1.45}
#root .lpr-navigation-footer{padding:16px;border-top:1px solid #292d35;color:#777e8d;font-size:10px;line-height:1.5}
#root .lpr-main{flex:1;min-width:0;display:flex;flex-direction:column;min-height:0;background:#181a20}
#root .lpr-workspace-bar{display:flex;align-items:center;gap:12px;min-height:62px;padding:0 18px;border-bottom:1px solid #292d36;background:#1a1d24}
#root .lpr-heading{display:flex;flex-direction:column;min-width:96px;gap:2px}
#root .lpr-page-title{font-size:17px;font-weight:760;letter-spacing:-.025em;color:#f9fafc}
#root .lpr-page-subtitle{font-size:10px;color:#9196a2}
#root .lpr-search{display:flex;align-items:center;flex:1;min-width:120px;max-width:420px;height:34px;background:#101218;border:1px solid #333741;border-radius:8px;padding:0 8px;gap:8px}
#root .lpr-search:focus-within{border-color:#7786a7}
#root .lpr-search-icon{font-size:20px;color:#969ead}
#root .lpr-search input{flex:1;min-width:0;background:transparent!important;border:0!important;outline:0!important;color:#fff;font:inherit;font-size:12px;padding:0}
#root .lpr-search button{border:0;background:transparent;color:#c7cbda;cursor:pointer;font-size:11px}
#root .lpr-search button:disabled{opacity:.45}
#root .lpr-browser-toolbar{flex:0 0 auto;min-height:49px;display:flex;align-items:center;gap:7px;padding:7px 14px;border-bottom:1px solid #2b2e37;background:#1a1c23}
#root .lpr-tool{background:#242731;border:1px solid #393d48;border-radius:7px;min-width:32px;height:32px;padding:0 8px;color:#eff1f7;cursor:pointer;font-size:15px}
#root .lpr-tool:hover:not(:disabled){background:#343946}
#root .lpr-tool:disabled{opacity:.35;cursor:default}
#root .lpr-browser-address{display:flex;flex:1;min-width:120px}
#root .lpr-browser-address input{min-width:0;flex:1;min-height:32px;background:#101218!important;border:1px solid #313640!important;border-radius:6px;padding:0 10px;color:#b7bdc9;font:inherit;font-size:11px;outline:none}
#root .lpr-browser-address input:focus{border-color:#7885a3!important}
#root .lpr-browser-viewport{flex:1;min-height:180px;width:100%;background:#101218;position:relative}
#root .lpr-browser-footer{display:flex;align-items:center;justify-content:space-between;gap:10px;min-height:34px;padding:6px 15px;border-top:1px solid #292e38;background:#191c24;color:#8a93a2;font-size:10px}
#root .lpr-browser-status{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
#root .lpr-external{background:transparent;color:#aeb6c9;border:0;font:inherit;font-size:10px;cursor:pointer;flex-shrink:0}
#root .lpr-external:hover{color:#fff}
#root .lpr-workspace button:focus-visible{outline:2px solid #e6a3c1;outline-offset:2px}
@media(max-width:850px){#root .lpr-navigation{width:142px;min-width:142px}#root .lpr-brand{padding:0 12px}#root .lpr-workspace-bar{padding:0 10px}#root .lpr-browser-toolbar{padding:6px}#root .lpr-search{max-width:300px}}
@media(max-width:630px){#root .lpr-navigation{width:48px;min-width:48px}#root .lpr-nav-link{padding:0 14px}#root .lpr-nav-link span:nth-child(2),#root .lpr-group-heading,#root .lpr-brand span,#root .lpr-navigation-footer,#root .lpr-saved{display:none}#root .lpr-brand{padding:0 11px}#root .lpr-search{max-width:none}}
`
