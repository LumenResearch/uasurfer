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

`BrowserBot`..`BrowserYahooBot` must stay a contiguous block at the end of the
`BrowserName` list — `IsBot` is a range check over it.

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
- Benchmark before and after any hot-path change:
  `go test -run=XXX -bench=Parse -benchmem -count=6`. Compare against a worktree
  at the base commit; report medians honestly, including regressions.

## Tests

- A test lives in the `_test.go` file mirroring the implementation file it
  covers. `device.go` → `device_test.go`. Never create catch-all test files.
- Non-trivial logic leaves a runnable check behind. When hand rolling a
  replacement for something standard, keep the original as a test oracle and
  compare differentially — see `TestIsAmazonFire` against the regexp it replaced.
- When correctness depends on an invariant, assert the invariant, not just the
  behaviour: `TestIPadScreenSizesAreLandscape` guards the table ordering that
  `isiPad` relies on.

### Enum terminators

Each `iota` block ends in an unexported terminator (`_deviceTypeFinal` and
friends). Because it shifts whenever a constant is added, tests can enumerate the
lists at runtime — which Go constants otherwise do not allow. Keep it last.

This is why there is no `stringer`: `String()` methods are hand written as
explicit switches next to their constants, and `assertNamed` compares each list
against a golden set of names. Adding a constant means adding its `case` and
extending that list.

The terminator is doing the real work here. `String()` has no numeric fallback —
an out-of-range value reads as the type's Unknown name, which is the useful
answer for a value cast from a newer release, but it also means a constant
missing its `case` is invisible from its output alone. `assertNamed` catches it
by length: the terminator moved, so the list no longer matches.

## Tooling

Everything below must be clean before work is considered done:

```sh
gofmt -l .
go vet ./...
golangci-lint run ./...   # CI runs this via bsm/misc; it fails the build
go fix -diff ./...        # Go 1.26 modernizers, run as standard
go test ./...
```

## Workflow

- Do not weaken a test to make it pass.
