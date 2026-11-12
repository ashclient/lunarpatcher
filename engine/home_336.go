package main

import (
	"errors"
	"fmt"
	"strings"
)

// patchHome336 handles the user-supplied Lunar 3.7.36-ow renderer.
// Every anchor is checked inside the unique native homepage component before
// touching it. Unknown or ambiguous layouts fail closed.
func patchHome336(s string) ([]byte, error) {
	const home = "tq=()=>{"
	const navBefore = "{type:`external`,id:`news`,icon:Kee,tooltip:e(`layout.sidebar.news.tooltip`),url:`https://www.lunarclient.com/news`}"
	const navAfter = "{type:`route`,id:`news`,inactiveIcon:Kee,activeIcon:Kee,tooltip:`Updates`,path:`/updates`}"
	const routeBefore = "(0,V.jsx)(A,{element:(0,V.jsx)(tq,{}),path:`/`})"
	const routeAfter = routeBefore + ",(0,V.jsx)(A,{element:(0,V.jsx)(lunarPatcherUpdatesPage,{}),path:`/updates`})"
	const wrapperBefore = "(0,V.jsx)(g,{display:`flex`,flexDirection:`column`,flexGrow:1,gap:`4`,minHeight:`0`,children:(0,V.jsxs)(Js,{children:["
	const wrapperAfter = "(0,V.jsx)(g,{className:`lunar-patcher-home`,display:`flex`,flexDirection:`column`,flexGrow:1,gap:`4`,minHeight:`0`,children:(0,V.jsxs)(Js,{className:`lunar-patcher-home-content`,children:["
	const heroBefore = "(0,V.jsxs)(rK,{background:u.src,onLoad:"
	const heroAfter = "(0,V.jsxs)(rK,{className:`lunar-patcher-hero`,background:u.src,onLoad:"
	const newsBefore = "(0,V.jsx)(yK,{})"
	const newsAfter = "(0,V.jsx)(lunarPatcherHomeTools,{})"
	const friendsBefore = "(0,V.jsx)(JG,{full:n===!1,mode:D,rootRef:E})"
	const quickPlay = "(0,V.jsx)(GK,{})"
	const right = "(0,V.jsxs)(g,{css:{\"&&::-webkit-scrollbar\":{display:`none`}},display:`flex`,flexDirection:`column`,flexShrink:0,gridColumn:`span 2`,height:`100%`"
	const adMain = "(0,V.jsx)(wf,{...C.ad.main"
	const adSecond = "(0,V.jsx)(wf,{...C.ad.secondary"

	p := strings.Index(s, home)
	if p < 0 || strings.Count(s, home) != 1 {
		return nil, errors.New("3.7.36 home component missing or ambiguous; no patch applied")
	}
	region := s[p:min(len(s), p+5400)]
	required := map[string]string{"wrapper": wrapperBefore, "hero": heroBefore, "news": newsBefore, "friends": friendsBefore, "quick play": quickPlay, "right column": right, "main ad": adMain, "secondary ad": adSecond}
	for name, needle := range required {
		if n := strings.Count(region, needle); n != 1 {
			return nil, fmt.Errorf("3.7.36 home %s signature changed (found %d); no patch applied", name, n)
		}
	}
	if strings.Count(s, navBefore) != 1 || strings.Count(s, routeBefore) != 1 {
		return nil, errors.New("3.7.36 sidebar/route changed; no patch applied")
	}
	// Parse the right column as a complete JSX invocation. This avoids slicing
	// by a brittle chain of closing minified brackets or swallowing launch UI.
	start := p + strings.Index(region, right)
	closeParen, err := matchJSParen(s, start+len("(0,V.jsxs)"), 4200)
	if err != nil {
		return nil, fmt.Errorf("3.7.36 right column could not be parsed: %w; no patch applied", err)
	}
	originalRight := s[start : closeParen+1]
	for _, needle := range []string{friendsBefore, quickPlay, adMain, adSecond} {
		if strings.Count(originalRight, needle) != 1 {
			return nil, errors.New("3.7.36 right column no longer contains uniquely matched friends, quick play, and ads; no patch applied")
		}
	}
	friend := "(0,V.jsx)(g,{className:`lunar-patcher-friends`,display:`flex`,flexDirection:`column`,flexShrink:0,gridColumn:`span 2`,height:`100%`,minHeight:`0`,width:`400px`,ref:i,children:(0,V.jsx)(JG,{full:!0,mode:`list`,rootRef:E})})"
	s = s[:start] + friend + s[closeParen+1:]
	// Rebind news/quick-play only inside the validated native homepage.
	p = strings.Index(s, home)
	region = s[p:min(len(s), p+5400)]
	if strings.Count(region, newsBefore) != 1 {
		return nil, errors.New("3.7.36 news moved during rewrite; no patch applied")
	}
	newsAt := p + strings.Index(region, newsBefore)
	s = s[:newsAt] + newsAfter + s[newsAt+len(newsBefore):]
	for _, edit := range [][3]string{{navBefore, navAfter, "Updates sidebar"}, {routeBefore, routeAfter, "Updates route"}, {wrapperBefore, wrapperAfter, "home wrapper"}, {heroBefore, heroAfter, "hero artwork"}} {
		s, err = replaceUnique(s, edit[0], edit[1], edit[2])
		if err != nil {
			return nil, err
		}
	}
	helpers := "lunarPatcherHomeTools=()=>{const go=r=>window.lunar.misc.switchRoute({route:r,initiator:Ur.sidebar_click});return(0,V.jsxs)(`div`,{className:`lunar-patcher-home-tools`,children:[(0,V.jsxs)(`section`,{className:`lunar-patcher-section`,children:[(0,V.jsx)(`h3`,{children:`Quick Play`}),(0,V.jsx)(GK,{})]}),(0,V.jsxs)(`section`,{className:`lunar-patcher-section`,children:[(0,V.jsx)(`h3`,{children:`Shortcuts`}),(0,V.jsxs)(`div`,{className:`lunar-patcher-shortcuts`,children:[(0,V.jsx)(`button`,{type:`button`,className:`lunar-patcher-shortcut`,onClick:()=>go(`/versions`),children:`Versions`}),(0,V.jsx)(`button`,{type:`button`,className:`lunar-patcher-shortcut`,onClick:()=>go(`/servers`),children:`Servers`}),(0,V.jsx)(`button`,{type:`button`,className:`lunar-patcher-shortcut`,onClick:()=>go(`/explore`),children:`Explore Mods`})]})]})]})},"
	updates := "lunarPatcherUpdatesPage=()=> (0,V.jsxs)(`div`,{className:`lunar-patcher-updates`,children:[(0,V.jsx)(`h1`,{children:`Updates`}),(0,V.jsx)(yK,{})]}),"
	s, err = replaceUnique(s, home, helpers+updates+home, "3.7.36 home tools and Updates components")
	if err != nil {
		return nil, err
	}
	for _, token := range []string{"lunarPatcherHomeTools=", "lunarPatcherUpdatesPage=", "className:`lunar-patcher-friends`", "(0,V.jsx)(JG,{full:!0,mode:`list`,rootRef:E})"} {
		if strings.Count(s, token) != 1 {
			return nil, fmt.Errorf("3.7.36 output missing or duplicated %s; no patch applied", token)
		}
	}
	if strings.Count(s, newsBefore) != 1 || strings.Contains(s, adMain) || strings.Contains(s, adSecond) {
		return nil, errors.New("3.7.36 output still contains homepage ads or misplaced news; no patch applied")
	}
	return []byte(s), nil
}
