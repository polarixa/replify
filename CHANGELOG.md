# Changelog

All notable changes to the replify ecosystem are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Releases are cut from `master` and tagged `vMAJOR.MINOR.PATCH`. Pre-releases use
`vMAJOR.MINOR.PATCH-beta.N` or `vMAJOR.MINOR.PATCH-rc.N`; the version scheme is
documented in [`docs/RELEASE.md`](docs/RELEASE.md).

**Breaking changes** are marked with a bold `Breaking:` prefix and always state what
changed, who is affected, and the required migration. A release entry that only
reorganises files internally is not marked as breaking.

<details>
<summary><strong>Version index</strong></summary>

- **v0.1.x** — [v0.1.56](#v0156---2026-10-21), [v0.1.55](#v0155---2026-10-16), [v0.1.54](#v0154---2026-10-14), [v0.1.53](#v0153---2026-10-05), [v0.1.52](#v0152---2026-10-05), [v0.1.51](#v0151---2026-10-04), [v0.1.50](#v0150---2026-10-03), [v0.1.49](#v0149---2026-10-02), [v0.1.48](#v0148---2026-10-01), [v0.1.47](#v0147---2026-09-29), [v0.1.46](#v0146---2026-09-19), [v0.1.45](#v0145---2026-09-18), [v0.1.44](#v0144---2026-09-07), [v0.1.43](#v0143---2026-09-06), [v0.1.42](#v0142---2026-09-05), [v0.1.41](#v0141---2026-09-02), [v0.1.40](#v0140---2026-09-01), [v0.1.39](#v0139---2026-08-31), [v0.1.38](#v0138---2026-08-31), [v0.1.37](#v0137---2026-08-22), [v0.1.36](#v0136---2026-08-19), [v0.1.35](#v0135---2026-08-19), [v0.1.34](#v0134---2026-08-16), [v0.1.33](#v0133---2026-08-15), [v0.1.32](#v0132---2026-08-14), [v0.1.31](#v0131---2026-08-12), [v0.1.30](#v0130---2026-08-10), [v0.1.29](#v0129---2026-08-09), [v0.1.28](#v0128---2026-07-27), [v0.1.27](#v0127---2026-07-25), [v0.1.26](#v0126---2026-07-24), [v0.1.25](#v0125---2026-07-22), [v0.1.24](#v0124---2026-06-14), [v0.1.23](#v0123---2026-05-29), [v0.1.22](#v0122---2026-05-29), [v0.1.21](#v0121---2026-05-29), [v0.1.20](#v0120---2026-05-27), [v0.1.19](#v0119---2026-05-24), [v0.1.18](#v0118---2026-05-02), [v0.1.17](#v0117---2026-04-18), [v0.1.16](#v0116---2026-04-18), [v0.1.15](#v0115---2026-04-16), [v0.1.14](#v0114---2026-04-15), [v0.1.13-rc.1](#v0113-rc1---2026-04-15), [v0.1.12](#v0112---2026-04-12), [v0.1.12-rc.2](#v0112-rc2---2026-04-12), [v0.1.12-rc.1](#v0112-rc1---2026-04-12), [v0.1.11](#v0111---2026-04-11), [v0.1.10](#v0110---2026-04-05), [v0.1.9](#v019---2026-04-05), [v0.1.8](#v018---2026-04-05), [v0.1.7](#v017---2026-04-05), [v0.1.6](#v016---2026-04-04), [v0.1.5](#v015---2026-04-03), [v0.1.4](#v014---2026-04-03), [v0.1.2](#v012---2026-04-02), [v0.1.1](#v011---2026-03-30), [v0.1.0](#v010---2026-03-09)
- **v0.0.x** — [v0.0.9](#v009---2026-03-08), [v0.0.8](#v008---2026-03-06), [v0.0.7](#v007---2026-03-04), [v0.0.6](#v006---2026-03-02), [v0.0.5](#v005---2026-03-02), [v0.0.5-beta.1](#v005-beta1---2026-02-28), [v0.0.4](#v004---2026-02-28), [v0.0.4-beta.1](#v004-beta1---2026-02-26), [v0.0.3](#v003---2026-02-26), [v0.0.2](#v002---2026-02-20), [v0.0.1](#v001---2026-01-29)

</details>

## [Unreleased]

- No unreleased changes. The working tree currently matches `v0.1.56`.

## [v0.1.56] - 2026-10-21

### Added

- Added the `WithMetaRequestIDGenerated(opts ...codegen.Option)` builder method to generate a
  cryptographically secure request ID through `pkg/codegen` and store it in the response
  metadata. The default is a 16-character alphanumeric string; `codegen.WithLength`,
  `codegen.WithCharset` and `codegen.WithPrefix` customise it. Users no longer need to call
  `randn` by hand for traceable request identifiers.

## [v0.1.55] - 2026-10-16

### Added

- Added `pkg/mapbuilder`, a fluent builder for nested `map[string]any` structures. It provides
  `New`, `From`, `Add`, `AddPath`, `AddIfAbsent`, `AddPathIfAbsent`, `Append`, `AppendPath`,
  `Merge`, `MergePath`, `Remove`, `RemovePath`, `Get`, `GetPath`, `Has`, `HasPath`, `Build`,
  `MustBuild` and `Reset`, so response bodies can be assembled from dotted paths instead of
  nested map literals.
- Added the response option `WithBodyBuilder` to plug a `pkg/mapbuilder` builder into a
  response, and the builder methods `EnableSpan`, `DisableSpan`, `EnableSkipBody` and
  `DisableSkipBody` to toggle the span and skip-body behaviours.

## [v0.1.54] - 2026-10-14

### Added

- Added `pkg/explorer`, an interactive file explorer for local inspection of the response
  artefacts produced by the dump helpers, including Windows-specific support through
  `mousetrap_windows.go` and `mousetrap_others.go`.
- Added `StartedByExplorer` to mark responses created from within the explorer.

### Changed

- Replaced the previous short `LICENSE` notice with the complete Apache License 2.0 text, so
  the repository now ships the full license body.

### Removed

- Removed the agent-guidance documents under `.skills/claude/` and `.skills/gemini/`
  (`cross-platform-compatibility.md`, `refactor-coding-convention.md`,
  `technical-documentation.md`). They are development-only helpers and are not part of the
  published module.

## [v0.1.53] - 2026-10-05

### Added

- Added the body helpers `EncodeBodyBase64URL`, `DecodeBodyBase64URL`, `EncodeBodyBase64Raw`
  and `DecodeBodyBase64Raw` to encode and decode a response body with the URL-safe and
  raw (unpadded) Base64 alphabets, complementing the standard Base64 helpers.
- Added the response options `WithBodyBase64` and `WithBodyBase64Bytes` to set an already
  Base64-encoded body.
- Added `Base64Encode`, `Base64Decode`, `Base64URLEncode` and `Base64URLDecode` to
  `pkg/b64` for encoding and decoding values outside of a response.

## [v0.1.52] - 2026-10-05

### Removed

- **Breaking: removed the HTML report renderer.** `DumpHTML` and `DumpResolveESTHTMLDoc`
  were deleted together with `report_html.go` and `templates/report.html`. Code that dumped
  responses as HTML must switch to the Markdown renderers (`DumpResolveESTMDDoc`,
  `DumpResolveESTMDDocTo`) or to `DumpJSON` / `DumpJSONTo`.

## [v0.1.51] - 2026-10-04

### Added

- Added `pkg/workergroup`, replify-native concurrent execution with two building blocks:
  a `Group` (fan-out over tasks with aggregated results) and a `WorkerPool` (persistent,
  live-scalable pool of workers). The constructors are `NewWorkerGroup`,
  `NewWorkerGroupWithContext`, `NewWorkerPool`, `NewPool`, `NewPoolJob` and
  `NewWorkerTask`; the entry points are `RunWorkerGroup` and `RunWorkerPool`.
- Added the group methods `Go`, `TryGo`, `Fail`, `Wait`, `SetLimit`, `Scale`, `Limit`,
  `Running`, `Err` and `Errors`, and the pool methods `Submit`, `TrySubmit`,
  `SubmitReply`, `ScaleBy`, `ScaleTo`, `Workers`, `Close`, `Wait`, `Err` and `Errors`.
- Added the worker options `WithJob`, `WithFn`, `WithIndex`, `WithDuration`, `WithLimit`,
  `WithQueueSize`, `WithErrorMode`, `WithPoolErrorMode`, `WithPanicRecovery` and
  `WithPoolPanicRecovery`, plus the `ErrorMode` values `FirstError` and `CollectErrors`
  and the sentinel errors `ErrPanic` and `ErrPoolClosed`.
- Added integration with the response envelope: a worker result can be attached to the
  response body via `Respond`, and logged with `Logging` / `Slogging`.

## [v0.1.50] - 2026-10-03

### Added

- Added the reason module, which attaches a machine-readable reason to a response. The public
  API is `NewReason`, `WithReason`, `WithReasonCode`, `WithReasonCategory`, `CastReasonCode`,
  `CastReasonCategory`, `CastReasonWithCode`, `CastReasonWithCategory`, `ReasonCategoryOf` and
  `IsReasonKnown`, together with the `ReasonCode` and `ReasonCategory` value types.
- Added `WithMetaFromRequest` and `WithPathFromRequest` so a reason can inherit the API
  version and the request path from the incoming `*http.Request`.
- Added the HTML dump renderers `DumpHTML` and `DumpResolveESTHTMLDoc` (removed again in
  `v0.1.52`).

## [v0.1.49] - 2026-10-02

### Changed

- **Breaking: renamed `DisableInjectStackTrace()` to `ReleaseInjectStackTrace()`.** The new
  name states that the injected stack trace is released rather than merely switched off.
  Update every call site; the previous name no longer exists.

## [v0.1.48] - 2026-10-01

### Fixed

- Fixed number handling in `UnwrapJSON`: JSON numbers are now decoded through
  `json.Decoder.UseNumber()` and kept as `json.Number` instead of being coerced eagerly,
  so large integers (for example Snowflake IDs) and high-precision decimals survive a
  parse/serialize round trip without precision loss.
- Fixed the signature generation so the request body is canonicalized before it is signed,
  the signature timestamp is taken from the current time, and an expired signature is
  detected before its remaining lifetime is computed.

## [v0.1.47] - 2026-09-29

### Added

- Added response signing. `NewSignature` and `NewSignatureConfig` create a signature, and
  `GenerateSignature`, `GenerateSignatureFromRequest`, `ApplySignature`,
  `ApplySignatureFromRequest` and `ApplySignatureHeader` produce and attach it.
- Added verification helpers `VerifySignature`, `VerifySignatureFromRequest` and
  `VerifySignatureHeader`, plus `SignatureFromHeader` for reading a signature back from an
  inbound request.
- Added the signing options `WithAlgorithm`, `WithSecretKey`, `WithSignature`, `WithHeaders`,
  `WithCanonicalize` / `EnableCanonicalize` / `DisableCanonicalize`,
  `WithIncludeTimestamp` / `EnableIncludeTimestamp` / `DisableIncludeTimestamp`,
  `WithMaxAge`, `WithNow`, `WithIgnoringResponseFields` and `WithIgnoringSignatureFields`.
- Added expiry inspection through `IsExpired`, `IsExpiredDuration`, `IsExpiredTime`,
  `IsExpiredUnix` and `RemainingMaxAge`, and configuration validation through
  `ValidateSignatureConfig`.
- Added signed response writers `WriteJSONSignature`, `WriteJSONSignatureHeader` and
  `WriteJSONSignatureFromRequest`, and the "ignoring"/"only" variants `RespondIgnoring`,
  `RespondOnly`, `JSONIgnoring`, `JSONOnly`, `JSONBytesIgnoring`, `JSONPrettyIgnoring`,
  `JSONPrettyOnly` and `LoggingIgnoring` for emitting a response with selected fields
  excluded.

## [v0.1.46] - 2026-09-19

### Added

- Added `pkg/b64` for Base64 detection and validation: `IsBase64`, `IsBase64Any`,
  `Base64Validator` / `NewBase64Validator`, `CompositeValidator` / `NewCompositeValidator`
  and `NewMultiEncodingValidator`. Use it to verify that a client-supplied body really is
  Base64 before decoding it.
- Added `EqualLinks` to compare the HATEOAS links of two responses.

### Changed

- Added the `base64.go` file with the Base64 body helpers `EncodeBodyBase64` and
  `DecodeBodyBase64`. The package path and the function signatures are unchanged, so no
  migration is required.

## [v0.1.45] - 2026-09-18

### Added

- Added body hashing helpers: `BodyBytes`, `Hash`, `HashBody`, `Hash256Body`, `MustHash`,
  `MustHashBody` and `MustHash256Body`. They compute a deterministic hash over the response
  body so responses can be cached or deduplicated.
- Added `EncodeBodyBase64` and `DecodeBodyBase64` to encode and decode the response body with
  the standard Base64 alphabet.

## [v0.1.44] - 2026-09-07

### Added

- Added HATEOAS link support. The new `links` component is built with `WithLink`,
  `WithLinkObject`, `WithLinks`, `WithSelf`, `WithFirst`, `WithLast`, `WithNext`, `WithPrev`,
  `WithHref`, `WithHreflang`, `WithMethod`, `WithName`, `WithTitle`, `WithType`, `WithProfile`,
  `WithTemplated`, `WithDeprecation` and `WithRequest`.
- Added the link inspection helpers `Has`, `Count`, `Items`, `Remove`, `IsLinkPresent`,
  `IsLinksPresent`, `IsSelfLinkPresent`, `IsHTTPRequestPresent` and `IsTemplated`, plus
  `Reply`, `ReplyPtr`, `Respond`, `JSON`, `JSONPretty`, `Logging` and `Slogging` to emit the
  links as a response.

## [v0.1.43] - 2026-09-06

### Added

- Added `EqualIssue` to compare the issue of two responses.

### Removed

- **Breaking: removed the `Toolbox` variable** from the root package. The utility functions
  it exposed are reachable through the dedicated packages; remove any `replify.Toolbox`
  reference.

### Changed

- Added the `ReleaseIssue`, `ReleaseCursor` and `ReleaseDebug` helpers to drop the
  corresponding component from a response before it is written.

## [v0.1.42] - 2026-09-05

### Added

- Added the issue module, which carries a machine-readable problem description with a
  response. The public API is `NewIssue`, `NewIssueID`, `NewFingerprint`, `NewPanicIssue`,
  `WithIssue`, `WithID`, `WithIDf`, `WithFingerprint`, `WithFingerprintf`, `AutoIssue`,
  `Issue`, `Fingerprint` and `ID`, together with the presence checks `IsIssuePresent`,
  `IsIDPresent`, `IsFingerprintPresent` and `IsMessagePresent`.
- Added `Reply`, `ReplyPtr`, `Respond`, `JSON`, `JSONPretty`, `Logging`, `Slogging` and
  `Equal` so an issue can be emitted or compared like any other response component.

## [v0.1.41] - 2026-09-02

### Added

- Added `DumpResolveESTMDDoc`, `DumpResolveESTMDDocTo` and `ResolveESTOrderedDoc` to render
  a response as a resolved Markdown document, with a deterministic section order.

## [v0.1.40] - 2026-09-01

### Added

- Added the `net/http` middleware `Recovery()`, which converts a panic in a downstream
  handler into a logged `500 Internal Server Error` response instead of dropping the
  connection.
- Added the `net/http` middleware `Logger()`, which logs the status code and duration of
  every request through `pkg/slogger`.
- Added `ResetGlobalLogger` to `pkg/slogger` for restoring the default global logger in
  tests.

### Fixed

- Fixed the `net/http` middleware so the wrapped `http.ResponseWriter` keeps its
  `http.Flusher` and `Unwrap` behaviour, and a panic in a downstream handler no longer
  produces a partial response body.

## [v0.1.39] - 2026-08-31

### Added

- Added the response option `WithSpan(span bool)` to enable or disable the tracing span
  attached to a response.

## [v0.1.38] - 2026-08-31

### Changed

- **Breaking: `pkg/encoding` merged the pretty-printing entry points into the compact ones.**
  `JSON` now accepts an optional `pretty` flag (`JSON(data, true)`) and `JSONE` does the
  same (`JSONE(data, true)`), while `JSONPretty` and `JSONPrettyE` were removed. Replace
  `encoding.JSONPretty(v)` with `encoding.JSON(v, true)` and `encoding.JSONPrettyE(v)`
  with `encoding.JSONE(v, true)`.
- **Breaking: the file-payload helpers of the root package were renamed.** `WithFilename`
  → `Filename` and `FileSizeHuman` → `SizeHumanReadable`. `WithJSONBody` was removed
  without a replacement — set the body directly with `WithBody`.

### Added

- Added `pkg/encoding/binary.go` with `IsBinary` and helpers for binary and file bodies
  (`Binary`, `File`, `FileAttachment`, `Filename`, `FromPath`, `IsBinaryBody`, `IsDir`,
  `IsSafeFilename`, `SizeHumanReadable`, `WithBodySuppressed`), so body encoding decisions
  can be made by the encoding package instead of by each caller.

## [v0.1.37] - 2026-08-22

### Added

- Added native `net/http` support. The wrapper can now write itself to an
  `http.ResponseWriter` through `Write`, `WriteJSON`, `WriteBinary`, `WriteFile` and
  `WriteHeader`, and can carry file payloads through `File`, `FileAttachment`, `FromPath` and
  `WithFilename`.
- Added the full set of HTTP status helpers on the wrapper — `OK`, `Created`, `Accepted`,
  `NoContent`, `NotFound`, `BadRequest`, `Unauthorized`, `Forbidden`, `Conflict`,
  `Gone`, `UnprocessableEntity`, `TooManyRequests`, `InternalServerError`,
  `ServiceUnavailable`, `GatewayTimeout` and the remaining RFC 9110 statuses — so a handler
  can set the status code without importing a status constant.

## [v0.1.36] - 2026-08-19

### Changed

- Improved the Markdown error-flow diagram: the root cause is now rendered in a fenced code
  block, and stack frames carry their shortened source location (`file.go:line`) as a note on
  the Mermaid diagram.

## [v0.1.35] - 2026-08-19

### Added

- Added Markdown documentation helpers that render a response as a document:
  `SummaryDoc`, `BasicDoc`, `HeaderDoc`, `MetaDoc`, `PagingDoc`, `CursorDoc`,
  `CustomFieldDoc`, `DebugDoc`, `ErrorFlowDoc`, `ErrorStackTraceDoc` and
  `SafeErrorStackTraceDoc`. Each returns a `*strchain.StringWeaver` so the result can be
  post-processed before it is written.
- Added Markdown formatting primitives to `pkg/strchain`: `Bold`, `Italic`, `Underline`,
  `Strikethrough`, `CodeBlock`, `InlineCode`, `InlineCodeF`, `Pipe`, `Pipes`, `Dash`,
  `Dashes`, `Underscore` and `Underscores`.
- Added `EqualCursor` to compare the cursor of two responses.
- Added `DumpMDDoc` and `DumpMDDocTo` to render a response as a Markdown document, and
  `HumanizeBytes` to format a byte count in human-readable units.
- Added `DisableInjectStackTrace()` to stop injecting the stack trace into an error
  response (renamed in `v0.1.49`).
- Added `Dump.FileSlogging` to log a dump and `Dump.FileSizeHuman` to report its size in
  human-readable units.
- Added `RequestedTimeUTC`, `RequestedTimeFormat` and `RequestedTimeUTCFormat` to read the
  request timestamp in UTC and in a formatted layout.

## [v0.1.34] - 2026-08-16

### Changed

- **Breaking: renamed the dump helpers to make their output explicit.** `Dump()` is now
  `DumpJSON()` and `DumpTo()` is now `DumpJSONTo()`. The behaviour is unchanged; update the
  call sites.
- Added `ReadAllBytes`, `MustReadAll` and `MustReadAllBytes` to read the dumped content
  directly, and `Content` to obtain a `sysx.ReadSeekCloser`.

## [v0.1.33] - 2026-08-15

### Added

- Added cursor pagination. The new `cursor` component is built with `WithCursor`,
  `WithCursorNext`, `WithCursorPrev`, `WithCursorHasNext`, `WithCursorHasPrev`,
  `WithCursorLimit`, `WithNext`, `WithPrevious`, `WithHasNext`, `WithHasPrevious` and
  `WithLimit`, and is read back with `Next`, `Previous`, `HasNext`, `HasPrevious`, `Limit`
  and `IsCursorPresent`. The component also exposes `Available`, `Equal`, `Reply`,
  `ReplyPtr`, `Respond`, `JSON`, `JSONPretty`, `String`, `Logging` and `Slogging`.
- Added the `Accepted`, `Rejected` and `RejectedOrAccepted` status helpers.

## [v0.1.32] - 2026-08-14

### Added

- Added tracing helpers to the wrapper: `S()` returns the `*slogger.Logger` bound to the
  response, `Span(name)` and `Scope(name)` open and close a tracing scope, and `Await(d)`
  records a wait duration in the log.

## [v0.1.31] - 2026-08-12

### Added

- Added the structured logging helpers `Checkpoint`, `Step` and `Flow` to record progress
  points, current/total steps and processing layers while a request is handled.

## [v0.1.30] - 2026-08-10

### Added

- Added `CarriageReturn` to `pkg/strchain` for building progress-style output.

## [v0.1.29] - 2026-08-09

### Added

- Added `Oneline` to `pkg/strutil` to collapse a multi-line string into a single line.

## [v0.1.28] - 2026-07-27

### Added

- Added `Reply()` and `ReplyPtr()` to the `header`, `meta` and `pagination` components, each
  returning its own typed interface (`H`, `M`, `P`). A component can now be returned to a
  caller without exposing the unexported concrete type.

## [v0.1.27] - 2026-07-25

### Changed

- **Breaking: the module path changed from `github.com/sivaosorg/replify` to
  `github.com/polarixa/replify`.** Every import of the library and of its sub-packages must
  be rewritten, and `go.mod` must require the new path:

  ```bash
  # old
  go get github.com/sivaosorg/replify
  # new
  go get github.com/polarixa/replify
  ```

  The package API itself is unchanged, so this is a mechanical rename.

- Updated the `pkg/slogger` documentation and examples to the new module path.

## [v0.1.26] - 2026-07-24

### Added

- Added `pkg/codegen`, a random-string generator for tokens, identifiers and fixtures. It
  exposes `New`, `MustNew`, `Generate`, `GenerateN`, `SetOptions` and `Options`, with the
  options `WithLength`, `WithPrefix`, `WithSuffix`, `WithCharset` and `WithCustomCharset`.

## [v0.1.25] - 2026-07-22

### Changed

- **Breaking: `pkg/encoding` renamed its JSON helpers to state the encoding they perform.**
  `Marshal` → `MarshalJSON`, `MarshalToString` → `MarshalJSONString`,
  `MarshalIndent` → `MarshalJSONIndent`, `Unmarshal` → `UnmarshalJSON` (taking `[]byte`),
  `UnmarshalFromString` → `UnmarshalJSONString`, `SafeUnmarshalBytes` →
  `StrictUnmarshalJSON`, `SafeUnmarshalJSON` → `StrictUnmarshalJSONString`,
  `IsValidJSON` → `IsValidJSONString`, `IsValidJSONBytes` → `IsValidJSON` (taking `[]byte`),
  `JSONToken` → `JSONE`, `JSONPrettyToken` → `JSONPrettyE`, `Ugly` → `CompactJSON`,
  `UglyInPlace` → `CompactJSONUnsafe` and `SpecInPlace` → `SpecUnsafe`. Note that
  `UnmarshalJSON` and `IsValidJSON` now accept `[]byte` where they previously accepted a
  string, and `UnmarshalJSONString` / `IsValidJSONString` cover the string variants.
- Added `pkg/translit` for Unicode transliteration, with `Unidecode`, `UnidecodeBody`,
  `ValidUTF8` and the translation `Tables`, so non-ASCII text can be normalised for search
  indexes, slugs or log output.

## [v0.1.24] - 2026-06-14

### Added

- Added the dump API: `Dump()`, `DumpTo(dst)`, `DumpBody()` and `DumpBodyTo(dst)` write a
  response (or only its body) to a self-cleaning temporary file and return a `*Dump` handle.
  The handle exposes `Resource`, `Filepath`, `Size`, `Rewind`, `File`, `FileInfo` and
  `Close`, and `Close` is thread-safe and idempotent so it can be deferred safely.
- Added the `pkg/sysx` streaming resources that back the dump API: `Resource` /
  `NewResource`, `NewTempFile`, `NewTempFileAt`, `NewTempFilename`, `MemBlob` /
  `NewMemBlob`, a thread-safe spill buffer, and MIME detection via `MimeFromName`.

- Hardened the atomic write path of `pkg/sysx` by adding the platform-specific
  implementations `opwrite_unix.go` and `opwrite_windows.go`, so concurrent writers cannot
  interleave partial files.

### Removed

- Removed `JoinUnary` from `pkg/strutil`.

## [v0.1.23] - 2026-05-29

### Added

- Added the `WithSkipBody` option and the `IsBodySkipped` check to suppress the body of a
  response while keeping its status code, metadata and headers — useful for `204`/`304`
  style answers and for log-only responses.

## [v0.1.22] - 2026-05-29

### Changed

- Reorganised the root package: the header, meta and pagination implementations moved from
  `replify.go` into `header.go`, `meta.go` and `page.go`. The exported API, the method
  receivers and the wire format are unchanged — no migration is required.

## [v0.1.21] - 2026-05-29

### Added

- Added stack-trace support for responses: `StackTrace`, `StackTraceString`,
  `WithStackTrace`, `WithInjectStackTrace` and `InjectStackTrace` capture where an error was
  raised, and `Slogging` renders the response element by element for tracing.

### Changed

- Renamed the internal `autoCause` helper to `autoAdjust` and taught it to inject the stack
  trace root cause into the error chain.

## [v0.1.20] - 2026-05-27

### Added

- Added `BodyString` and `String` to read the response body as a string.
- Added the `autoCause()` helper and wired it into `Reply()` and `ReplyPPtr()`, so every
  error response consistently carries the underlying cause instead of a bare message.

## [v0.1.19] - 2026-05-24

### Added

- Added a functional-options API for building a response in one expression. `Wrap(opts...)`
  accepts the new `ROption` values, and the following options are available:
  `WithStatusCode`, `WithBody`, `WithMessage` / `WithMessagef`, `WithPath` / `WithPathf`,
  `WithMeta`, `WithPagination`, `WithDebugging`, `WithDebuggingKV` / `WithDebuggingKVf`,
  `WithError` / `WithErrorf`, `WithErrorAck` / `WithErrorAckf`, `WithApiVersion` /
  `WithApiVersionf`, `WithRequestID` / `WithRequestIDf`, `WithLocale` / `WithLocaleValue`,
  `WithRequestedTime`, `WithCustomFields`, `WithCustomFieldKV` / `WithCustomFieldKVf`,
  `WithPage`, `WithPerPage`, `WithTotalPages`, `WithTotalItems`, `WithTotal`, `WithIsLast`
  and the error-accumulation options `WithAppendError`, `WithAppendErrorf`,
  `WithAppendErrorAck` and `WithAppendErrors`.
- Added `Logging(logger ...*slogger.Logger)` to dispatch a structured log entry describing
  the response through `pkg/slogger`.
- Added the response comparison helpers `EqualHeader`, `EqualMeta` and `EqualPages`, plus the
  component-level `Equal` methods on the header, meta and pagination types, and the error
  helpers `Is` and `IsOwned` to check whether an error is set and owned by a response.

### Fixed

- Fixed the response cache so a wrapper that is written twice does not serve a stale body,
  and made the status-aware logging report the final status code.

## [v0.1.18] - 2026-05-02

### Changed

- Updated the GitHub Actions workflows (`.github/workflows/gh_wrk_news.yml` and
  `.github/workflows/gh_wrk_news_go.yml`) to match the current build and release layout.

## [v0.1.17] - 2026-04-18

### Added

- Added a second, higher-level API to `pkg/strchain`: `Chain`, `ToWeaver` and
  `ToSafeWeaver` provide a value-oriented builder with `Append`, `Prepend`, `PrependIf`,
  `AppendIf`, `AppendIfF`, `Map`, `MapRunes`, `Filter`, `Each`, `EachCast`, `When`,
  `WhenCast`, `Unless`, `UnlessCast`, `IfEmpty`, `IfNotEmpty`, `Join`, `JoinSlice`, `Split`,
  `SplitN`, `Contains`, `Count`, `HasPrefix`, `HasSuffix`, `Index`, `Runes`, `Bytes`,
  `Equals`, `Empty`, `IsEmpty`, `IsNotEmpty`, `Reverse`, `Title`, `Upper`, `Lower`,
  `PadLeft`, `PadRight`, `Center`, `Trim`, `TrimChars`, `TrimLeft`, `TrimRight`,
  `TrimPrefix`, `TrimSuffix`, `Replace`, `ReplaceN`, `Truncate` and `TruncateWithSuffix`.

## [v0.1.16] - 2026-04-18

### Added

- Added `AppendErrors` to append several errors to a response in one call.

## [v0.1.15] - 2026-04-16

### Added

- Added `Stringf` to `pkg/slogger`, which formats a message with `fmt.Sprintf` semantics.

## [v0.1.14] - 2026-04-15

### Changed

- Promoted `v0.1.13-rc.1` to `v0.1.14` without any code change: both tags point at the same
  tree. Upgrade from `v0.1.12` with `go get github.com/polarixa/replify@v0.1.14`.

## [v0.1.13-rc.1] - 2026-04-15

### Added

- Added the accessors of `pkg/slogger.RotationOptions` and `pkg/slogger.SamplingOptions`
  (`Directory`, `MaxBytes`, `MaxAge`, `Period`, `Compress`, `First`, `Thereafter`, ...), so
  the options can be read back after they have been configured.

### Fixed

- Made the response cache concurrency-safe: the cached wrapper is now read under a read
  lock, rebuilt under a write lock with a double check, and the content hash is
  recomputed while the write lock is held so a stale entry can never be served.
- Fixed `pkg/sysx` to defer `Close()` on files, listeners and connections so no descriptor
  leaks, to compare `syscall.Errno` values directly instead of matching the
  locale-dependent error text, and to assert the file mutex type safely.
- Fixed the `pkg/slogger` rotation, sampling, writer and options handling.

## [v0.1.12] - 2026-04-12

### Added

- Added `ColorMode` and the `WithColorMode` option to `pkg/slogger`, so ANSI colour codes can
  be forced off for file output while remaining enabled for terminals.
- Added fuzz targets `FuzzItoa64`, `FuzzUtoa64` and `FuzzParseLevel` to `pkg/slogger`, and a
  broad edge-case test suite for `StripANSI`.
- Reworked `pkg/slogger`: the logger, options, rotation, sampling and writer
  implementations were hardened and extended with a large internal test suite, and
  `pkg/strchain` was reworked for the same reasons.

## [v0.1.12-rc.2] - 2026-04-12

### Changed

- **Breaking: the fields of `slogger.Options`, `slogger.SamplingOptions` and
  `slogger.RotationOptions` are now unexported.** Options can no longer be built with a
  struct literal; use the setters and getters instead — `SetLevel`, `SetFormatter`,
  `SetOutput`, `SetCaller`, `SetCallerSkip`, `SetFields`, `AddFields`, `SetName`,
  `SetSamplingOpts` and `SetRotationOpts` (plus `SetCompress`, `SetDirectory`,
  `SetFirst`, `SetMaxAge`, `SetMaxBytes`, `SetPeriod` and `SetThereafter` for the
  rotation options), and the matching accessors `Level`, `Formatter`, `Writer`,
  `Writers`, `Caller`, `CallerSkip`, `Fields`, `Name`, `SamplingOpts` and
  `RotationOpts`.

## [v0.1.12-rc.1] - 2026-04-12

### Fixed

- Fixed cross-platform compatibility, concurrency safety and panic paths in `pkg/slogger`.
- Fixed `pkg/encoding` to return sentinel errors, to bounds-check slice accesses and to
  handle errors instead of ignoring them; removed the dead code found during the audit.
- Fixed `pkg/strchain` to encode JSON strings with `encoding/json.Marshal` (RFC 8259
  compliant) instead of hand-rolled escaping, and replaced `fmt.Fprintf` with `strconv` and
  `strings.Repeat` on the hot paths.
- Reorganised `pkg/encoding` into `constant.go`, `errors.go`, `type.go` and `utilities.go`.
  The exported API is unchanged.

## [v0.1.11] - 2026-04-11

### Changed

- **Breaking: the `pkg/sysx` file helpers were renamed to state the data they exchange.**
  `ReadFile` → `ReadBytes`, `ReadFileString` → `ReadString`, `WriteFile` → `WriteBytes`,
  `WriteFileString` → `WriteString`, `WriteFileLocked` → `WriteBytesWithLocked`,
  `AppendFile` → `AppendBytes` and `AtomicWriteFile` → `AtomicWriteBytes`. Update every
  call site; the old names no longer exist.
- **Breaking: the `pkg/hashy` helpers were renamed around their output width.** `HashBase10`
  → `Hash10`, `HashBase16` → `Hash16`, `HashBase32` → `Hash32`, `HashEncoded` → `Hash64`,
  `HashHex` → `HashHex16` and `HashHexShort` → `Hash16Padded`. The new `NewHash` and
  `NewHash64` constructors return a standard `hash.Hash` / `hash.Hash64`, and `HashAlgorithm`
  selects the algorithm.
- **Breaking: `pkg/randn` dropped the `UUIDJoin` helper**, which concatenated UUID strings.
  Use `fmt` or a dedicated string builder instead.
### Added

- Added the `pkg/randn` XID implementation: `NewXID`, `NewXIDWithTime` and the `XID` type,
  plus `RandIDHex` and `RandUint32`.
- Added `pkg/common` helpers (`Transform`, `Filter`, `Unique`, `Partition`, `RotateLeft`,
  `RotateRight`) and extended `pkg/sysx` with `CopyDir`, `ClearDir`, `IsDirEmpty`,
  `IsSafeDirEmpty`, `Move`, `Touch`, `FileMD5`, `FileSHA256`, `IsBinary`,
  `RemoveDirIfExist`, `RemoveFileIfExist`, `SafeRemove`, `SafeRemoveAll`, `CreateDirs` and
  `CreateSubDirs`.
- Added the `.skills/` agent-guidance documents for Claude and Gemini.

### Fixed

- Fixed `pkg/randn` to generate RFC 4122 compliant UUIDs across platforms, removed the data
  race in `RandIntr`, deprecated `Seed`, and replaced the `log.Fatalf` call in `CryptoID`
  with an error return.
- Fixed `pkg/coll`: the `KeySet` bug, asymmetric `Difference`, duplicate values in
  `Intersect`, the `Shuffle` seed, the `Join`/`String` builders, the `Flatten` assertion
  and the `JoinKeySep` naming.
- Fixed `pkg/conv` to wrap errors, guard against integer overflow, and keep the date-format
  table safe under concurrent use.

## [v0.1.10] - 2026-04-05

### Added

- Added the `cmd/replify` command-line entry point, and a `main/main.go` development entry
  point.
- Added the block helpers `StartBrace`, `EndBrace`, `StartBracket`, `EndBracket`,
  `StartParenthesize`, `EndParenthesize`, `IndentOnly` and `IndentOnlyLine` to
  `pkg/strchain`.

### Changed

- Improved the `Makefile`: `make dev`, `make run`, `make build`, `make test` (with coverage),
  `make tidy`, `make deps-upgrade`, `make deps-clean-cache`, `make clean`, `make tree` and
  `make tree-dirs`, plus Windows binary extension handling.

## [v0.1.9] - 2026-04-05

### Added

- Added `JSONFieldArrayEnd` to `pkg/strchain`.

## [v0.1.8] - 2026-04-05

### Added

- Added `JSONFieldArrayStart` to `pkg/strchain`.

## [v0.1.7] - 2026-04-05

### Added

- Added typed JSON field helpers to `pkg/strchain`: `JSONFieldBool`, `JSONFieldFloat32`,
  `JSONFieldFloat64`, `JSONFieldInt8`, `JSONFieldInt16`, `JSONFieldInt32`, `JSONFieldInt64`,
  `JSONFieldString`, `JSONFieldUint`, `JSONFieldUint16`, `JSONFieldUint32` and
  `JSONFieldUint64`.
- Added the cast helpers `EachCast`, `WhenCast` and `UnlessCast` to `pkg/strchain`.

### Removed

- **Breaking: removed the generic `JSONField` helper** from `pkg/strchain`. Use the typed
  `JSONField*` variants instead, for example `JSONFieldString` instead of `JSONField`.

## [v0.1.6] - 2026-04-04

### Added

- Added JSON and formatting helpers to `pkg/strchain`: `JSONArrayStart`, `JSONArrayEnd`,
  `JSONObjectStart`, `JSONObjectEnd`, `JSONField`, `JSONKey`, `JSONKeyString`,
  `JSONKeyInt`, `JSONKeyFloat`, `JSONKeyBool`, `JSONString`, `CommaIfNotLast`, `IndentF` and
  `IndentLineF`.

## [v0.1.5] - 2026-04-03

### Added

- Added `FromPtr` and `SafeFromPtr` to `pkg/strchain` to build a `StringWeaver` from an
  existing `*strings.Builder`.

## [v0.1.4] - 2026-04-03

### Added

- Added the `Builder` constructor to `pkg/strchain`.

## [v0.1.2] - 2026-04-02

### Added

- Added `pkg/strchain`, a thread-safe `strings.Builder` wrapper with a fluent API. It offers
  `New`, `NewSafe`, `NewWithCapacity`, `NewSafeWithCapacity`, `Append`, `AppendF`, `AppendIf`,
  `AppendIfF`, `AppendBool`, `AppendByte`, `AppendBytes`, `AppendRune`, `AppendInt`,
  `AppendInt8`, `AppendInt16`, `AppendInt32`, `AppendInt64`, `AppendUint`, `AppendUint8`,
  `AppendUint16`, `AppendUint32`, `AppendUint64`, `AppendUintptr`, `AppendFloat32`,
  `AppendFloat64`, plus the punctuation helpers `Arrow`, `FatArrow`, `Brace`, `Bracket`,
  `Colon`, `Comma`, `Dot`, `Semicolon`, `Space`, `Spaces`, `Tab`, `Tabs`, `Quote`,
  `SingleQuote`, `Parenthesize`, `Wrap`, `Indent`, `IndentLine`, `Line`, `LineF`, `NewLine`,
  `NewLines`, `Join`, `Repeat`, `Each`, `When`, `Unless`, `Inspect`, `Build`, `String`,
  `Len`, `Cap`, `Grow`, `Reset` and `Clone`.

## [v0.1.1] - 2026-03-30

### Added

- Added the pointer-aware string checks to `pkg/strutil`: `IsBlankPtr`, `IsEmptyPtr`,
  `IsAnyBlankPtr`, `IsAnyEmptyPtr`, `IsNoneBlankPtr`, `IsNoneEmptyPtr`, `IsNotBlankPtr`,
  `IsNotEmptyPtr`, `IsNumericPtr`, `IsNumericSpacePtr` and `IsWhitespacePtr`.

### Changed

- **Breaking: `strutil.SlugifySpecial` was replaced by the `SlugifySpec` type**, which
  describes the slug rules explicitly instead of relying on a boolean flag. Replace
  `strutil.SlugifySpecial(s, "-")` with `strutil.SlugifySpec(s, "-")`.

## [v0.1.0] - 2026-03-09

### Added

- Added the `LICENSE` file and the package documentation in `doc.go`.
- Added YAML configuration support to `pkg/slogger`: `SloggerConfig`, `ApplyGlobalConfig`,
  `FileConfig`, `OutputConfig`, `RotationConfig`, `CallerConfig`, `ColorConfig` and
  `ArchiveConfig`, with ready-made files under `pkg/slogger/configs`.
- Added the `pkg/sysx` filesystem utilities `CopyFile`, `CountLines`, `CreateDir`,
  `ListDir`, `ListDirFiles`, `ListDirDirs`, `JoinPath`, `AbsPath`, `CleanPath`,
  `SplitPath`, `BaseName`, `DirName`, `Ext`, `Head`, `FileMode`, `FileModTime`,
  `TruncateFile` and `RemoveDir`.

### Changed

- **Breaking: `slogger.TextFormatter.WithDisableColors()` was renamed to
  `WithDisableColor()`.** Update every call site; the old name no longer exists. The
  matching `WithEnableColor()` was introduced earlier, in `v0.0.9`.

## [v0.0.9] - 2026-03-08

### Added

- Added `pkg/slogger`, a structured logger built on the standard library `log/slog`
  concepts: levels (`Trace` to `Fatal`), typed fields (`String`, `Int`, `Bool`, `Duration`,
  `Time`, `Any`, `Err`), hooks (`AddHook`, `NewHooks`, `LevelWriterHook`), formatters
  (`JSONFormatter`, `TextFormatter`, `NewJSONFormatter`, `NewTextFormatter`), log rotation
  (`NewLevelFileWriter`, `RotationOptions`) and sampling (`SamplingOptions`).
- Added the `With*` option builders of `pkg/slogger` (`WithLevel`, `WithOutput`,
  `WithFormatter`, `WithFields`, `WithContextFields`, `WithName`, `WithCaller`,
  `WithCallerSkip`, `WithTimeFormat`, `WithSampling`, `WithRotation`, `WithColor`,
  `WithColorStyle`, `WithDisableColors`) and the global logger helpers
  `GlobalLogger`, `SetGlobalLogger` and `GlobalWithContextFields`.
- Added optional ANSI colour support to the JSON formatter.
- Added the colour options `WithEnableColor` and `WithDisableColors` to toggle the ANSI
  colour codes of the text and JSON formatters.

### Fixed

- Fixed double-encoding of JSON string values in `slogger.AnyType`.

## [v0.0.8] - 2026-03-06

### Added

- Added `pkg/sysx`, a lightweight system-utility package: process inspection (`Arch`,
  `OSName`, `OSVersion`, `GoVersion`, `NumCPU`, `IsLinux`, `IsDarwin`, `IsWindows`,
  `Is64Bit`, `IsArm`), environment helpers (`Getenv`, `Setenv`, `Unsetenv`, `Hasenv`,
  `EnvMap`, `JSONEnvMap`, the typed `GetenvInt`, `GetenvBool`, `GetenvDuration` family),
  file helpers (`ReadFile`, `WriteFile`, `AppendFile`, `AtomicWriteFile`, `SafeFileWriter`)
  and command execution (`ExecCommand`, `ExecCommandContext`, `ExecCommandWithTimeout`,
  `ExecOutput`, `ExecPipeline`, `ExecStreaming`, `NewCommand`).
- Added `pkg/netx`, an IPv4/IPv6 subnetting toolkit with fixed-length and variable-length
  subnet masks (`ParseCIDR`, `Split`, `SplitIntoN`, `DivideByHosts`, `NextSubnet`,
  `Overlaps`, `BroadcastAddress`, `NetworkAddress`) plus host and network utilities.

## [v0.0.7] - 2026-03-04

### Added

- Added `pkg/crontask`, a cron scheduler: `Scheduler`, `New`, `Start`, `Stop`, `Shutdown`,
  `Schedule`, `Remove`, `Parse`, `MustParse`, `Explain`, `NextRun`, `NextRuns` and
  `IsValidCronExpr`, with the job options `WithJobID`, `WithJobName`, `WithLocation`,
  `WithSeconds`, `WithTimeout`, `WithContext`, `WithMaxRetries`, `WithBackoff`,
  `WithJitter` and `WithErrorHandler`.
- Added the `pkg/crontask` hook system (`Hooks`, `ChainHooks`, `NoopHooks`, `LoggingHook`,
  `MetricsHook`, `RetryHook`, `PanicHook`, `RecoverPanicHook`,
  `RecoverPanicHookWithHandler`, `ConcurrencyLimiterHook`, `TimeoutLoggerHook`) and the
  backoff policies `ConstantBackoff` and `ExponentialBackoff`.
- Added 40+ business-oriented cron aliases (for example `@businessDaily`,
  `@businessHourly`, `@marketOpen`, `@marketClose`, `@lunchtime`, `@payroll`,
  `@taxDeadline` and `@nightlyMaintenance`) through `RegisterAlias`, `DeleteAlias` and
  `Explain`.
- Added package documentation (`doc.go`) for `pkg/assert`, `pkg/coll`, `pkg/common`,
  `pkg/conv`, `pkg/encoding`, `pkg/hashy`, `pkg/match`, `pkg/msort`, `pkg/randn`,
  `pkg/ref`, `pkg/strutil` and `pkg/truncate`, plus a Gin integration guide under
  `pkg/crontask/enterprise-grades/`.

## [v0.0.6] - 2026-03-02

### Changed

- **Breaking: the JSON accessors were renamed to the `JSON` spelling.** `Json()` became
  `JSON()` and `JsonPretty()` became `JSONPretty()` on the root wrapper, the header, the
  meta, the pagination, `StreamConfig`, `StreamingStats` and `StreamProgress`. Update every
  call site; the old names no longer exist.

## [v0.0.5] - 2026-03-02

### Changed

- **Breaking: the `pkg/encoding` helpers were renamed to state their behaviour.**
  `Marshal` → `MarshalJSONb`, `MarshalToString` → `MarshalJSONs`,
  `MarshalIndent` → `MarshalJSONIndent`, `Unmarshal` → `UnmarshalBytes`,
  `UnmarshalFromString` → `UnmarshalJSON`, `JsonSafe` → `JSON` and
  `JsonSafePretty` → `JSONPretty`. Update every call site; the old names no longer exist.
- Added `JSONToken` and `JSONPrettyToken` to `pkg/encoding`, which return the encoding error
  instead of swallowing it.
- Added `SafeUnmarshalBytes` and `SafeUnmarshalJSON` to `pkg/encoding`, which never panic on
  malformed input.

### Added

- Added `NormalizeJSON` to `pkg/encoding` and the `WithNormalizedBody` wrapper option, which
  repairs escaped structural quotes, trailing commas and stray null bytes before a body is
  encoded.
- Added the root-level `UnwrapJSON` helper to parse a JSON document back into a wrapper.

### Fixed

- Fixed `w.JSON()` and `w.JSONPretty()` to avoid double-encoding a body that is already
  valid JSON.
- Added validation to `UnwrapJSON` so malformed input is rejected instead of producing a
  half-initialised wrapper.

## [v0.0.5-beta.1] - 2026-02-28

### Changed

- **Breaking: the root-level `Json()` and `JsonPretty()` accessors were renamed to `JSON()`
  and `JSONPretty()`.** This pre-release prepares the rename that ships in `v0.0.6`.
- Added the `JSONDebugging*` accessors to read values from the debug section of a response.

## [v0.0.4] - 2026-02-28

### Added

- Added a large set of built-in `@` transformers to `pkg/fj`, usable directly in a path
  expression: `@pretty`, `@minify`, `@keys`, `@values`, `@valid`, `@flatten`, `@join`,
  `@group`, `@search`, `@reverse`, `@flip`, `@trim`, `@uppercase`, `@lowercase`,
  `@snakecase`, `@camelcase`, `@kebabcase`, `@replace`, `@replaceAll`, `@insertAt`,
  `@hex`, `@binary`, `@countWords`, `@count`, `@first`, `@last`, `@sum`, `@min`, `@max`,
  `@pluck`, `@padLeft`, `@padRight`, `@project`, `@filter`, `@default` and `@coerce`.
  They take an optional JSON argument, for example
  `user.@project:{"pick":["name","age"],"rename":{"age":"years"}}`.

## [v0.0.4-beta.1] - 2026-02-26

### Changed

- **Breaking: `fj.AddTransformer` now accepts the `fj.Transformer` interface instead of a
  plain function.** Wrap a function with `fj.TransformerFunc` to keep the old behaviour:

  ```go
  // before
  fj.AddTransformer("upper", func(json, arg string) string { return strings.ToUpper(json) })
  // after
  fj.AddTransformer("upper", fj.TransformerFunc(func(json, arg string) string { return strings.ToUpper(json) }))
  ```

  A struct that implements `Apply(json, arg string) string` can now be registered directly,
  which allows stateful transformers.

### Added

- Added the `fj.Transformer` interface and the `fj.TransformerFunc` adapter.

## [v0.0.3] - 2026-02-26

### Added

- Added `pkg/fj`, a JSON path engine that extracts, searches and aggregates data without
  unmarshalling the whole document. It provides `Get`, `GetMulti`, `GetBytes`, `FindPath`,
  `FindPaths`, `FindPathMatch`, `FindPathsMatch`, `Filter`, `First`, `Last`, `Pluck`,
  `Project`, `Distinct`, `GroupBy`, `SortBy`, `Search`, `SearchByKey`,
  `SearchByKeyPattern`, `Contains`, `Count`, `Sum`, `Avg`, `Min`, `Max`, `CoerceTo`, the
  `Foreach` iterator and the transformer registry (`AddTransformer`,
  `IsTransformerRegistered`).
- Added `pkg/common` with the reader and tee helpers `ReadAll`, `SlurpAll`, `SlurpLine`,
  `SlurpLines`, `TailLines`, `TeeCopy` and `TeeTap`.
- Added the root-level `parser.go` that parses a JSON document into a wrapper.
- Added the typed debugging accessors on the meta component (`DebuggingBool`,
  `DebuggingInt`, `DebuggingString`, `DebuggingTime`, `DebuggingUint`, ...) and the
  `JSONBody*` helpers (`FindJSONBodyPath`, `FilterJSONBody`, `GroupByJSONBody`,
  `SortJSONBody`, `SumJSONBody`, `MinJSONBody`, `MaxJSONBody`, `PluckJSONBody`,
  `DistinctJSONBody`, `CountJSONBody`, `QueryJSONBody`, `SearchJSONBody`, ...).

### Changed

- **Breaking: the debugging accessor of the meta component was renamed.**
  `OnKeyDebugging` → `OnDebugging`.

### Removed

- **Breaking: removed `OptimizeSafe`** from the encoding helpers. Use `NormalizeJSON`
  (added in `v0.0.5`) or `pkg/encoding`'s `Pretty` / `CompactJSON` helpers instead.

## [v0.0.2] - 2026-02-20

### Added

- Added `pkg/truncate`, a strategy-based string truncation package. `Truncate`,
  `TruncateWithLength`, `NewTruncator` and `TruncatorBuilder` are configured with
  `WithMaxLength`, `WithOmission` and `WithPosition`, and the strategies
  `NewCutStrategy`, `NewCutEllipsisStrategy`, `NewCutEllipsisLeadingStrategy` and
  `NewEllipsisMiddleStrategy` decide where the cut happens.
- Added `pkg/conv`, a type-conversion package with the `Must*` helpers (`MustBool`,
  `MustInt`, `MustFloat64`, `MustTime`, `MustParseJSON`, ...), the `*OrDefault` family
  (`BoolOrDefault`, `Int64OrDefault`, `TimeOrDefault`, ...), the `Converter` type with
  `NewConverter`, and the error helpers `ConvError`, `IsConvError` and `AsConvError`.
- Added the `pkg/strutil` helpers `Apply`, `Center`, `PadLeft`, `PadRight`, `TrimLeft`,
  `TrimRight`, `Replace`, `ReplaceAll`, `ReplaceIgnoreCase`, `StartsWithAny`,
  `EndsWithAny`, `Equals`, `EqualsIgnoreCase`, `IsAllEmpty`, `DefaultIfEmpty` and
  `DefaultIfBlank`.
- Added the typed custom-field accessors on the meta component: `CustomBool`,
  `CustomInt`, `CustomInt8`, `CustomInt16`, `CustomInt32`, `CustomInt64`, `CustomUint`,
  `CustomUint8`, `CustomUint16`, `CustomUint32`, `CustomUint64`, `CustomFloat32`,
  `CustomFloat64`, `CustomString`, `CustomTime` and `CustomDuration`.

### Changed

- **Breaking: the custom-field accessors of the meta component were renamed.**
  `IsCustomFieldPresent` → `IsCustomPresent`, `IsCustomFieldKeyPresent` →
  `IsCustomKeyPresent` and `OnKeyCustomField` → `OnCustom`. Update every call site; the
  old names no longer exist.

## [v0.0.1] - 2026-01-29

### Added

- Initial release of the replify ecosystem: a zero-dependency, fluent and chainable
  response wrapper for Go REST services.
- Added the response envelope and its fluent builders: `New()` plus the `With*` setters for
  the body, status code, message, metadata, pagination, custom fields and debugging data,
  and the status helpers `WrapOk`, `WrapCreated`, `WrapAccepted`, `WrapNoContent`,
  `WrapBadRequest`, `WrapUnauthorized`, `WrapForbidden`, `WrapNotFound`, `WrapConflict`,
  `WrapGone`, `WrapUnprocessableEntity`, `WrapTooManyRequest`, `WrapLocked`,
  `WrapProcessing`, `WrapUpgradeRequired`, `WrapRequestTimeout`,
  `WrapRequestEntityTooLarge`, `WrapUnsupportedMediaType`, `WrapHTTPVersionNotSupported`,
  `WrapPaymentRequired`, `WrapPreconditionFailed`, `WrapNotImplemented`,
  `WrapServiceUnavailable`, `WrapInternalServerError`, `WrapBadGateway`,
  `WrapGatewayTimeout` and `WrapMethodNotAllowed`.
- Added the response components `Header()`, `Meta()` and `Pages()` / `FromPages()`, the
  error handling with stack traces (`errstack.go`, `Error`, `Cause`), the JSON normaliser
  (`norm.go`), the buffer pool (`NewBufferPool`) and the streaming support (`stream.go`,
  `NewStreamConfig`, `NewStreaming`).
- Added the first sub-packages: `pkg/assert`, `pkg/coll`, `pkg/common`, `pkg/conv`,
  `pkg/encoding`, `pkg/hashy`, `pkg/match`, `pkg/msort`, `pkg/randn`, `pkg/ref` and
  `pkg/strutil`.

[v0.1.56]: https://github.com/polarixa/replify/compare/v0.1.55...v0.1.56
[v0.1.55]: https://github.com/polarixa/replify/compare/v0.1.54...v0.1.55
[v0.1.54]: https://github.com/polarixa/replify/compare/v0.1.53...v0.1.54
[v0.1.53]: https://github.com/polarixa/replify/compare/v0.1.52...v0.1.53
[v0.1.52]: https://github.com/polarixa/replify/compare/v0.1.51...v0.1.52
[v0.1.51]: https://github.com/polarixa/replify/compare/v0.1.50...v0.1.51
[v0.1.50]: https://github.com/polarixa/replify/compare/v0.1.49...v0.1.50
[v0.1.49]: https://github.com/polarixa/replify/compare/v0.1.48...v0.1.49
[v0.1.48]: https://github.com/polarixa/replify/compare/v0.1.47...v0.1.48
[v0.1.47]: https://github.com/polarixa/replify/compare/v0.1.46...v0.1.47
[v0.1.46]: https://github.com/polarixa/replify/compare/v0.1.45...v0.1.46
[v0.1.45]: https://github.com/polarixa/replify/compare/v0.1.44...v0.1.45
[v0.1.44]: https://github.com/polarixa/replify/compare/v0.1.43...v0.1.44
[v0.1.43]: https://github.com/polarixa/replify/compare/v0.1.42...v0.1.43
[v0.1.42]: https://github.com/polarixa/replify/compare/v0.1.41...v0.1.42
[v0.1.41]: https://github.com/polarixa/replify/compare/v0.1.40...v0.1.41
[v0.1.40]: https://github.com/polarixa/replify/compare/v0.1.39...v0.1.40
[v0.1.39]: https://github.com/polarixa/replify/compare/v0.1.38...v0.1.39
[v0.1.38]: https://github.com/polarixa/replify/compare/v0.1.37...v0.1.38
[v0.1.37]: https://github.com/polarixa/replify/compare/v0.1.36...v0.1.37
[v0.1.36]: https://github.com/polarixa/replify/compare/v0.1.35...v0.1.36
[v0.1.35]: https://github.com/polarixa/replify/compare/v0.1.34...v0.1.35
[v0.1.34]: https://github.com/polarixa/replify/compare/v0.1.33...v0.1.34
[v0.1.33]: https://github.com/polarixa/replify/compare/v0.1.32...v0.1.33
[v0.1.32]: https://github.com/polarixa/replify/compare/v0.1.31...v0.1.32
[v0.1.31]: https://github.com/polarixa/replify/compare/v0.1.30...v0.1.31
[v0.1.30]: https://github.com/polarixa/replify/compare/v0.1.29...v0.1.30
[v0.1.29]: https://github.com/polarixa/replify/compare/v0.1.28...v0.1.29
[v0.1.28]: https://github.com/polarixa/replify/compare/v0.1.27...v0.1.28
[v0.1.27]: https://github.com/polarixa/replify/compare/v0.1.26...v0.1.27
[v0.1.26]: https://github.com/polarixa/replify/compare/v0.1.25...v0.1.26
[v0.1.25]: https://github.com/polarixa/replify/compare/v0.1.24...v0.1.25
[v0.1.24]: https://github.com/polarixa/replify/compare/v0.1.23...v0.1.24
[v0.1.23]: https://github.com/polarixa/replify/compare/v0.1.22...v0.1.23
[v0.1.22]: https://github.com/polarixa/replify/compare/v0.1.21...v0.1.22
[v0.1.21]: https://github.com/polarixa/replify/compare/v0.1.20...v0.1.21
[v0.1.20]: https://github.com/polarixa/replify/compare/v0.1.19...v0.1.20
[v0.1.19]: https://github.com/polarixa/replify/compare/v0.1.18...v0.1.19
[v0.1.18]: https://github.com/polarixa/replify/compare/v0.1.17...v0.1.18
[v0.1.17]: https://github.com/polarixa/replify/compare/v0.1.16...v0.1.17
[v0.1.16]: https://github.com/polarixa/replify/compare/v0.1.15...v0.1.16
[v0.1.15]: https://github.com/polarixa/replify/compare/v0.1.14...v0.1.15
[v0.1.14]: https://github.com/polarixa/replify/compare/v0.1.13-rc.1...v0.1.14
[v0.1.13-rc.1]: https://github.com/polarixa/replify/compare/v0.1.12...v0.1.13-rc.1
[v0.1.12]: https://github.com/polarixa/replify/compare/v0.1.12-rc.2...v0.1.12
[v0.1.12-rc.2]: https://github.com/polarixa/replify/compare/v0.1.12-rc.1...v0.1.12-rc.2
[v0.1.12-rc.1]: https://github.com/polarixa/replify/compare/v0.1.11...v0.1.12-rc.1
[v0.1.11]: https://github.com/polarixa/replify/compare/v0.1.10...v0.1.11
[v0.1.10]: https://github.com/polarixa/replify/compare/v0.1.9...v0.1.10
[v0.1.9]: https://github.com/polarixa/replify/compare/v0.1.8...v0.1.9
[v0.1.8]: https://github.com/polarixa/replify/compare/v0.1.7...v0.1.8
[v0.1.7]: https://github.com/polarixa/replify/compare/v0.1.6...v0.1.7
[v0.1.6]: https://github.com/polarixa/replify/compare/v0.1.5...v0.1.6
[v0.1.5]: https://github.com/polarixa/replify/compare/v0.1.4...v0.1.5
[v0.1.4]: https://github.com/polarixa/replify/compare/v0.1.2...v0.1.4
[v0.1.2]: https://github.com/polarixa/replify/compare/v0.1.1...v0.1.2
[v0.1.1]: https://github.com/polarixa/replify/compare/v0.1.0...v0.1.1
[v0.1.0]: https://github.com/polarixa/replify/compare/v0.0.9...v0.1.0
[v0.0.9]: https://github.com/polarixa/replify/compare/v0.0.8...v0.0.9
[v0.0.8]: https://github.com/polarixa/replify/compare/v0.0.7...v0.0.8
[v0.0.7]: https://github.com/polarixa/replify/compare/v0.0.6...v0.0.7
[v0.0.6]: https://github.com/polarixa/replify/compare/v0.0.5...v0.0.6
[v0.0.5]: https://github.com/polarixa/replify/compare/v0.0.5-beta.1...v0.0.5
[v0.0.5-beta.1]: https://github.com/polarixa/replify/compare/v0.0.4...v0.0.5-beta.1
[v0.0.4]: https://github.com/polarixa/replify/compare/v0.0.4-beta.1...v0.0.4
[v0.0.4-beta.1]: https://github.com/polarixa/replify/compare/v0.0.3...v0.0.4-beta.1
[v0.0.3]: https://github.com/polarixa/replify/compare/v0.0.2...v0.0.3
[v0.0.2]: https://github.com/polarixa/replify/compare/v0.0.1...v0.0.2
[v0.0.1]: https://github.com/polarixa/replify/releases/tag/v0.0.1
