package uasurfer

import (
	"strings"
)

// isAmazonFire reports whether s carries a whitespace delimited Amazon Fire
// tablet or phone model token. It is equivalent to the regexp it replaced,
// `\s(k[a-z]{3,5}|sd\d{4}ur)\s`, which TestIsAmazonFire pins down by
// differential testing against that pattern.
//
// Hand rolled rather than compiled because handing the agent to regexp makes it
// escape, which in turn forces parse's normalise buffer onto the heap - the
// regexp was single-handedly costing every parse an allocation.
func isAmazonFire(s string) bool {
	for i := 0; i+1 < len(s); i++ {
		if !isSpace(s[i]) {
			continue
		}
		n := amazonFireToken(s[i+1:])
		if n > 0 && i+1+n < len(s) && isSpace(s[i+1+n]) {
			return true
		}
	}
	return false
}

// amazonFireToken returns the length of the model token at the start of s, or 0
// if there is none. Matching is greedy: for the "k" form the letter run is
// consumed up to the 5 the pattern allows, and a longer run simply fails the
// caller's trailing whitespace check, exactly as the regexp's backtracking did
// (every shorter prefix ends on a letter, never on whitespace).
func amazonFireToken(s string) int {
	if len(s) > 0 && s[0] == 'k' {
		n := 1
		for n < len(s) && n <= 5 && isLower(s[n]) {
			n++
		}
		if n >= 4 { // "k" plus at least three letters
			return n
		}
		return 0
	}

	// sd<4 digits>ur
	if len(s) >= 8 && s[0] == 's' && s[1] == 'd' && s[6] == 'u' && s[7] == 'r' &&
		isDigit(s[2]) && isDigit(s[3]) && isDigit(s[4]) && isDigit(s[5]) {
		return 8
	}
	return 0
}

// isSpace matches regexp's \s class: [\t\n\f\r ].
func isSpace(c byte) bool { return c == ' ' || c == '\t' || c == '\n' || c == '\f' || c == '\r' }
func isLower(c byte) bool { return c >= 'a' && c <= 'z' }
func isDigit(c byte) bool { return c >= '0' && c <= '9' }

// platformGroup returns the text between the first parentheses. When the closing
// paren is missing or precedes the opening one, the group runs to the end of the
// agent instead. Only e moves: s still points at the opening paren (or -1 when
// there is none), so the s+1 below skips that paren rather than the agent's
// first byte.
func platformGroup(ua string) string {
	s := strings.IndexByte(ua, '(')
	e := strings.IndexByte(ua, ')')
	if e == -1 || s > e {
		e = len(ua)
	}
	return ua[s+1 : e]
}

func (u *UserAgent) parseOS(ua string, hints *Hints) bool {
	agentPlatform := platformGroup(ua)
	// Cut returns the whole string when there is no ";", which is what we want.
	specs, _, _ := strings.Cut(agentPlatform, ";")

	// strict OS & version identification first, by exact specs match; the
	// broader whole-agent scans below only run when none of them hit.
	switch {
	case specs == "android" || specs == "x11" || specs == "linux":
		u.parseLinux(ua, agentPlatform)

	case specs == "bb10" || specs == "playbook":
		u.OS.Platform = PlatformBlackberry
		u.OS.Name = OSBlackberry

	case strings.HasPrefix(specs, "ipad") || strings.HasPrefix(specs, "iphone") || strings.HasPrefix(specs, "ipod"):
		u.parseiOS(specs, agentPlatform)

	case specs == "macintosh":
		u.parseMacintosh(ua, hints)

	// Blackberry
	case strings.Contains(ua, "blackberry") || strings.Contains(ua, "playbook"):
		u.OS.Platform = PlatformBlackberry
		u.OS.Name = OSBlackberry

	// Windows Phone
	case strings.Contains(agentPlatform, "windows phone "):
		u.parseWindowsPhone(agentPlatform)

	// Windows, Xbox
	case strings.Contains(ua, "windows ") || strings.Contains(ua, "microsoft-cryptoapi"):
		u.parseWindows(ua)

	// Kindle
	case strings.Contains(ua, "kindle/") || isAmazonFire(agentPlatform):
		u.OS.Platform = PlatformLinux
		u.OS.Name = OSKindle

	// Linux (broader attempt)
	case strings.Contains(ua, "linux"):
		u.parseLinux(ua, agentPlatform)

	// WebOS (non-linux flagged)
	case isWebOS(ua):
		u.OS.Platform = PlatformLinux
		u.OS.Name = OSWebOS

	// Nintendo
	case strings.Contains(ua, "nintendo"):
		u.OS.Platform = PlatformNintendo
		u.OS.Name = OSNintendo

	// Playstation
	// "vita" is deliberately absent: every real Vita agent says
	// "PlayStation Vita", and unanchored it reads "VitaMahjong", an iOS
	// game, as a console.
	case strings.Contains(ua, "playstation") || strings.Contains(ua, "psp"):
		u.OS.Platform = PlatformPlaystation
		u.OS.Name = OSPlaystation

	// Android
	case strings.Contains(ua, "android"):
		u.parseLinux(ua, agentPlatform)

	// Apple TV, ahead of the CFNetwork case below: its native apps carry the
	// same Darwin signature as an iPhone's, and the model is what separates
	// them. The version follows the model generation ("AppleTV6,2/11.1") or,
	// for the media player agents, sits in the iOS style "CPU OS" field.
	case strings.Contains(ua, "appletv") || strings.Contains(ua, "apple tv"):
		u.OS.Platform = PlatformAppleTV
		u.OS.Name = OSTvOS
		u.parseTvOSVersion(ua)

	// Apps that name iOS rather than the device they run on:
	// "Hulu/9.34.0 (iOS 26.5.1; en_US; iPhone18,2; 23F81)",
	// "TuneIn Radio/27.1.0 (iPadOS/16.6)". Ahead of CFNetwork, which some of
	// them also carry, because the version here is the real one.
	case strings.Contains(agentPlatform, "ipados") || strings.Contains(agentPlatform, "ios ") ||
		strings.Contains(agentPlatform, "ios/"):
		u.parseiOSNative(agentPlatform)

	// Apple CFNetwork
	case strings.Contains(ua, "cfnetwork") && strings.Contains(ua, "darwin"):
		u.parseAppleNative(ua, hints)

	// Roku, after CFNetwork: the players state their OS version after the
	// model ("Roku/DVP-12.5", "Roku4640X/DVP-7.70") and never carry Darwin,
	// while "Roku" on its own is also the remote control app, which runs on
	// a phone and is caught above.
	case strings.Contains(ua, "roku"):
		u.OS.Platform = PlatformLinux
		u.OS.Name = OSRoku
		u.OS.Version.parseAfter(ua, "/dvp-", "roku/")

	default:
		u.OS.Platform = PlatformUnknown
		u.OS.Name = OSUnknown
	}

	return u.applyBotDefaults()
}

// applyBotDefaults reports whether the agent is a bot, and if so overwrites the OS and
// device fields with the bot defaults.
func (u *UserAgent) applyBotDefaults() bool {
	if u.IsBot() {
		u.OS.Platform = PlatformBot
		u.OS.Name = OSBot
		u.DeviceType = DeviceComputer
		return true
	}
	return false
}

// parseLinux returns the `Platform`, `OSName` and Version of UAs with
// 'linux' listed as their platform.
func (u *UserAgent) parseLinux(ua, agentPlatform string) {

	switch {
	// Kindle Fire
	case strings.Contains(ua, "kindle") || isAmazonFire(agentPlatform):
		// get the version of Android if available, though we don't call this OSAndroid
		u.OS.Platform = PlatformLinux
		u.OS.Name = OSKindle
		u.OS.Version.parseAfter(agentPlatform, "android ")

	// Android, Kindle Fire
	case strings.Contains(ua, "android") || strings.Contains(ua, "googletv"):
		// Android
		u.OS.Platform = PlatformLinux
		u.OS.Name = OSAndroid
		// A few old agents wrote it "Android-4.0.3".
		u.OS.Version.parseAfter(agentPlatform, "android ", "android-")

	// ChromeOS
	case strings.Contains(ua, "cros"):
		u.OS.Platform = PlatformLinux
		u.OS.Name = OSChromeOS

	// Tizen: Samsung's smart TVs, which report "SMART-TV; LINUX; Tizen 6.0",
	// and the few Tizen phones, which report "Linux; Tizen 2.3".
	case strings.Contains(ua, "tizen"):
		u.OS.Platform = PlatformLinux
		u.OS.Name = OSTizen
		u.OS.Version.parseAfter(ua, "tizen ", "tizen/")

	// WebOS
	case isWebOS(ua):
		u.OS.Platform = PlatformLinux
		u.OS.Name = OSWebOS

	// Linux, "Linux-like"
	default:
		u.OS.Platform = PlatformLinux
		u.OS.Name = OSLinux
	}
}

// parseiOS returns the `Platform`, `OSName` and Version of UAs with
// 'ipad', 'iphone' or 'ipod' listed as their platform. The caller's prefix
// check guarantees specs starts with one of the three.
func (u *UserAgent) parseiOS(specs, agentPlatform string) {
	switch {
	case strings.HasPrefix(specs, "iphone"):
		u.OS.Platform = PlatformiPhone
		u.OS.Name = OSiOS

	case strings.HasPrefix(specs, "ipad"):
		u.OS.Platform = PlatformiPad
		u.OS.Name = OSiPadOS

	default: // ipod, "ipod touch" included
		u.OS.Platform = PlatformiPod
		u.OS.Name = OSiOS
	}
	u.OS.parseiOSVersion(agentPlatform)
}

// parseiOSNative reads the OS out of an app agent that states iOS as a field of
// its own. The model follows in the same group - "iPhone18,2", "iPad11,3" - and
// is the only thing separating an iPhone from an iPad here.
func (u *UserAgent) parseiOSNative(agentPlatform string) {
	if strings.Contains(agentPlatform, "ipad") {
		u.OS.Platform = PlatformiPad
		u.OS.Name = OSiPadOS
	} else {
		u.OS.Platform = PlatformiPhone
		u.OS.Name = OSiOS
	}
	u.OS.Version.parseAfter(agentPlatform, "ipados/", "ipados ", "ios ", "ios/")
}

// isWebOS covers the three spellings in the wild: Palm and HP's original
// "webOS" and "hpwOS", and the "Web0S" - with a zero - that LG ships on its
// smart TVs.
func isWebOS(ua string) bool {
	return strings.Contains(ua, "webos") ||
		strings.Contains(ua, "web0s") ||
		strings.Contains(ua, "hpwos")
}

// parseTvOSVersion reads the tvOS version, which sits after the slash that
// follows the hardware generation: "AppleTV6,2/11.1", "AppleTV/1.1". The media
// player agents carry no model and state the version the iOS way instead.
func (u *UserAgent) parseTvOSVersion(ua string) {
	if _, after, ok := strings.Cut(ua, "appletv"); ok {
		if _, version, ok := strings.Cut(after, "/"); ok && u.OS.Version.parse(version) {
			return
		}
	}
	u.OS.Version.parseAfter(ua, "cpu os ", "os x ")
}

func (u *UserAgent) parseWindowsPhone(agentPlatform string) {
	u.OS.Platform = PlatformWindowsPhone

	if u.OS.Version.parseAfter(agentPlatform, "windows phone os ", "windows phone ") {
		u.OS.Name = OSWindowsPhone
	} else {
		u.OS.Name = OSUnknown
	}
}

func (u *UserAgent) parseWindows(ua string) {
	// Xbox -- it reads just like Windows
	if strings.Contains(ua, "xbox") {
		u.OS.Platform = PlatformXbox
		u.OS.Name = OSXbox
		if !u.OS.Version.parseAfter(ua, "windows nt ") {
			u.OS.Version = Version{Major: 6}
		}
		return
	}

	u.OS.Platform = PlatformWindows
	switch {
	case u.OS.Version.parseAfter(ua, "windows nt "):
		u.OS.Name = OSWindows

	case strings.Contains(ua, "windows xp"):
		u.OS.Name = OSWindows
		u.OS.Version = Version{Major: 5, Minor: 1}

	default: // no windows version at all, microsoft-cryptoapi included
		u.OS.Name = OSUnknown
	}
}

// parseMacintosh takes the whole agent rather than the platform group: the
// "os x " marker and the CFNetwork path that also calls this both need it.
func (u *UserAgent) parseMacintosh(ua string, hints *Hints) {
	u.OS.Platform = PlatformMac
	// "os x " for the long-standing spelling, "macos " for the one Apple has
	// started sending: "(Macintosh; Apple macOS 15_7_3)".
	after, ok := cutAfter(ua, "os x ", "macos ")
	if ok {
		u.OS.Name = OSMacOSX
		u.OS.Version.parse(after)

		if hints != nil && hints.ScreenSize != nil && hints.ScreenSize.isiPad() {
			u.OS.Name = OSiPadOS
			u.OS.Platform = PlatformiPad
		}

		return
	}
	u.OS.Name = OSUnknown
}

// cutAfter returns the text following the first of markers that appears in s.
func cutAfter(s string, markers ...string) (string, bool) {
	for _, m := range markers {
		if _, after, ok := strings.Cut(s, m); ok {
			return after, true
		}
	}
	return "", false
}

// parseAfter parses the version following the first of markers that appears in
// s, trying each marker in order, and stores it on v. It reports whether one
// parsed.
func (v *Version) parseAfter(s string, markers ...string) bool {
	for _, m := range markers {
		if _, after, ok := strings.Cut(s, m); ok && v.parse(after) {
			return true
		}
	}
	return false
}

// parseiOSVersion reads the iOS version out of the platform group and stores
// it on o.
func (o *OS) parseiOSVersion(agentPlatform string) {
	// "ipados/16.6" reaches here through the strict ipad prefix above, and its
	// version sits where no "cpu os" field does.
	if !o.Version.parseAfter(agentPlatform, "cpu iphone os ", "cpu os ", "ipados/", "ipados ") {
		o.Version.parse(agentPlatform)
	}
}

// maxVersionPart caps each component of a version.
//
// Nothing has ever shipped a version number this large, and accumulating one
// unchecked overflows: FuzzParse found "roku/10000000000000000000", which gave a
// negative major that then compares as older than every release there has been.
// The cap is also small enough that the multiply below cannot overflow an int on
// a 32 bit platform.
const maxVersionPart = 1 << 24

// strToVer accepts a string and returns a Version,
// with {0, 0, 0} being default.
func (v *Version) parse(str string) bool {
	if len(str) == 0 || str[0] < '0' || str[0] > '9' {
		return false
	}
	for i := range 3 {
		empty := true
		val := 0
		l := len(str) - 1

		for k, c := range str {
			if c >= '0' && c <= '9' {
				if empty {
					val = int(c) - 48
					empty = false
					if k == l {
						str = str[:0]
					}
					continue
				}

				if val == 0 {
					if c == '0' {
						if k == l {
							str = str[:0]
						}
						continue
					}
					str = str[k:]
					break
				}

				if val <= maxVersionPart {
					val = 10*val + int(c) - 48
				}
				if k == l {
					str = str[:0]
				}
				continue
			}
			str = str[k+1:]
			break
		}

		switch i {
		case 0:
			v.Major = val

		case 1:
			v.Minor = val

		case 2:
			v.Patch = val
		}
	}
	return true
}
