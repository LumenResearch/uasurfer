package uasurfer

import "testing"

// TestIPadScreenSizesAreLandscape guards the invariant isiPad now relies on:
// the table is compared directly, so every entry must already be landscape.
func TestIPadScreenSizesAreLandscape(t *testing.T) {
	for _, size := range iPadScreenSizes {
		if size.Width < size.Height {
			t.Errorf("iPadScreenSizes entry %+v is portrait, want longest edge as Width", size)
		}
	}
}

func TestScreenSizeIsiPad(t *testing.T) {
	for _, size := range iPadScreenSizes {
		if !(ScreenSize{size.Width, size.Height}).isiPad() {
			t.Errorf("%+v landscape: want iPad", size)
		}
		// the same panel reported portrait must match too
		if !(ScreenSize{size.Height, size.Width}).isiPad() {
			t.Errorf("%+v portrait: want iPad", size)
		}
	}

	for _, size := range []ScreenSize{{0, 0}, {1920, 1080}, {390, 844}, {1024, 769}} {
		if size.isiPad() {
			t.Errorf("%+v: want not iPad", size)
		}
	}
}

// TestParseWithNilHints covers the nil handling that moved out to the call sites
// when isiPad became a value method.
func TestParseWithNilHints(t *testing.T) {
	const mac = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15"
	for _, hints := range []*Hints{nil, {}, {ScreenSize: nil}} {
		if got := ParseWithHints(mac, hints); got.OS.Name != OSMacOSX {
			t.Errorf("ParseWithHints(mac, %+v).OS.Name = %v, want %v", hints, got.OS.Name, OSMacOSX)
		}
	}

	// a supplied iPad panel still promotes the agent
	if got := ParseWithHints(mac, &Hints{ScreenSize: &ScreenSize{1024, 768}}); got.OS.Name != OSiPadOS {
		t.Errorf("ParseWithHints(mac, iPad size).OS.Name = %v, want %v", got.OS.Name, OSiPadOS)
	}
}

// Client hints refine what the agent could not settle: the form factor outright,
// and the OS version where the agent reports a frozen fiction.
func TestParseWithClientHints(t *testing.T) {
	const (
		androidTablet = "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
		androidPhone = "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Chrome/126.0.0.0 Mobile Safari/537.36"
		mac = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
		fireTV = "Mozilla/5.0 (Linux; Android 9; AFTKM) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Silk/122.4.2 like Chrome/122.0 Safari/537.36"
	)

	tests := []struct {
		name       string
		agent      string
		hints      Hints
		wantDevice DeviceType
		wantOS     Version
	}{
		{"no hints leaves the agent alone", androidPhone, Hints{}, DevicePhone, Version{10, 0, 0}},

		{"form factor settles a phone that reads as a tablet", androidTablet,
			Hints{FormFactors: `"Mobile"`}, DevicePhone, Version{10, 0, 0}},
		{"and a tablet that reads as a phone", androidPhone,
			Hints{FormFactors: `"Tablet"`}, DeviceTablet, Version{10, 0, 0}},
		{"a watch says so outright", androidPhone,
			Hints{FormFactors: `"Watch"`}, DeviceWearable, Version{10, 0, 0}},
		{"the first value we know of a list wins", androidPhone,
			Hints{FormFactors: `"EInk", "Tablet"`}, DeviceTablet, Version{10, 0, 0}},
		{"a form factor with no constant changes nothing", androidPhone,
			Hints{FormFactors: `"Automotive"`}, DevicePhone, Version{10, 0, 0}},
		{"and none of them overrule a television", fireTV,
			Hints{FormFactors: `"Desktop"`, Mobile: "?0"}, DeviceTV, Version{9, 0, 0}},

		{"?0 makes a tablet of an Android phone", androidPhone,
			Hints{Mobile: "?0"}, DeviceTablet, Version{10, 0, 0}},
		{"?1 makes a phone of a desktop", mac,
			Hints{Mobile: "?1"}, DevicePhone, Version{10, 15, 7}},

		// Chromium freezes both of these in the agent, so the hint is the only
		// place the real version appears.
		{"the platform version unfreezes macOS", mac,
			Hints{Platform: `"macOS"`, PlatformVersion: `"14.5.0"`}, DeviceComputer, Version{14, 5, 0}},
		{"and Android", androidPhone,
			Hints{Platform: `"Android"`, PlatformVersion: `"14.0.0"`}, DevicePhone, Version{14, 0, 0}},
		{"but not when the platform disagrees with the agent", mac,
			Hints{Platform: `"Android"`, PlatformVersion: `"14.0.0"`}, DeviceComputer, Version{10, 15, 7}},
		{"Windows is left to the NT version the agent states", mac,
			Hints{Platform: `"Windows"`, PlatformVersion: `"13.0.0"`}, DeviceComputer, Version{10, 15, 7}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ua := ParseWithHints(tt.agent, &tt.hints)
			if ua.DeviceType != tt.wantDevice {
				t.Errorf("device = %v, want %v", ua.DeviceType, tt.wantDevice)
			}
			if ua.OS.Version != tt.wantOS {
				t.Errorf("os version = %v, want %v", ua.OS.Version, tt.wantOS)
			}
		})
	}
}
