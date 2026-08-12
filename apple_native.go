package uasurfer

// darwinToIOS maps a Darwin kernel version onto the iOS release it shipped
// with: Darwin ran iOS + 6 from Darwin 16 (iOS 10) through Darwin 24 (iOS 18),
// then Apple unified its version numbers on the year and Darwin 25 shipped
// iOS 26, one ahead, which is where the numbering now stays.
func darwinToIOS(darwinMajor, darwinMinor int) Version {
	switch {
	case darwinMajor >= 25:
		return Version{Major: darwinMajor + 1, Minor: darwinMinor}
	case darwinMajor >= 16:
		return Version{Major: darwinMajor - 6, Minor: darwinMinor}
	}
	return Version{}
}

const iosCFNetworkMin = 2000

func (u *UserAgent) parseAppleNative(ua string, hints *Hints) {
	var cfNetwork Version
	cfNetwork.parseAfter(ua, "cfnetwork/")

	if cfNetwork.Major < iosCFNetworkMin {
		u.parseMacintosh(ua, hints)
		return
	}

	var darwin Version
	darwin.parseAfter(ua, "darwin/")

	u.OS.Version = darwinToIOS(darwin.Major, darwin.Minor)

	if hints != nil && hints.ScreenSize != nil && hints.ScreenSize.isiPad() {
		u.OS.Platform = PlatformiPad
		u.OS.Name = OSiPadOS
		return
	}

	u.OS.Platform = PlatformiPhone
	u.OS.Name = OSiOS
}
