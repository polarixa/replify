# translit

Package `translit` transliterates Unicode text into plain 7-bit ASCII.

```go
import "github.com/polarixa/replify/pkg/translit"
```

`translit` is a from-scratch, performance-oriented reimplementation of
[`github.com/mozillazg/go-unidecode`](https://github.com/mozillazg/go-unidecode).
It produces **byte-for-byte identical output** to the original library while
eliminating all heap allocations on the hot path and running **2.4x-10x faster**
across every script category.

---

## Features

- **Zero allocations** on the `Append`/`AppendBytes` path — no `strings.Builder`,
  no `bytes.Buffer`, no intermediate conversions.
- **O(1) per-rune lookup** — four flat array reads into immutable, compile-time
  tables; no map hashing, no pointer chasing through scattered heap objects.
- **ASCII fast path** — bytes below U+007F are identified with a single comparison
  and appended directly, with no table lookup.
- **Exact pre-sizing** via `SizeHint`, enabling a guaranteed zero-reallocation
  transliteration with a single `make` up front.
- **Full concurrency safety** — all package-level state is immutable after
  compilation; no mutex, no `sync.Once`, no locking of any kind is required.
- **Compatible** — produces byte-for-byte identical output to the original
  library across every legal Unicode codepoint, verified by exhaustive golden
  testing of all 1,114,112 codepoints.

---

## Installation

```bash
go get github.com/polarixa/replify/pkg/translit
```

---

## Quick start

```go
// Simplest usage: transliterate a string.
result := translit.Unidecode("kožušček") // "kozuscek"
result  = translit.Unidecode("北京")      // "Bei Jing "
result  = translit.Unidecode("Москва")   // "Moskva"

// Zero-allocation path using a caller-supplied buffer.
buf := make([]byte, 0, translit.SizeHint(src))
buf  = translit.Append(buf, src)

// Reuse the same buffer across multiple calls.
var buf []byte
for _, s := range inputs {
    buf = translit.Append(buf[:0], s)
    process(buf)
}
```

---

## API

### Modes

`Mode` controls how unmapped codepoints — those outside the table's covered range
or assigned no replacement — are handled.

| Constant   | Behavior                                                                                         |
| ---------- | ------------------------------------------------------------------------------------------------ |
| `ModeSkip` | Drop unmapped codepoints (default). Matches the original library's behavior.                     |
| `ModeKeep` | Pass unmapped codepoints through, re-encoded as UTF-8. Malformed bytes are emitted as raw bytes. |

> **Note:** Codepoints whose table entry is explicitly empty (combining marks and
> similar) are always dropped regardless of mode, consistent with the original library.

### Functions

#### Transliteration

```go
func Append(dst []byte, src string) []byte
```

Transliterates `src` and appends the result to `dst`. Uses `ModeSkip`. Never
allocates when `dst` has sufficient spare capacity.

```go
func AppendBytes(dst []byte, src []byte) []byte
```

`Append` for a `[]byte` source. Avoids a string conversion when the caller
already holds the input as bytes.

```go
func AppendMode(dst []byte, src string, mode Mode) []byte
```

`Append` with an explicit `Mode` for unmapped codepoints.

```go
func AppendBytesMode(dst []byte, src []byte, mode Mode) []byte
```

`AppendBytes` with an explicit `Mode` for unmapped codepoints.

#### Pre-sizing

```go
func SizeHint(src string) int
```

Returns the exact number of bytes `Append(nil, src)` would produce. Use it to
pre-size `dst` for a guaranteed zero-allocation call:

```go
buf := make([]byte, 0, translit.SizeHint(src))
buf  = translit.Append(buf, src)
```

```go
func SizeHintMode(src string, mode Mode) int
```

`SizeHint` for a non-default `Mode`.

#### Convenience

```go
func Unidecode(s string) string
```

Transliterates `s` and returns a new string. Matches the original library's API.
Performs exactly one allocation, sized precisely via `SizeHint`. For
high-throughput paths where a `[]byte` result is acceptable, prefer `Append`
with a reused buffer.

#### Validation

```go
func ValidUTF8(src string) bool
```

Reports whether `src` is valid UTF-8. Allocation-free wrapper around
`utf8.ValidString`. Passing invalid UTF-8 to `Append` is safe — malformed bytes
are handled gracefully — but callers that need to surface encoding errors before
transliterating can do so here.

---

## Behavior notes

- **ASCII identity**: bytes U+0000-U+007E pass through unchanged.
- **DEL (U+007F)**: dropped in `ModeSkip`, matching the original library's
  `unicode.MaxASCII` fast-path boundary.
- **Invalid UTF-8**: malformed bytes are skipped in `ModeSkip` and emitted as
  raw bytes in `ModeKeep`.
- **Covered range**: the transliteration table covers up to U+EFFFF. Codepoints
  above that threshold follow the active `Mode`.

---

## Performance

Measured on an Intel Xeon @ 2.10 GHz, `GOMAXPROCS=1`, Go 1.22,
`-benchtime=1s`. Full raw output is in `compare/`; reproduce with:

```bash
go test ./compare/... -bench . -benchmem
```

| Input category | Original ns/op | translit ns/op |   Speedup | Original allocs/op | translit allocs/op |
| -------------- | -------------: | -------------: | --------: | -----------------: | -----------------: |
| ASCII          |          772.8 |           77.5 | **10.0x** |                  5 |              **0** |
| Latin-extended |          865.4 |          281.5 |  **3.1x** |                  5 |              **0** |
| Vietnamese     |          868.8 |          252.5 |  **3.4x** |                  5 |              **0** |
| Chinese        |          703.2 |          266.9 |  **2.6x** |                  5 |              **0** |
| Japanese       |          745.7 |          313.0 |  **2.4x** |                  5 |              **0** |
| Korean         |          774.1 |          252.2 |  **3.1x** |                  5 |              **0** |
| Mixed scripts  |          959.9 |          295.7 |  **3.2x** |                  5 |              **0** |
| 1 KB input     |           7384 |           2563 |  **2.9x** |                  8 |              **0** |
| 10 KB input    |          76434 |          31530 |  **2.4x** |                 15 |              **0** |
| 100 KB input   |         747018 |         247176 |  **3.0x** |                 23 |              **0** |

The `Append`/`AppendBytes` hot path is `0 B/op, 0 allocs/op` for every category,
confirmed both by benchmark output and by `TestZeroAllocations`, which uses
`testing.AllocsPerRun` and fails the build if this ever regresses.

The `Unidecode` convenience wrapper beats the original by **1.5x-3.2x** while
using 2 allocations instead of 5-23, because it sizes its single allocation
exactly via `SizeHint` rather than growing a `strings.Builder` repeatedly.

---

## How it works

The original library stores transliteration data as a `map[rune][]string` built
at `init()` time. Resolving one rune requires: hash computation, bucket probing,
a pointer to a `[]string`, an index into that slice, and a pointer to the
replacement bytes — four dependent pointer dereferences scattered across the heap.

`translit` stores the same data in four flat, contiguous, read-only arrays
generated at compile time:

```
data        — 157 KB  flat byte string: all replacement text concatenated
sectionStart — 503 × int32:  section → start index in entryOff/entryLen, or −1
sectionLen   — 503 × uint16: section → number of valid positions in this section
entryOff     — 48,597 × uint32: entry → byte offset into data
entryLen     — 48,597 × uint8:  entry → byte length within data
```

A lookup for rune `r` is four array reads and one string slice — O(1), cache-
friendly, and completely branch-free in the common in-range case. ASCII bytes
never touch the table at all.

See [ARCHITECTURE.md](ARCHITECTURE.md) for the full write-up: data layout
rationale, complexity analysis, cache-locality analysis, and a line-by-line
comparison against the original implementation.

---

## Correctness

Output is verified to be byte-for-byte identical to `github.com/mozillazg/go-unidecode`
by `TestGoldenEveryRune` in `compare/`, which iterates all 1,114,112 legal
Unicode codepoints and asserts agreement on each one. Property-based tests
(`TestPropertyConcatenation`, `TestPropertyASCIISubsetIsIdentity`) and a
continuous fuzz target (`FuzzAppend`) provide additional coverage.

---

## Thread safety

All package-level state is declared as compile-time array literals with no
`init()`-time mutation. Every exported function reads only immutable memory and
writes only to the caller-supplied `dst` slice. There is no shared mutable
state; no mutex, `sync.Once`, or any other synchronization primitive is used
or needed. `go test -race` passes across the full suite.

---

## Code generation

The lookup tables in `tables_gen.go` are produced by the generator in `gen/`.
Do not hand-edit `tables_gen.go`. To regenerate after modifying the source data:

```bash
go generate ./pkg/translit/...
```
