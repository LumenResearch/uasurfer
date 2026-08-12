[![Go](https://github.com/LumenResearch/uasurfer/actions/workflows/test.yaml/badge.svg)](https://github.com/LumenResearch/uasurfer/actions/workflows/test.yaml) [![Go Reference](https://pkg.go.dev/badge/github.com/LumenResearch/uasurfer.svg)](https://pkg.go.dev/github.com/LumenResearch/uasurfer)  [![Go Report Card](https://goreportcard.com/badge/github.com/LumenResearch/uasurfer)](https://goreportcard.com/report/github.com/LumenResearch/uasurfer)

# uasurfer

![uasurfer-100px](https://cloud.githubusercontent.com/assets/597902/16172506/9debc136-357a-11e6-90fb-c7c46f50dff0.png)

**User Agent Surfer** (uasurfer) is a lightweight Golang package that parses and abstracts [HTTP User-Agent strings](https://en.wikipedia.org/wiki/User_agent) with particular attention to device type.

The following information is returned by uasurfer from a raw HTTP User-Agent string:

| Name | Example |
|---|---|
| Browser name | `chrome` |
| Browser version | `53` |
| Platform | `ipad` |
| OS name | `ios` |
| OS version | `10` |
| Device type | `tablet` |
| Bot | `false` |

Layout engine, browser language, device brand and device model are not parsed.

Mainstream desktop and mobile are effectively solved. Two caveats are worth
knowing before relying on a field:

* **Android phone versus tablet** rests on the `Mobile` token, which every
  browser since Android 4 states on a phone and omits on a tablet. Tablets that
  state it anyway are caught by model, and handsets from before the convention
  are read by version. Where the agent is a native app rather than a browser it
  states no form factor at all and reads as a phone; the
  `Sec-CH-UA-Form-Factors` hint settles it outright.
* **Bot detection** trades the long tail for a name table nobody has to
  maintain: it catches 83% of a 1,975 agent crawler fixture set with no false
  positive across 2,906 real device agents. See [doc/bots.md](doc/bots.md).
* **Chrome's user agent reduction** hollowed out several fields at the source.
  Chromium browsers report a frozen `10.15.7` for macOS, NT 10.0 for both
  Windows 10 and 11, `Android 10` on every Android device, and `0` for every
  version component below the major. None of that is recoverable from the agent;
  the client hints carry it instead.

## Usage

### Parse(ua string) Function

The `Parse()` function accepts a user agent `string` and returns a `*UserAgent`, with named constants for the browser, OS, platform and device, and integers for versions (major, minor and patch separately). A string can be retrieved by adding `.String()` to a variable, such as `uasurfer.BrowserName.String()`, or `.StringTrimPrefix()` to drop the type name.

```go
// Define a user agent string
myUA := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/45.0.2454.85 Safari/537.36"

ua := uasurfer.Parse(myUA)
```

For a request loop, `ParseUserAgent(ua string, dest *UserAgent)` fills a
`UserAgent` the caller owns; call `Reset()` before reusing one. `ParseWithHints`
takes a `Hints` struct carrying what the agent cannot say for itself: the screen
size, which is how an iPad in desktop mode is told apart from a Mac, and the
`Sec-CH-UA-*` client hints, which are where Chromium now puts the device type
and OS version it has stopped stating in the agent. See
[doc/hints.md](doc/hints.md).

where example UserAgent is:
```
{
    Browser {
        BrowserName: BrowserChrome,
        Version: {
            Major: 45,
            Minor: 0,
            Patch: 2454,
        },
    },
    OS {
        Platform: PlatformMac,
        Name: OSMacOSX,
        Version: {
            Major: 10,
            Minor: 10,
            Patch: 5,
        },
    },
    DeviceType: DeviceComputer,
}
```

**Usage note:** There are some OSes that do not return a version, see docs below. Linux is typically not reported with a specific Linux distro name or version.

#### Constants

The full lists of `BrowserName`, `OSName`, `Platform` and `DeviceType` constants
live in the package reference:
[`BrowserName`](https://pkg.go.dev/github.com/LumenResearch/uasurfer#BrowserName),
[`OSName`](https://pkg.go.dev/github.com/LumenResearch/uasurfer#OSName),
[`Platform`](https://pkg.go.dev/github.com/LumenResearch/uasurfer#Platform),
[`DeviceType`](https://pkg.go.dev/github.com/LumenResearch/uasurfer#DeviceType).
Grouping worth knowing about:

* `BrowserChrome` covers Chromium and Android WebView 4.4 or newer.
* `BrowserIE` covers Edge as well, including Chromium Edge; version >= 79 is how
  you tell them apart.
* `BrowserFirefox` covers IceCat, Iceweasel and SeaMonkey.
* `BrowserSafari` covers the Google Search app on iOS, which is Safari in
  disguise.
* Crawlers have their own constants and their own document: see
  [doc/bots.md](doc/bots.md); in-app webviews and the Chromium browsers that are
  not Chrome have [doc/inapp.md](doc/inapp.md).

#### Browser Version

Browser version is a `Version{Major, Minor, Patch}` of ints. For example Chrome
45.0.23423 gives `{45, 0, 23423}`, so `ua.Browser.Version.Major > 23` is the
usual test. Versions compare with `Less`: `if ver1.Less(ver2) {}`.

An unknown version is `{0, 0, 0}`. Chromium browsers now report `0` for
everything below the major version, whatever they are actually running.

#### OS Version

OS X major version is always 10 for releases prior to Big Sur with consecutive minor versions indicating releases (10 - Yosemite, 11 - El Capitan, 12 Sierra, etc). macOS Big Sur is indicated as `{11, 1, 0}`, though Chrome and Firefox freeze what they report at `10.15.7` whatever the machine runs. Windows version is the NT version. `Version{0, 0, 0}` means the version is unknown or not evaluated.
Versions can be compared using `Less` function: `if ver1.Less(ver2) {}`

Here are some examples across the platform, os.name, and os.version:

* For Windows XP (Windows NT 5.1), "`PlatformWindows`" is the platform, "`OSWindows`" is the name, and `{5, 1, 0}` the version.
* For OS X 10.5.1, "`PlatformMac`" is the platform, "`OSMacOSX`" the name, and `{10, 5, 1}` the version.
* For Android 5.1, "`PlatformLinux`" is the platform, "`OSAndroid`" is the name, and `{5, 1, 0}` the version.
* For iOS 5.1, "`PlatformiPhone`" or "`PlatformiPad`" is the platform, "`OSiOS`" is the name, and `{5, 1, 0}` the version.

###### Windows Version Guide

* Windows 10 - `{10, 0, 0}`
* Windows 8.1 - `{6, 3, 0}`
* Windows 8 - `{6, 2, 0}`
* Windows 7 - `{6, 1, 0}`
* Windows Vista - `{6, 0, 0}`
* Windows XP - `{5, 1, 0}` or `{5, 2, 0}`
* Windows 2000 - `{5, 0, 0}`

Windows 95, 98, and ME represent 0.01% of traffic worldwide and are not available through this package at this time.

#### DeviceType
DeviceType is typically quite accurate, though determining between phones and tablets on Android is not always possible due to how some vendors design their UA strings. A mobile Android device without a tablet indicator defaults to being classified as a phone.

`DeviceTV` covers the major TV brands and the streaming sticks and boxes from
Apple, Google, Roku and Amazon, most of which also report an OS of their own:
see [doc/tv.md](doc/tv.md).

## Example Combinations of Attributes
* Surface RT -> `OSWindows8`, `DeviceTablet`, OSVersion >= `6`
* Android Tablet -> `OSAndroid`, `DeviceTablet`
* Microsoft Edge -> `BrowserIE`, BrowserVersion >= `12.0.0`

## Deliberately not supported

Device brand and model, layout engine, CPU architecture and browser language.
Brave and Arc are not named either, because they ship an agent identical to
Chrome's on purpose.

## Adding new user agents

1. Source user agent strings which identify a device type, system or a browser you want to add
2. Identify a unique part of the user agent string which identifies a device
3. Add a condition to a switch statement inside `browser.go`, `device.go` or `system.go`
4. Add rows to the fixture sets in `testdata/` that fail without the change

For example, to identify a Google TV user agent as device type TV, we identify that all user agents contain "googletv" string and we add `strings.Contains(ua, "googletv")` to the `device.go` switch condition for identifying TVs.

There is a Makefile for all of it:

```sh
make all         # go fix, goimports, lint, test: everything CI enforces
make benchstat   # compare the parse path against master
make benchguard  # the same comparison, as CI grades it
```

Anything touching the parse path wants a `make benchstat` in the PR description.
It measures master and your tree and compares them through
[benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat), which knows the
difference between a real change and a busy laptop. A percentage on its own is
not a blocker - a parse is a few hundred nanoseconds, and spending some of that
budget on accuracy is a fair trade - but an unexplained one is. CI fails only
when a benchmark is slower by both a third and a few hundred nanoseconds, or when
a parse allocates more than the single copy it is allowed. `CLAUDE.md` has the
conventions in full.