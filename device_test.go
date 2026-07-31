package uasurfer

import "testing"

func TestIsTV(t *testing.T) {
	for _, ua := range []string{"smarttv", "hbbtv/1.1.1", "aftsss build", "roku4640x", "mbox android"} {
		if !isTV(ua) {
			t.Errorf("want TV: %q", ua)
		}
	}
	for _, ua := range []string{"iphone os 17", "xbox one; mbox", "windows nt 10.0"} {
		if isTV(ua) {
			t.Errorf("want not TV: %q", ua)
		}
	}
}

// TestIsTVMarkers guards the bucketed single-pass scan in isTV: every marker
// must still match both on its own and embedded in a full agent, whatever
// bucket it lands in.
func TestIsTVMarkers(t *testing.T) {
	for _, m := range tvMarkers {
		if !isTV(m) {
			t.Errorf("isTV(%q) = false, want true", m)
		}
		if !isTV("mozilla/5.0 (" + m + " build/x) applewebkit/537.36") {
			t.Errorf("isTV(...%q...) = false, want true", m)
		}
	}

	// Markers dropped from tvMarkers as redundant must still match via the
	// shorter marker that subsumes them.
	for _, m := range []string{"googletv", "dtv", "appletv", "smarttv", "smart-tv", "hbbtv", "tv box", "aftsss", "aftkauk"} {
		if !isTV(m) {
			t.Errorf("isTV(%q) = false, want true", m)
		}
	}

	for _, ua := range []string{
		"",
		"xbox",
		"mozilla/5.0 (windows nt 10.0; win64; x64) applewebkit/537.36 chrome/120.0",
		"mozilla/5.0 (iphone; cpu iphone os 17_0 like mac os x) applewebkit/605.1.15",
		"mozilla/5.0 (linux; android 13; sm-s918b) applewebkit/537.36 chrome/120.0 mobile",
	} {
		if isTV(ua) {
			t.Errorf("isTV(%q) = true, want false", ua)
		}
	}
}

// TestTVFixtures runs the connected TV fixture set. Device type and OS are
// asserted together on purpose: a stick that reports the right OS and the wrong
// device type is no more useful to a caller than one that reports neither.
func TestTVFixtures(t *testing.T) {
	for _, row := range readFixtures(t, "tv.tsv", 3) {
		wantDevice, wantOS, agent := row[0], row[1], row[2]
		ua := Parse(agent)
		if got := ua.DeviceType.StringTrimPrefix(); got != wantDevice {
			t.Errorf("device = %s, want %s: %s", got, wantDevice, agent)
		}
		if got := ua.OS.Name.StringTrimPrefix(); got != wantOS {
			t.Errorf("os = %s, want %s: %s", got, wantOS, agent)
		}
	}
}

func TestIsFireTV(t *testing.T) {
	// Real model tokens, as they appear in the platform group. Amazon has shipped
	// dozens; the point of the rule is that it does not need to know them.
	for _, specs := range []string{
		"linux; android 9; aftb", "linux; android 11; aftkm", "linux; android 9; aftmm",
		"linux; android 7.1.2; aftsss", "linux; android 9; aftka build/ps7233.3244n",
		"linux; android 11; aftkmst12", "linux; android 9; aftbamr311", "linux; android 5.1; aftt",
		"linux; android 9; aftdck", "aftb", "linux, aftgazl",
	} {
		if !isFireTV(specs) {
			t.Errorf("isFireTV(%q) = false, want true", specs)
		}
	}

	// "aft" also starts ordinary words, which is why the match is anchored to a
	// field of the platform group and never run over the whole agent.
	for _, specs := range []string{
		"", "aft", "linux; android 9; craft",
		"macintosh; intel mac os x 10_15_7", "linux; android 13; sm-s918b",
		"compatible; wordpress/5.3.3; https://belasting-aftrekken.nl",
		"linux; android 9; kfmuwi", // a Fire tablet, not a Fire TV
	} {
		if isFireTV(specs) {
			t.Errorf("isFireTV(%q) = true, want false", specs)
		}
	}

	// A bare "after" as its own field does match, and is left matching: a
	// platform group carries OS, model and build fields, never prose. The trap
	// worth guarding is the whole-agent one below.
	if !isFireTV("windows nt 10.0; after") {
		t.Error("isFireTV: expected the documented false positive on a bare \"after\" field")
	}

	// The whole-agent form of the same trap: a URL in the tail must not promote
	// a desktop browser to a television.
	notTV := "mozilla/5.0 (windows nt 10.0; win64; x64) applewebkit/537.36 " +
		"(khtml, like gecko) chrome/126.0.0.0 safari/537.36 http://afterice.se"
	if isTV(notTV) {
		t.Error("isTV: a URL containing \"after\" reads as a television")
	}
}

func TestPlatformGroup(t *testing.T) {
	tests := []struct{ ua, want string }{
		{"mozilla/5.0 (macintosh; intel mac os x 10_15_7) applewebkit/605", "macintosh; intel mac os x 10_15_7"},
		{"roku/dvp-12.5 (12.5.5.4405-46)", "12.5.5.4405-46"},
		// no parens, so the group is the whole agent
		{"curl/8.6.0", "curl/8.6.0"},
		// unterminated, and reversed: both run to the end rather than panicking
		{"mozilla/5.0 (macintosh", "macintosh"},
		// the parens are in the wrong order, so the group starts after the "("
		// and runs to the end, which here is nothing at all
		{"mozilla/5.0 ) macintosh (", ""},
		{"", ""},
	}
	for _, tt := range tests {
		if got := platformGroup(tt.ua); got != tt.want {
			t.Errorf("platformGroup(%q) = %q, want %q", tt.ua, got, tt.want)
		}
	}
}
