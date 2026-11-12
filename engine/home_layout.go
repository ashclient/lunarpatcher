package main

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

// patchHomeLayout replaces the 3.7.32 home composition with a full native
// friends list plus real server quick-play and navigation shortcuts. It keeps
// the original launch toolbar and news renderer, exposing news under Updates.
// These are version-specific, checked transformations: unknown builds fail
// without installing a partially styled or empty home page.
func patchHomeLayout(source []byte) ([]byte, error) {
	src := string(source)
	if strings.Contains(src, "(0,V.jsx)(A,{element:(0,V.jsx)(tq,{}),path:`/`})") {
		return patchHome336(src)
	}
	if strings.Contains(src, "(0,V.jsx)(j,{element:(0,V.jsx)(eq,{}),path:`/`})") {
		return patchHome334(src)
	}
	if strings.Contains(src, "(0,V.jsx)(y,{element:(0,V.jsx)(dW,{}),path:`/`})") {
		return patchHome332(src)
	}
	if strings.Contains(src, "(0,H.jsx)(b,{element:(0,H.jsx)(IK,{}),path:`/`})") {
		return patchHome331(src)
	}
	return nil, errors.New("unsupported home layout: friends and shortcuts not changed; no patch applied")
}

func replaceUnique(s, old, next, name string) (string, error) {
	n := strings.Count(s, old)
	if n != 1 {
		return "", fmt.Errorf("%s signature changed (found %d, expected 1); no patch applied", name, n)
	}
	return strings.Replace(s, old, next, 1), nil
}

// injectHomeTools332 finds the first-column home composition and appends the
// shortcuts beneath its launch panel. We deliberately do not match the exact
// minifier-generated JSX closing sequence: Lunar can change that while keeping
// the relevant component structure intact.
// It is bounded to the known 3.7.32 home function and refuses ambiguous input.
func injectHomeTools332(s string) (string, error) {
	const home = "dW=()=>{"
	const column = "(0,V.jsxs)(El,{children:["
	homeIndex := strings.Index(s, home)
	if homeIndex < 0 || strings.Count(s, home) != 1 {
		return "", errors.New("home component signature changed; no patch applied")
	}
	// Do not let a non-home component with a matching column name get patched.
	regionEnd := homeIndex + 10000
	if regionEnd > len(s) {
		regionEnd = len(s)
	}
	region := s[homeIndex:regionEnd]
	if strings.Count(region, column) != 1 {
		return "", errors.New("home column is missing or ambiguous; no patch applied")
	}
	start := homeIndex + strings.Index(region, column) + len(column) - 1
	end, err := matchJSArray(s, start, 6000)
	if err != nil {
		return "", fmt.Errorf("home column could not be parsed: %w; no patch applied", err)
	}
	contents := s[start+1 : end]
	if strings.Count(contents, "(0,V.jsxs)(pU,{className:`lunar-patcher-hero`,background:u.src,onLoad:") != 1 ||
		strings.Count(contents, "onOptionsClick:()=>{window.lunar.misc.switchRoute") != 1 ||
		strings.Count(contents, "(0,V.jsx)(aU,") != 1 ||
		strings.Count(contents, "(0,V.jsx)(gU,") != 1 {
		return "", fmt.Errorf("launch panel or controls changed (hero=%d, options=%d, toolbar=%d/%d); no patch applied", strings.Count(contents, "(0,V.jsxs)(pU,{className:`lunar-patcher-hero`,background:u.src,onLoad:"), strings.Count(contents, "onOptionsClick:()=>{window.lunar.misc.switchRoute"), strings.Count(contents, "(0,V.jsx)(aU,"), strings.Count(contents, "(0,V.jsx)(gU,"))
	}
	// The first column must still be immediately followed by the separate
	// friends column; otherwise this would insert into an unrelated JSX array.
	after := s[end:]
	if !strings.HasPrefix(after, "]}),(0,V.jsxs)(p,{") ||
		!strings.Contains(after[:min(len(after), 1200)], "(0,V.jsx)(iU,{full:n===!1,mode:O,rootRef:D})") {
		return "", errors.New("friends/home column boundary changed; no patch applied")
	}
	return s[:end] + ",(0,V.jsx)(lunarPatcherHomeTools,{})" + s[end:], nil
}

// matchJSArray finds the closing ] for a source-level array, ignoring bracket
// characters within quoted JS strings/templates. Limit the scan to avoid
// accidentally matching a later component after an upstream syntax change.
func matchJSArray(s string, start, maxBytes int) (int, error) {
	if start < 0 || start >= len(s) || s[start] != '[' {
		return -1, errors.New("expected JS children array")
	}
	depth := 0
	var quote byte
	escaped := false
	limit := start + maxBytes
	if limit > len(s) {
		limit = len(s)
	}
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
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return i, nil
			}
			if depth < 0 {
				return -1, errors.New("unexpected closing bracket")
			}
		}
	}
	return -1, errors.New("unterminated or unusually large JS children array")
}

func patchHome332(s string) ([]byte, error) {
	var err error
	s, err = replaceUnique(s,
		"{type:`external`,id:`news`,icon:ts,tooltip:e(`layout.sidebar.news.tooltip`),url:`https://www.lunarclient.com/news`}",
		"{type:`route`,id:`news`,inactiveIcon:ts,activeIcon:ts,tooltip:`Updates`,path:`/updates`}", "Updates sidebar")
	if err != nil {
		return nil, err
	}
	s, err = replaceUnique(s,
		"(0,V.jsx)(y,{element:(0,V.jsx)(dW,{}),path:`/`})",
		"(0,V.jsx)(y,{element:(0,V.jsx)(dW,{}),path:`/`}),(0,V.jsx)(y,{element:(0,V.jsx)(lunarPatcherUpdatesPage,{}),path:`/updates`})", "Updates route")
	if err != nil {
		return nil, err
	}
	s, err = replaceUnique(s,
		"(0,V.jsx)(p,{display:`flex`,flexDirection:`column`,flexGrow:1,gap:`4`,minHeight:`0`,children:(0,V.jsxs)(Xc,{children:[",
		"(0,V.jsx)(p,{className:`lunar-patcher-home`,display:`flex`,flexDirection:`column`,flexGrow:1,gap:`4`,minHeight:`0`,children:(0,V.jsxs)(Xc,{className:`lunar-patcher-home-content`,children:[", "home container")
	if err != nil {
		return nil, err
	}
	s, err = replaceUnique(s, "(0,V.jsxs)(pU,{background:u.src,onLoad:", "(0,V.jsxs)(pU,{className:`lunar-patcher-hero`,background:u.src,onLoad:", "hero container")
	if err != nil {
		return nil, err
	}
	s, err = injectHomeTools332(s)
	if err != nil {
		return nil, err
	}
	// The existing friends component supports a real full-list display; do not
	// clone avatar DOM or replace it with fake entries.
	constStart := "(0,V.jsxs)(p,{css:{\"&&::-webkit-scrollbar\":{display:`none`}},display:`flex`,flexDirection:`column`,flexShrink:0,gridColumn:`span 2`"
	start := strings.Index(s, constStart)
	if start < 0 || strings.Count(s, constStart) != 1 {
		return nil, errors.New("3.7.32 friends container changed; no patch applied")
	}
	endMark := "willShowAd:n})]})]})]})})},fW="
	tail := strings.Index(s[start:], endMark)
	if tail < 0 || tail > 2500 {
		return nil, errors.New("3.7.32 friends boundary changed; no patch applied")
	}
	// Keep the parent home JSX closing tokens after the friends JSX.
	end := start + tail + len("willShowAd:n})]})]})")
	old := s[start:end]
	if !strings.Contains(old, "(0,V.jsx)(iU,{full:n===!1,mode:O,rootRef:D})") || !strings.Contains(old, "(0,V.jsx)(nW,{})") {
		return nil, errors.New("friends list / quick-play signature changed")
	}
	newFriends := "(0,V.jsx)(p,{className:`lunar-patcher-friends`,display:`flex`,flexDirection:`column`,flexShrink:0,gridColumn:`span 2`,height:`100%`,minHeight:`0`,width:`360px`,ref:i,children:(0,V.jsx)(iU,{full:!0,mode:`list`,rootRef:D})})"
	s = s[:start] + newFriends + s[end:]
	helpers := "lunarPatcherHomeTools=()=>{const go=(r)=>window.lunar.misc.switchRoute({route:r,initiator:Pn.sidebar_click});return(0,V.jsxs)(`div`,{className:`lunar-patcher-home-tools`,children:[(0,V.jsxs)(`section`,{className:`lunar-patcher-section`,children:[(0,V.jsx)(`h3`,{children:`Quick Play`}),(0,V.jsx)(nW,{})]}),(0,V.jsxs)(`section`,{className:`lunar-patcher-section`,children:[(0,V.jsx)(`h3`,{children:`Shortcuts`}),(0,V.jsxs)(`div`,{className:`lunar-patcher-shortcuts`,children:[(0,V.jsx)(`button`,{type:`button`,className:`lunar-patcher-shortcut`,onClick:()=>go(`/versions`),children:`Versions`}),(0,V.jsx)(`button`,{type:`button`,className:`lunar-patcher-shortcut`,onClick:()=>go(`/servers`),children:`Servers`}),(0,V.jsx)(`button`,{type:`button`,className:`lunar-patcher-shortcut`,onClick:()=>go(`/explore`),children:`Explore Mods`})]})]})]})},"
	updates := "lunarPatcherUpdatesPage=()=> (0,V.jsxs)(`div`,{className:`lunar-patcher-updates`,children:[(0,V.jsx)(`h1`,{children:`Updates`}),(0,V.jsx)(kU,{})]}),"
	s, err = replaceUnique(s, "uW=Ke,dW=()=>", "uW=Ke,"+helpers+updates+"dW=()=>", "home/updates component injection")
	if err != nil {
		return nil, err
	}
	if bytes.Contains([]byte(s), []byte("kU=()=>null")) {
		return nil, errors.New("news component unexpectedly absent from Updates route")
	}
	return []byte(s), nil
}

// Previous 3.7.31 format. Keep its working React-based layout transformation
// so a small update to the app never regresses existing users.
func patchHome331(s string) ([]byte, error) {
	var err error
	s, err = replaceUnique(s,
		"{type:`external`,id:`news`,icon:rs,tooltip:e(`layout.sidebar.news.tooltip`),url:`https://www.lunarclient.com/news`}",
		"{type:`route`,id:`news`,inactiveIcon:rs,activeIcon:rs,tooltip:`Updates`,path:`/updates`}", "Updates sidebar")
	if err != nil {
		return nil, err
	}
	s, err = replaceUnique(s, "(0,H.jsx)(b,{element:(0,H.jsx)(IK,{}),path:`/`})",
		"(0,H.jsx)(b,{element:(0,H.jsx)(IK,{}),path:`/`}),(0,H.jsx)(b,{element:(0,H.jsx)(lunarPatcherUpdatesPage,{}),path:`/updates`})", "Updates route")
	if err != nil {
		return nil, err
	}
	s, err = replaceUnique(s, "(0,H.jsx)(eK,{})", "(0,H.jsx)(lunarPatcherHomeTools,{})", "home news replacement")
	if err != nil {
		return nil, err
	}
	startMark := "(0,H.jsxs)(m,{css:{\"&&::-webkit-scrollbar\":{display:`none`}},display:`flex`,flexDirection:`column`,flexShrink:0,gridColumn:`span 2`"
	start := strings.Index(s, startMark)
	if start < 0 || strings.Count(s, startMark) != 1 {
		return nil, errors.New("3.7.31 friends layout changed")
	}
	endMark := "willShowAd:n})]})]})]})})},"
	tail := strings.Index(s[start:], endMark)
	if tail < 0 || tail > 2500 {
		return nil, errors.New("3.7.31 friends boundary changed")
	}
	end := start + tail + len("willShowAd:n})]})]})")
	old := s[start:end]
	if !strings.Contains(old, "(0,H.jsx)(kG,{full:n===!1,mode:O,rootRef:D})") {
		return nil, errors.New("3.7.31 friends component changed")
	}
	s = s[:start] + "(0,H.jsx)(m,{className:`lunar-patcher-friends`,display:`flex`,flexDirection:`column`,flexShrink:0,gridColumn:`span 2`,height:`100%`,minHeight:`0`,width:`360px`,ref:i,children:(0,H.jsx)(kG,{full:!0,mode:`list`,rootRef:D})})" + s[end:]
	helpers := "lunarPatcherHomeTools=()=>{const go=(r)=>window.lunar.misc.switchRoute({route:r,initiator:In.sidebar_click});return(0,H.jsxs)(`div`,{className:`lunar-patcher-home-tools`,children:[(0,H.jsxs)(`section`,{className:`lunar-patcher-section`,children:[(0,H.jsx)(`h3`,{children:`Quick Play`}),(0,H.jsx)(DK,{})]}),(0,H.jsxs)(`section`,{className:`lunar-patcher-section`,children:[(0,H.jsx)(`h3`,{children:`Shortcuts`}),(0,H.jsxs)(`div`,{className:`lunar-patcher-shortcuts`,children:[(0,H.jsx)(`button`,{type:`button`,className:`lunar-patcher-shortcut`,onClick:()=>go(`/versions`),children:`Versions`}),(0,H.jsx)(`button`,{type:`button`,className:`lunar-patcher-shortcut`,onClick:()=>go(`/servers`),children:`Servers`}),(0,H.jsx)(`button`,{type:`button`,className:`lunar-patcher-shortcut`,onClick:()=>go(`/explore`),children:`Explore Mods`})]})]})]})},"
	updates := "lunarPatcherUpdatesPage=()=> (0,H.jsxs)(`div`,{className:`lunar-patcher-updates`,children:[(0,H.jsx)(`h1`,{children:`Updates`}),(0,H.jsx)(eK,{})]}),"
	s, err = replaceUnique(s, "FK=Je,IK=()=>{", "FK=Je,"+helpers+updates+"IK=()=>{", "home and updates injection")
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}
