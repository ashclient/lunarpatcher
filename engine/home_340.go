package main

import (
	"errors"
	"fmt"
	"strings"
)

// patchHome340 supports the inspected Lunar Client 3.7.40-ow homepage.
// Every edit is scoped to checked native components. Unknown structures fail
// without returning a partially rewritten renderer.
func patchHome340(s string) ([]byte, error) {
	const home = "Fm=()=>{"
	const navOld = "{type:`external`,id:`news`,icon:yt,tooltip:e(`layout.sidebar.news.tooltip`),url:`https://www.lunarclient.com/news`}"
	const navNew = "{type:`route`,id:`news`,inactiveIcon:yt,activeIcon:yt,tooltip:`Updates`,path:`/updates`}"
	const routeOld = "(0,Z.jsx)(D,{element:(0,Z.jsx)(Fm,{}),path:`/`})"
	const routeNew = routeOld + ",(0,Z.jsx)(D,{element:(0,Z.jsx)(lunarPatcherUpdatesPage,{}),path:`/updates`})"
	const wrapperOld = "(0,Z.jsx)(I,{display:`flex`,flexDirection:`column`,flexGrow:1,gap:`4`,minHeight:`0`,children:(0,Z.jsxs)(Oo,{children:["
	const wrapperNew = "(0,Z.jsx)(I,{className:`lunar-patcher-home`,display:`flex`,flexDirection:`column`,flexGrow:1,gap:`4`,minHeight:`0`,children:(0,Z.jsxs)(Oo,{className:`lunar-patcher-home-content`,children:["
	const heroOld = "(0,Z.jsxs)(Lp,{background:l.src,onLoad:"
	const heroNew = "(0,Z.jsxs)(Lp,{className:`lunar-patcher-hero`,background:l.src,onLoad:"
	const newsOld = "(0,Z.jsx)(Zp,{})"
	const newsNew = "(0,Z.jsx)(lunarPatcherHomeTools,{})"
	const quickPlay = "(0,Z.jsx)(jm,{})"
	const friendsOld = "(0,Z.jsx)(yp,{full:t===!1,mode:x,rootRef:b})"
	const mainAd = "(0,Z.jsx)(Ha,{..._.ad.main"
	const secondAd = "(0,Z.jsx)(Ha,{..._.ad.secondary"
	const right = "(0,Z.jsxs)(I,{css:{\"&&::-webkit-scrollbar\":{display:`none`}},display:`flex`,flexDirection:`column`,flexShrink:0,gridColumn:`span 2`,height:`100%`"

	p := strings.Index(s, home)
	if p < 0 || strings.Count(s, home) != 1 {
		return nil, errors.New("3.7.40 homepage signature changed; no patch applied")
	}
	region := s[p:min(len(s), p+6500)]
	for name, needle := range map[string]string{
		"wrapper": wrapperOld, "launch panel": heroOld, "news": newsOld, "quick play": quickPlay,
		"friends": friendsOld, "right column": right, "main ad": mainAd, "secondary ad": secondAd,
	} {
		if n := strings.Count(region, needle); n != 1 {
			return nil, fmt.Errorf("3.7.40 %s signature changed (found %d); no patch applied", name, n)
		}
	}
	if strings.Count(s, navOld) != 1 || strings.Count(s, routeOld) != 1 {
		return nil, errors.New("3.7.40 sidebar or home route changed; no patch applied")
	}
	start := p + strings.Index(region, right)
	end, err := matchJSParen(s, start+len("(0,Z.jsxs)"), 4600)
	if err != nil {
		return nil, fmt.Errorf("3.7.40 right column parse: %w; no patch applied", err)
	}
	column := s[start : end+1]
	for _, token := range []string{quickPlay, friendsOld, mainAd, secondAd} {
		if strings.Count(column, token) != 1 {
			return nil, errors.New("3.7.40 right column changed; no patch applied")
		}
	}
	// Remove the original ad containers and compact Quick Play, retaining
	// Lunar's own Friends component and its tracked layout ref.
	friends := "(0,Z.jsx)(I,{className:`lunar-patcher-friends`,display:`flex`,flexDirection:`column`,flexShrink:0,gridColumn:`span 2`,height:`100%`,minHeight:`0`,width:`400px`,ref:r,children:(0,Z.jsx)(yp,{full:!0,mode:`list`,rootRef:b})})"
	s = s[:start] + friends + s[end+1:]

	p = strings.Index(s, home)
	region = s[p:min(len(s), p+6500)]
	if strings.Count(region, newsOld) != 1 {
		return nil, errors.New("3.7.40 news component moved; no patch applied")
	}
	pos := p + strings.Index(region, newsOld)
	s = s[:pos] + newsNew + s[pos+len(newsOld):]
	for _, edit := range [][3]string{
		{navOld, navNew, "3.7.40 Updates sidebar"},
		{routeOld, routeNew, "3.7.40 Updates route"},
		{wrapperOld, wrapperNew, "3.7.40 home wrapper"},
		{heroOld, heroNew, "3.7.40 hero"},
	} {
		s, err = replaceUnique(s, edit[0], edit[1], edit[2])
		if err != nil {
			return nil, err
		}
	}
	helpers := "lunarPatcherHomeTools=()=>{const go=r=>window.lunar.misc.switchRoute({route:r,initiator:z.sidebar_click});return(0,Z.jsxs)(`div`,{className:`lunar-patcher-home-tools`,children:[(0,Z.jsxs)(`section`,{className:`lunar-patcher-section`,children:[(0,Z.jsx)(`h3`,{children:`Quick Play`}),(0,Z.jsx)(jm,{})]}),(0,Z.jsxs)(`section`,{className:`lunar-patcher-section`,children:[(0,Z.jsx)(`h3`,{children:`Shortcuts`}),(0,Z.jsxs)(`div`,{className:`lunar-patcher-shortcuts`,children:[(0,Z.jsx)(`button`,{type:`button`,className:`lunar-patcher-shortcut`,onClick:()=>go(`/versions`),children:`Versions`}),(0,Z.jsx)(`button`,{type:`button`,className:`lunar-patcher-shortcut`,onClick:()=>go(`/servers`),children:`Servers`}),(0,Z.jsx)(`button`,{type:`button`,className:`lunar-patcher-shortcut`,onClick:()=>go(`/explore`),children:`Explore Mods`})]})]})]})},"
	updates := "lunarPatcherUpdatesPage=()=> (0,Z.jsxs)(`div`,{className:`lunar-patcher-updates`,children:[(0,Z.jsx)(`h1`,{children:`Updates`}),(0,Z.jsx)(Zp,{})]}),"
	s, err = replaceUnique(s, home, helpers+updates+home, "3.7.40 tool and news declarations")
	if err != nil {
		return nil, err
	}
	for _, token := range []string{"lunarPatcherHomeTools=", "lunarPatcherUpdatesPage=", "className:`lunar-patcher-friends`", "className:`lunar-patcher-hero`", "(0,Z.jsx)(yp,{full:!0,mode:`list`,rootRef:b})"} {
		if strings.Count(s, token) != 1 {
			return nil, fmt.Errorf("3.7.40 output %s count invalid; no patch applied", token)
		}
	}
	if strings.Count(s, newsOld) != 1 || strings.Contains(s, mainAd) || strings.Contains(s, secondAd) || strings.Count(s, quickPlay) != 1 {
		return nil, errors.New("3.7.40 output news/ads/quick play validation failed; no patch applied")
	}
	return []byte(s), nil
}
