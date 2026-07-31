package uasurfer

type ScreenSize struct {
	Width  int
	Height int
}

var iPadScreenSizes = []ScreenSize{
	{1024, 768},
	{1112, 834},
	{1366, 1024},
	{1080, 810},
	{1133, 744},
	{1194, 834},
}

func (s *ScreenSize) isiPad() bool {
	if s == nil {
		return false
	}

	long, short := s.Width, s.Height
	if short > long {
		long, short = short, long
	}

	for _, size := range iPadScreenSizes {
		sizeLong, sizeShort := size.Width, size.Height
		if sizeShort > sizeLong {
			sizeLong, sizeShort = sizeShort, sizeLong
		}

		if long == sizeLong && short == sizeShort {
			return true
		}
	}

	return false
}
