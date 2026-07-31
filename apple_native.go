package uasurfer

var darwinToIOSMajor = map[int]int{
	16: 10,
	17: 11,
	18: 12,
	19: 13,
	20: 14,
	21: 15,
	22: 16,
	23: 17,
	24: 18,
	25: 26,
}

func darwinToIOS(darwinMajor, darwinMinor int) Version {
	if major, ok := darwinToIOSMajor[darwinMajor]; ok {
		return Version{Major: major, Minor: darwinMinor}
	}

	if darwinMajor > 25 {
		return Version{Major: darwinMajor + 1, Minor: darwinMinor}
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
