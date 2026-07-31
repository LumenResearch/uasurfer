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
