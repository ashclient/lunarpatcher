package main

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"
)

//go:embed assets/roblox_widget.js
var robloxWidgetJS string

// patchRobloxSidebar optionally adds a native React route and a sidebar icon.
// It does not replace or proxy Roblox authentication: the already installed
// Windows Roblox protocol handler performs the actual game launch.
func patchRobloxSidebar(source []byte) ([]byte, error) {
	s := string(source)
	var jsx, react, anchor, route, entry string
	if strings.Contains(s, "(0,V.jsx)(y,{element:(0,V.jsx)(dW,{}),path:`/`})") {
		jsx, react, anchor = "V", "B", "dW=()=>{"
		route = "(0,V.jsx)(y,{element:(0,V.jsx)(dW,{}),path:`/`})"
		entry = "{type:`route`,id:`roblox`,inactiveIcon:lunarPatcherRobloxIcon,activeIcon:lunarPatcherRobloxIcon,tooltip:`Roblox`,path:`/roblox`}"
		external := "{type:`external`,id:`news`,icon:ts,tooltip:e(`layout.sidebar.news.tooltip`),url:`https://www.lunarclient.com/news`}"
		internal := "{type:`route`,id:`news`,inactiveIcon:ts,activeIcon:ts,tooltip:`Updates`,path:`/updates`}"
		if strings.Count(s, internal) == 1 {
			s = strings.Replace(s, internal, internal+","+entry, 1)
		} else if strings.Count(s, external) == 1 {
			s = strings.Replace(s, external, external+","+entry, 1)
		} else {
			return nil, errors.New("Roblox sidebar anchor changed in 3.7.32; no patch applied")
		}
	} else if strings.Contains(s, "(0,H.jsx)(b,{element:(0,H.jsx)(IK,{}),path:`/`})") {
		jsx, react, anchor = "H", "V", "IK=()=>{"
		route = "(0,H.jsx)(b,{element:(0,H.jsx)(IK,{}),path:`/`})"
		entry = "{type:`route`,id:`roblox`,inactiveIcon:lunarPatcherRobloxIcon,activeIcon:lunarPatcherRobloxIcon,tooltip:`Roblox`,path:`/roblox`}"
		external := "{type:`external`,id:`news`,icon:rs,tooltip:e(`layout.sidebar.news.tooltip`),url:`https://www.lunarclient.com/news`}"
		internal := "{type:`route`,id:`news`,inactiveIcon:rs,activeIcon:rs,tooltip:`Updates`,path:`/updates`}"
		if strings.Count(s, internal) == 1 {
			s = strings.Replace(s, internal, internal+","+entry, 1)
		} else if strings.Count(s, external) == 1 {
			s = strings.Replace(s, external, external+","+entry, 1)
		} else {
			return nil, errors.New("Roblox sidebar anchor changed in 3.7.31; no patch applied")
		}
	} else {
		return nil, errors.New("unsupported Roblox sidebar layout; no patch applied")
	}
	if strings.Count(s, route) != 1 {
		return nil, fmt.Errorf("Roblox route signature changed (%d matches)", strings.Count(s, route))
	}
	routeAdd := strings.Replace(route, "path:`/`})", "path:`/`})", 1) // defensive unchanged anchor
	if jsx == "V" {
		s = strings.Replace(s, routeAdd, routeAdd+",(0,V.jsx)(y,{element:(0,V.jsx)(lunarPatcherRobloxPage,{}),path:`/roblox`})", 1)
	} else {
		s = strings.Replace(s, routeAdd, routeAdd+",(0,H.jsx)(b,{element:(0,H.jsx)(lunarPatcherRobloxPage,{}),path:`/roblox`})", 1)
	}
	if strings.Count(s, anchor) != 1 {
		return nil, errors.New("Roblox component insertion signature changed")
	}
	widget := strings.TrimSpace(robloxWidgetJS)
	widget = strings.TrimPrefix(widget, "// Injected into Lunar's own renderer by LunarPatcher; no Roblox credentials or client binaries are touched.\n")
	widget = strings.ReplaceAll(widget, "__JSX__", jsx)
	widget = strings.ReplaceAll(widget, "__REACT__", react)
	if jsx == "V" {
		widget = strings.ReplaceAll(widget, "__INITIATOR__", "tp.sidebar")
	} else {
		widget = strings.ReplaceAll(widget, "__INITIATOR__", "tp.sidebar")
	}
	s = strings.Replace(s, anchor, widget+anchor, 1)
	if strings.Count(s, "lunarPatcherRobloxPage=") != 1 || strings.Count(s, "path:`/roblox`") != 2 {
		return nil, errors.New("Roblox route injection verification failed")
	}
	return []byte(s), nil
}

const robloxCSS = `
#root .lunar-patcher-roblox{height:100%;min-height:0;min-width:0;display:flex;flex-direction:column;color:var(--colors-space-12,#e9ecf3);font-family:inherit;box-sizing:border-box}
#root .lunar-patcher-roblox *{box-sizing:border-box}
#root .lunar-patcher-roblox-header{display:flex;align-items:center;justify-content:space-between;gap:16px;padding:24px 28px 15px}
#root .lunar-patcher-roblox-eyebrow{font-size:11px;letter-spacing:.035em;color:var(--colors-space-11,#a5acbe);margin:0 0 5px}
#root .lunar-patcher-roblox-header h1{margin:0 0 3px;font-size:24px;font-weight:700;line-height:1.3;color:inherit}
#root .lunar-patcher-roblox-header p,#root .lunar-patcher-roblox-hint{font-size:12px;line-height:1.55;color:var(--colors-space-11,#a6aebd);margin:0 0 12px}
#root .lunar-patcher-roblox-tabs{display:flex;gap:20px;border-bottom:1px solid var(--colors-space-a4,rgba(180,190,210,.15));padding:0 28px;flex-shrink:0}
#root .lunar-patcher-roblox-tabs button{font:inherit;background:transparent;border:0;border-bottom:2px solid transparent;color:var(--colors-space-11,#a5acbb);padding:11px 2px 12px;cursor:pointer;font-size:13px;font-weight:600}
#root .lunar-patcher-roblox-tabs button.is-active{color:var(--colors-space-12,#f0f3f8);border-bottom-color:#6ba6ff}
#root .lunar-patcher-roblox-scroll{flex:1;min-height:0;overflow-y:auto;overscroll-behavior:contain}
#root .lunar-patcher-roblox-body{padding:20px 28px 28px;max-width:1200px;margin:0 auto;display:flex;flex-direction:column;gap:14px}
#root .lunar-patcher-roblox-section{background:var(--colors-space-2,#1c1e27);border:1px solid var(--colors-space-a4,rgba(180,190,210,.15));border-radius:12px;padding:17px 18px}
#root .lunar-patcher-roblox-heading{display:flex;align-items:center;justify-content:space-between;gap:10px;margin-bottom:12px}
#root .lunar-patcher-roblox-heading h2{margin:0;font-size:14px;font-weight:650;color:inherit}
#root .lunar-patcher-roblox-inputrow{display:flex;gap:9px;align-items:center;flex-wrap:wrap}
#root .lunar-patcher-roblox input[type=text],#root .lunar-patcher-roblox select{font:inherit;font-size:12.5px;min-width:0;flex:1;background:var(--colors-space-1,#14161d);border:1px solid var(--colors-space-a4,rgba(180,190,210,.22));color:inherit;padding:10px 11px;border-radius:7px;min-height:38px}
#root .lunar-patcher-roblox-play,#root .lunar-patcher-roblox-secondary,#root .lunar-patcher-roblox-iconbtn,#root .lunar-patcher-roblox-textbtn{font:inherit;cursor:pointer;white-space:nowrap;border-radius:7px;min-height:35px;padding:7px 11px;font-size:12px;font-weight:600}
#root .lunar-patcher-roblox-play{background:#327bdc;color:#fff;border:1px solid #327bdc}
#root .lunar-patcher-roblox-play:hover{background:#4b8ded}
#root .lunar-patcher-roblox-secondary,#root .lunar-patcher-roblox-iconbtn,#root .lunar-patcher-roblox-textbtn{background:var(--colors-space-3,#292c38);color:inherit;border:1px solid var(--colors-space-a4,rgba(180,190,210,.18))}
#root .lunar-patcher-roblox-iconbtn{padding:6px 8px}
#root .lunar-patcher-roblox-secondary:hover,#root .lunar-patcher-roblox-iconbtn:hover{filter:brightness(1.13)}
#root .lunar-patcher-roblox-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:9px}
#root .lunar-patcher-roblox-card{min-width:0;display:flex;align-items:center;gap:10px;background:var(--colors-space-1,#161820);border:1px solid var(--colors-space-a4,rgba(180,190,210,.13));border-radius:9px;padding:9px}
#root .lunar-patcher-roblox-tile{background:#303c57;width:34px;height:34px;flex-shrink:0;display:grid;place-items:center;border-radius:7px;color:white}
#root .lunar-patcher-roblox-card-copy{flex:1;min-width:0;display:flex;flex-direction:column;gap:3px}
#root .lunar-patcher-roblox-card-copy strong{font-size:12px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-weight:650}
#root .lunar-patcher-roblox-card-copy span{font-size:10.5px;color:var(--colors-space-11,#a9b0bd)}
#root .lunar-patcher-roblox-inline{display:flex;gap:4px}
#root .lunar-patcher-roblox-setting{display:flex;align-items:center;justify-content:space-between;gap:20px;min-height:62px;border-bottom:1px solid var(--colors-space-a4,rgba(180,190,210,.10));padding:12px 0;cursor:pointer}
#root .lunar-patcher-roblox-setting:last-child{border-bottom:0}
#root .lunar-patcher-roblox-setting span{display:flex;flex-direction:column;gap:4px}
#root .lunar-patcher-roblox-setting strong{font-size:13px;font-weight:600}
#root .lunar-patcher-roblox-setting small{font-size:11px;color:var(--colors-space-11,#aab1bf);line-height:1.45}
#root .lunar-patcher-roblox-setting input{accent-color:#6ba6ff;width:16px;height:16px;flex-shrink:0}
#root .lunar-patcher-roblox-setting select{max-width:230px;flex:0 0 230px}
#root .lunar-patcher-roblox-account-list{display:flex;flex-direction:column;gap:8px}
#root .lunar-patcher-roblox-account{display:flex;align-items:center;gap:9px;min-width:0;background:var(--colors-space-1,#161820);border-radius:8px;padding:9px}
#root .lunar-patcher-roblox-avatar{width:32px;height:32px;border-radius:50%;background:#394d72;color:#fff;display:grid;place-items:center;font-size:14px;font-weight:700;flex-shrink:0}
#root .lunar-patcher-roblox-account-copy{display:flex;flex-direction:column;gap:3px;min-width:0;flex:1}
#root .lunar-patcher-roblox-account-copy strong{font-size:12px}
#root .lunar-patcher-roblox-account-copy span{font-size:11px;color:var(--colors-space-11,#9ba6ba)}
#root .lunar-patcher-roblox-actions{display:flex;flex-wrap:wrap;gap:8px}
#root .lunar-patcher-roblox-notice{border-top:1px solid var(--colors-space-a4,rgba(180,190,210,.13));padding:10px 28px;font-size:11px;line-height:1.4;min-height:37px;color:var(--colors-space-11,#9fa7b8);flex-shrink:0}
@media(max-width:1020px){#root .lunar-patcher-roblox-grid{grid-template-columns:1fr}}
@media(max-width:740px){#root .lunar-patcher-roblox-body{padding:15px}#root .lunar-patcher-roblox-header{padding:17px 15px}#root .lunar-patcher-roblox-tabs{padding:0 15px}#root .lunar-patcher-roblox-notice{padding:10px 15px}#root .lunar-patcher-roblox-account{flex-wrap:wrap}#root .lunar-patcher-roblox-setting select{max-width:170px;flex-basis:170px}}
`
