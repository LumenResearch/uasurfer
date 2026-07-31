// Package uasurfer provides fast and reliable abstraction
// of HTTP User-Agent strings. The philosophy is to identify
// technologies that holds >1% market share, and to avoid
// expending resources and accuracy on guessing at esoteric UA
// strings.
package uasurfer

import (
	"cmp"
	"strings"
)

// DeviceType (int) returns a constant.
type DeviceType int

// A complete list of supported devices in the
// form of constants.
const (
	DeviceUnknown DeviceType = iota
	DeviceComputer
	DeviceTablet
	DevicePhone
	DeviceConsole
	DeviceWearable
	DeviceTV

	// _deviceTypeFinal terminates the list so tests can enumerate it; keep it last.
	_deviceTypeFinal
)

func (d DeviceType) String() string {
	switch d {
	case DeviceComputer:
		return "DeviceComputer"
	case DeviceTablet:
		return "DeviceTablet"
	case DevicePhone:
		return "DevicePhone"
	case DeviceConsole:
		return "DeviceConsole"
	case DeviceWearable:
		return "DeviceWearable"
	case DeviceTV:
		return "DeviceTV"
	default:
		// anything out of range, including a value cast from a newer
		// release, reads as unknown rather than a numeric placeholder
		return "DeviceUnknown"
	}
}

// StringTrimPrefix is like String() but trims the "Device" prefix
func (d DeviceType) StringTrimPrefix() string {
	return strings.TrimPrefix(d.String(), "Device")
}

// BrowserName (int) returns a constant.
type BrowserName int

// A complete list of supported web browsers in the
// form of constants.
const (
	BrowserUnknown BrowserName = iota
	BrowserChrome
	BrowserIE
	BrowserSafari
	BrowserFirefox
	BrowserAndroid
	BrowserOpera
	BrowserBlackberry
	BrowserUCBrowser
	BrowserSilk
	BrowserNokia
	BrowserNetFront
	BrowserQQ
	BrowserMaxthon
	BrowserSogouExplorer
	BrowserSpotify
	BrowserNintendo
	BrowserSamsung
	BrowserYandex
	BrowserCocCoc
	BrowserBot // Bot list begins here
	BrowserAppleBot
	BrowserBaiduBot
	BrowserBingBot
	BrowserDuckDuckGoBot
	BrowserFacebookBot
	BrowserGoogleBot
	BrowserLinkedInBot
	BrowserMsnBot
	BrowserPingdomBot
	BrowserTwitterBot
	BrowserYandexBot
	BrowserCocCocBot
	BrowserYahooBot
	BrowserOpenAIBot
	BrowserAnthropicBot
	BrowserPerplexityBot
	BrowserAmazonBot
	BrowserBytedanceBot
	BrowserCommonCrawlBot
	BrowserAhrefsBot
	BrowserSemrushBot
	BrowserPetalBot

	// In-app webviews. A tap on a link inside one of these apps renders in a
	// frame the app controls, with no address bar and its own idea of a viewport,
	// which is a different surface to measure than the browser it is built on.
	BrowserFacebook
	BrowserInstagram
	BrowserWeChat
	BrowserTikTok
	BrowserSnapchat
	BrowserLine

	// Chromium with another name on it. The engine is Chrome's; the vendor,
	// release cadence and default settings are not.
	BrowserVivaldi
	BrowserWhale
	BrowserMIUI
	BrowserHuawei
	BrowserDuckDuckGo

	// _browserNameFinal terminates the list so tests can enumerate it; keep it last.
	_browserNameFinal
)

func (b BrowserName) String() string {
	switch b {
	case BrowserChrome:
		return "BrowserChrome"
	case BrowserIE:
		return "BrowserIE"
	case BrowserSafari:
		return "BrowserSafari"
	case BrowserFirefox:
		return "BrowserFirefox"
	case BrowserAndroid:
		return "BrowserAndroid"
	case BrowserOpera:
		return "BrowserOpera"
	case BrowserBlackberry:
		return "BrowserBlackberry"
	case BrowserUCBrowser:
		return "BrowserUCBrowser"
	case BrowserSilk:
		return "BrowserSilk"
	case BrowserNokia:
		return "BrowserNokia"
	case BrowserNetFront:
		return "BrowserNetFront"
	case BrowserQQ:
		return "BrowserQQ"
	case BrowserMaxthon:
		return "BrowserMaxthon"
	case BrowserSogouExplorer:
		return "BrowserSogouExplorer"
	case BrowserSpotify:
		return "BrowserSpotify"
	case BrowserNintendo:
		return "BrowserNintendo"
	case BrowserSamsung:
		return "BrowserSamsung"
	case BrowserYandex:
		return "BrowserYandex"
	case BrowserCocCoc:
		return "BrowserCocCoc"
	case BrowserBot:
		return "BrowserBot"
	case BrowserAppleBot:
		return "BrowserAppleBot"
	case BrowserBaiduBot:
		return "BrowserBaiduBot"
	case BrowserBingBot:
		return "BrowserBingBot"
	case BrowserDuckDuckGoBot:
		return "BrowserDuckDuckGoBot"
	case BrowserFacebookBot:
		return "BrowserFacebookBot"
	case BrowserGoogleBot:
		return "BrowserGoogleBot"
	case BrowserLinkedInBot:
		return "BrowserLinkedInBot"
	case BrowserMsnBot:
		return "BrowserMsnBot"
	case BrowserPingdomBot:
		return "BrowserPingdomBot"
	case BrowserTwitterBot:
		return "BrowserTwitterBot"
	case BrowserYandexBot:
		return "BrowserYandexBot"
	case BrowserCocCocBot:
		return "BrowserCocCocBot"
	case BrowserYahooBot:
		return "BrowserYahooBot"
	case BrowserOpenAIBot:
		return "BrowserOpenAIBot"
	case BrowserAnthropicBot:
		return "BrowserAnthropicBot"
	case BrowserPerplexityBot:
		return "BrowserPerplexityBot"
	case BrowserAmazonBot:
		return "BrowserAmazonBot"
	case BrowserBytedanceBot:
		return "BrowserBytedanceBot"
	case BrowserCommonCrawlBot:
		return "BrowserCommonCrawlBot"
	case BrowserAhrefsBot:
		return "BrowserAhrefsBot"
	case BrowserSemrushBot:
		return "BrowserSemrushBot"
	case BrowserPetalBot:
		return "BrowserPetalBot"
	case BrowserFacebook:
		return "BrowserFacebook"
	case BrowserInstagram:
		return "BrowserInstagram"
	case BrowserWeChat:
		return "BrowserWeChat"
	case BrowserTikTok:
		return "BrowserTikTok"
	case BrowserSnapchat:
		return "BrowserSnapchat"
	case BrowserLine:
		return "BrowserLine"
	case BrowserVivaldi:
		return "BrowserVivaldi"
	case BrowserWhale:
		return "BrowserWhale"
	case BrowserMIUI:
		return "BrowserMIUI"
	case BrowserHuawei:
		return "BrowserHuawei"
	case BrowserDuckDuckGo:
		return "BrowserDuckDuckGo"
	default:
		// anything out of range, including a value cast from a newer
		// release, reads as unknown rather than a numeric placeholder
		return "BrowserUnknown"
	}
}

// StringTrimPrefix is like String() but trims the "Browser" prefix
func (b BrowserName) StringTrimPrefix() string {
	return strings.TrimPrefix(b.String(), "Browser")
}

// OSName (int) returns a constant.
type OSName int

// A complete list of supported OSes in the
// form of constants. For handling particular versions
// of operating systems (e.g. Windows 2000), see
// the README.md file.
const (
	OSUnknown OSName = iota
	OSWindowsPhone
	OSWindows
	OSMacOSX
	OSiOS
	OSiPadOS
	OSAndroid
	OSBlackberry
	OSChromeOS
	OSKindle
	OSWebOS
	OSLinux
	OSPlaystation
	OSXbox
	OSNintendo
	OSBot
	OSTizen // Samsung smart TVs, and the handful of Tizen phones
	OSRoku
	OSTvOS

	// _osNameFinal terminates the list so tests can enumerate it; keep it last.
	_osNameFinal
)

func (o OSName) String() string {
	switch o {
	case OSWindowsPhone:
		return "OSWindowsPhone"
	case OSWindows:
		return "OSWindows"
	case OSMacOSX:
		return "OSMacOSX"
	case OSiOS:
		return "OSiOS"
	case OSiPadOS:
		return "OSiPadOS"
	case OSAndroid:
		return "OSAndroid"
	case OSBlackberry:
		return "OSBlackberry"
	case OSChromeOS:
		return "OSChromeOS"
	case OSKindle:
		return "OSKindle"
	case OSWebOS:
		return "OSWebOS"
	case OSLinux:
		return "OSLinux"
	case OSPlaystation:
		return "OSPlaystation"
	case OSXbox:
		return "OSXbox"
	case OSNintendo:
		return "OSNintendo"
	case OSBot:
		return "OSBot"
	case OSTizen:
		return "OSTizen"
	case OSRoku:
		return "OSRoku"
	case OSTvOS:
		return "OSTvOS"
	default:
		// anything out of range, including a value cast from a newer
		// release, reads as unknown rather than a numeric placeholder
		return "OSUnknown"
	}
}

// StringTrimPrefix is like String() but trims the "OS" prefix
func (o OSName) StringTrimPrefix() string {
	return strings.TrimPrefix(o.String(), "OS")
}

// Platform (int) returns a constant.
type Platform int

// A complete list of supported platforms in the
// form of constants. Many OSes report their
// true platform, such as Android OS being Linux
// platform.
const (
	PlatformUnknown Platform = iota
	PlatformWindows
	PlatformMac
	PlatformLinux
	PlatformiPad
	PlatformiPhone
	PlatformiPod
	PlatformBlackberry
	PlatformWindowsPhone
	PlatformPlaystation
	PlatformXbox
	PlatformNintendo
	PlatformBot
	PlatformAppleTV

	// _platformFinal terminates the list so tests can enumerate it; keep it last.
	_platformFinal
)

func (p Platform) String() string {
	switch p {
	case PlatformWindows:
		return "PlatformWindows"
	case PlatformMac:
		return "PlatformMac"
	case PlatformLinux:
		return "PlatformLinux"
	case PlatformiPad:
		return "PlatformiPad"
	case PlatformiPhone:
		return "PlatformiPhone"
	case PlatformiPod:
		return "PlatformiPod"
	case PlatformBlackberry:
		return "PlatformBlackberry"
	case PlatformWindowsPhone:
		return "PlatformWindowsPhone"
	case PlatformPlaystation:
		return "PlatformPlaystation"
	case PlatformXbox:
		return "PlatformXbox"
	case PlatformNintendo:
		return "PlatformNintendo"
	case PlatformBot:
		return "PlatformBot"
	case PlatformAppleTV:
		return "PlatformAppleTV"
	default:
		// anything out of range, including a value cast from a newer
		// release, reads as unknown rather than a numeric placeholder
		return "PlatformUnknown"
	}
}

// StringTrimPrefix is like String() but trims the "Platform" prefix
func (p Platform) StringTrimPrefix() string {
	return strings.TrimPrefix(p.String(), "Platform")
}

type Version struct {
	Major int
	Minor int
	Patch int
}

// Less reports whether v sorts before c, comparing major, then minor, then patch.
func (v Version) Less(c Version) bool {
	return cmp.Or(
		cmp.Compare(v.Major, c.Major),
		cmp.Compare(v.Minor, c.Minor),
		cmp.Compare(v.Patch, c.Patch),
	) < 0
}

type UserAgent struct {
	Browser    Browser
	OS         OS
	DeviceType DeviceType
}

// Browser contains the name of the browser and its version. Browsers are
// grouped without consideration for device: Chrome (Chrome/43.0) and Chrome for
// iOS (CriOS/43.0) both report as BrowserChrome with version 43.0, and Internet
// Explorer 11 and Edge 12 both report as BrowserIE with version 11 or 12.
type Browser struct {
	Name    BrowserName
	Version Version
}

type OS struct {
	Platform Platform
	Name     OSName
	Version  Version
}

// Reset resets the UserAgent to it's zero value
func (u *UserAgent) Reset() {
	u.Browser = Browser{}
	u.OS = OS{}
	u.DeviceType = DeviceUnknown
}

// botNames marks the BrowserName values that identify a bot, by the one thing
// every such constant has in common: its name ends in "Bot".
//
// A set rather than the range this used to be. The range required the bot
// constants to be contiguous and last, which meant a new browser could only be
// added by inserting it ahead of them and shifting their values - and callers
// persist those values as ints. Now either list grows by appending, and a
// constant named "…Bot" is a bot with nothing else to remember;
// TestIsBot asserts exactly that correspondence.
var botNames = func() (set [_browserNameFinal]bool) {
	for b := range _browserNameFinal {
		set[b] = strings.HasSuffix(b.String(), "Bot")
	}
	return
}()

// IsBot returns true if the UserAgent represent a bot
func (u *UserAgent) IsBot() bool {
	// A name cast from a newer release can be out of range, and reads as no bot
	// rather than panicking.
	return (u.Browser.Name >= 0 && u.Browser.Name < _browserNameFinal && botNames[u.Browser.Name]) ||
		u.OS.Name == OSBot ||
		u.OS.Platform == PlatformBot
}

// Parse accepts a raw user agent (string) and returns the UserAgent.
func Parse(ua string) *UserAgent {
	dest := new(UserAgent)
	parse(ua, nil, dest)
	return dest
}

// ParseWithHints is the same as Parse, but accepts a Hints struct.
func ParseWithHints(ua string, hints *Hints) *UserAgent {
	dest := new(UserAgent)
	parse(ua, hints, dest)
	return dest
}

// ParseUserAgent is the same as Parse, but populates the supplied UserAgent.
// It is the caller's responsibility to call Reset() on the UserAgent before
// passing it to this function.
func ParseUserAgent(ua string, dest *UserAgent) {
	parse(ua, nil, dest)
}

// ParseUserAgentWithHints is the same as ParseUserAgent, but accepts a Hints struct.
func ParseUserAgentWithHints(ua string, hints *Hints, dest *UserAgent) {
	parse(ua, hints, dest)
}

func parse(ua string, hints *Hints, dest *UserAgent) {
	ua = normalise(ua)
	if len(ua) == 0 {
		return // dest keeps its zero (Unknown) values
	}
	// each parse* reports a bot, which needs nothing further parsed
	if dest.parseOS(ua, hints) || dest.parseBrowserName(ua) {
		return
	}
	dest.parseBrowserVersion(ua)
	dest.parseDevice(ua)
	hints.apply(dest)
}

// normalise normalises the user supplied agent string so that
// we can more easily parse it.
//
// The returned string is the parser's one remaining allocation: buf itself is
// stack allocated, but converting it to a string that outlives this frame is
// not something the compiler can keep on the stack. Removing it would mean
// parsing over []byte throughout rather than string.
func normalise(ua string) string {
	if len(ua) <= 1024 {
		var buf [1024]byte
		ascii := copyLower(buf[:len(ua)], ua)
		if !ascii {
			// Fall back for non ascii characters
			return strings.ToLower(ua)
		}
		return string(buf[:len(ua)])
	}
	// Fallback for unusually long strings
	return strings.ToLower(ua)
}

// copyLower copies a lowercase version of s to b. It assumes s contains only single byte characters
// and will panic if b is nil or is not long enough to contain all the bytes from s.
// It returns early with false if any characters were non ascii.
func copyLower(b []byte, s string) bool {
	for j := 0; j < len(s); j++ {
		c := s[j]
		if c > 127 {
			return false
		}

		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}

		b[j] = c
	}
	return true
}
