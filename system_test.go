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

// A version component that cannot fit is nonsense, but it must not come back
// negative: Less would then sort it before every real release. FuzzParse found
// this one, and testdata/fuzz keeps it as a seed.
func TestVersionParseOverflow(t *testing.T) {
	for _, agent := range []string{
		"roku/10000000000000000000",
		"chrome/99999999999999999999.99999999999999999999",
		"mozilla/5.0 (windows nt 184467440737095516151) applewebkit/537.36",
	} {
		ua := Parse(agent)
		for _, v := range []Version{ua.Browser.Version, ua.OS.Version} {
			if v.Major < 0 || v.Minor < 0 || v.Patch < 0 {
				t.Errorf("Parse(%q) = %v, want no negative component", agent, v)
			}
			if v.Major > maxVersionPart*10 {
				t.Errorf("Parse(%q) = %v, want each component capped", agent, v)
			}
		}
	}

	// and the cap must not disturb versions anyone actually ships
	var v Version
	v.parse("126.0.6478.126")
	if v != (Version{126, 0, 6478}) {
		t.Errorf("parse(%q) = %v, want {126 0 6478}", "126.0.6478.126", v)
	}
}

func TestVersionParse(t *testing.T) {
	tests := []struct {
		in   string
		want Version
		ok   bool
	}{
		// the separator is whatever byte ends the digit run
		{"126.0.6478.126", Version{126, 0, 6478}, true},
		{"10_15_7", Version{10, 15, 7}, true},
		{"6.1", Version{6, 1, 0}, true},
		{"11", Version{11, 0, 0}, true},
		// the run ends where the next field begins
		{"4.0 mobile safari/537.36", Version{4, 0, 0}, true},
		{"12.5 (12.5.5.4405-46)", Version{12, 5, 0}, true},
		// a component that starts on a non-digit is zero, and does not reach
		// forward for the next number in the string
		{"1..2", Version{1, 0, 2}, true},
		{"9", Version{9, 0, 0}, true},
		// no leading digit at all is a miss, which is how parseAfter tells a
		// marker worth following from one that led nowhere
		{"", Version{}, false},
		{"abc", Version{}, false},
		{".1", Version{}, false},
		{" 1", Version{}, false},
	}
	for _, tt := range tests {
		var v Version
		if ok := v.parse(tt.in); ok != tt.ok || v != tt.want {
			t.Errorf("parse(%q) = %v, %v; want %v, %v", tt.in, v, ok, tt.want, tt.ok)
		}
	}
}

// A zero-padded component is read as the zero, with what follows moving into the
// next place: "4.08" is {4, 0, 8}. This looks odd until you line the padded
// releases up against the bare ones - the browsers that wrote a version this way
// shipped both, and reading "08" as the single number 8 would collapse them:
//
//	Netscape 4.08 {4 0 8} against 4.8 {4 8 0}
//	Opera    9.01 {9 0 1} against 9.1 {9 1 0} and 9.10 {9 10 0}
//	IE       5.01 {5 0 1} against 5.5 {5 5 0}
//
// The corpus carries all three families, so this is not hypothetical; see the
// MSIE 5.01 row in uasurfer_test.go.
func TestVersionParseZeroPadded(t *testing.T) {
	tests := []struct {
		in   string
		want Version
	}{
		{"4.08", Version{4, 0, 8}},
		{"4.8", Version{4, 8, 0}},
		{"9.01", Version{9, 0, 1}},
		{"9.1", Version{9, 1, 0}},
		{"9.10", Version{9, 10, 0}},
		{"5.01", Version{5, 0, 1}},
		{"5.5", Version{5, 5, 0}},
		// padding in the major reads the same way
		{"04.1", Version{0, 4, 1}},
		{"0001", Version{0, 1, 0}},
		// a component of nothing but zeroes is just zero, which is most of what
		// actually ships: these mean what they look like
		{"5.00", Version{5, 0, 0}},
		{"126.0.0.0", Version{126, 0, 0}},
		{"0", Version{}},
		{"0.0.0", Version{}},
	}
	for _, tt := range tests {
		var v Version
		if !v.parse(tt.in) || v != tt.want {
			t.Errorf("parse(%q) = %v, want %v", tt.in, v, tt.want)
		}
	}
}

// parse overwrites all three components, so a Version reused across calls - as
// parseAfter does when an earlier marker matched and a later one is tried - does
// not keep a digit from the previous one.
func TestVersionParseOverwrites(t *testing.T) {
	v := Version{9, 9, 9}
	if !v.parse("1") || v != (Version{1, 0, 0}) {
		t.Errorf("parse over a used Version = %v, want {1 0 0}", v)
	}
}

// The version of a connected TV comes from a different field on each platform,
// which is worth pinning through the public API.
func TestTVOSVersions(t *testing.T) {
	tests := []struct {
		agent string
		want  Version
	}{
		{"AppleTV6,2/11.1", Version{11, 1, 0}},
		{"AppleTV/1.1", Version{1, 1, 0}},
		{"AppleCoreMedia/1.0.0.12F69 (Apple TV; U; CPU OS 8_3 like Mac OS X; en_us)", Version{8, 3, 0}},
		{"Roku/DVP-12.5 (12.5.5.4405-46)", Version{12, 5, 0}},
		{"Mozilla/5.0 (SMART-TV; LINUX; Tizen 6.0) AppleWebKit/537.36 (KHTML, like Gecko) TV Safari/537.36", Version{6, 0, 0}},
	}
	for _, tt := range tests {
		if got := Parse(tt.agent).OS.Version; got != tt.want {
			t.Errorf("Parse(%.50q).OS.Version = %v, want %v", tt.agent, got, tt.want)
		}
	}
}
