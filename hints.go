package uasurfer

import "strings"

// Hints carries what the agent cannot say for itself.
//
// The client hint fields take the raw header value, quotes and all, so that a
// caller can pass what it already has:
//
//	uasurfer.ParseWithHints(r.UserAgent(), &uasurfer.Hints{
//		Mobile:          r.Header.Get("Sec-CH-UA-Mobile"),
//		Platform:        r.Header.Get("Sec-CH-UA-Platform"),
//		PlatformVersion: r.Header.Get("Sec-CH-UA-Platform-Version"),
//		FormFactors:     r.Header.Get("Sec-CH-UA-Form-Factors"),
//	})
//
// Only Chromium browsers send these, and only for the hints a site has asked
// for with Accept-CH. Every field is optional: an empty one changes nothing, so
// there is no need to check before setting.
type Hints struct {
	// ScreenSize identifies an iPad running in desktop mode, which since
	// iPadOS 13 is indistinguishable from a Mac in the agent alone.
	ScreenSize *ScreenSize

	// Mobile is Sec-CH-UA-Mobile: "?1" on a phone, "?0" on anything else.
	// Chromium sends it by default, unasked.
	Mobile string

	// Platform is Sec-CH-UA-Platform: "Android", "Chrome OS", "iOS", "Linux",
	// "macOS", "Windows", "Unknown". Also sent by default.
	Platform string

	// PlatformVersion is Sec-CH-UA-Platform-Version, and is the only way to
	// recover an OS version that the agent no longer tells the truth about:
	// Chromium reports macOS frozen at 10.15.7, and Android as 10, whatever the
	// device runs.
	//
	// Windows is deliberately left alone. OS.Version is the NT version there,
	// and this hint counts differently - 13 or more means Windows 11, which NT
	// numbering cannot express, so a caller wanting that distinction has to read
	// the header itself.
	PlatformVersion string

	// FormFactors is Sec-CH-UA-Form-Factors: one or more of "Desktop",
	// "Automotive", "Mobile", "Tablet", "XR", "EInk", "Watch". The most direct
	// answer to a question the agent has never answered well, and the only one
	// that settles an Android tablet outright.
	FormFactors string
}

// formFactors maps the header's values onto the device types we report.
// "Automotive", "XR" and "EInk" have no constant of their own, and an agent
// stating one is left as whatever it parsed as rather than forced into a
// neighbouring shape.
var formFactors = map[string]DeviceType{
	"desktop": DeviceComputer,
	"mobile":  DevicePhone,
	"tablet":  DeviceTablet,
	"watch":   DeviceWearable,
}

// apply refines an already parsed agent with whatever the hints carry. The agent
// is parsed first and only corrected here: hints are a supplement to it, absent
// far more often than present.
func (h *Hints) apply(u *UserAgent) {
	if h == nil {
		return
	}

	h.applyPlatformVersion(u)

	// Form factors are stated rather than inferred, so they win outright - but
	// not over a television, which no form factor value describes and which the
	// agent identifies perfectly well.
	if u.DeviceType != DeviceTV {
		for f := range strings.SplitSeq(h.FormFactors, ",") {
			if d, ok := formFactors[unquoteHint(f)]; ok {
				u.DeviceType = d
				return
			}
		}
	}

	// Failing that, the mobile hint separates a phone from a tablet, which is
	// what the agent is worst at.
	switch unquoteHint(h.Mobile) {
	case "?1":
		switch u.DeviceType {
		case DeviceComputer, DeviceTablet, DeviceUnknown:
			u.DeviceType = DevicePhone
		}
	case "?0":
		if u.DeviceType == DevicePhone {
			u.DeviceType = DeviceTablet
		}
	}
}

// applyPlatformVersion takes the OS version from the hint where the agent's own
// is known to be a fiction, and where both count the same way.
func (h *Hints) applyPlatformVersion(u *UserAgent) {
	if h.PlatformVersion == "" {
		return
	}

	switch unquoteHint(h.Platform) {
	case "macos":
		if u.OS.Name == OSMacOSX {
			u.OS.Version.parse(unquoteHint(h.PlatformVersion))
		}
	case "android":
		if u.OS.Name == OSAndroid {
			u.OS.Version.parse(unquoteHint(h.PlatformVersion))
		}
	case "chrome os", "chromium os":
		if u.OS.Name == OSChromeOS {
			u.OS.Version.parse(unquoteHint(h.PlatformVersion))
		}
	}
}

// unquoteHint trims the whitespace and double quotes a structured header field
// arrives wrapped in, and lowercases what is left so it compares like the rest
// of the parser's input.
func unquoteHint(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"`)
	return strings.ToLower(strings.TrimSpace(s))
}
