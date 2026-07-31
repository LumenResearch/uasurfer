# Client hints

Chromium has spent several releases removing information from the user agent
string. What it took out it now offers in headers instead, and `ParseWithHints`
reads them:

```go
ua := uasurfer.ParseWithHints(r.UserAgent(), &uasurfer.Hints{
    Mobile:          r.Header.Get("Sec-CH-UA-Mobile"),
    Platform:        r.Header.Get("Sec-CH-UA-Platform"),
    PlatformVersion: r.Header.Get("Sec-CH-UA-Platform-Version"),
    FormFactors:     r.Header.Get("Sec-CH-UA-Form-Factors"),
})
```

Pass the header values as they arrive, quotes and all. Every field is optional,
and an empty one changes nothing.

## What each one settles

| hint                         | what it fixes                                                                         |
| ---------------------------- | ------------------------------------------------------------------------------------- |
| `Sec-CH-UA-Form-Factors`     | the device type outright: `Desktop`, `Mobile`, `Tablet`, `Watch`                      |
| `Sec-CH-UA-Mobile`           | phone against tablet, the one thing an Android agent is worst at                      |
| `Sec-CH-UA-Platform-Version` | the OS version on macOS, Android and ChromeOS, where the agent reports a frozen value |
| `ScreenSize`                 | an iPad in desktop mode, which no header distinguishes from a Mac                     |

Chromium sends `Sec-CH-UA-Mobile` and `Sec-CH-UA-Platform` by default. The rest
arrive only if the site asks for them:

```
Accept-CH: Sec-CH-UA-Platform-Version, Sec-CH-UA-Form-Factors
```

`Sec-CH-UA-Form-Factors` also has `Automotive`, `XR` and `EInk` values, which
have no constant here. An agent stating one keeps whatever the agent itself
parsed as. A television keeps `DeviceTV` regardless: no form factor value
describes one, and the agent identifies it perfectly well.

## Windows

Left deliberately alone. `OS.Version` is the NT version on Windows, and this
hint counts differently — 13 or more means Windows 11, which NT numbering cannot
express, since both 10 and 11 report NT 10.0. A caller that needs the difference
has to read the header itself.

## Safari

Sends none of this. The screen size hint is the only thing that helps there, and
it is what tells an iPad in desktop mode from a Mac.
