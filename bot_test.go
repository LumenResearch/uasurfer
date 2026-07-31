package uasurfer

import (
	"embed"
	"strings"
	"testing"
)

//go:embed testdata/bots.tsv testdata/devices.tsv testdata/tv.tsv
var fixtures embed.FS

// readFixtures returns the rows of a testdata file, "#" comments and blank
// lines dropped, each row split on tabs.
func readFixtures(t *testing.T, name string, fields int) [][]string {
	t.Helper()

	b, err := fixtures.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("reading %s: %v", name, err)
	}

	var rows [][]string
	for i, line := range strings.Split(string(b), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		row := strings.Split(line, "\t")
		if len(row) != fields {
			t.Fatalf("%s line %d: got %d fields, want %d: %q", name, i+1, len(row), fields, line)
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		t.Fatalf("%s: no fixtures", name)
	}
	return rows
}

// botRecallFloor is the share of testdata/bots.tsv we detect. It is a floor, not
// a target: raise it when detection improves, and treat a change that pushes
// recall below it as a regression to explain rather than a number to lower.
const botRecallFloor = 0.82

func TestBotFixtures(t *testing.T) {
	rows := readFixtures(t, "bots.tsv", 2)

	var detected int
	for _, row := range rows {
		want, agent := row[0], row[1]
		ua := Parse(agent)
		if ua.IsBot() {
			detected++
		}

		if want == "!" {
			// A known miss. Not asserted: catching one is an improvement, and
			// the row should then be relabelled rather than left to fail.
			if ua.IsBot() {
				t.Logf("now detected, relabel this row from %q to %q: %s",
					"!", ua.Browser.Name.StringTrimPrefix(), agent)
			}
			continue
		}

		if !ua.IsBot() {
			t.Errorf("IsBot() = false, want true: %s", agent)
			continue
		}
		if got := ua.Browser.Name.StringTrimPrefix(); got != want {
			t.Errorf("browser = %s, want %s: %s", got, want, agent)
		}
		// A bot's OS and device are overwritten wholesale, so they are part of
		// the claim: a caller filtering on either must not see a real device.
		if ua.OS.Name != OSBot || ua.OS.Platform != PlatformBot || ua.DeviceType != DeviceComputer {
			t.Errorf("bot defaults = %v/%v/%v, want bot/bot/computer: %s",
				ua.OS.Platform, ua.OS.Name, ua.DeviceType, agent)
		}
	}

	if recall := float64(detected) / float64(len(rows)); recall < botRecallFloor {
		t.Errorf("bot recall = %.3f (%d/%d), below the floor of %.2f",
			recall, detected, len(rows), botRecallFloor)
	} else {
		t.Logf("bot recall %.3f (%d/%d), floor %.2f", recall, detected, len(rows), botRecallFloor)
	}
}

// The precision half of the bargain: generic tokens buy the recall above, and
// this is what stops them costing a real device.
func TestDeviceFixturesAreNotBots(t *testing.T) {
	for _, row := range readFixtures(t, "devices.tsv", 2) {
		want, agent := row[0], row[1]
		ua := Parse(agent)
		if ua.IsBot() {
			t.Errorf("IsBot() = true, want false (%s): %s", ua.Browser.Name.StringTrimPrefix(), agent)
			continue
		}
		if got := ua.DeviceType.StringTrimPrefix(); got != want {
			t.Errorf("device = %s, want %s: %s", got, want, agent)
		}
	}
}

func TestBotName(t *testing.T) {
	tests := []struct {
		agent string
		want  BrowserName
	}{
		// generic tokens
		{"Mozilla/5.0 (compatible; SomeNewBot/1.0)", BrowserBot},
		{"Mozilla/5.0 (compatible; Whatever/1.0; +http://example.com/robot)", BrowserBot},
		{"SomeSpider/2.0", BrowserBot},
		{"NewCrawler", BrowserBot},
		{"curl/8.6.0", BrowserBot},
		{"python-requests/2.32.3", BrowserBot},
		{"Go-http-client/1.1", BrowserBot},

		// a named crawler beats the generic token it contains
		{"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)", BrowserGoogleBot},
		{"Mozilla/5.0 (compatible; AhrefsBot/7.0; +http://ahrefs.com/robot/)", BrowserAhrefsBot},
		{"Mozilla/5.0 (compatible; GPTBot/1.2; +https://openai.com/gptbot)", BrowserOpenAIBot},
		{"AdsBot-Google (+http://www.google.com/adsbot.html)", BrowserGoogleBot},

		// crawlers wearing a whole browser agent: the token can sit anywhere
		{"Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; ClaudeBot/1.0; " +
			"+claudebot@anthropic.com) Chrome/W.X.Y.Z Safari/537.36", BrowserAnthropicBot},
		{"Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Chrome/126.0.0.0 Mobile Safari/537.36 (compatible; " +
			"Googlebot/2.1; +http://www.google.com/bot.html)", BrowserGoogleBot},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) " +
			"HeadlessChrome/126.0.6478.61 Safari/537.36", BrowserBot},

		// the "bot" word boundary
		{"Mozilla/5.0 (Linux; Android 4.2.1; CUBOT ONE Build/JOP40D) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Chrome/27.0.1453.90 Mobile Safari/537.36", BrowserUnknown},
		{"Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; Banca Caboto s.p.a.; rv:11.0) like Gecko",
			BrowserUnknown},

		// a name claimed by resemblance is not that crawler
		{"TelegramBot (like TwitterBot)", BrowserBot},
		{"Feedly/1.0 (+http://www.feedly.com/fetcher.html; like FeedFetcher-Google)", BrowserBot},
		{"FreshRSS/1.11.2 (Linux; https://freshrss.org) like Googlebot", BrowserBot},

		// a named crawler wins wherever in the agent it sits
		{"Mozilla/5.0 (seoanalyzer; compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)", BrowserBingBot},

		// vendor tokens that belong to apps rather than crawlers
		{"YahooJMobileApp/1.1 (Android emg; 1.3.4) (samsung; SC-02E; samsung; SC-02E; 4.1.1/JRO03C)",
			BrowserUnknown},
		{"WhatsApp/2.11.102 Android/2.3.3 Device/HTC-HTC_Wildfire", BrowserUnknown},
		{"HuaweiU3100/B000 Browser/Obigo-Browser/Q05A Java/HWJa/2.0 Profile/MIDP-2.0", BrowserUnknown},
		{"BaconReader/3.3.1(Android 4.4.2; LGE Nexus 4)", BrowserUnknown},

		// and a plain browser, which must never pay a bot's price
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) " +
			"Chrome/126.0.0.0 Safari/537.36", BrowserUnknown},
	}

	for _, tt := range tests {
		ua := normalise(tt.agent)
		got, ok := botName(ua)
		if tt.want == BrowserUnknown {
			if ok {
				t.Errorf("botName(%.60q) = %v, want no match", tt.agent, got)
			}
			continue
		}
		if !ok {
			t.Errorf("botName(%.60q) = no match, want %v", tt.agent, tt.want)
			continue
		}
		if got != tt.want {
			t.Errorf("botName(%.60q) = %v, want %v", tt.agent, got, tt.want)
		}
	}
}

func TestIsBotWord(t *testing.T) {
	// A crawler's name, in the four shapes it takes.
	for _, agent := range []string{
		"googlebot/2.1", "petalbot;", "discordbot", "semrushbot/7~bl", "adsbot-google",
		"contextad bot 1.0", "better uptime bot mozilla/5.0", "df bot 1.0",
	} {
		i := strings.Index(agent, "bot")
		if !isBotWord(agent, i, i+3) {
			t.Errorf("isBotWord(%q) = false, want true", agent)
		}
	}

	// Letters inside a longer word: a phone brand, an Italian bank, a plural.
	for _, agent := range []string{
		"cubot one build/jop40d", "banca caboto s.p.a.", "robots welcome",
	} {
		i := strings.Index(agent, "bot")
		if isBotWord(agent, i, i+3) {
			t.Errorf("isBotWord(%q) = true, want false", agent)
		}
	}
}

func TestHasContactURL(t *testing.T) {
	for _, agent := range []string{
		"wordpress/4.3.27; http://afterice.se",
		"y!j-asr/0.1 crawler (http://www.yahoo-help.jp/)",
		"someagent/1.0 (+https://example.com/policy)",
		"linkanalyser [www.example.org]",
		"tool/1.0 www.example.com",
	} {
		if !hasContactURL(agent) {
			t.Errorf("hasContactURL(%q) = false, want true", agent)
		}
	}

	for _, agent := range []string{
		"",
		"mozilla/5.0 (windows nt 10.0; win64; x64) applewebkit/537.36",
		// a proxy rewrote the agent and left its own URL at the front: the
		// device behind it is not a crawler
		"http://atamg.wup.ru/samsung-gt-s5233t/s5233txejf1 shp/vpp/r5 jasmine/0.8",
		// and a URL glued to a word is not a field of its own
		"someapp/1.0 seehttp://example.com",
	} {
		if hasContactURL(agent) {
			t.Errorf("hasContactURL(%q) = true, want false", agent)
		}
	}
}

// The buckets are the index botName reads, and a marker landing in the wrong one
// is invisible: the scan simply never tries it.
func TestBotMarkersAreIndexed(t *testing.T) {
	for _, m := range botMarkers {
		if m.s == "" {
			t.Error("empty bot marker")
			continue
		}
		if m.s != strings.ToLower(m.s) {
			t.Errorf("marker %q is not lowercase, so the normalised agent can never match it", m.s)
		}
		if len(m.s) < 2 {
			t.Errorf("marker %q is shorter than the digram gate, so it is never tried", m.s)
			continue
		}
		k := uint16(m.s[0])<<8 | uint16(m.s[1])
		if botDigrams[k>>6]&(1<<(k&63)) == 0 {
			t.Errorf("marker %q: opening pair missing from botDigrams, so it is never tried", m.s)
		}
		c := m.s[0]
		found := false
		for _, b := range botBuckets[c] {
			found = found || b.s == m.s
		}
		if !found {
			t.Errorf("marker %q missing from its bucket", m.s)
		}
	}

	// Longest first, which is what makes a named crawler win over the generic
	// token inside it.
	for c, bucket := range botBuckets {
		for i := 1; i < len(bucket); i++ {
			if len(bucket[i-1].s) < len(bucket[i].s) {
				t.Errorf("bucket %q is not ordered longest first: %q before %q",
					byte(c), bucket[i-1].s, bucket[i].s)
			}
		}
	}
}

func BenchmarkBotName(b *testing.B) {
	agent := normalise("Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	for b.Loop() {
		botName(agent)
	}
}

// The pass every parse pays for: a browser agent that matches nothing.
func BenchmarkBotNameMiss(b *testing.B) {
	agent := normalise("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	for b.Loop() {
		botName(agent)
	}
}
