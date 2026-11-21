# replify examples

Runnable programs showing the public `replify` API. Each one is self-contained
(in-process `httptest` servers, temp files cleaned up) and terminates on its own.

**Logging:** all examples log through `pkg/slogger` with typed fields (text formatter on stdout for readability; `20_slogger_setup` shows the JSON production setup).

**Prerequisites:** Go 1.24+. No environment variables or external services.

| #   | Case                  | What it shows                                                               | Run                                          |
| --- | --------------------- | --------------------------------------------------------------------------- | -------------------------------------------- |
| 01  | quickstart            | `New().OK()`, message, `mapbuilder` body, generated request ID, pretty JSON | `go run ./examples/01_quickstart`            |
| 02  | http_handler          | `WriteJSON` in a handler: pagination, links, request metadata               | `go run ./examples/02_http_handler`          |
| 03  | error_stacktrace      | `WithError`, `InjectStackTrace`, `ResolveESTOrderedDoc`                     | `go run ./examples/03_error_stacktrace`      |
| 04  | signature             | `ApplySignature` / `VerifySignature`, tamper detection                      | `go run ./examples/04_signature`             |
| 05  | unwrap_json           | `UnwrapJSON`, reading status, body and debug fields                         | `go run ./examples/05_unwrap_json`           |
| 06  | middleware            | `Logger` + `Recovery` turning a panic into a 500                            | `go run ./examples/06_middleware`            |
| 07  | dump                  | `DumpJSONTo` / `DumpBodyTo` into a temp dir                                 | `go run ./examples/07_dump`                  |
| 08  | file_attachment       | `FileAttachment` download with Content-Disposition                          | `go run ./examples/08_file_attachment`       |
| 09  | cursor_reason_headers | Cursor pagination, reason code, custom fields, debugging, `WithHeader`      | `go run ./examples/09_cursor_reason_headers` |
| 10  | wrap_options          | `Wrap(...)` functional options vs the fluent builder                        | `go run ./examples/10_wrap_options`          |
| 11  | base64_binary         | `IsJSONBody`/`IsBinaryBody`, `EncodeBodyBase64`/`DecodeBodyBase64`          | `go run ./examples/11_base64_binary`         |
| 12  | tracing               | `Span`, `Checkpoint`, `Step`, `Flow`, `Await`                               | `go run ./examples/12_tracing`               |
| 13  | markdown_report       | `PagingDoc` and `DumpMDDocTo` Markdown report                               | `go run ./examples/13_markdown_report`       |
| 14  | error_responses       | 400/404/204 responses with `WithErrorf`, reasons, `NoContent`               | `go run ./examples/14_error_responses`       |
| 15  | rest_crud_api         | Real users API: 201/200/404/409/422/400/204, list with pagination           | `go run ./examples/15_rest_crud_api`         |
| 16  | signed_webhook        | Sender signs via headers, receiver verifies; tampered payload gets 401      | `go run ./examples/16_signed_webhook`        |
| 17  | auth_middleware       | Custom 401/403 middleware composed with `Logger` and `Recovery`             | `go run ./examples/17_auth_middleware`       |
| 18  | paginated_client      | Client walks all pages using `UnwrapJSON` and `Pagination().IsLast()`       | `go run ./examples/18_paginated_client`      |
| 19  | service_error_mapping | Domain errors mapped to responses with `errors.Is`, 500 with stack trace    | `go run ./examples/19_service_error_mapping` |
| 20  | slogger_setup         | JSON formatter, bound fields, `Named`, request-scoped `WithContextFields`   | `go run ./examples/20_slogger_setup`         |

## Running everything

```sh
for d in examples/[0-9]*; do echo "== $d"; go run ./$d || exit 1; done
```
