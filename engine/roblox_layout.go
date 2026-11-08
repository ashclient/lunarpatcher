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
#root .lunar-patcher-roblox{height:100%;min-height:0;min-width:0;display:flex;flex-direction:column;background:var(--colors-space-1,#111318);color:var(--colors-space-12,#f4f5f7);font-family:inherit;box-sizing:border-box}
#root .lunar-patcher-roblox *{box-sizing:border-box}
#root .lunar-patcher-roblox button,#root .lunar-patcher-roblox input,#root .lunar-patcher-roblox select{font:inherit}
#root .lunar-patcher-roblox button{cursor:pointer}
#root .lpr-topbar{display:flex;align-items:center;gap:20px;padding:16px 25px;border-bottom:1px solid rgba(255,255,255,.08);flex-shrink:0}
#root .lpr-wordmark{display:flex;align-items:center;gap:10px;font-size:18px;letter-spacing:-.035em;min-width:132px}
#root .lpr-wordmark svg{transform:rotate(0deg)}
#root .lpr-search{display:flex;gap:10px;align-items:center;flex:1;max-width:600px;background:rgba(255,255,255,.065);border:1px solid rgba(255,255,255,.1);border-radius:9px;padding:3px 4px 3px 12px;min-width:0}
#root .lpr-search>span{color:#9da5b2;font-size:23px;line-height:1}
#root .lpr-search input{flex:1;background:transparent;border:0;outline:none;min-width:0;padding:8px 0;color:inherit;font-size:13px}
#root .lpr-search input::placeholder{color:#9da5b2}
#root .lpr-navigation{display:flex;gap:6px;align-items:center;padding:8px 24px;border-bottom:1px solid rgba(255,255,255,.08);overflow-x:auto;flex-shrink:0}
#root .lpr-nav-item{display:flex;align-items:center;gap:6px;border:0;background:transparent;border-radius:7px;color:#9ba5b3;font-size:12px;font-weight:600;white-space:nowrap;padding:10px 10px}
#root .lpr-nav-item:hover{background:rgba(255,255,255,.06);color:#fff}
#root .lpr-nav-item.is-active{background:rgba(255,255,255,.12);color:#fff}
#root .lpr-nav-icon{font-size:16px;line-height:.8;font-weight:500}
#root .lpr-scroll{flex:1;overflow-y:auto;min-height:0;overscroll-behavior:contain}
#root .lpr-content,#root .lpr-detail{max-width:1330px;margin:0 auto;padding:27px 29px 40px;display:flex;flex-direction:column;gap:24px}
#root .lpr-heading,#root .lpr-welcome{display:flex;align-items:center;justify-content:space-between;gap:24px}
#root .lpr-heading h1,#root .lpr-welcome h1{font-size:26px;letter-spacing:-.04em;line-height:1.25;margin:0 0 5px;color:#f4f5f7}
#root .lpr-heading p,#root .lpr-welcome p{margin:0;color:#a2a9b4;font-size:12.5px;line-height:1.5}
#root .lpr-eyebrow{font-size:10px;letter-spacing:.13em;font-weight:700;color:#c5bfd8;margin-bottom:8px}
#root .lpr-btn{border:1px solid rgba(255,255,255,.13);border-radius:8px;background:#272b34;color:#f4f5f7;padding:9px 13px;white-space:nowrap;font-size:12px;font-weight:650;min-height:35px}
#root .lpr-btn:hover{filter:brightness(1.15)}
#root .lpr-btn:focus-visible,#root .lpr-nav-item:focus-visible{outline:2px solid #80b5ff;outline-offset:2px}
#root .lpr-primary{background:#f7f7f8;color:#13161e;border-color:#f7f7f8}
#root .lpr-glass{background:rgba(255,255,255,.14);border-color:rgba(255,255,255,.26)}
#root .lpr-text{background:transparent;border:0;color:#b6bdc9}
#root .lpr-chip{border-radius:999px;background:#232630;padding:8px 15px}
#root .lpr-chip-active{background:#f0f1f4;color:#111318;border-radius:999px;padding:8px 15px}
#root .lpr-feature{position:relative;display:flex;align-items:center;min-height:230px;padding:30px 40px;border-radius:15px;overflow:hidden;border:1px solid rgba(255,255,255,.12)}
#root .lpr-feature:after{content:"";position:absolute;right:-20px;top:-35px;width:240px;height:240px;border:36px solid rgba(255,255,255,.07);transform:rotate(-22deg);pointer-events:none}
#root .lpr-feature-copy{position:relative;z-index:1;max-width:540px}
#root .lpr-feature-label{font-size:10px;letter-spacing:.13em;font-weight:700;color:rgba(255,255,255,.72)}
#root .lpr-feature h2{font-size:35px;margin:11px 0 9px;letter-spacing:-.045em}
#root .lpr-feature p{margin:0 0 22px;font-size:14px;color:rgba(255,255,255,.84)}
#root .lpr-button-row{display:flex;gap:9px;flex-wrap:wrap;align-items:center}
#root .lpr-rail{min-width:0}
#root .lpr-rail-head{display:flex;justify-content:space-between;align-items:center;gap:14px;margin-bottom:13px}
#root .lpr-rail h2,#root .lpr-panel h2{font-size:16px;margin:0;letter-spacing:-.018em;font-weight:700;color:#f4f5f7}
#root .lpr-game-grid{display:grid;grid-template-columns:repeat(5,minmax(0,1fr));gap:15px}
#root .lpr-game{min-width:0;position:relative}
#root .lpr-game-art{width:100%;aspect-ratio:1/1;display:grid;place-items:center;border-radius:11px;border:1px solid rgba(255,255,255,.14);overflow:hidden;position:relative;padding:0}
#root .lpr-game-glyph{font-size:48px;font-weight:800;color:rgba(255,255,255,.7);text-shadow:0 2px 10px rgba(0,0,0,.1)}
#root .lpr-game-art img{position:absolute;inset:0;width:100%;height:100%;object-fit:cover}
#root .lpr-game-meta{display:flex;flex-direction:column;gap:4px;min-width:0;padding:9px 30px 0 1px}
#root .lpr-game-name{font-size:12.5px;font-weight:650;background:transparent;border:0;color:#f0f2f6;text-align:left;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;padding:0}
#root .lpr-game-meta small{font-size:11px;color:#9ca7b8}
#root .lpr-game-quick{position:absolute;right:1px;bottom:0;width:28px;height:28px;border-radius:50%;background:#363a45;color:white;border:1px solid rgba(255,255,255,.17);font-size:11px}
#root .lpr-art-brookhaven{background:linear-gradient(130deg,#3c7fc2,#164168 90%)}
#root .lpr-art-blox{background:linear-gradient(130deg,#2a1f59,#7141b2 60%,#c078d7)}
#root .lpr-art-doors{background:linear-gradient(145deg,#514530,#171519 85%)}
#root .lpr-art-adopt{background:linear-gradient(145deg,#f093bb,#9369c6)}
#root .lpr-art-mystery{background:linear-gradient(145deg,#762a48,#342742)}
#root .lpr-art-tower{background:linear-gradient(145deg,#388c86,#18464e)}
#root .lpr-art-disaster{background:linear-gradient(145deg,#446674,#254259)}
#root .lpr-art-rivals{background:linear-gradient(145deg,#ca6547,#5e314b)}
#root .lpr-art-custom{background:linear-gradient(145deg,#393e59,#565780)}
#root .lpr-inline-launch,#root .lpr-panel{border:1px solid rgba(255,255,255,.10);border-radius:12px;background:#1c1f27;padding:19px 21px;display:flex;flex-direction:column;gap:13px}
#root .lpr-inline-launch strong{font-size:13px}
#root .lpr-launch-row{display:flex;gap:8px;min-width:0}
#root .lpr-launch-row input,#root .lpr-account-form input,#root .lpr-setting select{min-width:0;flex:1;background:#11151d;border:1px solid rgba(255,255,255,.15);border-radius:7px;padding:10px 12px;color:#fff;font-size:12.5px}
#root .lpr-chips{display:flex;flex-wrap:wrap;gap:8px}
#root .lpr-empty{padding:30px;background:#1b1f27;border:1px dashed rgba(255,255,255,.14);color:#a7b2c1;border-radius:10px;font-size:13px;display:flex;gap:15px;align-items:center}
#root .lpr-account-gate{border:1px solid rgba(255,255,255,.12);background:#1d212b;border-radius:13px;text-align:center;padding:52px 22px;display:flex;align-items:center;flex-direction:column;gap:12px}
#root .lpr-gate-mark{width:68px;height:68px;display:grid;place-items:center;border-radius:18px;background:#303644;font-size:32px}
#root .lpr-account-gate h2{font-size:21px;margin:0}
#root .lpr-account-gate p{margin:0 0 12px;line-height:1.65;color:#abb2c0;font-size:13px;max-width:550px}
#root .lpr-account-form{display:flex;gap:10px;flex-wrap:wrap}
#root .lpr-profile-list{display:flex;flex-direction:column;gap:8px}
#root .lpr-profile{display:flex;align-items:center;gap:10px;padding:11px;background:#242832;border-radius:8px;min-width:0}
#root .lpr-profile-avatar{width:37px;height:37px;background:#46526a;color:#fff;border-radius:50%;display:grid;place-items:center;font-weight:700}
#root .lpr-profile-name{flex:1;min-width:0;display:flex;flex-direction:column;gap:3px}
#root .lpr-profile-name strong{font-size:12px}
#root .lpr-profile-name small{font-size:11px;color:#a0aabd}
#root .lpr-pill{font-size:11px;background:#313a50;padding:5px 8px;border-radius:30px}
#root .lpr-setting{display:flex;align-items:center;justify-content:space-between;padding:13px 0;gap:15px;border-bottom:1px solid rgba(255,255,255,.08);cursor:pointer}
#root .lpr-setting:last-child{border-bottom:0}
#root .lpr-setting span{display:flex;flex-direction:column;gap:4px}
#root .lpr-setting strong{font-size:12.5px}
#root .lpr-setting small,#root .lpr-panel p,#root .lpr-muted{font-size:12px;line-height:1.55;color:#a9b2c1}
#root .lpr-setting input{width:16px;height:16px;accent-color:#78a3fd}
#root .lpr-setting select{flex:0 0 210px}
#root .lpr-detail-hero{min-height:215px;border-radius:14px;display:flex;align-items:center;justify-content:center;overflow:hidden;position:relative}
#root .lpr-detail-hero:after{content:"";position:absolute;border:45px solid rgba(255,255,255,.09);width:240px;height:240px;transform:rotate(25deg);right:5%;top:-90px}
#root .lpr-detail-glyph{font-size:110px;font-weight:800;color:rgba(255,255,255,.8)}
#root .lpr-detail-info{display:flex;justify-content:space-between;gap:15px;flex-wrap:wrap;align-items:center}
#root .lpr-detail-info h1{font-size:29px;margin:2px 0 7px}
#root .lpr-detail-info p{margin:0 0 6px;color:#b0bcc9}
#root .lpr-detail-info small{color:#8894a7}
#root .lpr-status{padding:11px 24px;font-size:11px;color:#9da8b8;border-top:1px solid rgba(255,255,255,.08);min-height:36px;flex-shrink:0}
@media(max-width:1150px){#root .lpr-game-grid{grid-template-columns:repeat(4,minmax(0,1fr))}}
@media(max-width:900px){#root .lpr-game-grid{grid-template-columns:repeat(3,minmax(0,1fr))}#root .lpr-topbar{gap:10px;padding:12px}#root .lpr-navigation{padding:7px 12px}}
@media(max-width:660px){#root .lpr-game-grid{grid-template-columns:repeat(2,minmax(0,1fr))}#root .lpr-content,#root .lpr-detail{padding:20px 16px}#root .lpr-feature{min-height:190px;padding:22px}#root .lpr-feature h2{font-size:29px}#root .lpr-wordmark{min-width:auto}#root .lpr-wordmark strong{display:none}#root .lpr-heading{flex-wrap:wrap}#root .lpr-profile{flex-wrap:wrap}}
`;
