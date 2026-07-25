// Package replify provides a structured, high-level toolkit for building and
// consuming HTTP API responses in Go. It ships with:
//
//   - A fluent [wrapper] / [R] type for constructing and inspecting API responses
//   - First-class JSON parsing, normalisation, and field-level access
//   - Pre-built response helpers for every standard HTTP status code
//   - Pagination, metadata, and header sub-structures
//   - Chunked / buffered / direct data streaming with progress tracking and compression
//   - Stack-aware error wrapping compatible with the standard errors package
//   - A rich constants catalogue (HTTP headers, media types, locales)
//
// # Installation
//
//	go get github.com/polarixa/replify
//
// # Getting Started
//
// The quickest way to create a successful response:
//
//	w := replify.WrapOk("Users retrieved", users)
//	json.NewEncoder(rw).Encode(w.JSON())
//
// # Response Construction
//
// Use [New] plus the fluent With* methods to build fully-featured responses:
//
//	w := replify.New().
//	    WithStatusCode(http.StatusOK).
//	    WithMessage("Resource retrieved").
//	    WithBody(payload).
//	    WithPagination(replify.FromPages(120, 10).WithPage(1)).
//	    WithMeta(replify.Meta().
//	        WithApiVersion("v1.0.0").
//	        WithLocale("en_US").
//	        WithRequestID("req_abc123"),
//	    ).
//	    WithHeader(replify.OK)
//
// Render the result:
//
//	fmt.Println(w.JSON())         // compact JSON string
//	fmt.Println(w.JSONPretty())   // indented JSON string
//	fmt.Println(w.StatusCode())   // 200
//	fmt.Println(w.Message())      // "Resource retrieved"
//
// # Pre-built HTTP Status Helpers
//
// Every standard HTTP status has a dedicated constructor that sets the correct
// status code and header automatically:
//
//	replify.WrapOk("OK", data)                     // 200
//	replify.WrapCreated("Created", data)           // 201
//	replify.WrapAccepted("Accepted", data)         // 202
//	replify.WrapNoContent("No Content", nil)       // 204
//	replify.WrapBadRequest("Bad Request", nil)     // 400
//	replify.WrapUnauthorized("Unauthorized", nil)  // 401
//	replify.WrapForbidden("Forbidden", nil)        // 403
//	replify.WrapNotFound("Not Found", nil)         // 404
//	replify.WrapConflict("Conflict", nil)          // 409
//	replify.WrapUnprocessableEntity("Invalid", errs) // 422
//	replify.WrapTooManyRequest("Rate limited", nil)  // 429
//	replify.WrapInternalServerError("Error", nil)    // 500
//	replify.WrapServiceUnavailable("Down", nil)      // 503
//	replify.WrapGatewayTimeout("Timeout", nil)       // 504
//
// # Parsing a JSON API Response
//
// [UnwrapJSON] normalises (strips comments, trailing commas) and parses a raw
// JSON string into a [wrapper], giving typed access to every standard field:
//
//	jsonStr := `{
//	    "status_code": 200,
//	    "message":     "OK",
//	    "path":        "/api/v1/users",
//	    "data": [
//	        {"id": "u1", "username": "alice"},
//	        {"id": "u2", "username": "bob"}
//	    ],
//	    "pagination": {
//	        "page": 1, "per_page": 2,
//	        "total_items": 100, "total_pages": 50,
//	        "is_last": false
//	    },
//	    "meta": {
//	        "request_id":     "req_abc",
//	        "api_version":    "v1.0.0",
//	        "locale":         "en_US",
//	        "requested_time": "2026-03-09T07:00:00Z"
//	    }
//	}`
//
//	w, err := replify.UnwrapJSON(jsonStr)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(w.StatusCode())                                        // 200
//	fmt.Println(w.Pagination().TotalItems())                           // 100
//	fmt.Println(w.JSONBodyParser().Get("0").Get("username").String())  // "alice"
//
// For map-based input use [WrapFrom]:
//
//	m := map[string]any{
//	    "status_code": 201,
//	    "message":     "Created",
//	    "data":        map[string]any{"id": "new_001"},
//	}
//	w, err := replify.WrapFrom(m)
//
// # Pagination
//
// Create and attach pagination using [Pages] or the convenience constructor [FromPages]:
//
//	p := replify.FromPages(500, 20). // 500 total items, 20 per page
//	    WithPage(3)
//
//	w := replify.WrapOk("Users", users).WithPagination(p)
//	fmt.Println(w.Pagination().TotalPages()) // 25
//	fmt.Println(w.Pagination().IsLast())     // false
//
// # Metadata
//
// Attach API metadata to any response:
//
//	m := replify.Meta().
//	    WithApiVersion("v2.1.0").
//	    WithLocale("vi_VN").
//	    WithRequestID("req_xyz").
//	    WithCustomField("trace_id", "abc123").
//	    WithCustomField("region",   "us-east-1")
//
//	w := replify.WrapOk("OK", data).WithMeta(m)
//	fmt.Println(w.Meta().ApiVersion())  // "v2.1.0"
//
// # HTTP Headers and Status Codes
//
// Pre-built header singletons cover all standard HTTP/1.1 and WebDAV codes:
//
//	replify.OK                    // 200 Successful
//	replify.Created               // 201 Successful
//	replify.NotFound              // 404 Client Error
//	replify.InternalServerError   // 500 Server Error
//	replify.TooManyRequests       // 429 Client Error
//
// Inspect a pre-built header:
//
//	h := replify.NotFound
//	fmt.Println(h.Code())        // 404
//	fmt.Println(h.Text())        // "Not Found"
//	fmt.Println(h.Type())        // "Client Error"
//
// Build a custom header with the fluent API:
//
//	h := replify.Header().
//	    WithCode(422).
//	    WithText("Validation Error").
//	    WithType("Client Error").
//	    WithDescription("One or more fields failed validation.")
//
// # HTTP Header Name Constants
//
// All standard HTTP header names are available as typed constants:
//
//	replify.HeaderAuthorization   // "Authorization"
//	replify.HeaderContentType     // "Content-Type"
//	replify.HeaderAccept          // "Accept"
//	replify.HeaderCacheControl    // "Cache-Control"
//	replify.HeaderXRequestedWith  // "X-Requested-With"
//
// # Media Type Constants
//
// Common MIME types are available as constants:
//
//	replify.MediaTypeApplicationJSON     // "application/json"
//	replify.MediaTypeApplicationJSONUTF8 // "application/json; charset=utf-8"
//	replify.MediaTypeTextPlain           // "text/plain"
//	replify.MediaTypeMultipartFormData   // "multipart/form-data"
//
// # Locale Constants
//
// IETF-style locale identifiers for content localisation:
//
//	replify.LocaleEnUS  // "en_US"
//	replify.LocaleViVN  // "vi_VN"
//	replify.LocaleZhCN  // "zh_CN"
//	replify.LocaleJaJP  // "ja_JP"
//
// # Streaming
//
// [NewStreaming] creates a [StreamingWrapper] that streams data from any
// [io.Reader] with configurable chunking, compression, progress hooks, and
// context-aware cancellation:
//
//	cfg := replify.NewStreamConfig()
//	cfg.Strategy    = replify.StrategyChunked
//	cfg.Compression = replify.CompressGzip
//	cfg.ChunkSize   = 128 * 1024 // 128 KB
//
//	sw := replify.NewStreaming(reader, cfg)
//	sw.WithStreamingCallback(func(p *replify.StreamProgress, err error) {
//	    fmt.Printf("%.0f%%  %d B/s\n", float64(p.Percentage), p.TransferRate)
//	})
//
//	if err := sw.Stream(writer); err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(sw.Stats().CompressionRatio) // e.g. 0.12 (88% reduction)
//
// Streaming strategies:
//
//	replify.StrategyDirect    // write bytes immediately as they arrive
//	replify.StrategyBuffered  // collect in an internal buffer (default)
//	replify.StrategyChunked   // split into fixed-size chunks
//
// Supported compression algorithms:
//
//	replify.CompressNone     // no compression (default)
//	replify.CompressGzip     // gzip
//	replify.CompressDeflate  // deflate
//	replify.CompressFlate    // flate
//
// # Error Handling
//
// replify provides stack-trace-aware error construction compatible with the
// standard [errors] package:
//
//	// Create a new error with a stack trace.
//	err := replify.NewError("something went wrong")
//
//	// Format an error with printf-style args.
//	err = replify.NewErrorf("user %q not found", userID)
//
//	// Wrap an existing error, preserving its message and adding a stack trace.
//	err = replify.NewErrorAck(originalErr)
//
//	// Inspect the stack trace.
//	var st replify.StackTrace
//	if errors.As(err, &st) {
//	    fmt.Printf("%+v\n", st)
//	}
//
// # R — High-Level Wrapper Alias
//
// [R] is a thin alias over [wrapper]. It is returned by streaming hooks and
// can be used wherever a [wrapper] is accepted:
//
//	sw.WithStreamingHook(func(p *replify.StreamProgress, r *replify.R) {
//	    fmt.Println(r.StatusCode(), p.Percentage)
//	})
//
// # Buffer Pool
//
// [NewBufferPool] provides a reusable byte-buffer pool to reduce GC pressure
// during high-throughput streaming:
//
//	pool := replify.NewBufferPool(64*1024, 8) // 8 × 64 KB buffers
//	cfg  := replify.NewStreamConfig()
//	cfg.UseBufferPool = true
//
// # Toolbox
//
// The package also exposes a [Toolbox] variable for utility operations that do
// not fit the fluent response API:
//
//	replify.Toolbox // tools{}
package replify
