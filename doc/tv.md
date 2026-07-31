# Connected TV

Televisions, sticks and set top boxes report `DeviceTV`. Where the platform
states an OS of its own, that is reported too.

| platform | OS name | platform value | OS version |
|---|---|---|---|
| Samsung smart TVs | `OSTizen` | `PlatformLinux` | yes |
| LG smart TVs | `OSWebOS` | `PlatformLinux` | not stated by the device |
| Roku players | `OSRoku` | `PlatformLinux` | yes |
| Apple TV | `OSTvOS` | `PlatformAppleTV` | yes |
| Amazon Fire TV | `OSAndroid` | `PlatformLinux` | the Android version |
| Android TV, Google TV, Chromecast | `OSAndroid` | `PlatformLinux` | the Android version |
| Vizio, Hisense VIDAA, HbbTV boxes, RDK, NetCast, Viera, Bravia | `OSLinux` or `OSUnknown` | `PlatformLinux` | not stated by the device |

The last row is not an omission: those platforms are Linux and state no version
of their own, so a constant per brand would buy a name and nothing else. Use
`DeviceTV` to know it is a television, and the browser and its version for what
the device can actually render.

Amazon's Fire TV models are recognised as a family, so a stick released after
this version of the library still reports `DeviceTV`. Fire tablets remain
`DeviceTablet`.

Two notes that look like bugs and are not:

* LG spells its own OS **`Web0S`**, with a zero, on everything since webOS 3.
  All three spellings map to `OSWebOS`.
* The Roku and Apple TV **remote control apps** run on a phone, and report as
  one. Only the players and the boxes report `DeviceTV`.
