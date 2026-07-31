package uasurfer

import "slices"

type ScreenSize struct {
	Width  int
	Height int
}

// iPadScreenSizes must be stored longest edge first so orientation does not
// matter when comparing; TestIPadScreenSizesAreLandscape enforces that.
var iPadScreenSizes = []ScreenSize{
	{1024, 768},
	{1112, 834},
	{1366, 1024},
	{1080, 810},
	{1133, 744},
	{1194, 834},
}

// landscape returns s with its longer edge as the width, so that two sizes
// compare equal regardless of the orientation they were reported in.
func (s ScreenSize) landscape() ScreenSize {
	return ScreenSize{
		Width:  max(s.Width, s.Height),
		Height: min(s.Width, s.Height),
	}
}

func (s ScreenSize) isiPad() bool {
	return slices.Contains(iPadScreenSizes, s.landscape())
}
