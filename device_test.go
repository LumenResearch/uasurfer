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

// The Android phone-versus-tablet rule, which is the "Mobile" token plus two
// facts about the platform's history.
func TestAndroidDeviceType(t *testing.T) {
	tests := []struct {
		agent string
		want  DeviceType
	}{
		// the token, stated and omitted, on the reduced agent Chrome sends today
		{"Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) " +
			"Chrome/126.0.0.0 Mobile Safari/537.36", DevicePhone},
		{"Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) " +
			"Chrome/126.0.0.0 Safari/537.36", DeviceTablet},

		// and on agents that still name the model
		{"Mozilla/5.0 (Linux; Android 13; SM-S911B) AppleWebKit/537.36 (KHTML, like Gecko) " +
			"Chrome/126.0.0.0 Mobile Safari/537.36", DevicePhone},
		{"Mozilla/5.0 (Linux; Android 13; SM-X710) AppleWebKit/537.36 (KHTML, like Gecko) " +
			"Chrome/126.0.0.0 Safari/537.36", DeviceTablet},
		// a tablet that states "mobile" anyway, which is what the model list is for
		{"Mozilla/5.0 (Linux; Android 4.4.2; SM-T230 Build/KOT49H) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Chrome/33.0.0.0 Mobile Safari/537.36", DeviceTablet},

		// Honeycomb shipped on tablets and nothing else, token or no token
		{"Mozilla/5.0 (Linux; U; Android 3.2.1; de-de; HTC Flyer Build/HTK75C) " +
			"AppleWebKit/537.31 (KHTML, like Gecko) Version/4.0 Safari/537.31", DeviceTablet},
		// before it the token was not yet a convention, and the era was phones
		{"Mozilla/5.0 (Linux; U; Android 2.3.3; en-us; Sensation_4G Build/GRI40) " +
			"AppleWebKit/533.1 (KHTML, like Gecko) Version/5.0 Safari/533.16", DevicePhone},

		// an app HTTP stack states no form factor at all, so it stays a phone
		{"Dalvik/2.1.0 (Linux; U; Android 13; SM-S911B Build/TP1A.220624.014)", DevicePhone},
		{"AndroidDownloadManager/4.1.2 (Linux; U; Android 4.1.2; MediaPad 7 Lite II " +
			"Build/HuaweiMediaPad)", DevicePhone},
		// nor does a vendor agent that merely mentions an engine
		{"Android 4.0.4;AppleWebKit/534.30;Build/HuaweiU8836D;U8836D Build/HuaweiU8836D",
			DevicePhone},

		// a watch names itself in the model field and says "mobile" like a phone
		{"Mozilla/5.0 (Linux; Android 13; Pixel Watch) AppleWebKit/537.36 (KHTML, like Gecko) " +
			"Chrome/126.0.0.0 Mobile Safari/537.36", DeviceWearable},
		{"Mozilla/5.0 (Linux; Android 11; SM-R910 Build/RP1A.201005.001) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Version/4.0 Chrome/83.0.4103.106 Mobile Safari/537.36", DeviceWearable},
		// but a Galaxy Tab is "sm-t" and a phone "sm-g", neither of which is "sm-r"
		{"Mozilla/5.0 (Linux; Android 13; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) " +
			"Chrome/126.0.0.0 Mobile Safari/537.36", DevicePhone},

		// a TV is still a TV: that check runs before any of this
		{"Mozilla/5.0 (Linux; Android 9; AFTKM) AppleWebKit/537.36 (KHTML, like Gecko) " +
			"Silk/122.4.2 like Chrome/122.0.6261.119 Safari/537.36", DeviceTV},
	}

	for _, tt := range tests {
		if got := Parse(tt.agent).DeviceType; got != tt.want {
			t.Errorf("device = %v, want %v: %.90s", got, tt.want, tt.agent)
		}
	}
}

// isAndroidTablet decides for the tablets that state "mobile" like a phone, so
// the case worth pinning is the marker that needs its dash to stay honest.
func TestIsAndroidTablet(t *testing.T) {
	for _, ua := range []string{
		"mozilla/5.0 (linux; android 13; sm-t970) applewebkit/537.36",
		"mozilla/5.0 (linux; android 4.4.2; nexus 7 build/x) applewebkit/537.36 mobile",
		"mozilla/5.0 (linux; android 4.4; lenovo tab 2 a7) applewebkit/537.36",
		"mozilla/5.0 (linux; android 4.4.3; t1-a21l build/x) applewebkit/537.36",
	} {
		if !isAndroidTablet(ua) {
			t.Errorf("isAndroidTablet(%.60q) = false, want true", ua)
		}
	}
	for _, ua := range []string{
		"",
		"mozilla/5.0 (linux; android 13; sm-g991b) applewebkit/537.36 mobile",
		// "t100" is a handset: "; t1" without the dash would claim it
		"mozilla/5.0 (linux; u; android 2.2.1; en-us; t100 build/frg83) applewebkit/533.1",
	} {
		if isAndroidTablet(ua) {
			t.Errorf("isAndroidTablet(%.60q) = true, want false", ua)
		}
	}
}
