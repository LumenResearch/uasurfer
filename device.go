package uasurfer

import (
	"strings"
)

func isTVKeyword(ua string) bool {
	return strings.Contains(ua, "tv") || strings.Contains(ua, "crkey") || strings.Contains(ua, "googletv") || strings.Contains(ua, "aftb") || strings.Contains(ua, "aftt") || strings.Contains(ua, "aftm") ||
		strings.Contains(ua, "adt-") || strings.Contains(ua, "roku") || strings.Contains(ua, "viera") || strings.Contains(ua, "aquos") || strings.Contains(ua, "dtv") ||
		strings.Contains(ua, "appletv") || strings.Contains(ua, "smarttv") || strings.Contains(ua, "tuner") || strings.Contains(ua, "smart-tv") || strings.Contains(ua, "hbbtv") ||
		strings.Contains(ua, "netcast") || strings.Contains(ua, "vizio") || strings.Contains(ua, "stb") || strings.Contains(ua, "swisscom-ip") || strings.Contains(ua, "youview") ||
		strings.Contains(ua, "aftkrt") || strings.Contains(ua, "aftsss") || strings.Contains(ua, "aftss") || strings.Contains(ua, "aftka") || strings.Contains(ua, "aftr") ||
		strings.Contains(ua, "aftgazl") || strings.Contains(ua, "aftanna") || strings.Contains(ua, "aftkauk") ||
		strings.Contains(ua, "bravia") || strings.Contains(ua, "mibox") || strings.Contains(ua, "chromecast") || strings.Contains(ua, "ott-g1") || strings.Contains(ua, "ottera") ||
		strings.Contains(ua, "tpm191e") || strings.Contains(ua, "nokia streaming box") || (strings.Contains(ua, "stableavb_telly")) || strings.Contains(ua, "lxbox51") ||
		strings.Contains(ua, "x96max") || strings.Contains(ua, "x96q_max_pro") || strings.Contains(ua, "canal plus box") || strings.Contains(ua, "vectra 4k box") ||
		strings.Contains(ua, "diw377") || strings.Contains(ua, "diw380") || strings.Contains(ua, "dv8555") || strings.Contains(ua, "dctiw362") || strings.Contains(ua, "gd1 4k") ||
		strings.Contains(ua, "tpm171e") || strings.Contains(ua, "ai pont") || strings.Contains(ua, "b-stream") || strings.Contains(ua, "tv box") ||
		(strings.Contains(ua, "mbox") && !strings.Contains(ua, "xbox"))
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
	case isTVKeyword(ua):
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
