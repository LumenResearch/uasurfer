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
