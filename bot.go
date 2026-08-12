package uasurfer

import (
	"slices"
	"strings"
)

// botMarker is a substring that identifies a bot, and the name to report for
// it. A named crawler keeps its own constant; everything else reports as the
// generic BrowserBot.
type botMarker struct {
	s    string
	name BrowserName

	// word requires s to stand as a word of its own rather than as letters
	// inside a longer one; see isBotWord.
	word bool
}

// botMarkers is deliberately not a list of crawler names. Keeping one entry per
// crawler is how the other parsers end up shipping (and having to maintain) a
// thousand of them; the generic tokens below - "bot" as a word, "spider",
// "crawl", "+http" - carry most of the recall on their own, and a name is only
// listed when it is either reported through its own constant or too quiet to
// carry a generic token.
//
// A marker is only added when it cannot plausibly appear in a real browser's
// agent. That rules out the tokens of non-browser clients that are still real
// devices with a real person behind them: dalvik, cfnetwork, okhttp and friends
// are app HTTP stacks, not crawlers.
var botMarkers = []botMarker{
	// generic. These carry roughly nine of every ten bots we detect.
	{s: "bot", word: true},
	{s: "spider"},
	{s: "crawl"},
	{s: "+http"}, // the contact URL convention, near universal among crawlers

	// named crawlers, reported through their own constant
	{s: "applebot", name: BrowserAppleBot},
	{s: "baiduspider", name: BrowserBaiduBot},
	{s: "adidxbot", name: BrowserBingBot},
	{s: "bingbot", name: BrowserBingBot},
	{s: "bingpreview", name: BrowserBingBot},
	{s: "msnbot", name: BrowserMsnBot},
	{s: "duckduckbot", name: BrowserDuckDuckGoBot},
	{s: "facebookexternalhit", name: BrowserFacebookBot},
	{s: "facebot", name: BrowserFacebookBot},
	{s: "linkedinbot", name: BrowserLinkedInBot},
	{s: "twitterbot", name: BrowserTwitterBot},
	{s: "pingdom", name: BrowserPingdomBot}, // also PingdomTMS
	{s: "coccocbot", name: BrowserCocCocBot},
	{s: "yandexbot", name: BrowserYandexBot},
	{s: "yadirectfetcher", name: BrowserYandexBot},

	// Yahoo, by token rather than by the bare vendor name: "yahoo" on its own
	// also appears in Yahoo Japan's apps and in the portal agents of Japanese
	// TVs, 46 of which read as a crawler in the corpus before this.
	{s: "slurp", name: BrowserYahooBot},
	{s: "y!j", name: BrowserYahooBot}, // Yahoo Japan's crawler fleet
	{s: "yahoocachesystem", name: BrowserYahooBot},
	{s: "yahoomailproxy", name: BrowserYahooBot},
	{s: "yahoo ad monitoring", name: BrowserYahooBot},
	{s: "yahoo link preview", name: BrowserYahooBot},

	// Google runs a fleet: the crawler, the ads crawlers, the inspection and
	// verification agents. Only "googlebot" carries the generic token.
	{s: "googlebot", name: BrowserGoogleBot},
	{s: "adsbot-google", name: BrowserGoogleBot},
	{s: "mediapartners-google", name: BrowserGoogleBot},
	{s: "googleother", name: BrowserGoogleBot},
	{s: "google-inspectiontool", name: BrowserGoogleBot},
	{s: "feedfetcher-google", name: BrowserGoogleBot},
	{s: "apis-google", name: BrowserGoogleBot},
	{s: "google-read-aloud", name: BrowserGoogleBot},
	{s: "google-safety", name: BrowserGoogleBot},

	// AI crawlers. Named because they are the fastest growing share of crawl
	// traffic and callers ask about them by name; the smaller ones (ai2bot,
	// timpibot, youbot, imagesiftbot, ...) stay generic.
	{s: "gptbot", name: BrowserOpenAIBot},
	{s: "chatgpt-user", name: BrowserOpenAIBot},
	{s: "oai-searchbot", name: BrowserOpenAIBot},
	{s: "claudebot", name: BrowserAnthropicBot},
	{s: "claude-web", name: BrowserAnthropicBot},
	{s: "claude-user", name: BrowserAnthropicBot},
	{s: "claude-searchbot", name: BrowserAnthropicBot},
	{s: "anthropic-ai", name: BrowserAnthropicBot},
	{s: "perplexitybot", name: BrowserPerplexityBot},
	{s: "perplexity-user", name: BrowserPerplexityBot},
	{s: "amazonbot", name: BrowserAmazonBot},
	{s: "bytespider", name: BrowserBytedanceBot},
	{s: "ccbot", name: BrowserCommonCrawlBot},

	// Meta's AI crawler and its link preview fetcher are one company, so they
	// share the constant that facebookexternalhit already uses.
	{s: "meta-externalagent", name: BrowserFacebookBot},
	{s: "meta-externalfetcher", name: BrowserFacebookBot},

	// SEO and market intelligence crawlers. High volume: on many sites these
	// two alone outweigh every AI crawler put together.
	{s: "ahrefsbot", name: BrowserAhrefsBot},
	{s: "semrushbot", name: BrowserSemrushBot},
	{s: "petalbot", name: BrowserPetalBot}, // Huawei's crawler, PetalSearch

	// generic AI and aggregator agents that carry no token of their own
	{s: "cohere-ai"},
	{s: "omgili"},
	{s: "webzio-extended"},

	// automation and headless browsers: a real engine, but nobody is watching
	{s: "headless"},
	{s: "puppeteer"},
	{s: "playwright"},
	{s: "selenium"},
	{s: "webdriver"},
	{s: "phantomjs"},
	{s: "lighthouse"},
	{s: "pagespeed"},
	{s: "gtmetrix"},

	// HTTP clients and scripting stacks. Server side only: an agent that names
	// one of these is a script, never a browser.
	//
	// Deliberately absent: java/, which J2ME feature phones announce
	// ("Java/HWJa/1.0"), and the app HTTP stacks - dalvik, cfnetwork, okhttp -
	// which carry ad requests from real apps on real devices.
	{s: "curl/"},
	{s: "wget"},
	{s: "python-requests"},
	{s: "urllib"},
	{s: "aiohttp"},
	{s: "httpx/"},
	{s: "scrapy"},
	{s: "go-http-client"},
	{s: "axios/"},
	{s: "node-fetch"},
	{s: "guzzle"},
	{s: "libwww-perl"},
	{s: "lwp::"},
	{s: "apache-httpclient"},
	{s: "httpclient"},
	{s: "restsharp"},
	{s: "postmanruntime"},
	{s: "insomnia/"},
	{s: "powershell"},
	{s: "winhttp"},
	{s: "wordpress/"},

	// What a tool calls itself. These are the -er and -ing forms on purpose: an
	// agent naming itself a scanner or a checker is describing a job, where the
	// bare verb is a word a device or an app can carry - "scan" is a document
	// scanner app, "check" could be anything, and "reader/" is BaconReader on a
	// Nexus 4.
	{s: "checker"},
	{s: "checklink"},
	{s: "sitecheck"},
	{s: "linkcheck"},
	{s: "deadlink"},
	{s: "monitoring"},
	{s: "uptime"},
	{s: "scanner"},
	{s: "scraper"},
	{s: "scrape"},
	{s: "analyzer"},
	{s: "auditor"},
	{s: "verifier"},
	{s: "inspector"},
	{s: "extractor"},
	{s: "harvester"},
	{s: "collector"},
	{s: "indexer"},
	{s: "downloader"},
	{s: "parser"},
	{s: "preview"},
	{s: "robot"},
	{s: "synthetic"},
	{s: "seo"},

	// Named tools and services whose agents carry nothing else to go on, each
	// seen repeatedly in the fixture set rather than guessed at.
	{s: "ptst"},      // WebPageTest
	{s: "ips-agent"}, // Trustwave
	{s: "statuscake"},
	{s: "binlar"},
	{s: "siteimprove"},
	{s: "hatena"},
	{s: "newspaper"}, // the Python article scraper
	{s: "feedparser"},
	{s: "nutch"},
	{s: "heritrix"},

	// link expanders, feed readers, monitors and validators
	{s: "fetcher"},
	{s: "archiver"},
	{s: "validator"},
	{s: "skypeuripreview"},
	// whatsapp/ is absent: the link preview fetcher and the messenger app on a
	// real handset both announce "WhatsApp/<version>".
	{s: "embedly"},
	{s: "vkshare"},
	{s: "datadog"},
}

// botBuckets indexes botMarkers by first byte, longest marker first, so that a
// named crawler wins over the generic token inside it.
//
// botDigrams gates the walk: it has a bit set for the first two bytes of each
// marker, so one lookup per position of the agent skips every byte pair no
// marker starts with - which is almost all of them. First bytes alone were
// too common to gate on. See digramIndex for the bit layout.
var botBuckets, botDigrams = func() (buckets [256][]botMarker, digrams [1024]uint64) {
	for _, m := range botMarkers {
		buckets[m.s[0]] = append(buckets[m.s[0]], m)
		if len(m.s) >= 2 {
			w, bit := digramIndex(m.s[0], m.s[1])
			digrams[w] |= bit
		}
	}
	for c := range buckets {
		slices.SortFunc(buckets[c], func(a, b botMarker) int { return len(b.s) - len(a.s) })
	}
	return
}()

// digramIndex says where the byte pair a,b lives in botDigrams. Conceptually
// botDigrams is one flat string of 65536 bits - one bit for every possible
// 16-bit value - packed into 1024 uint64s, and the pair itself is the bit's
// address. For 'c','r', the start of "crawl":
//
//	k    = 'c'<<8 | 'r' = 25458
//	word = k / 64       = 397    which uint64
//	bit  = k % 64       = 50     which bit of it
//
// k is unsigned, so the compiler emits the shift and mask itself; writing
// k>>6 and k&63 by hand buys nothing.
func digramIndex(a, b byte) (word int, bit uint64) {
	k := uint16(a)<<8 | uint16(b)
	return int(k / 64), 1 << (k % 64)
}

// isBotWord reports whether ua[start:end] is a name in its own right rather than
// letters inside a longer word. It exists for the "bot" marker, where the
// difference decides whether a phone is a crawler.
//
// Punctuation, a digit or the end of the agent after it is where a crawler puts
// its version or contact URL: "googlebot/2.1", "PetalBot;", "Discordbot". A
// space after it only counts when the token also starts a word, which is what
// separates the crawler writing its name as two words - "ContextAd Bot 1.0" -
// from the phone brand "CUBOT GT99". A letter after it is never a match, or the
// Italian bank "Banca Caboto" in an old IE agent reads as a crawler.
func isBotWord(ua string, start, end int) bool {
	if end == len(ua) {
		return true
	}
	if c := ua[end]; c != ' ' {
		// "_" binds like a letter: "CUBOT_POWER" is a phone, where no crawler
		// writes its name as "…bot_". The few that do - pingdom.com_bot,
		// sukibot_heritrix - are matched by their own markers instead.
		return !isLower(c) && c != '_'
	}
	return start == 0 || !isLower(ua[start-1])
}

// hasContactURL reports whether ua publishes a URL as a field of its own.
//
// Consulted only for agents that named no browser at all, where it is the
// crawler convention of pointing at a page that explains the crawl. A browser
// has no reason to carry one, and neither does an app SDK; a crawler that has
// bothered with a contact URL is announcing itself as clearly as one calling
// itself a bot.
//
// The URL has to follow a field edge. A proxy that rewrites the agent can leave
// the whole string starting with one, and that agent still belongs to whatever
// device sat behind the proxy.
func hasContactURL(ua string) bool {
	for _, m := range [...]string{"http://", "https://", "www."} {
		for i := 0; i < len(ua); {
			j := strings.Index(ua[i:], m)
			if j < 0 {
				break
			}
			at := i + j
			if at > 0 && isURLEdge(ua[at-1]) {
				return true
			}
			i = at + len(m)
		}
	}
	return false
}

// isURLEdge reports whether c can precede a contact URL: whitespace, or the
// punctuation an agent uses to open a field.
func isURLEdge(c byte) bool {
	switch c {
	case '(', ';', ',', '+', '[', '<', '"', '\'':
		return true
	}
	return isSpace(c)
}

// botSuspect reports whether ua is worth a full botName pass.
//
// botName walks the agent, and at roughly a nanosecond and a half per byte that
// is more than the browser matching it precedes - paid on every parse, for
// something almost no agent is. These five checks are each a single SIMD scan,
// an order of magnitude cheaper per byte, and every crawler that disguises
// itself as a browser trips at least one:
//
//   - '+' and "http", the contact URL conventions: "+http://example.com/bot.html"
//     and the bare URL or mailto a crawler leaves in its agent
//   - "bot", "spider", "crawl", the three names crawlers give themselves
//   - "headless", for the automation stacks, which claim none of the above
//
// A crawler wearing no browser token needs no gate: nothing below matches it
// either, so parseBrowserName's default arm runs the full pass instead. What
// this does give up is the crawler that both wears a full browser agent and
// names itself something quieter still - Chrome-Lighthouse, PTST, ips-agent.
// See doc/bots.md.
func botSuspect(ua string) bool {
	return strings.IndexByte(ua, '+') >= 0 ||
		strings.Contains(ua, "http") ||
		strings.Contains(ua, "bot") ||
		strings.Contains(ua, "spider") ||
		strings.Contains(ua, "crawl") ||
		strings.Contains(ua, "headless")
}

// botName returns the name to report for ua and whether it is a bot at all. It
// runs before any browser matching, because crawlers copy whole browser agents
// and append themselves: ClaudeBot carries a Chrome token and would otherwise
// read as a person using Chrome.
func botName(ua string) (BrowserName, bool) {
	// A named crawler wins wherever it sits, so a generic hit is remembered
	// rather than returned: "Mozilla/5.0 (seoanalyzer; compatible; bingbot/2.0)"
	// is BingBot, not an anonymous bot that happened to say "seo" first. Only
	// the generic case reads the whole agent, and only bots get that far.
	generic := false
	for i := 0; i+1 < len(ua); i++ {
		if w, bit := digramIndex(ua[i], ua[i+1]); botDigrams[w]&bit == 0 {
			continue
		}
		for _, m := range botBuckets[ua[i]] {
			if !strings.HasPrefix(ua[i:], m.s) {
				continue
			}
			if m.word && !isBotWord(ua, i, i+len(m.s)) {
				continue
			}
			// "like FeedFetcher-Google" is Feedly claiming a resemblance, the
			// same move as "like Gecko". It is still a bot, just not that one.
			if m.name != BrowserUnknown && !strings.HasSuffix(ua[:i], "like ") {
				return m.name, true
			}
			generic = true
			break
		}
	}
	if generic {
		return BrowserBot, true
	}
	return BrowserUnknown, false
}
