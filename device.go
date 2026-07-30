package uasurfer

import (
	"strings"
)

// ponytail: substring set — "tv" already subsumes googletv/appletv/smarttv/smart-tv/hbbtv/dtv/"tv box",
// "aftss" subsumes "aftsss", "aftka" subsumes "aftkauk".
var tvMarkers = []string{
	"tv", "roku", "crkey", "chromecast", "stb", "tuner", "vizio", "viera", "aquos", "bravia",
	"netcast", "youview", "adt-", "swisscom-ip", "mibox", "ott-g1", "ottera",
	"aftb", "aftt", "aftm", "aftr", "aftss", "aftka", "aftkrt", "aftgazl", "aftanna",
	"tpm191e", "tpm171e", "nokia streaming box", "stableavb_telly", "lxbox51",
	"x96max", "x96q_max_pro", "canal plus box", "vectra 4k box",
	"diw377", "diw380", "dv8555", "dctiw362", "gd1 4k", "ai pont", "b-stream",
}

func isTV(ua string) bool {
	for _, m := range tvMarkers {
		if strings.Contains(ua, m) {
			return true
		}
	}

	return strings.Contains(ua, "mbox") && !strings.Contains(ua, "xbox")
}

func (u *UserAgent) evalDevice(ua string) {
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
