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

func (u *UserAgent) evalAppleNative(ua string, hints *Hints) {
	var cfNetwork Version
	cfNetwork.findVersionNumber(ua, "cfnetwork/")

	if cfNetwork.Major < iosCFNetworkMin {
		u.evalMacintosh(ua, hints)
		return
	}

	var darwin Version
	darwin.findVersionNumber(ua, "darwin/")

	u.OS.Platform = PlatformiPhone
	u.OS.Name = OSiOS
	u.OS.Version = darwinToIOS(darwin.Major, darwin.Minor)
}
