package uasurfer

import (
	"strings"
)

// Markers that are a substring of another marker are omitted: "tv" already
// subsumes googletv/appletv/smarttv/smart-tv/hbbtv/dtv/"tv box". Amazon's Fire
// TV models are not listed at all; isFireTV matches the whole family.
var tvMarkers = []string{
	"tv", "roku", "crkey", "chromecast", "stb", "tuner", "vizio", "viera", "aquos", "bravia",
	"netcast", "youview", "adt-", "swisscom-ip", "mibox", "ott-g1", "ottera",
	"tpm191e", "tpm171e", "nokia streaming box", "stableavb_telly", "lxbox51",
	"x96max", "x96q_max_pro", "canal plus box", "vectra 4k box",
	"diw377", "diw380", "dv8555", "dctiw362", "gd1 4k", "ai pont", "b-stream",
}

// tvMarkersByFirstByte buckets tvMarkers so isTV can scan the agent once
// instead of running a full strings.Contains pass per marker. This is the
// hottest check in the parser: every non-desktop, non-iOS agent walks the
// whole set before being classified.
var tvMarkersByFirstByte = func() (buckets [256][]string) {
	for _, m := range tvMarkers {
		buckets[m[0]] = append(buckets[m[0]], m)
	}
	return
}()

func isTV(ua string) bool {
	for i := 0; i < len(ua); i++ {
		for _, m := range tvMarkersByFirstByte[ua[i]] {
			if strings.HasPrefix(ua[i:], m) {
				return true
			}
		}
	}

	if strings.Contains(ua, "mbox") && !strings.Contains(ua, "xbox") {
		return true
	}

	// Last, because it is the only check that has to look at where in the agent
	// the match sits: every other marker here is distinctive enough to match
	// anywhere.
	return isFireTV(platformGroup(ua))
}

// isFireTV reports whether specs carries an Amazon Fire TV model token: "aft"
// and up to ten more letters or digits, as its own field.
//
// Amazon has shipped dozens of these - aftb, aftmm, aftka, aftkm, aftdck,
// aftkmst12, aftbamr311 - and enumerating them meant every new stick read as a
// phone until someone noticed. The family prefix is stable, so match on that.
//
// Only the platform group is searched, never the whole agent: unanchored, "aft"
// also begins "after", which turns any agent quoting a URL like
// "afterice.se" into a television.
func isFireTV(specs string) bool {
	for i := 0; i+3 <= len(specs); i++ {
		if i > 0 && !isFieldEdge(specs[i-1]) {
			continue
		}
		if specs[i] != 'a' || specs[i+1] != 'f' || specs[i+2] != 't' {
			continue
		}
		n := i + 3
		for n < len(specs) && n-i <= 12 && (isLower(specs[n]) || isDigit(specs[n])) {
			n++
		}
		if n > i+3 && (n == len(specs) || isFieldEdge(specs[n])) {
			return true
		}
	}
	return false
}

// isFieldEdge reports whether c separates two fields of a platform group.
func isFieldEdge(c byte) bool { return isSpace(c) || c == ';' || c == ',' }

func (u *UserAgent) parseDevice(ua string) {
	switch {

	case u.OS.Platform == PlatformWindows || u.OS.Platform == PlatformMac || u.OS.Name == OSChromeOS:
		if strings.Contains(ua, "mobile") || strings.Contains(ua, "touch") {
			u.DeviceType = DeviceTablet // windows rt, linux haxor tablets
			return
		}
		u.DeviceType = DeviceComputer

	// long list of smarttv and tv dongle identifiers - above "phone" and "tablet" check to prevent TVs from being detected as phones/tablets
	case u.OS.Platform != PlatformiPhone && u.OS.Platform != PlatformiPad && isTV(ua):
		u.DeviceType = DeviceTV

	case u.OS.Platform == PlatformiPad || u.OS.Platform == PlatformiPod || strings.Contains(ua, "tablet") || strings.Contains(ua, "kindle/") || strings.Contains(ua, "playbook"):
		u.DeviceType = DeviceTablet

	case u.OS.Platform == PlatformiPhone || u.OS.Platform == PlatformBlackberry || strings.Contains(ua, "phone"):
		u.DeviceType = DevicePhone

	case u.OS.Name == OSAndroid:
		// android phones report as "mobile", android tablets should not but often do -- http://android-developers.blogspot.com/2010/12/android-browser-user-agent-issues.html
		if strings.Contains(ua, "mobile") {
			u.DeviceType = DevicePhone
			return
		}

		if strings.Contains(ua, "tablet") || strings.Contains(ua, "nexus 7") || strings.Contains(ua, "nexus 9") || strings.Contains(ua, "nexus 10") || strings.Contains(ua, "xoom") ||
			strings.Contains(ua, "sm-t") || strings.Contains(ua, "; kf") || strings.Contains(ua, "; t1") || strings.Contains(ua, "lenovo tab") {
			u.DeviceType = DeviceTablet
			return
		}

		u.DeviceType = DevicePhone // default to phone

	case u.OS.Platform == PlatformPlaystation || u.OS.Platform == PlatformXbox || u.OS.Platform == PlatformNintendo:
		u.DeviceType = DeviceConsole

	case strings.Contains(ua, "glass") || strings.Contains(ua, "watch") || strings.Contains(ua, "sm-v"):
		u.DeviceType = DeviceWearable

	// specifically above "mobile" string check as Kindle Fire tablets report as "mobile"
	case u.Browser.Name == BrowserSilk || u.OS.Name == OSKindle && !strings.Contains(ua, "sd4930ur"):
		u.DeviceType = DeviceTablet

	case strings.Contains(ua, "mobile") || strings.Contains(ua, "touch") || strings.Contains(ua, " mobi") || strings.Contains(ua, "webos"): //anything "mobile"/"touch" that didn't get captured as tablet, console or wearable is presumed a phone
		u.DeviceType = DevicePhone

	case u.OS.Name == OSLinux: // linux goes last since it's in so many other device types (tvs, wearables, android-based stuff)
		u.DeviceType = DeviceComputer

	default:
		u.DeviceType = DeviceUnknown
	}
}
