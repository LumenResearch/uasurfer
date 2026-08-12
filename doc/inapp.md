# In-app browsers

A tap on a link inside a native app usually renders in a frame the app controls:
no address bar, its own viewport, its own idea of when a page is visible. The
engine is Chrome's or Safari's, but the surface is not the browser's, which for
anything measuring what a person actually saw is a different thing.

These report as themselves rather than as the engine they embed:

`BrowserFacebook` (also Messenger and FB4A) · `BrowserInstagram` ·
`BrowserWeChat` · `BrowserTikTok` · `BrowserSnapchat` · `BrowserLine`

The version is the app's own, not a browser release: `BrowserInstagram` with
version 331 is Instagram 331, whatever WebKit it happens to ship.

## Chromium under another name

`BrowserVivaldi` · `BrowserWhale` · `BrowserMIUI` · `BrowserHuawei` ·
`BrowserDuckDuckGo`

Same engine as Chrome, different vendor, release cadence and defaults. Brave and
Arc are deliberately absent: both ship an agent identical to Chrome's, so no
parser can tell you it was them.

## What still reports as Chrome

An Android WebView that no app has named itself in - `; wv` and nothing else -
is reported as `BrowserChrome`, because that is exactly what it is. `Edg` and
`OPR` continue to report as `BrowserIE` and `BrowserOpera` respectively, as they
always have here.
