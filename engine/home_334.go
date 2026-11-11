package main

import (
	"errors"
	"fmt"
	"strings"
)

// patchHome334 applies checked, renderer-specific transformations for 3.7.34-ow.
// The source component is the launcher's native home, NOT a reconstructed fake UI.
func patchHome334(s string) ([]byte, error) {
	const home = "eq=()=>{"
	const oldNav = "{type:`external`,id:`news`,icon:Kee,tooltip:e(`layout.sidebar.news.tooltip`),url:`https://www.lunarclient.com/news`}"
	const newNav = "{type:`route`,id:`news`,inactiveIcon:Kee,activeIcon:Kee,tooltip:`Updates`,path:`/updates`}"
	const oldRoute = "(0,V.jsx)(j,{element:(0,V.jsx)(eq,{}),path:`/`})"
	const newRoute = oldRoute + ",(0,V.jsx)(j,{element:(0,V.jsx)(lunarPatcherUpdatesPage,{}),path:`/updates`})"
	const oldWrapper = "(0,V.jsx)(g,{display:`flex`,flexDirection:`column`,flexGrow:1,gap:`4`,minHeight:`0`,children:(0,V.jsxs)(Ys,{children:["
	const newWrapper = "(0,V.jsx)(g,{className:`lunar-patcher-home`,display:`flex`,flexDirection:`column`,flexGrow:1,gap:`4`,minHeight:`0`,children:(0,V.jsxs)(Ys,{className:`lunar-patcher-home-content`,children:["
	const oldHero = "(0,V.jsxs)(nK,{background:u.src,onLoad:"
	const newHero = "(0,V.jsxs)(nK,{className:`lunar-patcher-hero`,background:u.src,onLoad:"
	const oldNews = "(0,V.jsx)(vK,{})"
	const newNews = "(0,V.jsx)(lunarPatcherHomeTools,{})"
	var err error
	// Guard the entire home JSX mutation by confirming all signatures are uniquely
	// present in the bounded home component, before writing anything.
	i := strings.Index(s, home)
	if i < 0 || strings.Count(s, home) != 1 {
		return nil, errors.New("3.7.34 home component signature missing or ambiguous; no patch applied")
	}
	end := i + 4200
	if end > len(s) {
		end = len(s)
	}
	region := s[i:end]
	for name, needle := range map[string]string{"wrapper": oldWrapper, "hero": oldHero, "news": oldNews, "friends": "(0,V.jsx)(qG,{full:n===!1,mode:D,rootRef:E})", "ads": "(0,V.jsx)(xf,{...C.ad.main"} {
		if strings.Count(region, needle) != 1 {
			return nil, fmt.Errorf("3.7.34 home %s signature changed (found %d); no patch applied", name, strings.Count(region, needle))
		}
	}
	for _, item := range [][3]string{{oldNav, newNav, "Updates sidebar"}, {oldRoute, newRoute, "Updates route"}, {oldWrapper, newWrapper, "home wrapper"}, {oldHero, newHero, "launch panel"}} {
		s, err = replaceUnique(s, item[0], item[1], item[2])
		if err != nil {
			return nil, err
		}
	}
	// News only moves after the native home component was checked above.
	// Avoid rewriting a news instance elsewhere in Lunar.
	p := strings.Index(s, home)
	region = s[p:min(len(s), p+4400)]
	if strings.Count(region, oldNews) != 1 {
		return nil, errors.New("3.7.34 home news moved; no patch applied")
	}
	newsAt := p + strings.Index(region, oldNews)
	s = s[:newsAt] + newNews + s[newsAt+len(oldNews):]
	// Replace only the matched right column: the old one combines friends, a
	// horizontal Quick Play shelf, and two ad-only slots. This removes both empty
	// ad placeholders while using Lunar's real qG full friends component.
	const right = "(0,V.jsxs)(g,{css:{\"&&::-webkit-scrollbar\":{display:`none`}},display:`flex`,flexDirection:`column`,flexShrink:0,gridColumn:`span 2`,height:`100%`"
	p = strings.Index(s, home)
	region = s[p:min(len(s), p+6000)]
	if strings.Count(region, right) != 1 {
		return nil, errors.New("3.7.34 friends/ad column changed; no patch applied")
	}
	start := p + strings.Index(region, right)
	// The callee expression begins after the `(0,V.jsxs)` function pointer.
	paren := start + len("(0,V.jsxs)")
	finish, e := matchJSParen(s, paren, 4000)
	if e != nil {
		return nil, fmt.Errorf("3.7.34 home friends column parse: %w; no patch applied", e)
	}
	old := s[start : finish+1]
	for _, needle := range []string{"(0,V.jsx)(WK,{})", "(0,V.jsx)(qG,{full:n===!1,mode:D,rootRef:E})", "(0,V.jsx)(xf,{...C.ad.main", "(0,V.jsx)(xf,{...C.ad.secondary"} {
		if !strings.Contains(old, needle) {
			return nil, fmt.Errorf("3.7.34 friends/ad column missing %s; no patch applied", needle)
		}
	}
	friend := "(0,V.jsx)(g,{className:`lunar-patcher-friends`,display:`flex`,flexDirection:`column`,flexShrink:0,gridColumn:`span 2`,height:`100%`,minHeight:`0`,width:`360px`,ref:i,children:(0,V.jsx)(qG,{full:!0,mode:`list`,rootRef:E})})"
	s = s[:start] + friend + s[finish+1:]
	// Native Quick Play shelf WK is now in the left column, rather than lost
	// with the right-column ad section.
	helpers := "lunarPatcherHomeTools=()=>{const go=r=>window.lunar.misc.switchRoute({route:r,initiator:Hr.sidebar_click});return(0,V.jsxs)(`div`,{className:`lunar-patcher-home-tools`,children:[(0,V.jsxs)(`section`,{className:`lunar-patcher-section`,children:[(0,V.jsx)(`h3`,{children:`Quick Play`}),(0,V.jsx)(WK,{})]}),(0,V.jsxs)(`section`,{className:`lunar-patcher-section`,children:[(0,V.jsx)(`h3`,{children:`Shortcuts`}),(0,V.jsxs)(`div`,{className:`lunar-patcher-shortcuts`,children:[(0,V.jsx)(`button`,{type:`button`,className:`lunar-patcher-shortcut`,onClick:()=>go(`/versions`),children:`Versions`}),(0,V.jsx)(`button`,{type:`button`,className:`lunar-patcher-shortcut`,onClick:()=>go(`/servers`),children:`Servers`}),(0,V.jsx)(`button`,{type:`button`,className:`lunar-patcher-shortcut`,onClick:()=>go(`/explore`),children:`Explore Mods`})]})]})]})},"
	updates := "lunarPatcherUpdatesPage=()=> (0,V.jsxs)(`div`,{className:`lunar-patcher-updates`,children:[(0,V.jsx)(`h1`,{children:`Updates`}),(0,V.jsx)(vK,{})]}),"
	s, err = replaceUnique(s, home, helpers+updates+home, "3.7.34 home tools and Updates components")
	if err != nil {
		return nil, err
	}
	if strings.Count(s, "lunarPatcherHomeTools=") != 1 || strings.Count(s, "lunarPatcherUpdatesPage=") != 1 || strings.Count(s, "(0,V.jsx)(vK,{})") != 1 {
		return nil, errors.New("3.7.34 post-patch news/tools structure mismatch; no patch applied")
	}
	return []byte(s), nil
}

// matchJSParen matches the full JSX call's opening `(` without being fooled by
// quoted CSS or template strings in its props. It bounds the scan to one column.
func matchJSParen(s string, start, maxBytes int) (int, error) {
	if start < 0 || start >= len(s) || s[start] != '(' {
		return -1, errors.New("expected opening call parenthesis")
	}
	depth := 0
	var quote byte
	escaped := false
	limit := min(len(s), start+maxBytes)
	for i := start; i < limit; i++ {
		ch := s[i]
		if quote != 0 {
			if escaped {
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == quote {
				quote = 0
			}
			continue
		}
		switch ch {
		case '\'', '"', '`':
			quote = ch
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i, nil
			}
			if depth < 0 {
				return -1, errors.New("unexpected closing parenthesis")
			}
		}
	}
	return -1, errors.New("unterminated or unusually long JSX call")
}
