package uasurfer

import "strings"

// appMarkers are the names that appear in an agent alongside Chrome's or
// Safari's: the apps that render a link in a frame of their own, and the
// browsers that ship Chrome's engine under their own brand. Every check in
// parseBrowserName would claim one of these for the engine underneath, so they
// are read first.
var appMarkers = []struct {
	s    string
	name BrowserName
}{
	// in-app webviews
	{"fbav", BrowserFacebook}, // also FBAN and FB_IAB: one app, several spellings
	{"fban", BrowserFacebook},
	{"fb_iab", BrowserFacebook},
	{"instagram", BrowserInstagram},
	{"micromessenger", BrowserWeChat},
	{"bytedancewebview", BrowserTikTok},
	{"musical_ly", BrowserTikTok},
	{"trill_", BrowserTikTok},
	{"snapchat", BrowserSnapchat},
	{"line/", BrowserLine},

	// Chromium under another brand
	{"vivaldi", BrowserVivaldi},
	{"whale", BrowserWhale},
	{"miuibrowser", BrowserMIUI},
	{"huaweibrowser", BrowserHuawei},
	{"duckduckgo", BrowserDuckDuckGo},
}

// appBuckets indexes appMarkers by first byte, and appFirstBytes is the set of
// those bytes. One pass with a bit test per byte, rather than a strings.Contains
// per marker: Contains costs the same fifteen nanoseconds however short the
// haystack, which fifteen markers of it would put on every parse.
var appBuckets, appFirstBytes = func() (buckets [256][]int, first [4]uint64) {
	for i, m := range appMarkers {
		c := m.s[0]
		buckets[c] = append(buckets[c], i)
		first[c>>6] |= 1 << (c & 63)
	}
	return
}()

// appBrowser returns the app or brand named in s, or BrowserUnknown.
//
// A marker only counts at the start of a field. Unanchored, "line/" also ends
// "baseline/" and "miuibrowser" would be found in any longer word containing it;
// anchoring costs one comparison and removes the whole class of mistake.
func appBrowser(s string) BrowserName {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if appFirstBytes[c>>6]&(1<<(c&63)) == 0 {
			continue
		}
		if i > 0 && !isAppEdge(s[i-1]) {
			continue
		}
		for _, m := range appBuckets[c] {
			if strings.HasPrefix(s[i:], appMarkers[m].s) {
				return appMarkers[m].name
			}
		}
	}
	return BrowserUnknown
}

// isAppEdge reports whether c can precede a name: whitespace, or the punctuation
// agents use to open a field. "/" is here for Xiaomi, which writes its browser
// as "XiaoMi/MiuiBrowser/17.4.11".
func isAppEdge(c byte) bool {
	switch c {
	case ';', '(', '[', '/', ',', '-':
		return true
	}
	return isSpace(c)
}

// webkitApp returns the app rendering the page on an Apple engine, for the
// agents that carry no Chrome token for chromiumBrowser to work from.
//
// The gate is the "Mobile/<build>" token that marks the agent as iOS at all; a
// desktop agent pays for one comparison. Only the tail after it is searched,
// as in chromiumBrowser: an app writes its name at the end of the agent, and
// plain Safari has some twenty bytes of tail. "Version/<n>" is deliberately no
// gate: real Safari states it and most embedded WebKits do not, but the apps
// that build their agent by appending their name to Safari's, as Line does,
// state it too.
func webkitApp(ua string) BrowserName {
	_, tail, ok := strings.Cut(ua, "mobile/")
	if !ok {
		return BrowserUnknown
	}
	return appBrowser(tail)
}

// chromiumBrowser returns the browser behind a Chromium agent: Chrome unless
// something else put its name to the engine.
//
// Only the tail after the Chrome token is searched. Everything a skin or an
// Android webview adds sits there, either side of the Safari token, and a plain
// Chrome agent has some twenty bytes of tail.
func chromiumBrowser(ua string) BrowserName {
	_, tail, ok := strings.Cut(ua, "chrome/")
	if !ok {
		// crios, chromium and crmo: Chrome's own spellings, which nothing here
		// is built on.
		return BrowserChrome
	}
	if name := appBrowser(tail); name != BrowserUnknown {
		return name
	}
	return BrowserChrome
}
