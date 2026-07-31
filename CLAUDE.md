# uasurfer

Parses HTTP User-Agent strings into a `UserAgent` (browser, OS, device type).
The philosophy is to identify technologies holding >1% market share and to spend
neither cycles nor accuracy guessing at esoteric agents.

## Goals, in priority order

1. **Correct** — a wrong classification is worse than an unknown one.
2. **Fast** — this runs per request in ad-serving hot paths. Substring scans over
   the agent are the cost; the parser is judged in nanoseconds per parse.
3. **Consistent** — one name per concept, one verb per kind of operation.
4. **Concise** — prefer deleting to adding. The shortest version that stays clear.
5. **Stable** — the exported API is depended on downstream. See below.

## Public API is frozen

Exported names, their signatures and their semantics do not change without an
explicit decision. Constants are append-only and must keep their values: callers
persist them as ints. Internal naming, file layout and implementation are free to
change.

The bot constants must stay a contiguous block at the **end** of the
`BrowserName` list: `IsBot` is a range check from `BrowserBot` to the terminator,
so a non-bot appended after them silently becomes a bot.
`TestBrowserNameStrings` pins the last one by name and fails if that happens.

## Naming

Three names for three concepts, used everywhere:

| name | meaning |
|---|---|
| `ua` | the whole normalised (lowercased) agent |
| `agentPlatform` | the text inside the first parentheses |
| `specs` | the first `;`-delimited field of `agentPlatform` |

Verb prefixes:

- `parse*` — reads the agent (or a slice of it) and **sets fields on the
  receiver**: `parseOS`, `parseDevice`, `parseiOSVersion`, `Version.parseAfter`.
  Anything that mutates while reading the agent belongs here.
- `is*` — pure predicate, no mutation: `isTV`, `isAmazonFire`, `isiPad`.
- Anything else returns a value and names what it returns: `landscape`,
  `darwinToIOS`, `normalise`, `copyLower`.

Do not use `get*` for something that mutates, or `eval*` as a synonym for parse.
`applyBotDefaults` is deliberately outside the `parse*` family: it reads
already-parsed fields rather than the agent.

## Receivers

Every type here is a handful of ints, cheap to copy. So:

- **A pointer receiver means the method mutates.** `parse*`, `Reset`,
  `applyBotDefaults`.
- **Everything else takes a value receiver**, `is*` predicates included:
  `ScreenSize.isiPad`, `ScreenSize.landscape`, `Version.Less`, the `String`
  methods. Do not reach for a pointer just to avoid a copy at this size, and do
  not use one so a method can absorb a nil check — handle nil at the call site,
  as `parseMacintosh` and `parseAppleNative` do for `hints.ScreenSize`.

`UserAgent.IsBot` is the one exception: it does not mutate but keeps its pointer
receiver because it is exported and the API is frozen.

## Bot markers

`botMarkers` in `bot.go` is deliberately not a list of crawler names: the generic
tokens (`bot` as a word, `spider`, `crawl`, `+http`) carry most of the recall, and
a name is listed only when it is reported through its own constant or is too
quiet to carry a generic token.

A marker must be one **no real device can carry**. This is the whole discipline
of the file, and it is easy to get wrong: `java/` looks like a server-side stack
and is what J2ME feature phones announce; `whatsapp/` is the messenger as often
as the link fetcher; `dalvik`, `cfnetwork` and `okhttp` carry real ad requests
from real apps. `bot` needs its word check or the phone brand CUBOT matches.

A marker that must be found in a browser-shaped agent also has to trip
`botSuspect`, the cheap gate deciding who pays for the full scan; otherwise it is
only ever reached through `parseBrowserName`'s default arm.

## Performance

- **Avoid allocating where you reasonably can.** A parse currently costs one
  allocation, the lowercased copy in `normalise`. That is a target to beat, not a
  hard cap: add an allocation when it genuinely buys correctness or clarity, but
  know that you are adding it and say why. Watch the count with `-benchmem` so a
  change never moves it by accident.
- Nothing may cause the agent string to escape to the heap. `regexp` does — that
  is why `isAmazonFire` is hand rolled. Verify with `go build -gcflags='-m'`.
- Long `strings.Contains` chains over the whole agent are O(markers × len). When
  a list grows past a handful, bucket by first byte and scan once, as `isTV`
  does.
- Benchmark before and after any hot-path change with `make benchstat`, which
  measures master and the working tree and compares them with benchstat. Report
  what it says, including regressions. A percentage is not automatically a
  blocker: a parse is a few hundred nanoseconds, and buying accuracy with some of
  that budget is a legitimate trade - saying so out loud is the requirement.
- `make benchguard` is the gate CI runs, on a pull request, measuring the base
  commit and the change on the same runner. It fails a benchmark only when it is
  slower by **both** a third and a few hundred nanoseconds, and fails any increase
  in allocations outright. Needing both keeps a 15ns check growing to 85ns from
  failing a parse that still costs half a microsecond, while catching the change
  of shape - a compiled regexp, a scan gone quadratic. Do not chase percentages
  below that bar; do explain the ones above it.

## Tests

- **New non-trivial code ships with a test.** A branch, a loop, a table, a
  parser, a boundary rule: if it can be wrong, something has to fail when it is.
  Trivial one-liners and pure renames do not need one. This is not negotiable
  per-change — it is the reason a twenty year old parser can still be refactored
  at all.
- A test lives in the `_test.go` file mirroring the implementation file it
  covers. `device.go` → `device_test.go`. Never create catch-all test files. Two
  deliberate exceptions: **benchmarks** all live in `bench_test.go`, because the
  parse path is one budget and comparing a change means running the whole set,
  and a **helper used by more than one test file** lives in `uasurfer_test.go`
  rather than beside its first caller.
- Test the exported API. Reach for a private function only when it is
  package-level, non-trivial, and not already covered through `Parse` —
  `isFireTV` and `botName` earn a test, the byte predicates they call do not.
  Testing a marker table entry by entry is noise; testing the behaviour the table
  exists for is not.
- Non-trivial logic leaves a runnable check behind. When hand rolling a
  replacement for something standard, keep the original as a test oracle and
  compare differentially — see `TestIsAmazonFire` against the regexp it replaced.
- When correctness depends on an invariant, assert the invariant, not just the
  behaviour: `TestIPadScreenSizesAreLandscape` guards the table ordering that
  `isiPad` relies on, and `TestBotMarkersAreIndexed` guards the index `botName`
  reads, where a mis-bucketed marker is simply never tried.
- Anything that walks untrusted input by index belongs in `FuzzParse`'s
  properties rather than in another table of hand-picked strings. CI fuzzes for a
  minute per run; a crash found there is written to `testdata/fuzz` and becomes a
  permanent seed.

### Fixtures

Fixture sets of real agents live in `testdata/` and are documented in
[doc/fixtures.md](doc/fixtures.md). Two rules matter when touching them:

- A row is a claim about what an agent *means*, not a snapshot of current output.
  When a change makes one fail, work out which of the two is wrong first. Test
  suites that were regenerated to match a bug are how parsers rot.
- Sources must be permissively licensed (MIT, BSD, Apache-2.0) and credited in
  `testdata/NOTICE.md`. Take agent strings, never another parser's labels: those
  are its judgement in its categories, and copying them imports both.
- `botRecallFloor` in `bot_test.go` is the share of the bot fixture set detected.
  Raise it when detection improves; lowering it to make a change pass is how a
  detector rots.

### Enum terminators

Each `iota` block ends in an unexported terminator (`_deviceTypeFinal` and
friends). Because it shifts whenever a constant is added, tests can enumerate the
lists at runtime — which Go constants otherwise do not allow. Keep it last.

This is why there is no `stringer`: `String()` methods are hand written as
explicit switches next to their constants. Adding a constant means adding its
`case` — nothing else, as the tests do not pin the exact names.

`assertNames` checks that the names are **unique** and carry the type's prefix,
not that they match a golden list. Uniqueness is what makes that sufficient:
`String()` has no numeric fallback, so a constant with no `case` falls into the
default arm and repeats the Unknown name, and the duplicate gives it away. The
same check catches two cases returning the same name by copy-paste.

## Documentation

`README.md` stays short: what the package is, the API, the honest caveats, and
links out. Lists and special topics live in `doc/<topic>.md` — bot detection,
connected TV, in-app browsers, client hints. Constant lists are not duplicated
anywhere: they link to pkg.go.dev. A new constant goes in the relevant
`doc/` file, not the README.

## Tooling

Everything below must be clean before work is considered done:

```sh
make all   # go fix, goimports, vet, golangci-lint, tests
```

Which is:

```sh
go fix ./...              # Go 1.26 modernizers, run as standard
goimports -l .            # not gofmt: it also fixes the imports an edit moved
go vet ./...
golangci-lint run ./...   # CI runs this; it fails the build
go test ./...
```

## Workflow

- Do not weaken a test to make it pass.
