package uasurfer

import (
	"regexp"
	"testing"
)

// amazonFireOracle is the regexp isAmazonFire replaced. It lives in the test so
// the hand rolled scan can be checked against it without dragging regexp (and
// the allocation it causes) back into the parser.
var amazonFireOracle = regexp.MustCompile(`\s(k[a-z]{3,5}|sd\d{4}ur)\s`)

func TestIsAmazonFire(t *testing.T) {
	cases := []string{
		"",
		" ",
		"k",
		" k ",
		" kfa ",       // k + 2 letters, too short
		" kfab ",      // k + 3 letters, shortest match
		" kfabc ",     // k + 4
		" kfabcd ",    // k + 5, longest match
		" kfabcde ",   // k + 6, too long
		" kfabcdef ",  // k + 7, too long
		"kfabc ",      // no leading whitespace
		" kfabc",      // no trailing whitespace
		"\tkfabc\n",   // other whitespace forms
		" kfab1 ",     // digit breaks the letter run
		" kf-ab ",     // hyphen breaks the letter run
		" sd4930ur ",  // sd form
		" sd4930ur",   // no trailing whitespace
		"sd4930ur ",   // no leading whitespace
		" sd493ur ",   // only 3 digits
		" sd49300ur ", // 5 digits
		" sd4930uz ",  // wrong suffix
		" sdur ",      // no digits
		"linux; android 9; kfmuwi build/ps7316; wv",
		"linux; android 5.1.1; sd4930ur build/lvy48f",
		"macintosh; intel mac os x 10_15_7",
		"windows nt 10.0; win64; x64",
		"linux; android 13; sm-s918b",
		"iphone; cpu iphone os 17_0 like mac os x",
		"linux; android 11; kfkawi",
		"linux; android 9; kftrwi build/ps7316",
	}

	for _, s := range cases {
		if got, want := isAmazonFire(s), amazonFireOracle.MatchString(s); got != want {
			t.Errorf("isAmazonFire(%q) = %v, regexp says %v", s, got, want)
		}
	}
}

// TestIsAmazonFireFuzzOracle walks a generated corpus to catch any divergence
// the hand-picked cases above miss.
func TestIsAmazonFireFuzzOracle(t *testing.T) {
	alphabet := []string{"", " ", "\t", "k", "s", "d", "u", "r", "a", "4", "0", "-"}
	var build func(prefix string, depth int)
	build = func(prefix string, depth int) {
		if got, want := isAmazonFire(prefix), amazonFireOracle.MatchString(prefix); got != want {
			t.Fatalf("isAmazonFire(%q) = %v, regexp says %v", prefix, got, want)
		}
		if depth == 0 {
			return
		}
		for _, a := range alphabet {
			build(prefix+a, depth-1)
		}
	}
	build("", 5)
}

func TestParseUnbalancedParens(t *testing.T) {
	tests := []struct {
		name string
		ua   string
		want OSName
	}{
		{"no closing paren", "Mozilla/5.0 (Windows NT 10.0; Win64; x64", OSWindows},
		// the strict specs match needs the platform group to start at the paren:
		// with the old off-by-one this saw "ozilla/5.0 (macintosh" and gave up
		{"truncated macintosh", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7", OSMacOSX},
		{"no parens at all", "Mozilla/5.0 Windows NT 10.0", OSWindows},
		{"closing before opening", "Mozilla/5.0) Windows NT 10.0 (x64", OSWindows},
		{"only opening paren", "(", OSUnknown},
		{"only closing paren", ")", OSUnknown},
		{"empty", "", OSUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Parse(tt.ua); got.OS.Name != tt.want {
				t.Errorf("Parse(%q).OS.Name = %v, want %v", tt.ua, got.OS.Name, tt.want)
			}
		})
	}
}

func TestParseiPadUnterminated(t *testing.T) {
	ua := Parse("(iPad; CPU OS 17_0 like Mac OS X")
	if ua.OS.Platform != PlatformiPad {
		t.Errorf("OS.Platform = %v, want %v", ua.OS.Platform, PlatformiPad)
	}
	if ua.OS.Version.Major != 17 {
		t.Errorf("OS.Version.Major = %d, want 17", ua.OS.Version.Major)
	}
}

func TestParseWindowsVersions(t *testing.T) {
	tests := []struct {
		ua       string
		platform Platform
		name     OSName
		version  Version
	}{
		{"mozilla/5.0 (windows nt 10.0; win64; x64)", PlatformWindows, OSWindows, Version{10, 0, 0}},
		{"mozilla/5.0 (windows nt 6.1; wow64)", PlatformWindows, OSWindows, Version{6, 1, 0}},
		{"mozilla/4.0 (compatible; msie 6.0; windows xp)", PlatformWindows, OSWindows, Version{5, 1, 0}},
		// xbox reads as windows and defaults to 6.0 when no nt version is given
		{"mozilla/5.0 (windows nt 10.0; win64; x64; xbox; xbox one)", PlatformXbox, OSXbox, Version{10, 0, 0}},
		{"mozilla/5.0 (xbox; windows )", PlatformXbox, OSXbox, Version{6, 0, 0}},
		// "windows" with no version at all
		{"mozilla/5.0 (compatible; microsoft-cryptoapi/10.0)", PlatformWindows, OSUnknown, Version{}},
	}
	for _, tt := range tests {
		var ua UserAgent
		ua.parseWindows(tt.ua)
		if ua.OS.Platform != tt.platform || ua.OS.Name != tt.name || ua.OS.Version != tt.version {
			t.Errorf("parseWindows(%q) = {%v %v %+v}, want {%v %v %+v}",
				tt.ua, ua.OS.Platform, ua.OS.Name, ua.OS.Version, tt.platform, tt.name, tt.version)
		}
	}
}
