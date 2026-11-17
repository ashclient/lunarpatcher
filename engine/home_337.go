package main

import (
	"errors"
	"fmt"
	"strings"
)

// patchHome337 targets the specific homepage composition in Lunar 3.7.37-ow.
// Every replacement is guarded against the actual native launcher component.
// Failure returns no output; patchSource never saves partial changes.
func patchHome337(s string) ([]byte, error) {
	const home = "$K=()=>{"
	const navOld = "{type:`external`,id:`news`,icon:Bo,tooltip:e(`layout.sidebar.news.tooltip`),url:`https://www.lunarclient.com/news`}"
	const navNew = "{type:`route`,id:`news`,inactiveIcon:Bo,activeIcon:Bo,tooltip:`Updates`,path:`/updates`}"
	const routeOld = "(0,V.jsx)(S,{element:(0,V.jsx)($K,{}),path:`/`})"
	const routeNew = routeOld + ",(0,V.jsx)(S,{element:(0,V.jsx)(lunarPatcherUpdatesPage,{}),path:`/updates`})"
	const wrapperOld = "(0,V.jsx)(h,{display:`flex`,flexDirection:`column`,flexGrow:1,gap:`4`,minHeight:`0`,children:(0,V.jsxs)(Kc,{children:["
	const wrapperNew = "(0,V.jsx)(h,{className:`lunar-patcher-home`,display:`flex`,flexDirection:`column`,flexGrow:1,gap:`4`,minHeight:`0`,children:(0,V.jsxs)(Kc,{className:`lunar-patcher-home-content`,children:["
	const heroOld = "(0,V.jsxs)(tK,{background:u.src,onLoad:"
	const heroNew = "(0,V.jsxs)(tK,{className:`lunar-patcher-hero`,background:u.src,onLoad:"
	const newsOld = "(0,V.jsx)(_K,{})"
	const newsNew = "(0,V.jsx)(lunarPatcherHomeTools,{})"
	const quickPlay = "(0,V.jsx)(UK,{})"
	const friendsOld = "(0,V.jsx)(KG,{full:n===!1,mode:ee,rootRef:O})"
	const mainAd = "(0,V.jsx)(wf,{...T.ad.main"
	const secondAd = "(0,V.jsx)(wf,{...T.ad.secondary"
	const right = "(0,V.jsxs)(h,{css:{\"&&::-webkit-scrollbar\":{display:`none`}},display:`flex`,flexDirection:`column`,flexShrink:0,gridColumn:`span 2`,height:`100%`"

	p := strings.Index(s, home)
	if p < 0 || strings.Count(s, home) != 1 {
		return nil, errors.New("3.7.37 homepage signature missing or ambiguous; no patch applied")
	}
	region := s[p:min(len(s), p+6000)]
	for name, needle := range map[string]string{
		"home wrapper": wrapperOld, "launch hero": heroOld, "news": newsOld,
		"quick play": quickPlay, "full friends component": friendsOld,
		"friends/ad column": right, "main ad": mainAd, "secondary ad": secondAd,
	} {
		if count := strings.Count(region, needle); count != 1 {
			return nil, fmt.Errorf("3.7.37 %s signature changed (found %d); no patch applied", name, count)
		}
	}
	if strings.Count(s, navOld) != 1 || strings.Count(s, routeOld) != 1 {
		return nil, errors.New("3.7.37 sidebar or homepage route changed; no patch applied")
	}
	start := p + strings.Index(region, right)
	end, err := matchJSParen(s, start+len("(0,V.jsxs)"), 4200)
	if err != nil {
		return nil, fmt.Errorf("3.7.37 friends/ad column parse: %w; no patch applied", err)
	}
	column := s[start : end+1]
	for _, n := range []string{quickPlay, friendsOld, mainAd, secondAd} {
		if strings.Count(column, n) != 1 {
			return nil, fmt.Errorf("3.7.37 friends/ad column missing or duplicated %s; no patch applied", n)
		}
	}
	// Render Lunar's own Friends component in full mode. The ad slots and
	// compact Quick Play strip disappear with the native old right column.
	friends := "(0,V.jsx)(h,{className:`lunar-patcher-friends`,display:`flex`,flexDirection:`column`,flexShrink:0,gridColumn:`span 2`,height:`100%`,minHeight:`0`,width:`400px`,ref:i,children:(0,V.jsx)(KG,{full:!0,mode:`list`,rootRef:O})})"
	s = s[:start] + friends + s[end+1:]

	p = strings.Index(s, home)
	region = s[p:min(len(s), p+6000)]
	if strings.Count(region, newsOld) != 1 {
		return nil, errors.New("3.7.37 native home news moved unexpectedly; no patch applied")
	}
	at := p + strings.Index(region, newsOld)
	s = s[:at] + newsNew + s[at+len(newsOld):]
	for _, edit := range [][3]string{
		{navOld, navNew, "3.7.37 Updates sidebar"},
		{routeOld, routeNew, "3.7.37 Updates route"},
		{wrapperOld, wrapperNew, "3.7.37 home wrapper"},
		{heroOld, heroNew, "3.7.37 launch panel"},
	} {
		s, err = replaceUnique(s, edit[0], edit[1], edit[2])
		if err != nil {
			return nil, err
		}
	}
	// Never forge server entries. Native Quick Play is re-mounted directly.
	helpers := "lunarPatcherHomeTools=()=>{const go=r=>window.lunar.misc.switchRoute({route:r,initiator:gi.sidebar_click});return(0,V.jsxs)(`div`,{className:`lunar-patcher-home-tools`,children:[(0,V.jsxs)(`section`,{className:`lunar-patcher-section`,children:[(0,V.jsx)(`h3`,{children:`Quick Play`}),(0,V.jsx)(UK,{})]}),(0,V.jsxs)(`section`,{className:`lunar-patcher-section`,children:[(0,V.jsx)(`h3`,{children:`Shortcuts`}),(0,V.jsxs)(`div`,{className:`lunar-patcher-shortcuts`,children:[(0,V.jsx)(`button`,{type:`button`,className:`lunar-patcher-shortcut`,onClick:()=>go(`/versions`),children:`Versions`}),(0,V.jsx)(`button`,{type:`button`,className:`lunar-patcher-shortcut`,onClick:()=>go(`/servers`),children:`Servers`}),(0,V.jsx)(`button`,{type:`button`,className:`lunar-patcher-shortcut`,onClick:()=>go(`/explore`),children:`Explore Mods`})]})]})]})},"
	updates := "lunarPatcherUpdatesPage=()=> (0,V.jsxs)(`div`,{className:`lunar-patcher-updates`,children:[(0,V.jsx)(`h1`,{children:`Updates`}),(0,V.jsx)(_K,{})]}),"
	s, err = replaceUnique(s, home, helpers+updates+home, "3.7.37 home tools/Updates component declaration")
	if err != nil {
		return nil, err
	}
	for _, token := range []string{"lunarPatcherHomeTools=", "lunarPatcherUpdatesPage=", "className:`lunar-patcher-friends`", "className:`lunar-patcher-hero`", "(0,V.jsx)(KG,{full:!0,mode:`list`,rootRef:O})", "path:`/updates`"} {
		if strings.Count(s, token) != 1 && token != "path:`/updates`" {
			return nil, fmt.Errorf("3.7.37 output component %s missing or duplicated; no patch applied", token)
		}
	}
	if strings.Count(s, newsOld) != 1 || strings.Contains(s, mainAd) || strings.Contains(s, secondAd) || strings.Count(s, quickPlay) != 1 {
		return nil, errors.New("3.7.37 news/ads/quick play output did not validate; no patch applied")
	}
	return []byte(s), nil
}
