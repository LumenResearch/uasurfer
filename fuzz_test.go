package uasurfer

import (
	"strings"
	"testing"
)

// The parser walks untrusted input by index throughout - platformGroup slices on
// the parentheses it finds, Version.parse and the marker scans step through
// bytes - so the property worth fuzzing is that no agent, however malformed,
// gets it to step outside a string or to disagree with itself.
//
// Run longer than the seed corpus with:
//
//	go test -run=XXX -fuzz=FuzzParse -fuzztime=2m
func FuzzParse(f *testing.F) {
	// Seeds: the shapes that have historically broken index arithmetic here,
	// rather than a sample of well formed agents.
	seeds := []string{
		"",
		"(",
		")",
		")(",
		"()",
		"mozilla/5.0 (",
		"mozilla/5.0 )",
		"mozilla/5.0 ) macintosh (",
		"(;;;;;;)",
		";",
		"/",
		"+",
		"aft",
		"bot",
		"cpu os ",
		"android ",
		"windows nt ",
		"version/",
		"appletv/",
		"appletv6,2/",
		"roku/dvp-",
		"tizen ",
		"trident/",
		"cfnetwork/ darwin/",
		"mozilla/5.0 (iphone; cpu iphone os 17_5 like mac os x) applewebkit/605.1.15",
		"mozilla/5.0 (windows nt 10.0; win64; x64) applewebkit/537.36 chrome/126.0.0.0 safari/537.36",
		"mozilla/5.0 (linux; android 10; k) applewebkit/537.36 chrome/126.0.0.0 mobile safari/537.36",
		"mozilla/5.0 (compatible; googlebot/2.1; +http://www.google.com/bot.html)",
		"roku/dvp-12.5 (12.5.5.4405-46)",
		"mozilla/5.0 (smart-tv; linux; tizen 6.0) applewebkit/537.36 version/6.0 tv safari/537.36",
		"mozilla/5.0 (web0s; linux/smarttv) applewebkit/537.36",
		"mozilla/5.0 (linux; android 11; aftkm) applewebkit/537.36 silk/122.4.2 like chrome/122.0",
		// non ASCII, which sends normalise down its fallback path
		"mozilla/5.0 (Ünicode ÅÄÖ)",
		// and an agent past the stack buffer, the other fallback
		strings.Repeat("a", 1100),
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, agent string) {
		ua := Parse(agent)

		// Every field has to be a constant the package defines. A value outside
		// the list means the parser invented one, and String() would then report
		// it as Unknown while a caller's switch fell through.
		if ua.Browser.Name < 0 || ua.Browser.Name >= _browserNameFinal {
			t.Fatalf("browser name %d out of range", ua.Browser.Name)
		}
		if ua.OS.Name < 0 || ua.OS.Name >= _osNameFinal {
			t.Fatalf("os name %d out of range", ua.OS.Name)
		}
		if ua.OS.Platform < 0 || ua.OS.Platform >= _platformFinal {
			t.Fatalf("platform %d out of range", ua.OS.Platform)
		}
		if ua.DeviceType < 0 || ua.DeviceType >= _deviceTypeFinal {
			t.Fatalf("device type %d out of range", ua.DeviceType)
		}

		// Versions are parsed by hand out of arbitrary text; negatives would
		// break Less and any caller comparing against a release number.
		for _, v := range []Version{ua.Browser.Version, ua.OS.Version} {
			if v.Major < 0 || v.Minor < 0 || v.Patch < 0 {
				t.Fatalf("negative version %v", v)
			}
		}

		// A bot is reported consistently across all four fields, because callers
		// filter on whichever one they happen to hold.
		if ua.IsBot() && (ua.OS.Name != OSBot || ua.OS.Platform != PlatformBot || ua.DeviceType != DeviceComputer) {
			t.Fatalf("bot reported inconsistently: %+v", ua)
		}

		// Parsing is a pure function of the agent, and normalisation is the only
		// thing that may differ between two spellings of one agent: an upper
		// case agent has to land on the same answer as its lowercase twin.
		if lower := Parse(strings.ToLower(agent)); *lower != *ua {
			t.Fatalf("case changed the result:\n lower %+v\n orig  %+v", lower, ua)
		}

		// And the reuse path has to agree with the allocating one, or a caller
		// pooling UserAgents gets different answers from the same input.
		var reused UserAgent
		ParseUserAgent(agent, &reused)
		if reused != *ua {
			t.Fatalf("ParseUserAgent disagrees with Parse:\n reused %+v\n parse  %+v", reused, ua)
		}
	})
}
