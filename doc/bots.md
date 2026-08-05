# Bot detection

`IsBot()` reports whether an agent belongs to a crawler, a script or an
automation stack rather than to a person. When it does, the OS, platform and
device fields are set to `OSBot`, `PlatformBot` and `DeviceComputer`, so a caller
filtering on any one of them sees the same answer.

```go
ua := uasurfer.Parse(agent)
if ua.IsBot() {
    return // not a person
}
```

## What it catches

Crawlers announce themselves conventionally, and detection is built on those
conventions rather than on a list of names: anything calling itself `…bot`,
`…spider` or `…crawler`, anything publishing a contact URL as `+http…`, the
HTTP client libraries (curl, python-requests, Go-http-client, Scrapy, …), the
headless and automation browsers, and the link preview fetchers.

The crawlers with the volume to be worth naming get their own constant -
`BrowserGoogleBot`, `BrowserBingBot`, `BrowserOpenAIBot`, `BrowserAnthropicBot`,
`BrowserPerplexityBot`, `BrowserAmazonBot`, `BrowserBytedanceBot`,
`BrowserAhrefsBot`, `BrowserSemrushBot` and the rest of the block in
[`BrowserName`](https://pkg.go.dev/github.com/LumenResearch/uasurfer#BrowserName).
Everything else reports as `BrowserBot`. A named constant covers the whole fleet
behind it: `BrowserGoogleBot` is AdsBot, Mediapartners, GoogleOther and
Google-InspectionTool as well as Googlebot itself.

An agent that names no browser at all but publishes a contact URL is also taken
as a crawler. Nothing that reaches that point is a browser, and a crawler that
has bothered to link a page explaining itself is announcing itself as clearly as
one calling itself a bot.

## What it does not catch

Roughly one crawler in six, by count of distinct agents: the ones that wear a
complete browser agent and then name themselves something unguessable, with no
URL and no generic token - `Datanyze`, `Rigor`, `Scope3/2.0`, `binlar`. Catching
those means shipping a list of every crawler name in existence and keeping it
current, which this package deliberately does not.

Measured against 1,975 real crawler agents, detection sits at **83%**, with no
false positive across the 2,906 real device agents in the device fixture set (nor
across the 17k agents it was sampled from). Both fixture sets are in `testdata/`.

Impostors are named by what they are, not what they claim: `TelegramBot (like
TwitterBot)` and `Feedly/1.0 (… like FeedFetcher-Google)` are `BrowserBot`, since
"like X" is a claim of resemblance rather than identity.

Some agents are non-browser but not bots, and are reported as the devices they
are: app HTTP stacks such as Dalvik, CFNetwork and OkHttp carry requests from
real apps on real phones, and `Java/…` is what a J2ME feature phone announces.
