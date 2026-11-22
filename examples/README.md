# replify examples

Runnable programs showing the public `replify` API. Each one is self-contained
(in-process `httptest` servers, temp files cleaned up) and terminates on its own.

**Logging:** all examples log through `pkg/slogger` with typed fields (text formatter on stdout for readability; `20_slogger_setup` shows the JSON production setup).

**Prerequisites:** Go 1.24+. No environment variables or external services.

| #   | Case                  | What it shows                                                                                 | Run                                          |
| --- | --------------------- | --------------------------------------------------------------------------------------------- | -------------------------------------------- |
| 01  | quickstart            | `New().OK()`, message, `mapbuilder` body, generated request ID, pretty JSON                   | `go run ./examples/01_quickstart`            |
| 02  | http_handler          | `WriteJSON` in a handler: pagination, links, request metadata                                 | `go run ./examples/02_http_handler`          |
| 03  | error_stacktrace      | `WithError`, `InjectStackTrace`, `ResolveESTOrderedDoc`                                       | `go run ./examples/03_error_stacktrace`      |
| 04  | signature             | `ApplySignature` / `VerifySignature`, tamper detection                                        | `go run ./examples/04_signature`             |
| 05  | unwrap_json           | `UnwrapJSON`, reading status, body and debug fields                                           | `go run ./examples/05_unwrap_json`           |
| 06  | middleware            | `Logger` + `Recovery` turning a panic into a 500                                              | `go run ./examples/06_middleware`            |
| 07  | dump                  | `DumpJSONTo` / `DumpBodyTo` into a temp dir                                                   | `go run ./examples/07_dump`                  |
| 08  | file_attachment       | `FileAttachment` download with Content-Disposition                                            | `go run ./examples/08_file_attachment`       |
| 09  | cursor_reason_headers | Cursor pagination, reason code, custom fields, debugging, `WithHeader`                        | `go run ./examples/09_cursor_reason_headers` |
| 10  | wrap_options          | `Wrap(...)` functional options vs the fluent builder                                          | `go run ./examples/10_wrap_options`          |
| 11  | base64_binary         | `IsJSONBody`/`IsBinaryBody`, `EncodeBodyBase64`/`DecodeBodyBase64`                            | `go run ./examples/11_base64_binary`         |
| 12  | tracing               | `Span`, `Checkpoint`, `Step`, `Flow`, `Await`                                                 | `go run ./examples/12_tracing`               |
| 13  | markdown_report       | `PagingDoc` and `DumpMDDocTo` Markdown report                                                 | `go run ./examples/13_markdown_report`       |
| 14  | error_responses       | 400/404/204 responses with `WithErrorf`, reasons, `NoContent`                                 | `go run ./examples/14_error_responses`       |
| 15  | rest_crud_api         | Real users API: 201/200/404/409/422/400/204, list with pagination                             | `go run ./examples/15_rest_crud_api`         |
| 16  | signed_webhook        | Sender signs via headers, receiver verifies; tampered payload gets 401                        | `go run ./examples/16_signed_webhook`        |
| 17  | auth_middleware       | Custom 401/403 middleware composed with `Logger` and `Recovery`                               | `go run ./examples/17_auth_middleware`       |
| 18  | paginated_client      | Client walks all pages using `UnwrapJSON` and `Pagination().IsLast()`                         | `go run ./examples/18_paginated_client`      |
| 19  | service_error_mapping | Domain errors mapped to responses with `errors.Is`, 500 with stack trace                      | `go run ./examples/19_service_error_mapping` |
| 20  | slogger_setup         | JSON formatter, bound fields, `Named`, request-scoped `WithContextFields`                     | `go run ./examples/20_slogger_setup`         |
| 21  | stream_quickstart     | `AsStreaming` / `WithStreaming`, `WithWriter`, `Start`, `GetStats`, `GetProgress`             | `go run ./examples/21_stream_quickstart`     |
| 22  | stream_strategies     | `StrategyDirect` and `StrategyChunked` produce identical, ordered output                      | `go run ./examples/22_stream_strategies`     |
| 23  | stream_compression    | `CompressGzip` / `CompressFlate`, compression statistics, decoding the gzip output            | `go run ./examples/23_stream_compression`    |
| 24  | stream_progress       | `WithTotalBytes`, `WithCallback`, `WithHook`, a final `GetProgress` snapshot                  | `go run ./examples/24_stream_progress`       |
| 25  | stream_throttle       | `WithThrottleRate` limiting throughput                                                        | `go run ./examples/25_stream_throttle`       |
| 26  | stream_cancel         | `Cancel()` stopping a running stream, `Errors()` / `HasErrors()`                              | `go run ./examples/26_stream_cancel`         |
| 27  | stream_errors         | nil reader 400, invalid setters 400, unknown strategy 500, mid-stream reader error            | `go run ./examples/27_stream_errors`         |
| 28  | stream_receive        | `WithReceiveMode(true)` reading back an uncompressed stream                                   | `go run ./examples/28_stream_receive`        |
| 29  | stream_http_download  | Streaming a large body from an HTTP handler, checksum verified by the client, request context | `go run ./examples/29_stream_http_download`  |
| 30  | stream_buffer_pool    | `NewBufferPool`, `Get` / `Put`, full-pool discard, `WithBufferPooling`                        | `go run ./examples/30_stream_buffer_pool`    |
| 31  | stream_config         | `NewStreamConfig` defaults, every setter, the `JSON()` helpers                                | `go run ./examples/31_stream_config`         |
| 32  | stream_lifecycle      | `IsStreaming`, `StreamingContext`, `GetWrapper`, `Close`                                      | `go run ./examples/32_stream_lifecycle`      |

## Streaming examples: not covered and known limits

Examples 21 to 32 cover the public streaming API in `stream.go` and run clean under `go run -race`. They avoid the behaviors
below, which are recorded as findings F1 to F13 in `docs/development/stream-refactor-plan.md` (reported, not fixed):

- **`StrategyBuffered` is not shown.** Its workers write to the destination concurrently, so output can be out of order and the
  library's own bookkeeping races (F1, F2).
- **Receiving compressed streams is not shown.** Per-chunk compression members are not re-framed on receive (F5); only the
  uncompressed receive (example 28) round-trips. `CompressDeflate` does not compress (F13); use `CompressFlate`.
- **Callbacks and hooks must not read the progress they are given.** They get the live object, from goroutines that can outlive
  `Start` (example 24 counts calls and reads a snapshot instead) (F1).
- **`TotalChunks` / `CurrentChunk` are not printed.** They are off by one for `Direct` and zero for `Chunked` (F8).
- **`Cancel()` needs `Start(nil)`.** With an explicit context, `Cancel` has no effect (F9); example 26 uses `nil`. A stream whose
  `Read` blocks cannot be cancelled until the read returns.
- **`WithReadTimeout` / `WithWriteTimeout` store values but are not enforced (F10)**, and the throttle wait ignores cancellation (F6b).
- **A receive in which chunks fail to decompress can still return 200 (F11).**

## Running everything

```sh
for d in examples/[0-9]*; do echo "== $d"; go run ./$d || exit 1; done
```
