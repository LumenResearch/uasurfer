package uasurfer

import "testing"

func TestIsTV(t *testing.T) {
	for _, ua := range []string{"smarttv", "hbbtv/1.1.1", "aftsss build", "roku4640x", "mbox android"} {
		if !isTV(ua) {
			t.Errorf("want TV: %q", ua)
		}
	}
	for _, ua := range []string{"iphone os 17", "xbox one; mbox", "windows nt 10.0"} {
		if isTV(ua) {
			t.Errorf("want not TV: %q", ua)
		}
	}
}
