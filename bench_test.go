package uasurfer

import (
	"strings"
	"testing"
)

// Benchmarks live together rather than beside the code they measure: the parse
// path is one budget, and comparing a change means running the whole set.
//
//	go test -run=XXX -bench=Parse -benchmem -count=6

func BenchmarkInternal_parseOS(b *testing.B) {
	var v UserAgent
	for i := 0; b.Loop(); i++ {
		v.parseOS(testUAVars[i%len(testUAVars)].UA, nil)
	}
}

func BenchmarkInternal_parseBrowserName(b *testing.B) {
	var v UserAgent
	for i := 0; b.Loop(); i++ {
		v.parseBrowserName(testUAVars[i%len(testUAVars)].UA)
	}
}

func BenchmarkInternal_parseBrowserVersion(b *testing.B) {
	var v UserAgent
	for i := 0; b.Loop(); i++ {
		want := testUAVars[i%len(testUAVars)]
		v.Browser.Name = want.Browser.Name
		v.parseBrowserVersion(want.UA)
	}
}

func BenchmarkInternal_parseDevice(b *testing.B) {
	var v UserAgent
	for i := 0; b.Loop(); i++ {
		want := testUAVars[i%len(testUAVars)]
		v.OS.Name = want.OS.Name
		v.OS.Platform = want.OS.Platform
		v.Browser.Name = want.Browser.Name
		v.parseDevice(want.UA)
	}
}

func BenchmarkInternal_botName(b *testing.B) {
	agent := normalise("Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	for b.Loop() {
		botName(agent)
	}
}

// The pass every parse pays for: a browser agent that matches nothing.
func BenchmarkInternal_botNameMiss(b *testing.B) {
	agent := normalise("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	for b.Loop() {
		botName(agent)
	}
}

// ----------------------------------------------------------------------------

func BenchmarkParse(b *testing.B) {
	for i := 0; b.Loop(); i++ {
		Parse(testUAVars[i%len(testUAVars)].UA)
	}
}

func BenchmarkParseReuse(b *testing.B) {
	dest := new(UserAgent)
	for i := 0; b.Loop(); i++ {
		dest.Reset()
		ParseUserAgent(testUAVars[i%len(testUAVars)].UA, dest)
	}
}

// Chrome for Mac
func BenchmarkParseChromeMac(b *testing.B) {
	for b.Loop() {
		Parse("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.130 Safari/537.36")
	}
}

// Chrome for Windows
func BenchmarkParseChromeWin(b *testing.B) {
	for b.Loop() {
		Parse("Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.134 Safari/537.36")
	}
}

// Chrome for Android
func BenchmarkParseChromeAndroid(b *testing.B) {
	for b.Loop() {
		Parse("Mozilla/5.0 (Linux; Android 4.4.2; GT-P5210 Build/KOT49H) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/43.0.2357.93 Safari/537.36")
	}
}

// Safari for Mac
func BenchmarkParseSafariMac(b *testing.B) {
	for b.Loop() {
		Parse("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_4) AppleWebKit/600.7.12 (KHTML, like Gecko) Version/8.0.7 Safari/600.7.12")
	}
}

// Safari for iPad
func BenchmarkParseSafariiPad(b *testing.B) {
	for b.Loop() {
		Parse("Mozilla/5.0 (iPad; CPU OS 8_1_2 like Mac OS X) AppleWebKit/600.1.4 (KHTML, like Gecko) Version/8.0 Mobile/12B440 Safari/600.1.4")
	}
}

// The agents above are the common cases. These are the slow ones, measured
// across the fixture sets: the median agent parses in about 750ns, and each of
// these costs several times that. They are here so a change to the scans shows
// up where it hurts most, not only where traffic is heaviest.

// The worst real agent in the fixture sets, and the shape of the worst case in
// general: long, and matching nothing until the very last check. Every scan in
// the parser is linear in the length of the agent, and this one runs all of them.
func BenchmarkParseSmartTV(b *testing.B) {
	for b.Loop() {
		Parse("Mozilla/5.0 (Linux mipsel; U; HbbTV/1.1.1 (; TOSHIBA; DTV_TL938; 7.0.27.9; a5; ) ; " +
			"ToshibaTP/1.3.0 (+VIDEO_MP4+VIDEO_X_MS_ASF+AUDIO_MPEG+AUDIO_MP4+DRM+3D+NATIVELAUNCH" +
			"+WEBSTORAGE+OFFLINEAPP+HAS_CMD_HTTP_SERVER) ; en) AppleWebKit/534.1 (KHTML, like Gecko)")
	}
}

// An agent no check claims, which is the expensive way through: the browser
// switch falls to its default arm, and the bot pass then reads the whole agent
// rather than stopping at a match.
func BenchmarkParseUnknownAgent(b *testing.B) {
	for b.Loop() {
		Parse("Mozilla/5.0 (SymbianOS/9.4; U; Series60/5.0 SonyEricssonU1i/R1BB; " +
			"Profile/MIDP-2.1 Configuration/CLDC-1.1) AppleWebKit/525 (KHTML, like Gecko) " +
			"Version/3.0 Safari/525")
	}
}

// Anything outside ASCII sends normalise down its fallback, which lowercases with
// strings.ToLower instead of the stack buffer.
func BenchmarkParseNonASCII(b *testing.B) {
	for b.Loop() {
		Parse("Mozilla/5.0 (Windows NT 10.0; Ünicode ÅÄÖ) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	}
}

// Past the stack buffer, normalise has to allocate. Agents this long are junk or
// an attack, and the point of measuring one is that neither should be expensive.
func BenchmarkParseOversized(b *testing.B) {
	agent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/126.0.0.0 Safari/537.36" + strings.Repeat(" x", 500)
	for b.Loop() {
		Parse(agent)
	}
}
