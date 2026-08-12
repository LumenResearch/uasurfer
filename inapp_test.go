package uasurfer

import "testing"

// One real agent per appMarkers entry, reached through Parse: on iOS the
// webkitApp path, on Android the tail behind the Chrome token. The first Line
// row pins the case the old "version/" gate lost: an app that appends its name
// to Safari's full agent, Version/ included.
func TestInAppBrowsers(t *testing.T) {
	tests := []struct {
		agent string
		want  BrowserName
	}{
		// Line iOS, both shapes: appended to Safari's agent, and a bare webview
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 13_5_1 like Mac OS X) AppleWebKit/605.1.15 " +
			"(KHTML, like Gecko) Version/13.1.1 Mobile/15E148 Safari/604.1 Line/10.9.1", BrowserLine},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 16_5 like Mac OS X) AppleWebKit/605.1.15 " +
			"(KHTML, like Gecko) Mobile/15E148 Line/13.5.0", BrowserLine},
		{"Mozilla/5.0 (Linux; Android 13; SM-S901B Build/TP1A.220624.014; wv) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Version/4.0 Chrome/108.0.5359.128 Mobile Safari/537.36 Line/13.1.0/IAB", BrowserLine},

		// Facebook: FBAN/FBIOS on iOS, FB_IAB on Android
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 17_5_1 like Mac OS X) AppleWebKit/605.1.15 " +
			"(KHTML, like Gecko) Mobile/15E148 [FBAN/FBIOS;FBAV/456.0.0.36.107;FBBV/577025896]", BrowserFacebook},
		{"Mozilla/5.0 (Linux; Android 5.1.1; R7sf Build/LMY47V; wv) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Version/4.0 Chrome/80.0.3987.149 Mobile Safari/537.36 " +
			"[FB_IAB/FB4A;FBAV/263.0.0.46.121;]", BrowserFacebook},

		// Instagram
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 16_5 like Mac OS X) AppleWebKit/605.1.15 " +
			"(KHTML, like Gecko) Mobile/15E148 Instagram 289.0.0.13.62 (iPhone14,2; iOS 16_5; " +
			"en_US; en; scale=3.00; 1170x2532; 489720557)", BrowserInstagram},

		// WeChat
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 " +
			"(KHTML, like Gecko) Mobile/15E148 MicroMessenger/8.0.16(0x18001023) NetType/WIFI " +
			"Language/zh_CN", BrowserWeChat},
		{"Mozilla/5.0 (Linux; Android 11; SM-G991B Build/RP1A.200720.012; wv) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Version/4.0 Chrome/88.0.4324.181 Mobile Safari/537.36 " +
			"MicroMessenger/8.0.2.1860(0x28000234) Process/toolsmp WeChat/arm64", BrowserWeChat},

		// TikTok: musical_ly on iOS, trill_ and BytedanceWebview on Android
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 16_6 like Mac OS X) AppleWebKit/605.1.15 " +
			"(KHTML, like Gecko) Mobile/15E148 musical_ly_30.5.0 JsSdk/2.0 NetType/WIFI " +
			"Channel/App Store ByteLocale/en Region/US", BrowserTikTok},
		{"Mozilla/5.0 (Linux; Android 10; SM-A505F Build/QP1A.190711.020; wv) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Version/4.0 Chrome/95.0.4638.74 Mobile Safari/537.36 " +
			"trill_2022107030 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/musical_ly " +
			"app_version/21.7.3 ByteLocale/en Region/GB BytedanceWebview/d8a21c6", BrowserTikTok},

		// Snapchat
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 16_1_1 like Mac OS X) AppleWebKit/605.1.15 " +
			"(KHTML, like Gecko) Mobile/15E148 Snapchat/12.10.0.36 (like Safari/8614.2.9.0.10)", BrowserSnapchat},

		// DuckDuckGo: appended to Safari's agent on iOS, behind Chrome on Android
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 13_5 like Mac OS X) AppleWebKit/605.1.15 " +
			"(KHTML, like Gecko) Version/13.1.1 Mobile/15E148 DuckDuckGo/7 Safari/605.1.15", BrowserDuckDuckGo},
		{"Mozilla/5.0 (Linux; Android 11) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 " +
			"Chrome/87.0.4280.141 Mobile DuckDuckGo/5 Safari/537.36", BrowserDuckDuckGo},

		// Chromium under another brand
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) " +
			"Chrome/124.0.0.0 Safari/537.36 Vivaldi/6.7.3329.31", BrowserVivaldi},
		{"Mozilla/5.0 (Linux; Android 13; SM-G991N) AppleWebKit/537.36 (KHTML, like Gecko) " +
			"Chrome/115.0.0.0 Whale/3.22.9.10 Mobile Safari/537.36", BrowserWhale},
		{"Mozilla/5.0 (Linux; U; Android 13; 2201123G Build/TKQ1.220829.002) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Version/4.0 Chrome/112.0.5615.136 Mobile Safari/537.36 " +
			"XiaoMi/MiuiBrowser/17.4.11", BrowserMIUI},
		{"Mozilla/5.0 (Linux; Android 12; NAM-LX9 Build/HUAWEINAM-L29) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Chrome/99.0.4844.88 HuaweiBrowser/13.0.5.303 Mobile Safari/537.36", BrowserHuawei},

		// plain Safari keeps its name: no marker, however Safari-shaped the agent
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 17_5 like Mac OS X) AppleWebKit/605.1.15 " +
			"(KHTML, like Gecko) Version/17.5 Mobile/15E148 Safari/604.1", BrowserSafari},
	}

	for _, tt := range tests {
		if got := Parse(tt.agent).Browser.Name; got != tt.want {
			t.Errorf("browser = %v, want %v: %.90s", got, tt.want, tt.agent)
		}
	}
}

// A marker only counts at the start of a field: "baseline/" must not read as
// Line, and Snapchat's own quirk of gluing its name to "Safari/537.36" stays
// unclaimed rather than matched by accident.
func TestAppBrowserAnchoring(t *testing.T) {
	for _, ua := range []string{
		"mozilla/5.0 (iphone; cpu iphone os 16_5 like mac os x) applewebkit/605.1.15 " +
			"(khtml, like gecko) mobile/15e148 baseline/2.0",
		"mozilla/5.0 (linux; android 12; sm-g991u; wv) applewebkit/537.36 " +
			"(khtml, like gecko) version/4.0 chrome/101.0.4951.61 mobile safari/537.36snapchat11.79.0.33",
	} {
		if got := appBrowser(ua); got != BrowserUnknown {
			t.Errorf("appBrowser = %v, want BrowserUnknown: %.90s", got, ua)
		}
	}
}
