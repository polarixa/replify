package replify

// Standard HTTP headers related to content negotiation and encoding.
const (
	// Accept specifies the media types that are acceptable for the response.
	// 	Example: "application/json, text/html"
	HeaderAccept HeaderType = "Accept"

	// AcceptCharset specifies the character sets that are acceptable.
	// 	Example: "utf-8, iso-8859-1"
	HeaderAcceptCharset HeaderType = "Accept-Charset"

	// AcceptEncoding specifies the content encodings that are acceptable.
	//	Example: "gzip, deflate, br"
	HeaderAcceptEncoding HeaderType = "Accept-Encoding"

	// AcceptLanguage specifies the acceptable languages for the response.
	// 	Example: "en-US, en;q=0.9, fr;q=0.8"
	HeaderAcceptLanguage HeaderType = "Accept-Language"

	// Authorization contains the credentials for authenticating the client with the server.
	// 	Example: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6..."
	HeaderAuthorization HeaderType = "Authorization"

	// CacheControl specifies directives for caching mechanisms in both requests and responses.
	// 	Example: "no-cache, no-store, must-revalidate"
	HeaderCacheControl HeaderType = "Cache-Control"

	// ContentDisposition specifies if the content should be displayed inline or treated as an attachment.
	// 	Example: "attachment; filename=\"document.pdf\""
	HeaderContentDisposition HeaderType = "Content-Disposition"

	// ContentEncoding specifies the encoding transformations that have been applied to the body of the response.
	// 	Example: "gzip"
	HeaderContentEncoding HeaderType = "Content-Encoding"

	// ContentLength specifies the size of the response body in octets.
	// 	Example: "1024"
	HeaderContentLength HeaderType = "Content-Length"

	// ContentType specifies the media type of the resource.
	// 	Example: "application/json; charset=utf-8"
	HeaderContentType HeaderType = "Content-Type"

	// Cookie contains stored HTTP cookies sent to the server by the client.
	// 	Example: "sessionId=abc123; userId=456"
	HeaderCookie HeaderType = "Cookie"

	// Host specifies the domain name of the server (for virtual hosting) and the TCP port number.
	// 	Example: "www.example.com:8080"
	HeaderHost HeaderType = "Host"

	// Origin specifies the origin of the cross-origin request or preflight request.
	// 	Example: "https://www.example.com"
	HeaderOrigin HeaderType = "Origin"

	// Referer contains the address of the previous web page from which a link to the currently requested page was followed.
	// 	Example: "https://www.example.com/page1.html"
	HeaderReferer HeaderType = "Referer"

	// UserAgent contains information about the user agent (browser or client) making the request.
	// 	Example: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	HeaderUserAgent HeaderType = "User-Agent"

	// IfMatch makes the request conditional on the target resource having the same entity tag as the one provided.
	// 	Example: "\"686897696a7c876b7e\""
	HeaderIfMatch HeaderType = "If-Match"

	// IfNoneMatch makes the request conditional on the target resource not having the same entity tag as the one provided.
	// 	Example: "\"686897696a7c876b7e\""
	HeaderIfNoneMatch HeaderType = "If-None-Match"

	// ETag provides the entity tag for the resource.
	// 	Example: "\"33a64df551425fcc55e4d42a148795d9f25f89d4\""
	HeaderETag HeaderType = "ETag"

	// LastModified specifies the last modified date of the resource.
	// 	Example: "Wed, 21 Oct 2015 07:28:00 GMT"
	HeaderLastModified HeaderType = "Last-Modified"

	// Location specifies the URL to redirect a client to.
	// 	Example: "https://www.example.com/new-location"
	HeaderLocation HeaderType = "Location"

	// Pragma specifies implementation-specific directives that might affect caching.
	// 	Example: "no-cache"
	HeaderPragma HeaderType = "Pragma"

	// RetryAfter specifies the time after which the client should retry the request after receiving a 503 Service Unavailable status code.
	// 	Example: "120" or "Fri, 07 Nov 2014 23:59:59 GMT"
	HeaderRetryAfter HeaderType = "Retry-After"

	// Server contains information about the software used by the origin server to handle the request.
	// 	Example: "Apache/2.4.41 (Ubuntu)"
	HeaderServer HeaderType = "Server"

	// WWWAuthenticate indicates that the client must authenticate to access the requested resource.
	// 	Example: "Basic realm=\"Access to staging site\""
	HeaderWWWAuthenticate HeaderType = "WWW-Authenticate"

	// Date specifies the date and time at which the message was sent.
	// 	Example: "Tue, 15 Nov 1994 08:12:31 GMT"
	HeaderDate HeaderType = "Date"

	// Expires specifies the date/time after which the response is considered stale.
	// 	Example: "Thu, 01 Dec 1994 16:00:00 GMT"
	HeaderExpires HeaderType = "Expires"

	// Age specifies the age of the response in seconds.
	// 	Example: "3600"
	HeaderAge HeaderType = "Age"

	// Connection specifies control options for the current connection (e.g., keep-alive or close).
	// 	Example: "keep-alive"
	HeaderConnection HeaderType = "Connection"

	// ContentLanguage specifies the language of the content.
	// 	Example: "en-US"
	HeaderContentLanguage HeaderType = "Content-Language"

	// Forwarded contains information about intermediate proxies or gateways that have forwarded the request.
	// 	Example: "for=192.0.2.60;proto=http;by=203.0.113.43"
	HeaderForwarded HeaderType = "Forwarded"

	// IfModifiedSince makes the request conditional on the target resource being modified since the specified date.
	// 	Example: "Wed, 21 Oct 2015 07:28:00 GMT"
	HeaderIfModifiedSince HeaderType = "If-Modified-Since"

	// Upgrade requests the server to switch to a different protocol.
	// 	Example: "websocket"
	HeaderUpgrade HeaderType = "Upgrade"

	// Via provides information about intermediate protocols and recipients between the user agent and the server.
	// 	Example: "1.1 proxy1.example.com, 1.0 proxy2.example.org"
	HeaderVia HeaderType = "Via"

	// Warning carries additional information about the status or transformation of a message.
	// 	Example: "110 anderson/1.3.37 \"Response is stale\""
	HeaderWarning HeaderType = "Warning"

	// XForwardedFor contains the originating IP address of a client connecting to a web server through an HTTP proxy or load balancer.
	// 	Example: "203.0.113.195, 70.41.3.18, 150.172.238.178"
	HeaderXForwardedFor HeaderType = "X-Forwarded-For"

	// XForwardedHost contains the original host requested by the client in the Host HTTP request header.
	// 	Example: "example.com"
	HeaderXForwardedHost HeaderType = "X-Forwarded-Host"

	// XForwardedProto specifies the protocol (HTTP or HTTPS) used by the client.
	// 	Example: "https"
	HeaderXForwardedProto HeaderType = "X-Forwarded-Proto"

	// XRequestedWith identifies the type of request being made (e.g., Ajax requests).
	// 	Example: "XMLHttpRequest"
	HeaderXRequestedWith HeaderType = "X-Requested-With"

	// XFrameOptions specifies whether the browser should be allowed to render the page in a <frame>, <iframe>, <object>, <embed>, or <applet>.
	// 	Example: "DENY" or "SAMEORIGIN"
	HeaderXFrameOptions HeaderType = "X-Frame-Options"

	// XXSSProtection controls browser's built-in XSS (Cross-Site Scripting) filter.
	// 	Example: "1; mode=block"
	HeaderXXSSProtection HeaderType = "X-XSS-Protection"

	// XContentTypeOpts prevents browsers from interpreting files as a different MIME type than what is specified.
	// 	Example: "nosniff"
	HeaderXContentTypeOpts HeaderType = "X-Content-Type-Options"

	// ContentSecurity specifies security policy for web applications, helping to prevent certain types of attacks.
	// 	Example: "default-src 'self'; script-src 'self' 'unsafe-inline'"
	HeaderContentSecurity HeaderType = "Content-Security-Policy"

	// StrictTransport enforces the use of HTTPS for the website to reduce security risks.
	// 	Example: "max-age=31536000; includeSubDomains"
	HeaderStrictTransport HeaderType = "Strict-Transport-Security"

	// PublicKeyPins specifies public key pins to prevent man-in-the-middle attacks.
	// 	Example: "pin-sha256=\"base64+primary==\"; pin-sha256=\"base64+backup==\"; max-age=5184000"
	HeaderPublicKeyPins HeaderType = "Public-Key-Pins"

	// ExpectCT allows websites to specify a Certificate Transparency policy.
	// 	Example: "max-age=86400, enforce"
	HeaderExpectCT HeaderType = "Expect-CT"

	// AccessControlAllowOrigin specifies which domains are allowed to access the resources.
	// 	Example: "*" or "https://example.com"
	HeaderAccessControlAllowOrigin HeaderType = "Access-Control-Allow-Origin"

	// AccessControlAllowMethods specifies which HTTP methods are allowed when accessing the resource.
	// 	Example: "GET, POST, PUT, DELETE"
	HeaderAccessControlAllowMethods HeaderType = "Access-Control-Allow-Methods"

	// AccessControlAllowHeaders specifies which HTTP headers can be used during the actual request.
	// 	Example: "Content-Type, Authorization"
	HeaderAccessControlAllowHeaders HeaderType = "Access-Control-Allow-Headers"

	// AccessControlMaxAge specifies how long the results of a preflight request can be cached.
	// 	Example: "86400"
	HeaderAccessControlMaxAge HeaderType = "Access-Control-Max-Age"

	// AccessControlExposeHeaders specifies which headers can be exposed as part of the response.
	// 	Example: "Content-Length, X-JSON"
	HeaderAccessControlExposeHeaders HeaderType = "Access-Control-Expose-Headers"

	// AccessControlRequestMethod indicates which HTTP method will be used during the actual request.
	// 	Example: "POST"
	HeaderAccessControlRequestMethod HeaderType = "Access-Control-Request-Method"

	// AccessControlRequestHeaders specifies which headers can be sent with the actual request.
	// 	Example: "Content-Type, X-Custom-Header"
	HeaderAccessControlRequestHeaders HeaderType = "Access-Control-Request-Headers"

	// AcceptPatch specifies which patch document formats are acceptable in the response.
	// 	Example: "application/json-patch+json"
	HeaderAcceptPatch HeaderType = "Accept-Patch"

	// DeltaBase specifies the URI of the delta information.
	// 	Example: "\"abc123\""
	HeaderDeltaBase HeaderType = "Delta-Base"

	// IfUnmodifiedSince makes the request conditional on the resource not being modified since the specified date.
	// 	Example: "Wed, 21 Oct 2015 07:28:00 GMT"
	HeaderIfUnmodifiedSince HeaderType = "If-Unmodified-Since"

	// AcceptRanges specifies the range of the resource that the client is requesting.
	// 	Example: "bytes"
	HeaderAcceptRanges HeaderType = "Accept-Ranges"

	// ContentRange specifies the range of the resource being sent in the response.
	// 	Example: "bytes 200-1000/5000"
	HeaderContentRange HeaderType = "Content-Range"

	// Allow specifies the allowed methods for a resource.
	// 	Example: "GET, HEAD, PUT"
	HeaderAllow HeaderType = "Allow"

	// AccessControlAllowCredentials indicates whether the response to the request can expose credentials.
	// 	Example: "true"
	HeaderAccessControlAllowCredentials HeaderType = "Access-Control-Allow-Credentials"

	// XCSRFToken is used to prevent Cross-Site Request Forgery (CSRF) attacks.
	// 	Example: "i8XNjC4b8KVok4uw5RftR38Wgp2BF"
	HeaderXCSRFToken HeaderType = "X-CSRF-Token"

	// XRealIP contains the real IP address of the client, often used in proxies or load balancers.
	// 	Example: "203.0.113.195"
	HeaderXRealIP HeaderType = "X-Real-IP"

	// ContentSecurityPolicy specifies content security policies to prevent certain attacks.
	// 	Example: "default-src 'self'; img-src *; media-src media1.com media2.com"
	HeaderContentSecurityPolicy HeaderType = "Content-Security-Policy"

	// ReferrerPolicy controls how much information about the referring page is sent.
	// 	Example: "no-referrer-when-downgrade"
	HeaderReferrerPolicy HeaderType = "Referrer-Policy"

	// ExpectCt specifies a Certificate Transparency policy for the web server.
	// 	Example: "max-age=86400, enforce, report-uri=\"https://example.com/report\""
	HeaderExpectCt HeaderType = "Expect-CT"

	// StrictTransportSecurity enforces HTTPS to reduce the chance of security breaches.
	// 	Example: "max-age=63072000; includeSubDomains; preload"
	HeaderStrictTransportSecurity HeaderType = "Strict-Transport-Security"

	// UpgradeInsecureRequests requests the browser to upgrade any insecure requests to secure HTTPS requests.
	// 	Example: "1"
	HeaderUpgradeInsecureRequests HeaderType = "Upgrade-Insecure-Requests"

	// X-Signature contains the signature of the request, often used for verifying the integrity and authenticity of the request.
	// 	Example: "HMAC-SHA256=base64encodedhash"
	HeaderXSignature HeaderType = "X-Signature"

	// X-Signature-Algorithm specifies the algorithm used to generate the signature in the X-Signature header.
	// 	Example: "HMAC-SHA256"
	HeaderXSignatureAlgorithm HeaderType = "X-Signature-Algorithm"

	// X-Signature-Timestamp specifies the timestamp when the signature was generated.
	// 	Example: "1790442053"
	HeaderXSignatureTimestamp HeaderType = "X-Signature-Timestamp"
)

// Media Type constants define commonly used MIME types for different content types in HTTP requests and responses.
const (
	// ApplicationJSON specifies that the content is JSON-formatted data.
	//  Example: "application/json"
	MediaTypeApplicationJSON MediaType = "application/json"

	// ApplicationJSONUTF8 specifies that the content is JSON-formatted data with UTF-8 character encoding.
	//  Example: "application/json; charset=utf-8"
	MediaTypeApplicationJSONUTF8 MediaType = "application/json; charset=utf-8"

	// ApplicationXML specifies that the content is XML-formatted data.
	//  Example: "application/xml"
	MediaTypeApplicationXML MediaType = "application/xml"

	// ApplicationForm specifies that the content is URL-encoded form data.
	//  Example: "application/x-www-form-urlencoded"
	MediaTypeApplicationForm MediaType = "application/x-www-form-urlencoded"

	// ApplicationOctetStream specifies that the content is binary data (not interpreted by the browser).
	//  Example: "application/octet-stream"
	MediaTypeApplicationOctetStream MediaType = "application/octet-stream"

	// TextPlain specifies that the content is plain text.
	//  Example: "text/plain"
	MediaTypeTextPlain MediaType = "text/plain"

	// TextHTML specifies that the content is HTML-formatted data.
	//  Example: "text/html"
	MediaTypeTextHTML MediaType = "text/html"

	// ImageJPEG specifies that the content is a JPEG image.
	//  Example: "image/jpeg"
	MediaTypeImageJPEG MediaType = "image/jpeg"

	// ImagePNG specifies that the content is a PNG image.
	//  Example: "image/png"
	MediaTypeImagePNG MediaType = "image/png"

	// ImageGIF specifies that the content is a GIF image.
	//  Example: "image/gif"
	MediaTypeImageGIF MediaType = "image/gif"

	// AudioMP3 specifies that the content is an MP3 audio file.
	//  Example: "audio/mpeg"
	MediaTypeAudioMP3 MediaType = "audio/mpeg"

	// AudioWAV specifies that the content is a WAV audio file.
	//  Example: "audio/wav"
	MediaTypeAudioWAV MediaType = "audio/wav"

	// VideoMP4 specifies that the content is an MP4 video file.
	//  Example: "video/mp4"
	MediaTypeVideoMP4 MediaType = "video/mp4"

	// VideoAVI specifies that the content is an AVI video file.
	//  Example: "video/x-msvideo"
	MediaTypeVideoAVI MediaType = "video/x-msvideo"

	// ApplicationPDF specifies that the content is a PDF file.
	//  Example: "application/pdf"
	MediaTypeApplicationPDF MediaType = "application/pdf"

	// ApplicationMSWord specifies that the content is a Microsoft Word document (.doc).
	//  Example: "application/msword"
	MediaTypeApplicationMSWord MediaType = "application/msword"

	// ApplicationMSPowerPoint specifies that the content is a Microsoft PowerPoint presentation (.ppt).
	//  Example: "application/vnd.ms-powerpoint"
	MediaTypeApplicationMSPowerPoint MediaType = "application/vnd.ms-powerpoint"

	// ApplicationExcel specifies that the content is a Microsoft Excel spreadsheet (.xls).
	//  Example: "application/vnd.ms-excel"
	MediaTypeApplicationExcel MediaType = "application/vnd.ms-excel"

	// ApplicationZip specifies that the content is a ZIP archive.
	//  Example: "application/zip"
	MediaTypeApplicationZip MediaType = "application/zip"

	// ApplicationGzip specifies that the content is a GZIP-compressed file.
	//  Example: "application/gzip"
	MediaTypeApplicationGzip MediaType = "application/gzip"

	// MultipartFormData specifies that the content is a multipart form, typically used for file uploads.
	//  Example: "multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW"
	MediaTypeMultipartFormData MediaType = "multipart/form-data"

	// ImageBMP specifies that the content is a BMP image.
	//  Example: "image/bmp"
	MediaTypeImageBMP MediaType = "image/bmp"

	// ImageTIFF specifies that the content is a TIFF image.
	//  Example: "image/tiff"
	MediaTypeImageTIFF MediaType = "image/tiff"

	// TextCSS specifies that the content is CSS (Cascading Style Sheets).
	//  Example: "text/css"
	MediaTypeTextCSS MediaType = "text/css"

	// TextJavaScript specifies that the content is JavaScript code.
	//  Example: "text/javascript"
	MediaTypeTextJavaScript MediaType = "text/javascript"

	// ApplicationJSONLD specifies that the content is a JSON-LD (JSON for Linked Data) document.
	//  Example: "application/ld+json"
	MediaTypeApplicationJSONLD MediaType = "application/ld+json"

	// ApplicationRDFXML specifies that the content is in RDF (Resource Description Framework) XML format.
	//  Example: "application/rdf+xml"
	MediaTypeApplicationRDFXML MediaType = "application/rdf+xml"

	// ApplicationGeoJSON specifies that the content is a GeoJSON (geostatics data) document.
	//  Example: "application/geo+json"
	MediaTypeApplicationGeoJSON MediaType = "application/geo+json"

	// ApplicationMsgpack specifies that the content is in MessagePack format (binary JSON).
	//  Example: "application/msgpack"
	MediaTypeApplicationMsgpack MediaType = "application/msgpack"

	// ApplicationOgg specifies that the content is an Ogg multimedia container format.
	//  Example: "application/ogg"
	MediaTypeApplicationOgg MediaType = "application/ogg"

	// ApplicationGraphQL specifies that the content is in GraphQL format.
	//  Example: "application/graphql"
	MediaTypeApplicationGraphQL MediaType = "application/graphql"

	// ApplicationProtobuf specifies that the content is in Protocol Buffers format (binary serialization).
	//  Example: "application/protobuf"
	MediaTypeApplicationProtobuf MediaType = "application/protobuf"

	// ImageWebP specifies that the content is a WebP image.
	//  Example: "image/webp"
	MediaTypeImageWebP MediaType = "image/webp"

	// FontWOFF specifies that the content is a WOFF (Web Open Font Format) font.
	//  Example: "font/woff"
	MediaTypeFontWOFF MediaType = "font/woff"

	// FontWOFF2 specifies that the content is a WOFF2 (Web Open Font Format 2) font.
	//  Example: "font/woff2"
	MediaTypeFontWOFF2 MediaType = "font/woff2"

	// AudioFLAC specifies that the content is a FLAC audio file (Free Lossless Audio Codec).
	//  Example: "audio/flac"
	MediaTypeAudioFLAC MediaType = "audio/flac"

	// VideoWebM specifies that the content is a WebM video file.
	//  Example: "video/webm"
	MediaTypeVideoWebM MediaType = "video/webm"

	// ApplicationDart specifies that the content is a Dart programming language file.
	//  Example: "application/dart"
	MediaTypeApplicationDart MediaType = "application/dart"

	// ApplicationXLSX specifies that the content is an Excel file in XLSX format.
	//  Example: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	MediaTypeApplicationXLSX MediaType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

	// ApplicationPPTX specifies that the content is a PowerPoint file in PPTX format.
	//  Example: "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	MediaTypeApplicationPPTX MediaType = "application/vnd.openxmlformats-officedocument.presentationml.presentation"

	// ApplicationGRPC specifies that the content is in gRPC format (a high-performance RPC framework).
	//  Example: "application/grpc"
	MediaTypeApplicationGRPC MediaType = "application/grpc"
)

// HTTP status code constants typed as [StatusCode].
// These mirror every standard constant from [net/http] so callers can write
// replify.WithStatusCode(replify.StatusOK) with full compile-time safety.
const (
	// 1xx — Informational

	// StatusContinue maps to HTTP 100.
	StatusContinue StatusCode = 100

	// StatusSwitchingProtocols maps to HTTP 101.
	StatusSwitchingProtocols StatusCode = 101

	// StatusProcessing maps to HTTP 102.
	StatusProcessing StatusCode = 102

	// StatusEarlyHints maps to HTTP 103.
	StatusEarlyHints StatusCode = 103

	// 2xx — Success

	// StatusOK maps to HTTP 200.
	StatusOK StatusCode = 200

	// StatusCreated maps to HTTP 201.
	StatusCreated StatusCode = 201

	// StatusAccepted maps to HTTP 202.
	StatusAccepted StatusCode = 202

	// StatusNonAuthoritativeInfo maps to HTTP 203.
	StatusNonAuthoritativeInfo StatusCode = 203

	// StatusNoContent maps to HTTP 204.
	StatusNoContent StatusCode = 204

	// StatusResetContent maps to HTTP 205.
	StatusResetContent StatusCode = 205

	// StatusPartialContent maps to HTTP 206.
	StatusPartialContent StatusCode = 206

	// StatusMultiStatus maps to HTTP 207.
	StatusMultiStatus StatusCode = 207

	// StatusAlreadyReported maps to HTTP 208.
	StatusAlreadyReported StatusCode = 208

	// StatusIMUsed maps to HTTP 226.
	StatusIMUsed StatusCode = 226

	// 3xx — Redirection

	// StatusMultipleChoices maps to HTTP 300.
	StatusMultipleChoices StatusCode = 300

	// StatusMovedPermanently maps to HTTP 301.
	StatusMovedPermanently StatusCode = 301

	// StatusFound maps to HTTP 302.
	StatusFound StatusCode = 302

	// StatusSeeOther maps to HTTP 303.
	StatusSeeOther StatusCode = 303

	// StatusNotModified maps to HTTP 304.
	StatusNotModified StatusCode = 304

	// StatusUseProxy maps to HTTP 305.
	StatusUseProxy StatusCode = 305

	// StatusTemporaryRedirect maps to HTTP 307.
	StatusTemporaryRedirect StatusCode = 307

	// StatusPermanentRedirect maps to HTTP 308.
	StatusPermanentRedirect StatusCode = 308

	// 4xx — Client Error

	// StatusBadRequest maps to HTTP 400.
	StatusBadRequest StatusCode = 400

	// StatusUnauthorized maps to HTTP 401.
	StatusUnauthorized StatusCode = 401

	// StatusPaymentRequired maps to HTTP 402.
	StatusPaymentRequired StatusCode = 402

	// StatusForbidden maps to HTTP 403.
	StatusForbidden StatusCode = 403

	// StatusNotFound maps to HTTP 404.
	StatusNotFound StatusCode = 404

	// StatusMethodNotAllowed maps to HTTP 405.
	StatusMethodNotAllowed StatusCode = 405

	// StatusNotAcceptable maps to HTTP 406.
	StatusNotAcceptable StatusCode = 406

	// StatusProxyAuthRequired maps to HTTP 407.
	StatusProxyAuthRequired StatusCode = 407

	// StatusRequestTimeout maps to HTTP 408.
	StatusRequestTimeout StatusCode = 408

	// StatusConflict maps to HTTP 409.
	StatusConflict StatusCode = 409

	// StatusGone maps to HTTP 410.
	StatusGone StatusCode = 410

	// StatusLengthRequired maps to HTTP 411.
	StatusLengthRequired StatusCode = 411

	// StatusPreconditionFailed maps to HTTP 412.
	StatusPreconditionFailed StatusCode = 412

	// StatusRequestEntityTooLarge maps to HTTP 413.
	StatusRequestEntityTooLarge StatusCode = 413

	// StatusRequestURITooLong maps to HTTP 414.
	StatusRequestURITooLong StatusCode = 414

	// StatusUnsupportedMediaType maps to HTTP 415.
	StatusUnsupportedMediaType StatusCode = 415

	// StatusRequestedRangeNotSatisfiable maps to HTTP 416.
	StatusRequestedRangeNotSatisfiable StatusCode = 416

	// StatusExpectationFailed maps to HTTP 417.
	StatusExpectationFailed StatusCode = 417

	// StatusTeapot maps to HTTP 418.
	StatusTeapot StatusCode = 418

	// StatusMisdirectedRequest maps to HTTP 421.
	StatusMisdirectedRequest StatusCode = 421

	// StatusUnprocessableEntity maps to HTTP 422.
	StatusUnprocessableEntity StatusCode = 422

	// StatusLocked maps to HTTP 423.
	StatusLocked StatusCode = 423

	// StatusFailedDependency maps to HTTP 424.
	StatusFailedDependency StatusCode = 424

	// StatusTooEarly maps to HTTP 425.
	StatusTooEarly StatusCode = 425

	// StatusUpgradeRequired maps to HTTP 426.
	StatusUpgradeRequired StatusCode = 426

	// StatusPreconditionRequired maps to HTTP 428.
	StatusPreconditionRequired StatusCode = 428

	// StatusTooManyRequests maps to HTTP 429.
	StatusTooManyRequests StatusCode = 429

	// StatusRequestHeaderFieldsTooLarge maps to HTTP 431.
	StatusRequestHeaderFieldsTooLarge StatusCode = 431

	// StatusUnavailableForLegalReasons maps to HTTP 451.
	StatusUnavailableForLegalReasons StatusCode = 451

	// 5xx — Server Error

	// StatusInternalServerError maps to HTTP 500.
	StatusInternalServerError StatusCode = 500

	// StatusNotImplemented maps to HTTP 501.
	StatusNotImplemented StatusCode = 501

	// StatusBadGateway maps to HTTP 502.
	StatusBadGateway StatusCode = 502

	// StatusServiceUnavailable maps to HTTP 503.
	StatusServiceUnavailable StatusCode = 503

	// StatusGatewayTimeout maps to HTTP 504.
	StatusGatewayTimeout StatusCode = 504

	// StatusHTTPVersionNotSupported maps to HTTP 505.
	StatusHTTPVersionNotSupported StatusCode = 505

	// StatusVariantAlsoNegotiates maps to HTTP 506.
	StatusVariantAlsoNegotiates StatusCode = 506

	// StatusInsufficientStorage maps to HTTP 507.
	StatusInsufficientStorage StatusCode = 507

	// StatusLoopDetected maps to HTTP 508.
	StatusLoopDetected StatusCode = 508

	// StatusNotExtended maps to HTTP 510.
	StatusNotExtended StatusCode = 510

	// StatusNetworkAuthenticationRequired maps to HTTP 511.
	StatusNetworkAuthenticationRequired StatusCode = 511

	// Non-standard / vendor-extended codes

	// StatusReserved is the deprecated HTTP 306, reserved for future use.
	StatusReserved StatusCode = 306

	// StatusEnhanceYourCalm is a non-standard HTTP 420 used by some APIs to signal rate-limiting.
	StatusEnhanceYourCalm StatusCode = 420

	// StatusNoResponse is a non-standard HTTP 444 used by Nginx to indicate no response.
	StatusNoResponse StatusCode = 444

	// StatusRetryWith is a non-standard HTTP 449 used by Microsoft to request a retry with different parameters.
	StatusRetryWith StatusCode = 449

	// StatusBlockedByParentalControls is a non-standard HTTP 450 used by Microsoft.
	StatusBlockedByParentalControls StatusCode = 450

	// StatusClientClosedRequest is a non-standard HTTP 499 used by Nginx when the client closes the connection.
	StatusClientClosedRequest StatusCode = 499

	// StatusBandwidthLimitExceeded is a non-standard HTTP 509 used by Apache/cPanel.
	StatusBandwidthLimitExceeded StatusCode = 509

	// StatusNetworkReadTimeout is a non-standard HTTP 598 indicating a network read timeout.
	StatusNetworkReadTimeout StatusCode = 598

	// StatusNetworkConnectTimeout is a non-standard HTTP 599 indicating a network connect timeout.
	StatusNetworkConnectTimeout StatusCode = 599
)

var (
	// 1xx Informational responses
	// ////////////////////////

	// Continue indicates that the initial part of the request has been received and has not yet been rejected by the server.
	Continue = Header().WithCode(StatusContinue.Value()).WithText("Continue").WithType("Informational")

	// SwitchingProtocols indicates that the server will switch protocols as requested by the client.
	SwitchingProtocols = Header().WithCode(StatusSwitchingProtocols.Value()).WithText("Switching Protocols").WithType("Informational")

	// Processing indicates that the server has received and is processing the request but no response is available yet.
	Processing = Header().WithCode(StatusProcessing.Value()).WithText("Processing").WithType("Informational")

	// 2xx Successful responses
	// ////////////////////////

	// OK indicates that the request has succeeded.
	OK = Header().WithCode(StatusOK.Value()).WithText("OK").WithType("Successful")

	// Created indicates that the request has been fulfilled and has resulted in a new resource being created.
	Created = Header().WithCode(StatusCreated.Value()).WithText("Created").WithType("Successful")

	// Accepted indicates that the request has been accepted for processing, but the processing has not been completed.
	Accepted = Header().WithCode(StatusAccepted.Value()).WithText("Accepted").WithType("Successful")

	// NonAuthoritativeInformation indicates that the request was successful, but the enclosed metadata may be from a different source.
	NonAuthoritativeInformation = Header().WithCode(StatusNonAuthoritativeInfo.Value()).WithText("Non-Authoritative Information").WithType("Successful")

	// NoContent indicates that the server successfully processed the request, but is not returning any content.
	NoContent = Header().WithCode(StatusNoContent.Value()).WithText("No Content").WithType("Successful")

	// ResetContent indicates that the server successfully processed the request and requests the client to reset the document view.
	ResetContent = Header().WithCode(StatusResetContent.Value()).WithText("Reset Content").WithType("Successful")

	// PartialContent indicates that the server is delivering only part of the resource due to a range request.
	PartialContent = Header().WithCode(StatusPartialContent.Value()).WithText("Partial Content").WithType("Successful")

	// MultiStatus provides status for multiple independent operations.
	MultiStatus = Header().WithCode(StatusMultiStatus.Value()).WithText("Multi-Status").WithType("Successful")

	// AlreadyReported indicates that the members of a DAV binding have already been enumerated in a previous reply.
	AlreadyReported = Header().WithCode(StatusAlreadyReported.Value()).WithText("Already Reported").WithType("Successful")

	// IMUsed indicates that the server has fulfilled a GET request for the resource and the response is a representation of the result.
	IMUsed = Header().WithCode(StatusIMUsed.Value()).WithText("IM Used").WithType("Successful")

	// 3xx Redirection responses
	// ////////////////////////

	// MultipleChoices indicates multiple options for the resource are available.
	MultipleChoices = Header().WithCode(StatusMultipleChoices.Value()).WithText("Multiple Choices").WithType("Redirection")

	// MovedPermanently indicates that the resource has been permanently moved to a new URI.
	MovedPermanently = Header().WithCode(StatusMovedPermanently.Value()).WithText("Moved Permanently").WithType("Redirection")

	// Found indicates that the resource has been temporarily moved to a different URI.
	Found = Header().WithCode(StatusFound.Value()).WithText("Found").WithType("Redirection")

	// SeeOther indicates that the response to the request can be found under another URI.
	SeeOther = Header().WithCode(StatusSeeOther.Value()).WithText("See Other").WithType("Redirection")

	// NotModified indicates that the resource has not been modified since the last request.
	NotModified = Header().WithCode(StatusNotModified.Value()).WithText("Not Modified").WithType("Redirection")

	// UseProxy indicates that the requested resource must be accessed through the proxy given by the location field.
	UseProxy = Header().WithCode(StatusUseProxy.Value()).WithText("Use Proxy").WithType("Redirection")

	// Reserved is a deprecated status code reserved for future use.
	Reserved = Header().WithCode(StatusReserved.Value()).WithText("Reserved").WithType("Redirection")

	// TemporaryRedirect indicates that the resource has been temporarily moved to a different URI and will return to the original URI later.
	TemporaryRedirect = Header().WithCode(StatusTemporaryRedirect.Value()).WithText("Temporary Redirect").WithType("Redirection")

	// PermanentRedirect indicates that the resource has been permanently moved to a new URI and future requests should use this URI.
	PermanentRedirect = Header().WithCode(StatusPermanentRedirect.Value()).WithText("Permanent Redirect").WithType("Redirection")

	// 4xx Client error responses
	// ////////////////////////

	// BadRequest indicates that the server could not understand the request due to invalid syntax.
	BadRequest = Header().WithCode(StatusBadRequest.Value()).WithText("Bad Request").WithType("Client Error")

	// Unauthorized indicates that the client must authenticate itself to get the requested response.
	Unauthorized = Header().WithCode(StatusUnauthorized.Value()).WithText("Unauthorized").WithType("Client Error")

	// PaymentRequired is reserved for future use, indicating payment is required to access the resource.
	PaymentRequired = Header().WithCode(StatusPaymentRequired.Value()).WithText("Payment Required").WithType("Client Error")

	// Forbidden indicates that the server understands the request but refuses to authorize it.
	Forbidden = Header().WithCode(StatusForbidden.Value()).WithText("Forbidden").WithType("Client Error")

	// NotFound indicates that the server can't find the requested resource.
	NotFound = Header().WithCode(StatusNotFound.Value()).WithText("Not Found").WithType("Client Error")

	// MethodNotAllowed indicates that the server knows the request method but the target resource doesn't support this method.
	MethodNotAllowed = Header().WithCode(StatusMethodNotAllowed.Value()).WithText("Method Not Allowed").WithType("Client Error")

	// NotAcceptable indicates that the server cannot produce a response matching the list of acceptable values defined in the request's headers.
	NotAcceptable = Header().WithCode(StatusNotAcceptable.Value()).WithText("Not Acceptable").WithType("Client Error")

	// ProxyAuthenticationRequired indicates that the client must first authenticate itself with the proxy.
	ProxyAuthenticationRequired = Header().WithCode(StatusProxyAuthRequired.Value()).WithText("Proxy Authentication Required").WithType("Client Error")

	// RequestTimeout indicates that the server timed out waiting for the request.
	RequestTimeout = Header().WithCode(StatusRequestTimeout.Value()).WithText("Request Timeout").WithType("Client Error")

	// Conflict indicates that the request conflicts with the current state of the server.
	Conflict = Header().WithCode(StatusConflict.Value()).WithText("Conflict").WithType("Client Error")

	// Gone indicates that the requested resource is no longer available and will not be available again.
	Gone = Header().WithCode(StatusGone.Value()).WithText("Gone").WithType("Client Error")

	// LengthRequired indicates that the server requires the request to be sent with a Content-Length header.
	LengthRequired = Header().WithCode(StatusLengthRequired.Value()).WithText("Length Required").WithType("Client Error")

	// PreconditionFailed indicates that the server does not meet one of the preconditions set by the client.
	PreconditionFailed = Header().WithCode(StatusPreconditionFailed.Value()).WithText("Precondition Failed").WithType("Client Error")

	// RequestEntityTooLarge indicates that the request entity is larger than what the server is willing or able to process.
	RequestEntityTooLarge = Header().WithCode(StatusRequestEntityTooLarge.Value()).WithText("Request Entity Too Large").WithType("Client Error")

	// RequestURITooLong indicates that the URI provided was too long for the server to process.
	RequestURITooLong = Header().WithCode(StatusRequestURITooLong.Value()).WithText("Request-URI Too Long").WithType("Client Error")

	// UnsupportedMediaType indicates that the media format of the requested data is not supported by the server.
	UnsupportedMediaType = Header().WithCode(StatusUnsupportedMediaType.Value()).WithText("Unsupported Media Type").WithType("Client Error")

	// RequestedRangeNotSatisfiable indicates that the range specified by the Range header cannot be satisfied.
	RequestedRangeNotSatisfiable = Header().WithCode(StatusRequestedRangeNotSatisfiable.Value()).WithText("Requested Range Not Satisfiable").WithType("Client Error")

	// ExpectationFailed indicates that the server cannot meet the requirements of the Expect request-header field.
	ExpectationFailed = Header().WithCode(StatusExpectationFailed.Value()).WithText("Expectation Failed").WithType("Client Error")

	// ImATeapot is a humorous response code indicating that the server is a teapot and refuses to brew coffee.
	ImATeapot = Header().WithCode(StatusTeapot.Value()).WithText("I'm a teapot").WithType("Client Error")

	// EnhanceYourCalm is a non-standard response code used to ask the client to reduce its request rate.
	EnhanceYourCalm = Header().WithCode(StatusEnhanceYourCalm.Value()).WithText("Enhance Your Calm").WithType("Client Error")

	// UnprocessableEntity indicates that the request was well-formed but could not be followed due to semantic errors.
	UnprocessableEntity = Header().WithCode(StatusUnprocessableEntity.Value()).WithText("Unprocessable Entity").WithType("Client Error")

	// Locked indicates that the resource being accessed is locked.
	Locked = Header().WithCode(StatusLocked.Value()).WithText("Locked").WithType("Client Error")

	// FailedDependency indicates that the request failed due to failure of a previous request.
	FailedDependency = Header().WithCode(StatusFailedDependency.Value()).WithText("Failed Dependency").WithType("Client Error")

	// UnorderedCollection is a non-standard response code indicating an unordered collection.
	UnorderedCollection = Header().WithCode(StatusTooEarly.Value()).WithText("Unordered Collection").WithType("Client Error")

	// UpgradeRequired indicates that the client should switch to a different protocol.
	UpgradeRequired = Header().WithCode(StatusUpgradeRequired.Value()).WithText("Upgrade Required").WithType("Client Error")

	// PreconditionRequired indicates that the origin server requires the request to be conditional.
	PreconditionRequired = Header().WithCode(StatusPreconditionRequired.Value()).WithText("Precondition Required").WithType("Client Error")

	// TooManyRequests indicates that the user has sent too many requests in a given time.
	TooManyRequests = Header().WithCode(StatusTooManyRequests.Value()).WithText("Too Many Requests").WithType("Client Error")

	// RequestHeaderFieldsTooLarge indicates that one or more header fields in the request are too large.
	RequestHeaderFieldsTooLarge = Header().WithCode(StatusRequestHeaderFieldsTooLarge.Value()).WithText("Request Header Fields Too Large").WithType("Client Error")

	// NoResponse is a non-standard code indicating that the server has no response to provide.
	NoResponse = Header().WithCode(StatusNoResponse.Value()).WithText("No Response").WithType("Client Error")

	// RetryWith is a non-standard code indicating that the client should retry with different parameters.
	RetryWith = Header().WithCode(StatusRetryWith.Value()).WithText("Retry With").WithType("Client Error")

	// BlockedByWindowsParentalControls is a non-standard code indicating that the request was blocked by parental controls.
	BlockedByWindowsParentalControls = Header().WithCode(StatusBlockedByParentalControls.Value()).WithText("Blocked by Windows Parental Controls").WithType("Client Error")

	// UnavailableForLegalReasons indicates that the server is denying access to the resource for legal reasons.
	UnavailableForLegalReasons = Header().WithCode(StatusUnavailableForLegalReasons.Value()).WithText("Unavailable For Legal Reasons").WithType("Client Error")

	// ClientClosedRequest is a non-standard code indicating that the client closed the connection before the server's response.
	ClientClosedRequest = Header().WithCode(StatusClientClosedRequest.Value()).WithText("Client Closed Request").WithType("Client Error")

	// 5xx Server error responses
	// ////////////////////////

	// InternalServerError indicates that the server encountered an unexpected condition that prevented it from fulfilling the request.
	InternalServerError = Header().WithCode(StatusInternalServerError.Value()).WithText("Internal Server Error").WithType("Server Error")

	// NotImplemented indicates that the server does not support the functionality required to fulfill the request.
	NotImplemented = Header().WithCode(StatusNotImplemented.Value()).WithText("Not Implemented").WithType("Server Error")

	// BadGateway indicates that the server received an invalid response from an upstream server.
	BadGateway = Header().WithCode(StatusBadGateway.Value()).WithText("Bad Gateway").WithType("Server Error")

	// ServiceUnavailable indicates that the server is currently unavailable (e.g., overloaded or under maintenance).
	ServiceUnavailable = Header().WithCode(StatusServiceUnavailable.Value()).WithText("Service Unavailable").WithType("Server Error")

	// GatewayTimeout indicates that the server did not receive a timely response from an upstream server.
	GatewayTimeout = Header().WithCode(StatusGatewayTimeout.Value()).WithText("Gateway Timeout").WithType("Server Error")

	// HTTPVersionNotSupported indicates that the server does not support the HTTP protocol version used in the request.
	HTTPVersionNotSupported = Header().WithCode(StatusHTTPVersionNotSupported.Value()).WithText("HTTP Version Not Supported").WithType("Server Error")

	// VariantAlsoNegotiates indicates an internal server configuration error leading to circular references.
	VariantAlsoNegotiates = Header().WithCode(StatusVariantAlsoNegotiates.Value()).WithText("Variant Also Negotiates").WithType("Server Error")

	// InsufficientStorage indicates that the server is unable to store the representation needed to complete the request.
	InsufficientStorage = Header().WithCode(StatusInsufficientStorage.Value()).WithText("Insufficient Storage").WithType("Server Error")

	// LoopDetected indicates that the server detected an infinite loop while processing the request.
	LoopDetected = Header().WithCode(StatusLoopDetected.Value()).WithText("Loop Detected").WithType("Server Error")

	// BandwidthLimitExceeded is a non-standard code indicating that the server's bandwidth limit has been exceeded.
	BandwidthLimitExceeded = Header().WithCode(StatusBandwidthLimitExceeded.Value()).WithText("Bandwidth Limit Exceeded").WithType("Server Error")

	// NotExtended indicates that further extensions to the request are required for the server to fulfill it.
	NotExtended = Header().WithCode(StatusNotExtended.Value()).WithText("Not Extended").WithType("Server Error")

	// NetworkAuthenticationRequired indicates that the client needs to authenticate to gain network access.
	NetworkAuthenticationRequired = Header().WithCode(StatusNetworkAuthenticationRequired.Value()).WithText("Network Authentication Required").WithType("Server Error")

	// NetworkReadTimeoutError is a non-standard code indicating a network read timeout error.
	NetworkReadTimeoutError = Header().WithCode(StatusNetworkReadTimeout.Value()).WithText("Network Read Timeout Error").WithType("Server Error")

	// NetworkConnectTimeoutError is a non-standard code indicating a network connection timeout error.
	NetworkConnectTimeoutError = Header().WithCode(StatusNetworkConnectTimeout.Value()).WithText("Network Connect Timeout Error").WithType("Server Error")
)

// StreamingStrategy defines the strategy used for streaming data.
// It determines how data is sent or received in a streaming manner.
const (
	// Direct streaming without buffering
	// Each piece of data is sent immediately as it becomes available.
	StrategyDirect StreamingStrategy = "direct"

	// Buffered streaming with internal buffer
	// Data is collected in a buffer and sent in larger chunks to optimize performance.
	StrategyBuffered StreamingStrategy = "buffered"

	// Chunked streaming with explicit chunk handling
	// Data is divided into chunks of a specified size and sent sequentially.
	StrategyChunked StreamingStrategy = "chunked"
)

// CompressionType defines the type of compression applied to data.
// It specifies the algorithm used to compress or decompress data.
const (
	// No compression applied
	CompressNone CompressionType = "none"

	// GZIP compression algorithm
	CompressGzip CompressionType = "gzip"

	// Deflate compression algorithm
	CompressDeflate CompressionType = "deflate"

	// Flate compression algorithm
	CompressFlate CompressionType = "flate"
)

// Common constants used across the package.
const (
	// ErrUnknown is a constant string used to represent an unknown or unspecified value in the context of XC (cross-cutting) concerns.
	// It is typically used as a placeholder when the actual value is not available or not applicable.
	ErrUnknown string = "replify: error unknown"

	// defaultChunkSize defines the maximum number of bytes in each chunk.
	// defaultChunkSize is used to limit the size of data chunks when processing large responses or requests.
	defaultChunkSize int = 1024
)

// Locale defines the language and regional settings for content localization.
// It specifies the language and country/region code.
const (
	// English (United States)
	//  Example: en_US
	LocaleEnUS Locale = "en_US"

	// English (United Kingdom)
	//  Example: en_GB
	LocaleEnGB Locale = "en_GB"

	// English (Australia)
	//  Example: en_AU
	LocaleEnAU Locale = "en_AU"

	// English (Canada)
	//  Example: en_CA
	LocaleEnCA Locale = "en_CA"

	// Vietnamese (Vietnam)
	//  Example: vi_VN
	LocaleViVN Locale = "vi_VN"

	// French (France)
	//  Example: fr_FR
	LocaleFrFR Locale = "fr_FR"

	// French (Canada)
	//  Example: fr_CA
	LocaleFrCA Locale = "fr_CA"

	// German (Germany)
	//  Example: de_DE
	LocaleDeDE Locale = "de_DE"

	// Spanish (Spain)
	//  Example: es_ES
	LocaleEsES Locale = "es_ES"

	// Spanish (Mexico)
	//  Example: es_MX
	LocaleEsMX Locale = "es_MX"

	// Portuguese (Brazil)
	//  Example: pt_BR
	LocalePtBR Locale = "pt_BR"

	// Portuguese (Portugal)
	//  Example: pt_PT
	LocalePtPT Locale = "pt_PT"

	// Italian (Italy)
	//  Example: it_IT
	LocaleItIT Locale = "it_IT"

	// Dutch (Netherlands)
	//  Example: nl_NL
	LocaleNlNL Locale = "nl_NL"

	// Russian (Russia)
	//  Example: ru_RU
	LocaleRuRU Locale = "ru_RU"

	// Japanese (Japan)
	//  Example: ja_JP
	LocaleJaJP Locale = "ja_JP"

	// Korean (South Korea)
	//  Example: ko_KR
	LocaleKoKR Locale = "ko_KR"

	// Chinese (Simplified, China)
	//  Example: zh_CN
	LocaleZhCN Locale = "zh_CN"

	// Chinese (Traditional, Taiwan)
	//  Example: zh_TW
	LocaleZhTW Locale = "zh_TW"

	// Thai (Thailand)
	//  Example: th_TH
	LocaleThTH Locale = "th_TH"

	// Indonesian (Indonesia)
	//  Example: id_ID
	LocaleIdID Locale = "id_ID"

	// Malay (Malaysia)
	//  Example: ms_MY
	LocaleMsMY Locale = "ms_MY"

	// Hindi (India)
	//  Example: hi_IN
	LocaleHiIN Locale = "hi_IN"

	// Arabic (Saudi Arabia)
	//  Example: ar_SA
	LocaleArSA Locale = "ar_SA"

	// Turkish (Turkey)
	//  Example: tr_TR
	LocaleTrTR Locale = "tr_TR"

	// Polish (Poland)
	//  Example: pl_PL
	LocalePlPL Locale = "pl_PL"

	// Swedish (Sweden)
	//  Example: sv_SE
	LocaleSvSE Locale = "sv_SE"

	// Danish (Denmark)
	//  Example: da_DK
	LocaleDaDK Locale = "da_DK"

	// Norwegian (Norway)
	//  Example: nb_NO
	LocaleNbNO Locale = "nb_NO"

	// Finnish (Finland)
	//  Example: fi_FI
	LocaleFiFI Locale = "fi_FI"
)

// SignatureAlgorithm constants for common cryptographic algorithms used in signatures.
const (
	// HMACSHA256 is widely supported and provides a good balance between security and performance.
	// It is suitable for most applications where both security and performance are important.
	HMACSHA256 SignatureAlgorithm = "HMAC-SHA256"

	// HMACSHA512 provides stronger security than HMACSHA256 but results in larger signatures.
	// It is suitable for applications where security is a higher priority than performance.
	HMACSHA512 SignatureAlgorithm = "HMAC-SHA512"

	// HMACSHA384 is a truncated version of SHA-512 (384-bit output). It offers strong security
	// with slightly smaller signature sizes than HMACSHA512, commonly used in TLS and JWTs.
	HMACSHA384 SignatureAlgorithm = "HMAC-SHA384"

	// HMACSHA512_256 uses the SHA-512 algorithm but truncates the output to 256 bits.
	// It provides the performance benefits of SHA-512 on 64-bit architectures while keeping
	// the signature size equivalent to SHA-256.
	HMACSHA512_256 SignatureAlgorithm = "HMAC-SHA512/256"

	// HMACSHA224 is a truncated version of SHA-256 (224-bit output).
	// It is occasionally used in environments with strict payload size constraints.
	HMACSHA224 SignatureAlgorithm = "HMAC-SHA224"

	// HMACSHA1 is provided primarily for backward compatibility with older systems (e.g., OAuth 1.0).
	// WARNING: SHA-1 is considered cryptographically broken against collision attacks and
	// should NOT be used for new security-critical signatures.
	HMACSHA1 SignatureAlgorithm = "HMAC-SHA1"

	// HMACMD5 is provided strictly for legacy system integration.
	// WARNING: MD5 is completely insecure for cryptographic signatures and must be avoided
	// in any modern application.
	HMACMD5 SignatureAlgorithm = "HMAC-MD5"
)

// ReasonCategory constants define various categories for reasons associated with wrapper instances.
const (
	// CategoryCommon represents general/common reasons that do not fit into other specific categories.
	CategoryCommon ReasonCategory = "GENERAL_COMMON"

	// CategoryValidation represents reasons related to input validation failures.
	CategoryValidation ReasonCategory = "VALIDATION"

	// CategoryAuthentication represents reasons related to authentication failures.
	CategoryAuthentication ReasonCategory = "AUTHENTICATION"

	// CategoryAuthorization represents reasons related to authorization or permission failures.
	CategoryAuthorization ReasonCategory = "AUTHORIZATION_PERMISSION"

	// CategoryUserAccount represents reasons related to user account issues.
	CategoryUserAccount ReasonCategory = "USER_ACCOUNT"

	// CategoryResourceLifecycle represents reasons related to the lifecycle of resources.
	CategoryResourceLifecycle ReasonCategory = "RESOURCE_LIFECYCLE"

	// CategoryStateTransition represents reasons related to state transitions of resources.
	CategoryStateTransition ReasonCategory = "STATE_STATE_TRANSITION"

	// CategoryConcurrencyIdempotency represents reasons related to concurrency control and idempotency issues.
	CategoryConcurrencyIdempotency ReasonCategory = "CONCURRENCY_IDEMPOTENCY"

	// CategoryRateLimitQuota represents reasons related to rate limiting and quota enforcement.
	CategoryRateLimitQuota ReasonCategory = "RATE_LIMIT_QUOTA"

	// CategoryPaginationQuery represents reasons related to pagination and query operations.
	CategoryPaginationQuery ReasonCategory = "PAGINATION_QUERY"

	// CategoryFileUpload represents reasons related to file upload operations.
	CategoryFileUpload ReasonCategory = "FILE_UPLOAD"

	// CategoryPayment represents reasons related to payment operations.
	CategoryPayment ReasonCategory = "PAYMENT"

	// CategoryOrder represents reasons related to order operations.
	CategoryOrder ReasonCategory = "ORDER"

	// CategoryInventory represents reasons related to inventory management.
	CategoryInventory ReasonCategory = "INVENTORY"

	// CategoryShippingDelivery represents reasons related to shipping and delivery operations.
	CategoryShippingDelivery ReasonCategory = "SHIPPING_DELIVERY"

	// CategoryCouponPromotion represents reasons related to coupon and promotion operations.
	CategoryCouponPromotion ReasonCategory = "COUPON_PROMOTION"

	// CategoryVerification represents reasons related to verification processes.
	CategoryVerification ReasonCategory = "VERIFICATION"

	// CategoryBusinessRule represents reasons related to business rule violations.
	CategoryBusinessRule ReasonCategory = "BUSINESS_RULE"

	// CategoryDependencyExternalService represents reasons related to external service dependencies.
	CategoryDependencyExternalService ReasonCategory = "DEPENDENCY_EXTERNAL_SERVICE"

	// CategoryDatabasePersistence represents reasons related to database persistence operations.
	CategoryDatabasePersistence ReasonCategory = "DATABASE_PERSISTENCE"

	// CategoryCache represents reasons related to caching operations.
	CategoryCache ReasonCategory = "CACHE"

	// CategorySecurity represents reasons related to security issues.
	CategorySecurity ReasonCategory = "SECURITY"

	// CategoryWebhookEvent represents reasons related to webhook events.
	CategoryWebhookEvent ReasonCategory = "WEBHOOK_EVENT"

	// CategoryAsyncJob represents reasons related to asynchronous job operations.
	CategoryAsyncJob ReasonCategory = "ASYNC_JOB"

	// CategoryImportExport represents reasons related to import and export operations.
	CategoryImportExport ReasonCategory = "IMPORT_EXPORT"

	// CategoryNotification represents reasons related to notification operations.
	CategoryNotification ReasonCategory = "NOTIFICATION"

	// CategorySystem represents reasons related to system operations.
	CategorySystem ReasonCategory = "SYSTEM"
)

// ReasonCode represents the specific reason for an error or failure.
const (
	// GENERAL / COMMON
	ReasonCodeInvalidRequest        ReasonCode = "INVALID_REQUEST"
	ReasonCodeMissingRequiredField  ReasonCode = "MISSING_REQUIRED_FIELD"
	ReasonCodeInvalidField          ReasonCode = "INVALID_FIELD"
	ReasonCodeInvalidParameter      ReasonCode = "INVALID_PARAMETER"
	ReasonCodeInvalidQueryParameter ReasonCode = "INVALID_QUERY_PARAMETER"
	ReasonCodeInvalidPathParameter  ReasonCode = "INVALID_PATH_PARAMETER"
	ReasonCodeInvalidHeader         ReasonCode = "INVALID_HEADER"
	ReasonCodeInvalidBody           ReasonCode = "INVALID_BODY"
	ReasonCodeInvalidFormat         ReasonCode = "INVALID_FORMAT"
	ReasonCodeInvalidValue          ReasonCode = "INVALID_VALUE"
	ReasonCodeValueOutOfRange       ReasonCode = "VALUE_OUT_OF_RANGE"
	ReasonCodeValueTooLong          ReasonCode = "VALUE_TOO_LONG"
	ReasonCodeValueTooShort         ReasonCode = "VALUE_TOO_SHORT"
	ReasonCodeDuplicateValue        ReasonCode = "DUPLICATE_VALUE"
	ReasonCodeUnsupportedValue      ReasonCode = "UNSUPPORTED_VALUE"
	ReasonCodeUnsupportedOperation  ReasonCode = "UNSUPPORTED_OPERATION"
	ReasonCodeOperationNotAllowed   ReasonCode = "OPERATION_NOT_ALLOWED"
	ReasonCodeOperationNotSupported ReasonCode = "OPERATION_NOT_SUPPORTED"
	ReasonCodeResourceNotFound      ReasonCode = "RESOURCE_NOT_FOUND"
	ReasonCodeResourceAlreadyExists ReasonCode = "RESOURCE_ALREADY_EXISTS"
	ReasonCodeResourceConflict      ReasonCode = "RESOURCE_CONFLICT"
	ReasonCodeResourceUnavailable   ReasonCode = "RESOURCE_UNAVAILABLE"
	ReasonCodeResourceExpired       ReasonCode = "RESOURCE_EXPIRED"
	ReasonCodeResourceLocked        ReasonCode = "RESOURCE_LOCKED"
	ReasonCodeResourceArchived      ReasonCode = "RESOURCE_ARCHIVED"
	ReasonCodeResourceDeleted       ReasonCode = "RESOURCE_DELETED"

	// VALIDATION
	ReasonCodeValidationFailed     ReasonCode = "VALIDATION_FAILED"
	ReasonCodeRequiredFieldMissing ReasonCode = "REQUIRED_FIELD_MISSING"
	ReasonCodeFieldInvalid         ReasonCode = "FIELD_INVALID"
	ReasonCodeFieldEmpty           ReasonCode = "FIELD_EMPTY"
	ReasonCodeFieldNull            ReasonCode = "FIELD_NULL"
	ReasonCodeFieldTypeInvalid     ReasonCode = "FIELD_TYPE_INVALID"
	ReasonCodeFieldFormatInvalid   ReasonCode = "FIELD_FORMAT_INVALID"
	ReasonCodeInvalidEmail         ReasonCode = "INVALID_EMAIL"
	ReasonCodeInvalidPhoneNumber   ReasonCode = "INVALID_PHONE_NUMBER"
	ReasonCodeInvalidURL           ReasonCode = "INVALID_URL"
	ReasonCodeInvalidUUID          ReasonCode = "INVALID_UUID"
	ReasonCodeInvalidDate          ReasonCode = "INVALID_DATE"
	ReasonCodeInvalidDateTime      ReasonCode = "INVALID_DATETIME"
	ReasonCodeInvalidTimezone      ReasonCode = "INVALID_TIMEZONE"
	ReasonCodeInvalidLocale        ReasonCode = "INVALID_LOCALE"
	ReasonCodeInvalidCurrency      ReasonCode = "INVALID_CURRENCY"
	ReasonCodeValueTooSmall        ReasonCode = "VALUE_TOO_SMALL"
	ReasonCodeValueTooLarge        ReasonCode = "VALUE_TOO_LARGE"
	ReasonCodeInvalidEnumValue     ReasonCode = "INVALID_ENUM_VALUE"
	ReasonCodeInvalidRegexp        ReasonCode = "INVALID_REGEX"
	ReasonCodeInvalidEncoding      ReasonCode = "INVALID_ENCODING"

	// AUTHENTICATION
	ReasonCodeAuthenticationRequired ReasonCode = "AUTHENTICATION_REQUIRED"
	ReasonCodeAuthenticationFailed   ReasonCode = "AUTHENTICATION_FAILED"
	ReasonCodeInvalidCredentials     ReasonCode = "INVALID_CREDENTIALS"
	ReasonCodeInvalidPassword        ReasonCode = "INVALID_PASSWORD"
	ReasonCodeInvalidUsername        ReasonCode = "INVALID_USERNAME"
	ReasonCodeInvalidToken           ReasonCode = "INVALID_TOKEN"
	ReasonCodeTokenExpired           ReasonCode = "TOKEN_EXPIRED"
	ReasonCodeTokenRevoked           ReasonCode = "TOKEN_REVOKED"
	ReasonCodeTokenInvalid           ReasonCode = "TOKEN_INVALID"
	ReasonCodeTokenMalformed         ReasonCode = "TOKEN_MALFORMED"
	ReasonCodeTokenMissing           ReasonCode = "TOKEN_MISSING"
	ReasonCodeTokenNotActive         ReasonCode = "TOKEN_NOT_ACTIVE"
	ReasonCodeSessionExpired         ReasonCode = "SESSION_EXPIRED"
	ReasonCodeSessionRevoked         ReasonCode = "SESSION_REVOKED"
	ReasonCodeSessionNotFound        ReasonCode = "SESSION_NOT_FOUND"
	ReasonCodeMFARequired            ReasonCode = "MFA_REQUIRED"
	ReasonCodeMFAFailed              ReasonCode = "MFA_FAILED"
	ReasonCodeMFAInvalid             ReasonCode = "MFA_INVALID"
	ReasonCodeMFAExpired             ReasonCode = "MFA_EXPIRED"
	ReasonCodeMFANotEnabled          ReasonCode = "MFA_NOT_ENABLED"
	ReasonCodeMFAAlreadyEnabled      ReasonCode = "MFA_ALREADY_ENABLED"
	ReasonCodeMFAAlreadyVerified     ReasonCode = "MFA_ALREADY_VERIFIED"
	ReasonCodeOTPRequired            ReasonCode = "OTP_REQUIRED"
	ReasonCodeOTPInvalid             ReasonCode = "OTP_INVALID"
	ReasonCodeOTPExpired             ReasonCode = "OTP_EXPIRED"
	ReasonCodeOTPMaxAttemptsExceeded ReasonCode = "OTP_MAX_ATTEMPTS_EXCEEDED"
	ReasonCodeOTPAlreadyUsed         ReasonCode = "OTP_ALREADY_USED"
	ReasonCodeOTPNotFound            ReasonCode = "OTP_NOT_FOUND"

	// AUTHORIZATION / PERMISSION
	ReasonCodeAccessDenied           ReasonCode = "ACCESS_DENIED"
	ReasonCodePermissionDenied       ReasonCode = "PERMISSION_DENIED"
	ReasonCodeInsufficientPermission ReasonCode = "INSUFFICIENT_PERMISSION"
	ReasonCodeRoleRequired           ReasonCode = "ROLE_REQUIRED"
	ReasonCodeResourceAccessDenied   ReasonCode = "RESOURCE_ACCESS_DENIED"
	ReasonCodeOperationForbidden     ReasonCode = "OPERATION_FORBIDDEN"
	ReasonCodeAccountSuspended       ReasonCode = "ACCOUNT_SUSPENDED"
	ReasonCodeAccountDisabled        ReasonCode = "ACCOUNT_DISABLED"
	ReasonCodeAccountLocked          ReasonCode = "ACCOUNT_LOCKED"

	// USER / ACCOUNT
	ReasonCodeUserNotFound                  ReasonCode = "USER_NOT_FOUND"
	ReasonCodeUserAlreadyExists             ReasonCode = "USER_ALREADY_EXISTS"
	ReasonCodeUserAlreadyActive             ReasonCode = "USER_ALREADY_ACTIVE"
	ReasonCodeUserAlreadyInactive           ReasonCode = "USER_ALREADY_INACTIVE"
	ReasonCodeUserAlreadyVerified           ReasonCode = "USER_ALREADY_VERIFIED"
	ReasonCodeUserNotVerified               ReasonCode = "USER_NOT_VERIFIED"
	ReasonCodeUserEmailAlreadyExists        ReasonCode = "USER_EMAIL_ALREADY_EXISTS"
	ReasonCodeUserEmailNotFound             ReasonCode = "USER_EMAIL_NOT_FOUND"
	ReasonCodeUserEmailAlreadyVerified      ReasonCode = "USER_EMAIL_ALREADY_VERIFIED"
	ReasonCodeUserEmailVerificationRequired ReasonCode = "USER_EMAIL_VERIFICATION_REQUIRED"
	ReasonCodeUserPhoneAlreadyExists        ReasonCode = "USER_PHONE_ALREADY_EXISTS"
	ReasonCodeUserPhoneNotFound             ReasonCode = "USER_PHONE_NOT_FOUND"
	ReasonCodeUserPhoneAlreadyVerified      ReasonCode = "USER_PHONE_ALREADY_VERIFIED"
	ReasonCodeUserPhoneVerificationRequired ReasonCode = "USER_PHONE_VERIFICATION_REQUIRED"
	ReasonCodeUserCannotBeDeleted           ReasonCode = "USER_CANNOT_BE_DELETED"
	ReasonCodeUserCannotBeDisabled          ReasonCode = "USER_CANNOT_BE_DISABLED"
	ReasonCodeUserCannotBeEnabled           ReasonCode = "USER_CANNOT_BE_ENABLED"
	ReasonCodeUserCannotBeUpdated           ReasonCode = "USER_CANNOT_BE_UPDATED"
	ReasonCodeUserCannotBeActivated         ReasonCode = "USER_CANNOT_BE_ACTIVATED"
	ReasonCodeUserCannotBeDeactivated       ReasonCode = "USER_CANNOT_BE_DEACTIVATED"
	ReasonCodeUsernameAlreadyExists         ReasonCode = "USERNAME_ALREADY_EXISTS"
	ReasonCodeUsernameUnavailable           ReasonCode = "USERNAME_UNAVAILABLE"
	ReasonCodePasswordIncorrect             ReasonCode = "PASSWORD_INCORRECT"
	ReasonCodePasswordExpired               ReasonCode = "PASSWORD_EXPIRED"
	ReasonCodePasswordTooWeak               ReasonCode = "PASSWORD_TOO_WEAK"
	ReasonCodePasswordReused                ReasonCode = "PASSWORD_REUSED"
	ReasonCodePasswordResetRequired         ReasonCode = "PASSWORD_RESET_REQUIRED"
	ReasonCodePasswordResetTokenInvalid     ReasonCode = "PASSWORD_RESET_TOKEN_INVALID"
	ReasonCodePasswordResetTokenExpired     ReasonCode = "PASSWORD_RESET_TOKEN_EXPIRED"
	ReasonCodeAccountAlreadyExists          ReasonCode = "ACCOUNT_ALREADY_EXISTS"
	ReasonCodeAccountNotFound               ReasonCode = "ACCOUNT_NOT_FOUND"
	ReasonCodeAccountAlreadyActive          ReasonCode = "ACCOUNT_ALREADY_ACTIVE"
	ReasonCodeAccountAlreadyDisabled        ReasonCode = "ACCOUNT_ALREADY_DISABLED"
	ReasonCodeAccountPending                ReasonCode = "ACCOUNT_PENDING"
	ReasonCodeAccountDeletionPending        ReasonCode = "ACCOUNT_DELETION_PENDING"

	// RESOURCE LIFECYCLE
	ReasonCodeResourceAlreadyCreated  ReasonCode = "RESOURCE_ALREADY_CREATED"
	ReasonCodeResourceAlreadyUpdated  ReasonCode = "RESOURCE_ALREADY_UPDATED"
	ReasonCodeResourceAlreadyDeleted  ReasonCode = "RESOURCE_ALREADY_DELETED"
	ReasonCodeResourceNotActive       ReasonCode = "RESOURCE_NOT_ACTIVE"
	ReasonCodeResourceAlreadyActive   ReasonCode = "RESOURCE_ALREADY_ACTIVE"
	ReasonCodeResourceNotInactive     ReasonCode = "RESOURCE_NOT_INACTIVE"
	ReasonCodeResourceAlreadyInactive ReasonCode = "RESOURCE_ALREADY_INACTIVE"
	ReasonCodeResourceNotExpired      ReasonCode = "RESOURCE_NOT_EXPIRED"
	ReasonCodeResourceAlreadyLocked   ReasonCode = "RESOURCE_ALREADY_LOCKED"
	ReasonCodeResourceNotLocked       ReasonCode = "RESOURCE_NOT_LOCKED"
	ReasonCodeResourceAlreadyArchived ReasonCode = "RESOURCE_ALREADY_ARCHIVED"
	ReasonCodeResourceNotArchived     ReasonCode = "RESOURCE_NOT_ARCHIVED"
	ReasonCodeResourcePending         ReasonCode = "RESOURCE_PENDING"
	ReasonCodeResourceProcessing      ReasonCode = "RESOURCE_PROCESSING"
	ReasonCodeResourceCompleted       ReasonCode = "RESOURCE_COMPLETED"
	ReasonCodeResourceFailed          ReasonCode = "RESOURCE_FAILED"

	// STATE / STATE TRANSITION
	ReasonCodeInvalidState                   ReasonCode = "INVALID_STATE"
	ReasonCodeInvalidStateTransition         ReasonCode = "INVALID_STATE_TRANSITION"
	ReasonCodeStateTransitionNotAllowed      ReasonCode = "STATE_TRANSITION_NOT_ALLOWED"
	ReasonCodeActionNotAllowedInCurrentState ReasonCode = "ACTION_NOT_ALLOWED_IN_CURRENT_STATE"
	ReasonCodeResourceNotReady               ReasonCode = "RESOURCE_NOT_READY"
	ReasonCodeResourceNotEligible            ReasonCode = "RESOURCE_NOT_ELIGIBLE"
	ReasonCodeAlreadyProcessed               ReasonCode = "ALREADY_PROCESSED"
	ReasonCodeAlreadyCompleted               ReasonCode = "ALREADY_COMPLETED"
	ReasonCodeAlreadyCancelled               ReasonCode = "ALREADY_CANCELLED"
	ReasonCodeAlreadyRejected                ReasonCode = "ALREADY_REJECTED"
	ReasonCodeAlreadyApproved                ReasonCode = "ALREADY_APPROVED"
	ReasonCodeCannotBeCancelled              ReasonCode = "CANNOT_BE_CANCELLED"
	ReasonCodeCannotBeUpdated                ReasonCode = "CANNOT_BE_UPDATED"
	ReasonCodeCannotBeDeleted                ReasonCode = "CANNOT_BE_DELETED"
	ReasonCodeCannotBeRetried                ReasonCode = "CANNOT_BE_RETRIED"
	ReasonCodeCannotBeReversed               ReasonCode = "CANNOT_BE_REVERSED"

	// CONCURRENCY / IDEMPOTENCY
	ReasonCodeConcurrentModification  ReasonCode = "CONCURRENT_MODIFICATION"
	ReasonCodeResourceVersionConflict ReasonCode = "RESOURCE_VERSION_CONFLICT"
	ReasonCodeOptimisticLockConflict  ReasonCode = "OPTIMISTIC_LOCK_CONFLICT"
	ReasonCodeDuplicateRequest        ReasonCode = "DUPLICATE_REQUEST"
	ReasonCodeDuplicateOperation      ReasonCode = "DUPLICATE_OPERATION"
	ReasonCodeRequestAlreadyProcessed ReasonCode = "REQUEST_ALREADY_PROCESSED"
	ReasonCodeIdempotencyKeyRequired  ReasonCode = "IDEMPOTENCY_KEY_REQUIRED"
	ReasonCodeIdempotencyKeyInvalid   ReasonCode = "IDEMPOTENCY_KEY_INVALID"
	ReasonCodeIdempotencyKeyReused    ReasonCode = "IDEMPOTENCY_KEY_REUSED"
	ReasonCodeIdempotencyConflict     ReasonCode = "IDEMPOTENCY_CONFLICT"

	// RATE LIMIT / QUOTA
	ReasonCodeRateLimitExceeded       ReasonCode = "RATE_LIMIT_EXCEEDED"
	ReasonCodeRequestLimitExceeded    ReasonCode = "REQUEST_LIMIT_EXCEEDED"
	ReasonCodeQuotaExceeded           ReasonCode = "QUOTA_EXCEEDED"
	ReasonCodeUsageLimitExceeded      ReasonCode = "USAGE_LIMIT_EXCEEDED"
	ReasonCodeConcurrentLimitExceeded ReasonCode = "CONCURRENT_LIMIT_EXCEEDED"
	ReasonCodeDailyLimitExceeded      ReasonCode = "DAILY_LIMIT_EXCEEDED"
	ReasonCodeMonthlyLimitExceeded    ReasonCode = "MONTHLY_LIMIT_EXCEEDED"
	ReasonCodeTooManyRequests         ReasonCode = "TOO_MANY_REQUESTS"
	ReasonCodeTooManyAttempts         ReasonCode = "TOO_MANY_ATTEMPTS"

	// PAGINATION / QUERY
	ReasonCodeInvalidPage           ReasonCode = "INVALID_PAGE"
	ReasonCodeInvalidPageSize       ReasonCode = "INVALID_PAGE_SIZE"
	ReasonCodeInvalidCursor         ReasonCode = "INVALID_CURSOR"
	ReasonCodeCursorExpired         ReasonCode = "CURSOR_EXPIRED"
	ReasonCodeCursorInvalid         ReasonCode = "CURSOR_INVALID"
	ReasonCodeCursorMalformed       ReasonCode = "CURSOR_MALFORMED"
	ReasonCodeInvalidSort           ReasonCode = "INVALID_SORT"
	ReasonCodeInvalidSortField      ReasonCode = "INVALID_SORT_FIELD"
	ReasonCodeInvalidSortDirection  ReasonCode = "INVALID_SORT_DIRECTION"
	ReasonCodeInvalidFilter         ReasonCode = "INVALID_FILTER"
	ReasonCodeInvalidFilterField    ReasonCode = "INVALID_FILTER_FIELD"
	ReasonCodeInvalidFilterOperator ReasonCode = "INVALID_FILTER_OPERATOR"
	ReasonCodeInvalidSearchQuery    ReasonCode = "INVALID_SEARCH_QUERY"

	// FILE / UPLOAD
	ReasonCodeFileNotFound           ReasonCode = "FILE_NOT_FOUND"
	ReasonCodeFileAlreadyExists      ReasonCode = "FILE_ALREADY_EXISTS"
	ReasonCodeFileTooLarge           ReasonCode = "FILE_TOO_LARGE"
	ReasonCodeFileTooSmall           ReasonCode = "FILE_TOO_SMALL"
	ReasonCodeFileTypeNotAllowed     ReasonCode = "FILE_TYPE_NOT_ALLOWED"
	ReasonCodeFileFormatNotSupported ReasonCode = "FILE_FORMAT_NOT_SUPPORTED"
	ReasonCodeFileCorrupted          ReasonCode = "FILE_CORRUPTED"
	ReasonCodeFileEmpty              ReasonCode = "FILE_EMPTY"
	ReasonCodeFileUploadFailed       ReasonCode = "FILE_UPLOAD_FAILED"
	ReasonCodeFileDownloadFailed     ReasonCode = "FILE_DOWNLOAD_FAILED"
	ReasonCodeInvalidFilename        ReasonCode = "INVALID_FILENAME"
	ReasonCodeFilenameTooLong        ReasonCode = "FILENAME_TOO_LONG"
	ReasonCodeStorageQuotaExceeded   ReasonCode = "STORAGE_QUOTA_EXCEEDED"
	ReasonCodeStorageUnavailable     ReasonCode = "STORAGE_UNAVAILABLE"

	// PAYMENT
	ReasonCodePaymentRequired           ReasonCode = "PAYMENT_REQUIRED"
	ReasonCodePaymentNotFound           ReasonCode = "PAYMENT_NOT_FOUND"
	ReasonCodePaymentAlreadyExists      ReasonCode = "PAYMENT_ALREADY_EXISTS"
	ReasonCodePaymentAlreadyPaid        ReasonCode = "PAYMENT_ALREADY_PAID"
	ReasonCodePaymentAlreadyFailed      ReasonCode = "PAYMENT_ALREADY_FAILED"
	ReasonCodePaymentAlreadyRefunded    ReasonCode = "PAYMENT_ALREADY_REFUNDED"
	ReasonCodePaymentAlreadyCancelled   ReasonCode = "PAYMENT_ALREADY_CANCELLED"
	ReasonCodePaymentFailed             ReasonCode = "PAYMENT_FAILED"
	ReasonCodePaymentDeclined           ReasonCode = "PAYMENT_DECLINED"
	ReasonCodePaymentExpired            ReasonCode = "PAYMENT_EXPIRED"
	ReasonCodePaymentCancelled          ReasonCode = "PAYMENT_CANCELLED"
	ReasonCodePaymentRequiresAction     ReasonCode = "PAYMENT_REQUIRES_ACTION"
	ReasonCodePaymentMethodInvalid      ReasonCode = "PAYMENT_METHOD_INVALID"
	ReasonCodePaymentMethodNotSupported ReasonCode = "PAYMENT_METHOD_NOT_SUPPORTED"
	ReasonCodePaymentMethodExpired      ReasonCode = "PAYMENT_METHOD_EXPIRED"
	ReasonCodePaymentMethodDeclined     ReasonCode = "PAYMENT_METHOD_DECLINED"
	ReasonCodePaymentMethodNotFound     ReasonCode = "PAYMENT_METHOD_NOT_FOUND"
	ReasonCodeInsufficientFunds         ReasonCode = "INSUFFICIENT_FUNDS"
	ReasonCodePaymentLimitExceeded      ReasonCode = "PAYMENT_LIMIT_EXCEEDED"
	ReasonCodeCurrencyNotSupported      ReasonCode = "CURRENCY_NOT_SUPPORTED"
	ReasonCodeAmountInvalid             ReasonCode = "AMOUNT_INVALID"
	ReasonCodeAmountTooSmall            ReasonCode = "AMOUNT_TOO_SMALL"
	ReasonCodeAmountTooLarge            ReasonCode = "AMOUNT_TOO_LARGE"
	ReasonCodeRefundNotAllowed          ReasonCode = "REFUND_NOT_ALLOWED"
	ReasonCodeRefundAlreadyProcessed    ReasonCode = "REFUND_ALREADY_PROCESSED"
	ReasonCodeRefundAmountExceeded      ReasonCode = "REFUND_AMOUNT_EXCEEDED"
	ReasonCodeRefundFailed              ReasonCode = "REFUND_FAILED"

	// ORDER
	ReasonCodeOrderNotFound            ReasonCode = "ORDER_NOT_FOUND"
	ReasonCodeOrderAlreadyExists       ReasonCode = "ORDER_ALREADY_EXISTS"
	ReasonCodeOrderAlreadyPaid         ReasonCode = "ORDER_ALREADY_PAID"
	ReasonCodeOrderAlreadyCancelled    ReasonCode = "ORDER_ALREADY_CANCELLED"
	ReasonCodeOrderAlreadyCompleted    ReasonCode = "ORDER_ALREADY_COMPLETED"
	ReasonCodeOrderAlreadyShipped      ReasonCode = "ORDER_ALREADY_SHIPPED"
	ReasonCodeOrderCannotBeCancelled   ReasonCode = "ORDER_CANNOT_BE_CANCELLED"
	ReasonCodeOrderCannotBeUpdated     ReasonCode = "ORDER_CANNOT_BE_UPDATED"
	ReasonCodeOrderCannotBeDeleted     ReasonCode = "ORDER_CANNOT_BE_DELETED"
	ReasonCodeOrderCannotBeRefunded    ReasonCode = "ORDER_CANNOT_BE_REFUNDED"
	ReasonCodeOrderItemNotFound        ReasonCode = "ORDER_ITEM_NOT_FOUND"
	ReasonCodeOrderItemOutOfStock      ReasonCode = "ORDER_ITEM_OUT_OF_STOCK"
	ReasonCodeOrderItemUnavailable     ReasonCode = "ORDER_ITEM_UNAVAILABLE"
	ReasonCodeOrderMinimumAmountNotMet ReasonCode = "ORDER_MINIMUM_AMOUNT_NOT_MET"
	ReasonCodeOrderExpired             ReasonCode = "ORDER_EXPIRED"
	ReasonCodeOrderPaymentRequired     ReasonCode = "ORDER_PAYMENT_REQUIRED"
	ReasonCodeOrderPaymentFailed       ReasonCode = "ORDER_PAYMENT_FAILED"

	// INVENTORY
	ReasonCodeProductNotFound       ReasonCode = "PRODUCT_NOT_FOUND"
	ReasonCodeProductUnavailable    ReasonCode = "PRODUCT_UNAVAILABLE"
	ReasonCodeProductInactive       ReasonCode = "PRODUCT_INACTIVE"
	ReasonCodeProductDiscontinued   ReasonCode = "PRODUCT_DISCONTINUED"
	ReasonCodeInventoryNotFound     ReasonCode = "INVENTORY_NOT_FOUND"
	ReasonCodeInventoryUnavailable  ReasonCode = "INVENTORY_UNAVAILABLE"
	ReasonCodeInsufficientInventory ReasonCode = "INSUFFICIENT_INVENTORY"
	ReasonCodeOutOfStock            ReasonCode = "OUT_OF_STOCK"
	ReasonCodeStockReserved         ReasonCode = "STOCK_RESERVED"
	ReasonCodeStockAlreadyReserved  ReasonCode = "STOCK_ALREADY_RESERVED"
	ReasonCodeReservationExpired    ReasonCode = "RESERVATION_EXPIRED"
	ReasonCodeReservationNotFound   ReasonCode = "RESERVATION_NOT_FOUND"

	// SHIPPING / DELIVERY
	ReasonCodeShippingAddressInvalid     ReasonCode = "SHIPPING_ADDRESS_INVALID"
	ReasonCodeShippingAddressNotFound    ReasonCode = "SHIPPING_ADDRESS_NOT_FOUND"
	ReasonCodeShippingMethodNotSupported ReasonCode = "SHIPPING_METHOD_NOT_SUPPORTED"
	ReasonCodeShippingMethodUnavailable  ReasonCode = "SHIPPING_METHOD_UNAVAILABLE"
	ReasonCodeDeliveryNotFound           ReasonCode = "DELIVERY_NOT_FOUND"
	ReasonCodeDeliveryUnavailable        ReasonCode = "DELIVERY_UNAVAILABLE"
	ReasonCodeDeliveryCancelled          ReasonCode = "DELIVERY_CANCELLED"
	ReasonCodeDeliveryFailed             ReasonCode = "DELIVERY_FAILED"
	ReasonCodeDeliveryAlreadyCompleted   ReasonCode = "DELIVERY_ALREADY_COMPLETED"
	ReasonCodeInvalidPostalCode          ReasonCode = "INVALID_POSTAL_CODE"
	ReasonCodeInvalidCountry             ReasonCode = "INVALID_COUNTRY"
	ReasonCodeInvalidRegion              ReasonCode = "INVALID_REGION"

	// COUPON / PROMOTION
	ReasonCodeCouponNotFound          ReasonCode = "COUPON_NOT_FOUND"
	ReasonCodeCouponInvalid           ReasonCode = "COUPON_INVALID"
	ReasonCodeCouponExpired           ReasonCode = "COUPON_EXPIRED"
	ReasonCodeCouponNotActive         ReasonCode = "COUPON_NOT_ACTIVE"
	ReasonCodeCouponAlreadyUsed       ReasonCode = "COUPON_ALREADY_USED"
	ReasonCodeCouponUsageLimitReached ReasonCode = "COUPON_USAGE_LIMIT_REACHED"
	ReasonCodePromotionNotFound       ReasonCode = "PROMOTION_NOT_FOUND"
	ReasonCodePromotionExpired        ReasonCode = "PROMOTION_EXPIRED"
	ReasonCodePromotionNotActive      ReasonCode = "PROMOTION_NOT_ACTIVE"
	ReasonCodePromotionNotApplicable  ReasonCode = "PROMOTION_NOT_APPLICABLE"
	ReasonCodeMinimumPurchaseNotMet   ReasonCode = "MINIMUM_PURCHASE_NOT_MET"
	ReasonCodeMaximumDiscountExceeded ReasonCode = "MAXIMUM_DISCOUNT_EXCEEDED"
	ReasonCodeProductNotEligible      ReasonCode = "PRODUCT_NOT_ELIGIBLE"
	ReasonCodeUserNotEligible         ReasonCode = "USER_NOT_ELIGIBLE"

	// VERIFICATION
	ReasonCodeVerificationRequired         ReasonCode = "VERIFICATION_REQUIRED"
	ReasonCodeVerificationFailed           ReasonCode = "VERIFICATION_FAILED"
	ReasonCodeVerificationExpired          ReasonCode = "VERIFICATION_EXPIRED"
	ReasonCodeVerificationInvalid          ReasonCode = "VERIFICATION_INVALID"
	ReasonCodeVerificationAlreadyCompleted ReasonCode = "VERIFICATION_ALREADY_COMPLETED"
	ReasonCodeVerificationNotFound         ReasonCode = "VERIFICATION_NOT_FOUND"
	ReasonCodeEmailVerificationRequired    ReasonCode = "EMAIL_VERIFICATION_REQUIRED"
	ReasonCodeEmailVerificationFailed      ReasonCode = "EMAIL_VERIFICATION_FAILED"
	ReasonCodeEmailVerificationExpired     ReasonCode = "EMAIL_VERIFICATION_EXPIRED"
	ReasonCodePhoneVerificationRequired    ReasonCode = "PHONE_VERIFICATION_REQUIRED"
	ReasonCodePhoneVerificationFailed      ReasonCode = "PHONE_VERIFICATION_FAILED"
	ReasonCodePhoneVerificationExpired     ReasonCode = "PHONE_VERIFICATION_EXPIRED"

	// BUSINESS RULE
	ReasonCodeBusinessRuleViolation      ReasonCode = "BUSINESS_RULE_VIOLATION"
	ReasonCodeBusinessRuleNotSatisfied   ReasonCode = "BUSINESS_RULE_NOT_SATISFIED"
	ReasonCodeBusinessConditionNotMet    ReasonCode = "BUSINESS_CONDITION_NOT_MET"
	ReasonCodeOperationNotEligible       ReasonCode = "OPERATION_NOT_ELIGIBLE"
	ReasonCodeMinimumRequirementNotMet   ReasonCode = "MINIMUM_REQUIREMENT_NOT_MET"
	ReasonCodeMaximumRequirementExceeded ReasonCode = "MAXIMUM_REQUIREMENT_EXCEEDED"
	ReasonCodeDependencyNotSatisfied     ReasonCode = "DEPENDENCY_NOT_SATISFIED"
	ReasonCodePrerequisiteNotMet         ReasonCode = "PREREQUISITE_NOT_MET"

	// DEPENDENCY / EXTERNAL SERVICE
	ReasonCodeDependencyUnavailable      ReasonCode = "DEPENDENCY_UNAVAILABLE"
	ReasonCodeDependencyTimeout          ReasonCode = "DEPENDENCY_TIMEOUT"
	ReasonCodeDependencyFailed           ReasonCode = "DEPENDENCY_FAILED"
	ReasonCodeDependencyRejected         ReasonCode = "DEPENDENCY_REJECTED"
	ReasonCodeDependencyInvalidResponse  ReasonCode = "DEPENDENCY_INVALID_RESPONSE"
	ReasonCodeExternalServiceUnavailable ReasonCode = "EXTERNAL_SERVICE_UNAVAILABLE"
	ReasonCodeExternalServiceTimeout     ReasonCode = "EXTERNAL_SERVICE_TIMEOUT"
	ReasonCodeExternalServiceError       ReasonCode = "EXTERNAL_SERVICE_ERROR"
	ReasonCodeThirdPartyError            ReasonCode = "THIRD_PARTY_ERROR"
	ReasonCodeThirdPartyTimeout          ReasonCode = "THIRD_PARTY_TIMEOUT"
	ReasonCodeThirdPartyUnavailable      ReasonCode = "THIRD_PARTY_UNAVAILABLE"

	// DATABASE / PERSISTENCE
	ReasonCodePersistenceError              ReasonCode = "PERSISTENCE_ERROR"
	ReasonCodeDatabaseError                 ReasonCode = "DATABASE_ERROR"
	ReasonCodeDatabaseUnavailable           ReasonCode = "DATABASE_UNAVAILABLE"
	ReasonCodeDatabaseTimeout               ReasonCode = "DATABASE_TIMEOUT"
	ReasonCodeRecordNotFound                ReasonCode = "RECORD_NOT_FOUND"
	ReasonCodeRecordAlreadyExists           ReasonCode = "RECORD_ALREADY_EXISTS"
	ReasonCodeUniqueConstraintViolation     ReasonCode = "UNIQUE_CONSTRAINT_VIOLATION"
	ReasonCodeForeignKeyConstraintViolation ReasonCode = "FOREIGN_KEY_CONSTRAINT_VIOLATION"
	ReasonCodeConstraintViolation           ReasonCode = "CONSTRAINT_VIOLATION"
	ReasonCodeTransactionFailed             ReasonCode = "TRANSACTION_FAILED"
	ReasonCodeTransactionTimeout            ReasonCode = "TRANSACTION_TIMEOUT"
	ReasonCodeTransactionRollback           ReasonCode = "TRANSACTION_ROLLBACK"

	// CACHE
	ReasonCodeCacheError       ReasonCode = "CACHE_ERROR"
	ReasonCodeCacheUnavailable ReasonCode = "CACHE_UNAVAILABLE"
	ReasonCodeCacheTimeout     ReasonCode = "CACHE_TIMEOUT"
	ReasonCodeCacheMiss        ReasonCode = "CACHE_MISS"

	// SECURITY
	ReasonCodeSecurityViolation ReasonCode = "SECURITY_VIOLATION"
	ReasonCodeSuspiciousRequest ReasonCode = "SUSPICIOUS_REQUEST"
	ReasonCodeInvalidSignature  ReasonCode = "INVALID_SIGNATURE"
	ReasonCodeSignatureExpired  ReasonCode = "SIGNATURE_EXPIRED"
	ReasonCodeSignatureInvalid  ReasonCode = "SIGNATURE_INVALID"
	ReasonCodeCSRFTokenInvalid  ReasonCode = "CSRF_TOKEN_INVALID"
	ReasonCodeCSRFTokenMissing  ReasonCode = "CSRF_TOKEN_MISSING"
	ReasonCodeEncryptionFailed  ReasonCode = "ENCRYPTION_FAILED"
	ReasonCodeDecryptionFailed  ReasonCode = "DECRYPTION_FAILED"
	ReasonCodeHashingFailed     ReasonCode = "HASHING_FAILED"

	// WEBHOOK / EVENT
	ReasonCodeWebhookNotFound         ReasonCode = "WEBHOOK_NOT_FOUND"
	ReasonCodeWebhookInvalid          ReasonCode = "WEBHOOK_INVALID"
	ReasonCodeWebhookDisabled         ReasonCode = "WEBHOOK_DISABLED"
	ReasonCodeWebhookFailed           ReasonCode = "WEBHOOK_FAILED"
	ReasonCodeWebhookTimeout          ReasonCode = "WEBHOOK_TIMEOUT"
	ReasonCodeEventNotFound           ReasonCode = "EVENT_NOT_FOUND"
	ReasonCodeEventInvalid            ReasonCode = "EVENT_INVALID"
	ReasonCodeEventAlreadyProcessed   ReasonCode = "EVENT_ALREADY_PROCESSED"
	ReasonCodeEventProcessingFailed   ReasonCode = "EVENT_PROCESSING_FAILED"
	ReasonCodeEventVersionUnsupported ReasonCode = "EVENT_VERSION_UNSUPPORTED"

	// ASYNC / JOB
	ReasonCodeJobNotFound         ReasonCode = "JOB_NOT_FOUND"
	ReasonCodeJobAlreadyExists    ReasonCode = "JOB_ALREADY_EXISTS"
	ReasonCodeJobAlreadyRunning   ReasonCode = "JOB_ALREADY_RUNNING"
	ReasonCodeJobAlreadyCompleted ReasonCode = "JOB_ALREADY_COMPLETED"
	ReasonCodeJobAlreadyFailed    ReasonCode = "JOB_ALREADY_FAILED"
	ReasonCodeJobCancelled        ReasonCode = "JOB_CANCELLED"
	ReasonCodeJobNotCancellable   ReasonCode = "JOB_NOT_CANCELLABLE"
	ReasonCodeJobTimeout          ReasonCode = "JOB_TIMEOUT"
	ReasonCodeJobFailed           ReasonCode = "JOB_FAILED"
	ReasonCodeJobRetryExhausted   ReasonCode = "JOB_RETRY_EXHAUSTED"
	ReasonCodeJobPayloadInvalid   ReasonCode = "JOB_PAYLOAD_INVALID"

	// IMPORT / EXPORT
	ReasonCodeImportFailed             ReasonCode = "IMPORT_FAILED"
	ReasonCodeImportInvalid            ReasonCode = "IMPORT_INVALID"
	ReasonCodeImportEmpty              ReasonCode = "IMPORT_EMPTY"
	ReasonCodeImportFormatNotSupported ReasonCode = "IMPORT_FORMAT_NOT_SUPPORTED"
	ReasonCodeImportDuplicateRecord    ReasonCode = "IMPORT_DUPLICATE_RECORD"
	ReasonCodeImportValidationFailed   ReasonCode = "IMPORT_VALIDATION_FAILED"
	ReasonCodeExportFailed             ReasonCode = "EXPORT_FAILED"
	ReasonCodeExportNotReady           ReasonCode = "EXPORT_NOT_READY"
	ReasonCodeExportExpired            ReasonCode = "EXPORT_EXPIRED"
	ReasonCodeExportTooLarge           ReasonCode = "EXPORT_TOO_LARGE"

	// NOTIFICATION
	ReasonCodeNotificationNotFound     ReasonCode = "NOTIFICATION_NOT_FOUND"
	ReasonCodeNotificationFailed       ReasonCode = "NOTIFICATION_FAILED"
	ReasonCodeNotificationNotDelivered ReasonCode = "NOTIFICATION_NOT_DELIVERED"
	ReasonCodeNotificationAlreadySent  ReasonCode = "NOTIFICATION_ALREADY_SENT"
	ReasonCodeEmailSendFailed          ReasonCode = "EMAIL_SEND_FAILED"
	ReasonCodeSMSSendFailed            ReasonCode = "SMS_SEND_FAILED"
	ReasonCodePushSendFailed           ReasonCode = "PUSH_SEND_FAILED"
	ReasonCodeRecipientInvalid         ReasonCode = "RECIPIENT_INVALID"
	ReasonCodeRecipientUnreachable     ReasonCode = "RECIPIENT_UNREACHABLE"

	// SYSTEM
	ReasonCodeInternalError          ReasonCode = "INTERNAL_ERROR"
	ReasonCodeServiceUnavailable     ReasonCode = "SERVICE_UNAVAILABLE"
	ReasonCodeServiceTimeout         ReasonCode = "SERVICE_TIMEOUT"
	ReasonCodeRequestTimeout         ReasonCode = "REQUEST_TIMEOUT"
	ReasonCodeOperationTimeout       ReasonCode = "OPERATION_TIMEOUT"
	ReasonCodeConfigurationError     ReasonCode = "CONFIGURATION_ERROR"
	ReasonCodeFeatureNotAvailable    ReasonCode = "FEATURE_NOT_AVAILABLE"
	ReasonCodeFeatureDisabled        ReasonCode = "FEATURE_DISABLED"
	ReasonCodeMaintenanceMode        ReasonCode = "MAINTENANCE_MODE"
	ReasonCodeNotImplemented         ReasonCode = "NOT_IMPLEMENTED"
	ReasonCodeUnsupportedVersion     ReasonCode = "UNSUPPORTED_VERSION"
	ReasonCodeAPIVersionNotSupported ReasonCode = "API_VERSION_NOT_SUPPORTED"
)

// ReasonCodes is a map of all defined reason codes to their corresponding [reason] struct, which includes the code and its category.
// This map allows for easy lookup of reason codes and their associated categories.
var ReasonCodes = map[ReasonCode]reason{
	// GENERAL / COMMON
	// Common reason codes for general errors
	ReasonCodeInvalidRequest:        {code: ReasonCodeInvalidRequest, category: CategoryCommon},
	ReasonCodeMissingRequiredField:  {code: ReasonCodeMissingRequiredField, category: CategoryCommon},
	ReasonCodeInvalidField:          {code: ReasonCodeInvalidField, category: CategoryCommon},
	ReasonCodeInvalidParameter:      {code: ReasonCodeInvalidParameter, category: CategoryCommon},
	ReasonCodeInvalidQueryParameter: {code: ReasonCodeInvalidQueryParameter, category: CategoryCommon},
	ReasonCodeInvalidPathParameter:  {code: ReasonCodeInvalidPathParameter, category: CategoryCommon},
	ReasonCodeInvalidHeader:         {code: ReasonCodeInvalidHeader, category: CategoryCommon},
	ReasonCodeInvalidBody:           {code: ReasonCodeInvalidBody, category: CategoryCommon},
	ReasonCodeInvalidFormat:         {code: ReasonCodeInvalidFormat, category: CategoryCommon},
	ReasonCodeInvalidValue:          {code: ReasonCodeInvalidValue, category: CategoryCommon},
	ReasonCodeValueOutOfRange:       {code: ReasonCodeValueOutOfRange, category: CategoryCommon},
	ReasonCodeValueTooLong:          {code: ReasonCodeValueTooLong, category: CategoryCommon},
	ReasonCodeValueTooShort:         {code: ReasonCodeValueTooShort, category: CategoryCommon},
	ReasonCodeDuplicateValue:        {code: ReasonCodeDuplicateValue, category: CategoryCommon},
	ReasonCodeUnsupportedValue:      {code: ReasonCodeUnsupportedValue, category: CategoryCommon},
	ReasonCodeUnsupportedOperation:  {code: ReasonCodeUnsupportedOperation, category: CategoryCommon},
	ReasonCodeOperationNotAllowed:   {code: ReasonCodeOperationNotAllowed, category: CategoryCommon},
	ReasonCodeOperationNotSupported: {code: ReasonCodeOperationNotSupported, category: CategoryCommon},
	ReasonCodeResourceNotFound:      {code: ReasonCodeResourceNotFound, category: CategoryCommon},
	ReasonCodeResourceAlreadyExists: {code: ReasonCodeResourceAlreadyExists, category: CategoryCommon},
	ReasonCodeResourceConflict:      {code: ReasonCodeResourceConflict, category: CategoryCommon},
	ReasonCodeResourceUnavailable:   {code: ReasonCodeResourceUnavailable, category: CategoryCommon},
	ReasonCodeResourceExpired:       {code: ReasonCodeResourceExpired, category: CategoryCommon},
	ReasonCodeResourceLocked:        {code: ReasonCodeResourceLocked, category: CategoryCommon},
	ReasonCodeResourceArchived:      {code: ReasonCodeResourceArchived, category: CategoryCommon},
	ReasonCodeResourceDeleted:       {code: ReasonCodeResourceDeleted, category: CategoryCommon},

	// VALIDATION
	// Validation reason codes for input validation errors
	ReasonCodeValidationFailed:     {code: ReasonCodeValidationFailed, category: CategoryValidation},
	ReasonCodeRequiredFieldMissing: {code: ReasonCodeRequiredFieldMissing, category: CategoryValidation},
	ReasonCodeFieldInvalid:         {code: ReasonCodeFieldInvalid, category: CategoryValidation},
	ReasonCodeFieldEmpty:           {code: ReasonCodeFieldEmpty, category: CategoryValidation},
	ReasonCodeFieldNull:            {code: ReasonCodeFieldNull, category: CategoryValidation},
	ReasonCodeFieldTypeInvalid:     {code: ReasonCodeFieldTypeInvalid, category: CategoryValidation},
	ReasonCodeFieldFormatInvalid:   {code: ReasonCodeFieldFormatInvalid, category: CategoryValidation},
	ReasonCodeInvalidEmail:         {code: ReasonCodeInvalidEmail, category: CategoryValidation},
	ReasonCodeInvalidPhoneNumber:   {code: ReasonCodeInvalidPhoneNumber, category: CategoryValidation},
	ReasonCodeInvalidURL:           {code: ReasonCodeInvalidURL, category: CategoryValidation},
	ReasonCodeInvalidUUID:          {code: ReasonCodeInvalidUUID, category: CategoryValidation},
	ReasonCodeInvalidDate:          {code: ReasonCodeInvalidDate, category: CategoryValidation},
	ReasonCodeInvalidDateTime:      {code: ReasonCodeInvalidDateTime, category: CategoryValidation},
	ReasonCodeInvalidTimezone:      {code: ReasonCodeInvalidTimezone, category: CategoryValidation},
	ReasonCodeInvalidLocale:        {code: ReasonCodeInvalidLocale, category: CategoryValidation},
	ReasonCodeInvalidCurrency:      {code: ReasonCodeInvalidCurrency, category: CategoryValidation},
	ReasonCodeValueTooSmall:        {code: ReasonCodeValueTooSmall, category: CategoryValidation},
	ReasonCodeValueTooLarge:        {code: ReasonCodeValueTooLarge, category: CategoryValidation},
	ReasonCodeInvalidEnumValue:     {code: ReasonCodeInvalidEnumValue, category: CategoryValidation},
	ReasonCodeInvalidRegexp:        {code: ReasonCodeInvalidRegexp, category: CategoryValidation},
	ReasonCodeInvalidEncoding:      {code: ReasonCodeInvalidEncoding, category: CategoryValidation},

	// AUTHENTICATION
	// Authentication reason codes for authentication and authorization errors
	ReasonCodeAuthenticationRequired: {code: ReasonCodeAuthenticationRequired, category: CategoryAuthentication},
	ReasonCodeAuthenticationFailed:   {code: ReasonCodeAuthenticationFailed, category: CategoryAuthentication},
	ReasonCodeInvalidCredentials:     {code: ReasonCodeInvalidCredentials, category: CategoryAuthentication},
	ReasonCodeInvalidPassword:        {code: ReasonCodeInvalidPassword, category: CategoryAuthentication},
	ReasonCodeInvalidUsername:        {code: ReasonCodeInvalidUsername, category: CategoryAuthentication},
	ReasonCodeInvalidToken:           {code: ReasonCodeInvalidToken, category: CategoryAuthentication},
	ReasonCodeTokenExpired:           {code: ReasonCodeTokenExpired, category: CategoryAuthentication},
	ReasonCodeTokenRevoked:           {code: ReasonCodeTokenRevoked, category: CategoryAuthentication},
	ReasonCodeTokenInvalid:           {code: ReasonCodeTokenInvalid, category: CategoryAuthentication},
	ReasonCodeTokenMalformed:         {code: ReasonCodeTokenMalformed, category: CategoryAuthentication},
	ReasonCodeTokenMissing:           {code: ReasonCodeTokenMissing, category: CategoryAuthentication},
	ReasonCodeTokenNotActive:         {code: ReasonCodeTokenNotActive, category: CategoryAuthentication},
	ReasonCodeSessionExpired:         {code: ReasonCodeSessionExpired, category: CategoryAuthentication},
	ReasonCodeSessionRevoked:         {code: ReasonCodeSessionRevoked, category: CategoryAuthentication},
	ReasonCodeSessionNotFound:        {code: ReasonCodeSessionNotFound, category: CategoryAuthentication},
	ReasonCodeMFARequired:            {code: ReasonCodeMFARequired, category: CategoryAuthentication},
	ReasonCodeMFAFailed:              {code: ReasonCodeMFAFailed, category: CategoryAuthentication},
	ReasonCodeMFAInvalid:             {code: ReasonCodeMFAInvalid, category: CategoryAuthentication},
	ReasonCodeMFAExpired:             {code: ReasonCodeMFAExpired, category: CategoryAuthentication},
	ReasonCodeMFANotEnabled:          {code: ReasonCodeMFANotEnabled, category: CategoryAuthentication},
	ReasonCodeMFAAlreadyEnabled:      {code: ReasonCodeMFAAlreadyEnabled, category: CategoryAuthentication},
	ReasonCodeMFAAlreadyVerified:     {code: ReasonCodeMFAAlreadyVerified, category: CategoryAuthentication},
	ReasonCodeOTPRequired:            {code: ReasonCodeOTPRequired, category: CategoryAuthentication},
	ReasonCodeOTPInvalid:             {code: ReasonCodeOTPInvalid, category: CategoryAuthentication},
	ReasonCodeOTPExpired:             {code: ReasonCodeOTPExpired, category: CategoryAuthentication},
	ReasonCodeOTPMaxAttemptsExceeded: {code: ReasonCodeOTPMaxAttemptsExceeded, category: CategoryAuthentication},
	ReasonCodeOTPAlreadyUsed:         {code: ReasonCodeOTPAlreadyUsed, category: CategoryAuthentication},
	ReasonCodeOTPNotFound:            {code: ReasonCodeOTPNotFound, category: CategoryAuthentication},

	// AUTHORIZATION / PERMISSION
	// Authorization reason codes for authorization and permission errors
	ReasonCodeAccessDenied:           {code: ReasonCodeAccessDenied, category: CategoryAuthorization},
	ReasonCodePermissionDenied:       {code: ReasonCodePermissionDenied, category: CategoryAuthorization},
	ReasonCodeInsufficientPermission: {code: ReasonCodeInsufficientPermission, category: CategoryAuthorization},
	ReasonCodeRoleRequired:           {code: ReasonCodeRoleRequired, category: CategoryAuthorization},
	ReasonCodeResourceAccessDenied:   {code: ReasonCodeResourceAccessDenied, category: CategoryAuthorization},
	ReasonCodeOperationForbidden:     {code: ReasonCodeOperationForbidden, category: CategoryAuthorization},
	ReasonCodeAccountSuspended:       {code: ReasonCodeAccountSuspended, category: CategoryAuthorization},
	ReasonCodeAccountDisabled:        {code: ReasonCodeAccountDisabled, category: CategoryAuthorization},
	ReasonCodeAccountLocked:          {code: ReasonCodeAccountLocked, category: CategoryAuthorization},

	// USER / ACCOUNT
	// User and account reason codes for user and account related errors
	ReasonCodeUserNotFound:                  {code: ReasonCodeUserNotFound, category: CategoryUserAccount},
	ReasonCodeUserAlreadyExists:             {code: ReasonCodeUserAlreadyExists, category: CategoryUserAccount},
	ReasonCodeUserAlreadyActive:             {code: ReasonCodeUserAlreadyActive, category: CategoryUserAccount},
	ReasonCodeUserAlreadyInactive:           {code: ReasonCodeUserAlreadyInactive, category: CategoryUserAccount},
	ReasonCodeUserAlreadyVerified:           {code: ReasonCodeUserAlreadyVerified, category: CategoryUserAccount},
	ReasonCodeUserNotVerified:               {code: ReasonCodeUserNotVerified, category: CategoryUserAccount},
	ReasonCodeUserEmailAlreadyExists:        {code: ReasonCodeUserEmailAlreadyExists, category: CategoryUserAccount},
	ReasonCodeUserEmailNotFound:             {code: ReasonCodeUserEmailNotFound, category: CategoryUserAccount},
	ReasonCodeUserEmailAlreadyVerified:      {code: ReasonCodeUserEmailAlreadyVerified, category: CategoryUserAccount},
	ReasonCodeUserEmailVerificationRequired: {code: ReasonCodeUserEmailVerificationRequired, category: CategoryUserAccount},
	ReasonCodeUserPhoneAlreadyExists:        {code: ReasonCodeUserPhoneAlreadyExists, category: CategoryUserAccount},
	ReasonCodeUserPhoneNotFound:             {code: ReasonCodeUserPhoneNotFound, category: CategoryUserAccount},
	ReasonCodeUserPhoneAlreadyVerified:      {code: ReasonCodeUserPhoneAlreadyVerified, category: CategoryUserAccount},
	ReasonCodeUserPhoneVerificationRequired: {code: ReasonCodeUserPhoneVerificationRequired, category: CategoryUserAccount},
	ReasonCodeUserCannotBeDeleted:           {code: ReasonCodeUserCannotBeDeleted, category: CategoryUserAccount},
	ReasonCodeUserCannotBeDisabled:          {code: ReasonCodeUserCannotBeDisabled, category: CategoryUserAccount},
	ReasonCodeUserCannotBeEnabled:           {code: ReasonCodeUserCannotBeEnabled, category: CategoryUserAccount},
	ReasonCodeUserCannotBeUpdated:           {code: ReasonCodeUserCannotBeUpdated, category: CategoryUserAccount},
	ReasonCodeUserCannotBeActivated:         {code: ReasonCodeUserCannotBeActivated, category: CategoryUserAccount},
	ReasonCodeUserCannotBeDeactivated:       {code: ReasonCodeUserCannotBeDeactivated, category: CategoryUserAccount},
	ReasonCodeUsernameAlreadyExists:         {code: ReasonCodeUsernameAlreadyExists, category: CategoryUserAccount},
	ReasonCodeUsernameUnavailable:           {code: ReasonCodeUsernameUnavailable, category: CategoryUserAccount},
	ReasonCodePasswordIncorrect:             {code: ReasonCodePasswordIncorrect, category: CategoryUserAccount},
	ReasonCodePasswordExpired:               {code: ReasonCodePasswordExpired, category: CategoryUserAccount},
	ReasonCodePasswordTooWeak:               {code: ReasonCodePasswordTooWeak, category: CategoryUserAccount},
	ReasonCodePasswordReused:                {code: ReasonCodePasswordReused, category: CategoryUserAccount},
	ReasonCodePasswordResetRequired:         {code: ReasonCodePasswordResetRequired, category: CategoryUserAccount},
	ReasonCodePasswordResetTokenInvalid:     {code: ReasonCodePasswordResetTokenInvalid, category: CategoryUserAccount},
	ReasonCodePasswordResetTokenExpired:     {code: ReasonCodePasswordResetTokenExpired, category: CategoryUserAccount},
	ReasonCodeAccountAlreadyExists:          {code: ReasonCodeAccountAlreadyExists, category: CategoryUserAccount},
	ReasonCodeAccountNotFound:               {code: ReasonCodeAccountNotFound, category: CategoryUserAccount},
	ReasonCodeAccountAlreadyActive:          {code: ReasonCodeAccountAlreadyActive, category: CategoryUserAccount},
	ReasonCodeAccountAlreadyDisabled:        {code: ReasonCodeAccountAlreadyDisabled, category: CategoryUserAccount},
	ReasonCodeAccountPending:                {code: ReasonCodeAccountPending, category: CategoryUserAccount},
	ReasonCodeAccountDeletionPending:        {code: ReasonCodeAccountDeletionPending, category: CategoryUserAccount},

	// RESOURCE LIFECYCLE
	// Resource lifecycle reason codes for resource creation, update, deletion, and state changes
	ReasonCodeResourceAlreadyCreated:  {code: ReasonCodeResourceAlreadyCreated, category: CategoryResourceLifecycle},
	ReasonCodeResourceAlreadyUpdated:  {code: ReasonCodeResourceAlreadyUpdated, category: CategoryResourceLifecycle},
	ReasonCodeResourceAlreadyDeleted:  {code: ReasonCodeResourceAlreadyDeleted, category: CategoryResourceLifecycle},
	ReasonCodeResourceNotActive:       {code: ReasonCodeResourceNotActive, category: CategoryResourceLifecycle},
	ReasonCodeResourceAlreadyActive:   {code: ReasonCodeResourceAlreadyActive, category: CategoryResourceLifecycle},
	ReasonCodeResourceNotInactive:     {code: ReasonCodeResourceNotInactive, category: CategoryResourceLifecycle},
	ReasonCodeResourceAlreadyInactive: {code: ReasonCodeResourceAlreadyInactive, category: CategoryResourceLifecycle},
	ReasonCodeResourceNotExpired:      {code: ReasonCodeResourceNotExpired, category: CategoryResourceLifecycle},
	ReasonCodeResourceAlreadyLocked:   {code: ReasonCodeResourceAlreadyLocked, category: CategoryResourceLifecycle},
	ReasonCodeResourceNotLocked:       {code: ReasonCodeResourceNotLocked, category: CategoryResourceLifecycle},
	ReasonCodeResourceAlreadyArchived: {code: ReasonCodeResourceAlreadyArchived, category: CategoryResourceLifecycle},
	ReasonCodeResourceNotArchived:     {code: ReasonCodeResourceNotArchived, category: CategoryResourceLifecycle},
	ReasonCodeResourcePending:         {code: ReasonCodeResourcePending, category: CategoryResourceLifecycle},
	ReasonCodeResourceProcessing:      {code: ReasonCodeResourceProcessing, category: CategoryResourceLifecycle},
	ReasonCodeResourceCompleted:       {code: ReasonCodeResourceCompleted, category: CategoryResourceLifecycle},
	ReasonCodeResourceFailed:          {code: ReasonCodeResourceFailed, category: CategoryResourceLifecycle},

	// STATE / STATE TRANSITION
	// Reason codes for invalid state or state transition errors
	ReasonCodeInvalidState:                   {code: ReasonCodeInvalidState, category: CategoryStateTransition},
	ReasonCodeInvalidStateTransition:         {code: ReasonCodeInvalidStateTransition, category: CategoryStateTransition},
	ReasonCodeStateTransitionNotAllowed:      {code: ReasonCodeStateTransitionNotAllowed, category: CategoryStateTransition},
	ReasonCodeActionNotAllowedInCurrentState: {code: ReasonCodeActionNotAllowedInCurrentState, category: CategoryStateTransition},
	ReasonCodeResourceNotReady:               {code: ReasonCodeResourceNotReady, category: CategoryStateTransition},
	ReasonCodeResourceNotEligible:            {code: ReasonCodeResourceNotEligible, category: CategoryStateTransition},
	ReasonCodeAlreadyProcessed:               {code: ReasonCodeAlreadyProcessed, category: CategoryStateTransition},
	ReasonCodeAlreadyCompleted:               {code: ReasonCodeAlreadyCompleted, category: CategoryStateTransition},
	ReasonCodeAlreadyCancelled:               {code: ReasonCodeAlreadyCancelled, category: CategoryStateTransition},
	ReasonCodeAlreadyRejected:                {code: ReasonCodeAlreadyRejected, category: CategoryStateTransition},
	ReasonCodeAlreadyApproved:                {code: ReasonCodeAlreadyApproved, category: CategoryStateTransition},
	ReasonCodeCannotBeCancelled:              {code: ReasonCodeCannotBeCancelled, category: CategoryStateTransition},
	ReasonCodeCannotBeUpdated:                {code: ReasonCodeCannotBeUpdated, category: CategoryStateTransition},
	ReasonCodeCannotBeDeleted:                {code: ReasonCodeCannotBeDeleted, category: CategoryStateTransition},
	ReasonCodeCannotBeRetried:                {code: ReasonCodeCannotBeRetried, category: CategoryStateTransition},
	ReasonCodeCannotBeReversed:               {code: ReasonCodeCannotBeReversed, category: CategoryStateTransition},

	// CONCURRENCY / IDEMPOTENCY
	// Reason codes for concurrency control and idempotency errors
	ReasonCodeConcurrentModification:  {code: ReasonCodeConcurrentModification, category: CategoryConcurrencyIdempotency},
	ReasonCodeResourceVersionConflict: {code: ReasonCodeResourceVersionConflict, category: CategoryConcurrencyIdempotency},
	ReasonCodeOptimisticLockConflict:  {code: ReasonCodeOptimisticLockConflict, category: CategoryConcurrencyIdempotency},
	ReasonCodeDuplicateRequest:        {code: ReasonCodeDuplicateRequest, category: CategoryConcurrencyIdempotency},
	ReasonCodeDuplicateOperation:      {code: ReasonCodeDuplicateOperation, category: CategoryConcurrencyIdempotency},
	ReasonCodeRequestAlreadyProcessed: {code: ReasonCodeRequestAlreadyProcessed, category: CategoryConcurrencyIdempotency},
	ReasonCodeIdempotencyKeyRequired:  {code: ReasonCodeIdempotencyKeyRequired, category: CategoryConcurrencyIdempotency},
	ReasonCodeIdempotencyKeyInvalid:   {code: ReasonCodeIdempotencyKeyInvalid, category: CategoryConcurrencyIdempotency},
	ReasonCodeIdempotencyKeyReused:    {code: ReasonCodeIdempotencyKeyReused, category: CategoryConcurrencyIdempotency},
	ReasonCodeIdempotencyConflict:     {code: ReasonCodeIdempotencyConflict, category: CategoryConcurrencyIdempotency},

	// RATE LIMIT / QUOTA
	// Reason codes for rate limiting and quota enforcement errors
	ReasonCodeRateLimitExceeded:       {code: ReasonCodeRateLimitExceeded, category: CategoryRateLimitQuota},
	ReasonCodeRequestLimitExceeded:    {code: ReasonCodeRequestLimitExceeded, category: CategoryRateLimitQuota},
	ReasonCodeQuotaExceeded:           {code: ReasonCodeQuotaExceeded, category: CategoryRateLimitQuota},
	ReasonCodeUsageLimitExceeded:      {code: ReasonCodeUsageLimitExceeded, category: CategoryRateLimitQuota},
	ReasonCodeConcurrentLimitExceeded: {code: ReasonCodeConcurrentLimitExceeded, category: CategoryRateLimitQuota},
	ReasonCodeDailyLimitExceeded:      {code: ReasonCodeDailyLimitExceeded, category: CategoryRateLimitQuota},
	ReasonCodeMonthlyLimitExceeded:    {code: ReasonCodeMonthlyLimitExceeded, category: CategoryRateLimitQuota},
	ReasonCodeTooManyRequests:         {code: ReasonCodeTooManyRequests, category: CategoryRateLimitQuota},
	ReasonCodeTooManyAttempts:         {code: ReasonCodeTooManyAttempts, category: CategoryRateLimitQuota},

	// PAGINATION / QUERY
	// Reason codes for pagination and query parameter errors
	ReasonCodeInvalidPage:           {code: ReasonCodeInvalidPage, category: CategoryPaginationQuery},
	ReasonCodeInvalidPageSize:       {code: ReasonCodeInvalidPageSize, category: CategoryPaginationQuery},
	ReasonCodeInvalidCursor:         {code: ReasonCodeInvalidCursor, category: CategoryPaginationQuery},
	ReasonCodeCursorExpired:         {code: ReasonCodeCursorExpired, category: CategoryPaginationQuery},
	ReasonCodeCursorInvalid:         {code: ReasonCodeCursorInvalid, category: CategoryPaginationQuery},
	ReasonCodeCursorMalformed:       {code: ReasonCodeCursorMalformed, category: CategoryPaginationQuery},
	ReasonCodeInvalidSort:           {code: ReasonCodeInvalidSort, category: CategoryPaginationQuery},
	ReasonCodeInvalidSortField:      {code: ReasonCodeInvalidSortField, category: CategoryPaginationQuery},
	ReasonCodeInvalidSortDirection:  {code: ReasonCodeInvalidSortDirection, category: CategoryPaginationQuery},
	ReasonCodeInvalidFilter:         {code: ReasonCodeInvalidFilter, category: CategoryPaginationQuery},
	ReasonCodeInvalidFilterField:    {code: ReasonCodeInvalidFilterField, category: CategoryPaginationQuery},
	ReasonCodeInvalidFilterOperator: {code: ReasonCodeInvalidFilterOperator, category: CategoryPaginationQuery},
	ReasonCodeInvalidSearchQuery:    {code: ReasonCodeInvalidSearchQuery, category: CategoryPaginationQuery},

	// FILE / UPLOAD
	// Reason codes for file and upload related errors
	ReasonCodeFileNotFound:           {code: ReasonCodeFileNotFound, category: CategoryFileUpload},
	ReasonCodeFileAlreadyExists:      {code: ReasonCodeFileAlreadyExists, category: CategoryFileUpload},
	ReasonCodeFileTooLarge:           {code: ReasonCodeFileTooLarge, category: CategoryFileUpload},
	ReasonCodeFileTooSmall:           {code: ReasonCodeFileTooSmall, category: CategoryFileUpload},
	ReasonCodeFileTypeNotAllowed:     {code: ReasonCodeFileTypeNotAllowed, category: CategoryFileUpload},
	ReasonCodeFileFormatNotSupported: {code: ReasonCodeFileFormatNotSupported, category: CategoryFileUpload},
	ReasonCodeFileCorrupted:          {code: ReasonCodeFileCorrupted, category: CategoryFileUpload},
	ReasonCodeFileEmpty:              {code: ReasonCodeFileEmpty, category: CategoryFileUpload},
	ReasonCodeFileUploadFailed:       {code: ReasonCodeFileUploadFailed, category: CategoryFileUpload},
	ReasonCodeFileDownloadFailed:     {code: ReasonCodeFileDownloadFailed, category: CategoryFileUpload},
	ReasonCodeInvalidFilename:        {code: ReasonCodeInvalidFilename, category: CategoryFileUpload},
	ReasonCodeFilenameTooLong:        {code: ReasonCodeFilenameTooLong, category: CategoryFileUpload},
	ReasonCodeStorageQuotaExceeded:   {code: ReasonCodeStorageQuotaExceeded, category: CategoryFileUpload},
	ReasonCodeStorageUnavailable:     {code: ReasonCodeStorageUnavailable, category: CategoryFileUpload},

	// PAYMENT
	// Reason codes for payment related errors
	ReasonCodePaymentRequired:           {code: ReasonCodePaymentRequired, category: CategoryPayment},
	ReasonCodePaymentNotFound:           {code: ReasonCodePaymentNotFound, category: CategoryPayment},
	ReasonCodePaymentAlreadyExists:      {code: ReasonCodePaymentAlreadyExists, category: CategoryPayment},
	ReasonCodePaymentAlreadyPaid:        {code: ReasonCodePaymentAlreadyPaid, category: CategoryPayment},
	ReasonCodePaymentAlreadyFailed:      {code: ReasonCodePaymentAlreadyFailed, category: CategoryPayment},
	ReasonCodePaymentAlreadyRefunded:    {code: ReasonCodePaymentAlreadyRefunded, category: CategoryPayment},
	ReasonCodePaymentAlreadyCancelled:   {code: ReasonCodePaymentAlreadyCancelled, category: CategoryPayment},
	ReasonCodePaymentFailed:             {code: ReasonCodePaymentFailed, category: CategoryPayment},
	ReasonCodePaymentDeclined:           {code: ReasonCodePaymentDeclined, category: CategoryPayment},
	ReasonCodePaymentExpired:            {code: ReasonCodePaymentExpired, category: CategoryPayment},
	ReasonCodePaymentCancelled:          {code: ReasonCodePaymentCancelled, category: CategoryPayment},
	ReasonCodePaymentRequiresAction:     {code: ReasonCodePaymentRequiresAction, category: CategoryPayment},
	ReasonCodePaymentMethodInvalid:      {code: ReasonCodePaymentMethodInvalid, category: CategoryPayment},
	ReasonCodePaymentMethodNotSupported: {code: ReasonCodePaymentMethodNotSupported, category: CategoryPayment},
	ReasonCodePaymentMethodExpired:      {code: ReasonCodePaymentMethodExpired, category: CategoryPayment},
	ReasonCodePaymentMethodDeclined:     {code: ReasonCodePaymentMethodDeclined, category: CategoryPayment},
	ReasonCodePaymentMethodNotFound:     {code: ReasonCodePaymentMethodNotFound, category: CategoryPayment},
	ReasonCodeInsufficientFunds:         {code: ReasonCodeInsufficientFunds, category: CategoryPayment},
	ReasonCodePaymentLimitExceeded:      {code: ReasonCodePaymentLimitExceeded, category: CategoryPayment},
	ReasonCodeCurrencyNotSupported:      {code: ReasonCodeCurrencyNotSupported, category: CategoryPayment},
	ReasonCodeAmountInvalid:             {code: ReasonCodeAmountInvalid, category: CategoryPayment},
	ReasonCodeAmountTooSmall:            {code: ReasonCodeAmountTooSmall, category: CategoryPayment},
	ReasonCodeAmountTooLarge:            {code: ReasonCodeAmountTooLarge, category: CategoryPayment},
	ReasonCodeRefundNotAllowed:          {code: ReasonCodeRefundNotAllowed, category: CategoryPayment},
	ReasonCodeRefundAlreadyProcessed:    {code: ReasonCodeRefundAlreadyProcessed, category: CategoryPayment},
	ReasonCodeRefundAmountExceeded:      {code: ReasonCodeRefundAmountExceeded, category: CategoryPayment},
	ReasonCodeRefundFailed:              {code: ReasonCodeRefundFailed, category: CategoryPayment},

	// ORDER
	// Reason codes for order related errors
	ReasonCodeOrderNotFound:            {code: ReasonCodeOrderNotFound, category: CategoryOrder},
	ReasonCodeOrderAlreadyExists:       {code: ReasonCodeOrderAlreadyExists, category: CategoryOrder},
	ReasonCodeOrderAlreadyPaid:         {code: ReasonCodeOrderAlreadyPaid, category: CategoryOrder},
	ReasonCodeOrderAlreadyCancelled:    {code: ReasonCodeOrderAlreadyCancelled, category: CategoryOrder},
	ReasonCodeOrderAlreadyCompleted:    {code: ReasonCodeOrderAlreadyCompleted, category: CategoryOrder},
	ReasonCodeOrderAlreadyShipped:      {code: ReasonCodeOrderAlreadyShipped, category: CategoryOrder},
	ReasonCodeOrderCannotBeCancelled:   {code: ReasonCodeOrderCannotBeCancelled, category: CategoryOrder},
	ReasonCodeOrderCannotBeUpdated:     {code: ReasonCodeOrderCannotBeUpdated, category: CategoryOrder},
	ReasonCodeOrderCannotBeDeleted:     {code: ReasonCodeOrderCannotBeDeleted, category: CategoryOrder},
	ReasonCodeOrderCannotBeRefunded:    {code: ReasonCodeOrderCannotBeRefunded, category: CategoryOrder},
	ReasonCodeOrderItemNotFound:        {code: ReasonCodeOrderItemNotFound, category: CategoryOrder},
	ReasonCodeOrderItemOutOfStock:      {code: ReasonCodeOrderItemOutOfStock, category: CategoryOrder},
	ReasonCodeOrderItemUnavailable:     {code: ReasonCodeOrderItemUnavailable, category: CategoryOrder},
	ReasonCodeOrderMinimumAmountNotMet: {code: ReasonCodeOrderMinimumAmountNotMet, category: CategoryOrder},
	ReasonCodeOrderExpired:             {code: ReasonCodeOrderExpired, category: CategoryOrder},
	ReasonCodeOrderPaymentRequired:     {code: ReasonCodeOrderPaymentRequired, category: CategoryOrder},
	ReasonCodeOrderPaymentFailed:       {code: ReasonCodeOrderPaymentFailed, category: CategoryOrder},

	// INVENTORY
	// Reason codes for inventory and product stock related errors
	ReasonCodeProductNotFound:       {code: ReasonCodeProductNotFound, category: CategoryInventory},
	ReasonCodeProductUnavailable:    {code: ReasonCodeProductUnavailable, category: CategoryInventory},
	ReasonCodeProductInactive:       {code: ReasonCodeProductInactive, category: CategoryInventory},
	ReasonCodeProductDiscontinued:   {code: ReasonCodeProductDiscontinued, category: CategoryInventory},
	ReasonCodeInventoryNotFound:     {code: ReasonCodeInventoryNotFound, category: CategoryInventory},
	ReasonCodeInventoryUnavailable:  {code: ReasonCodeInventoryUnavailable, category: CategoryInventory},
	ReasonCodeInsufficientInventory: {code: ReasonCodeInsufficientInventory, category: CategoryInventory},
	ReasonCodeOutOfStock:            {code: ReasonCodeOutOfStock, category: CategoryInventory},
	ReasonCodeStockReserved:         {code: ReasonCodeStockReserved, category: CategoryInventory},
	ReasonCodeStockAlreadyReserved:  {code: ReasonCodeStockAlreadyReserved, category: CategoryInventory},
	ReasonCodeReservationExpired:    {code: ReasonCodeReservationExpired, category: CategoryInventory},
	ReasonCodeReservationNotFound:   {code: ReasonCodeReservationNotFound, category: CategoryInventory},

	// SHIPPING / DELIVERY
	// Reason codes for shipping and delivery related errors
	ReasonCodeShippingAddressInvalid:     {code: ReasonCodeShippingAddressInvalid, category: CategoryShippingDelivery},
	ReasonCodeShippingAddressNotFound:    {code: ReasonCodeShippingAddressNotFound, category: CategoryShippingDelivery},
	ReasonCodeShippingMethodNotSupported: {code: ReasonCodeShippingMethodNotSupported, category: CategoryShippingDelivery},
	ReasonCodeShippingMethodUnavailable:  {code: ReasonCodeShippingMethodUnavailable, category: CategoryShippingDelivery},
	ReasonCodeDeliveryNotFound:           {code: ReasonCodeDeliveryNotFound, category: CategoryShippingDelivery},
	ReasonCodeDeliveryUnavailable:        {code: ReasonCodeDeliveryUnavailable, category: CategoryShippingDelivery},
	ReasonCodeDeliveryCancelled:          {code: ReasonCodeDeliveryCancelled, category: CategoryShippingDelivery},
	ReasonCodeDeliveryFailed:             {code: ReasonCodeDeliveryFailed, category: CategoryShippingDelivery},
	ReasonCodeDeliveryAlreadyCompleted:   {code: ReasonCodeDeliveryAlreadyCompleted, category: CategoryShippingDelivery},
	ReasonCodeInvalidPostalCode:          {code: ReasonCodeInvalidPostalCode, category: CategoryShippingDelivery},
	ReasonCodeInvalidCountry:             {code: ReasonCodeInvalidCountry, category: CategoryShippingDelivery},
	ReasonCodeInvalidRegion:              {code: ReasonCodeInvalidRegion, category: CategoryShippingDelivery},

	// COUPON / PROMOTION
	// Reason codes for coupon and promotion related errors
	ReasonCodeCouponNotFound:          {code: ReasonCodeCouponNotFound, category: CategoryCouponPromotion},
	ReasonCodeCouponInvalid:           {code: ReasonCodeCouponInvalid, category: CategoryCouponPromotion},
	ReasonCodeCouponExpired:           {code: ReasonCodeCouponExpired, category: CategoryCouponPromotion},
	ReasonCodeCouponNotActive:         {code: ReasonCodeCouponNotActive, category: CategoryCouponPromotion},
	ReasonCodeCouponAlreadyUsed:       {code: ReasonCodeCouponAlreadyUsed, category: CategoryCouponPromotion},
	ReasonCodeCouponUsageLimitReached: {code: ReasonCodeCouponUsageLimitReached, category: CategoryCouponPromotion},
	ReasonCodePromotionNotFound:       {code: ReasonCodePromotionNotFound, category: CategoryCouponPromotion},
	ReasonCodePromotionExpired:        {code: ReasonCodePromotionExpired, category: CategoryCouponPromotion},
	ReasonCodePromotionNotActive:      {code: ReasonCodePromotionNotActive, category: CategoryCouponPromotion},
	ReasonCodePromotionNotApplicable:  {code: ReasonCodePromotionNotApplicable, category: CategoryCouponPromotion},
	ReasonCodeMinimumPurchaseNotMet:   {code: ReasonCodeMinimumPurchaseNotMet, category: CategoryCouponPromotion},
	ReasonCodeMaximumDiscountExceeded: {code: ReasonCodeMaximumDiscountExceeded, category: CategoryCouponPromotion},
	ReasonCodeProductNotEligible:      {code: ReasonCodeProductNotEligible, category: CategoryCouponPromotion},
	ReasonCodeUserNotEligible:         {code: ReasonCodeUserNotEligible, category: CategoryCouponPromotion},

	// VERIFICATION
	// Reason codes for verification related errors
	ReasonCodeVerificationRequired:         {code: ReasonCodeVerificationRequired, category: CategoryVerification},
	ReasonCodeVerificationFailed:           {code: ReasonCodeVerificationFailed, category: CategoryVerification},
	ReasonCodeVerificationExpired:          {code: ReasonCodeVerificationExpired, category: CategoryVerification},
	ReasonCodeVerificationInvalid:          {code: ReasonCodeVerificationInvalid, category: CategoryVerification},
	ReasonCodeVerificationAlreadyCompleted: {code: ReasonCodeVerificationAlreadyCompleted, category: CategoryVerification},
	ReasonCodeVerificationNotFound:         {code: ReasonCodeVerificationNotFound, category: CategoryVerification},
	ReasonCodeEmailVerificationRequired:    {code: ReasonCodeEmailVerificationRequired, category: CategoryVerification},
	ReasonCodeEmailVerificationFailed:      {code: ReasonCodeEmailVerificationFailed, category: CategoryVerification},
	ReasonCodeEmailVerificationExpired:     {code: ReasonCodeEmailVerificationExpired, category: CategoryVerification},
	ReasonCodePhoneVerificationRequired:    {code: ReasonCodePhoneVerificationRequired, category: CategoryVerification},
	ReasonCodePhoneVerificationFailed:      {code: ReasonCodePhoneVerificationFailed, category: CategoryVerification},
	ReasonCodePhoneVerificationExpired:     {code: ReasonCodePhoneVerificationExpired, category: CategoryVerification},

	// BUSINESS RULE
	// Reason codes for business rule related errors
	ReasonCodeBusinessRuleViolation:      {code: ReasonCodeBusinessRuleViolation, category: CategoryBusinessRule},
	ReasonCodeBusinessRuleNotSatisfied:   {code: ReasonCodeBusinessRuleNotSatisfied, category: CategoryBusinessRule},
	ReasonCodeBusinessConditionNotMet:    {code: ReasonCodeBusinessConditionNotMet, category: CategoryBusinessRule},
	ReasonCodeOperationNotEligible:       {code: ReasonCodeOperationNotEligible, category: CategoryBusinessRule},
	ReasonCodeMinimumRequirementNotMet:   {code: ReasonCodeMinimumRequirementNotMet, category: CategoryBusinessRule},
	ReasonCodeMaximumRequirementExceeded: {code: ReasonCodeMaximumRequirementExceeded, category: CategoryBusinessRule},
	ReasonCodeDependencyNotSatisfied:     {code: ReasonCodeDependencyNotSatisfied, category: CategoryBusinessRule},
	ReasonCodePrerequisiteNotMet:         {code: ReasonCodePrerequisiteNotMet, category: CategoryBusinessRule},

	// DEPENDENCY / EXTERNAL SERVICE
	// Reason codes for dependency and external service related errors
	ReasonCodeDependencyUnavailable:      {code: ReasonCodeDependencyUnavailable, category: CategoryDependencyExternalService},
	ReasonCodeDependencyTimeout:          {code: ReasonCodeDependencyTimeout, category: CategoryDependencyExternalService},
	ReasonCodeDependencyFailed:           {code: ReasonCodeDependencyFailed, category: CategoryDependencyExternalService},
	ReasonCodeDependencyRejected:         {code: ReasonCodeDependencyRejected, category: CategoryDependencyExternalService},
	ReasonCodeDependencyInvalidResponse:  {code: ReasonCodeDependencyInvalidResponse, category: CategoryDependencyExternalService},
	ReasonCodeExternalServiceUnavailable: {code: ReasonCodeExternalServiceUnavailable, category: CategoryDependencyExternalService},
	ReasonCodeExternalServiceTimeout:     {code: ReasonCodeExternalServiceTimeout, category: CategoryDependencyExternalService},
	ReasonCodeExternalServiceError:       {code: ReasonCodeExternalServiceError, category: CategoryDependencyExternalService},
	ReasonCodeThirdPartyError:            {code: ReasonCodeThirdPartyError, category: CategoryDependencyExternalService},
	ReasonCodeThirdPartyTimeout:          {code: ReasonCodeThirdPartyTimeout, category: CategoryDependencyExternalService},
	ReasonCodeThirdPartyUnavailable:      {code: ReasonCodeThirdPartyUnavailable, category: CategoryDependencyExternalService},

	// DATABASE / PERSISTENCE
	// Reason codes for database and persistence related errors
	ReasonCodePersistenceError:              {code: ReasonCodePersistenceError, category: CategoryDatabasePersistence},
	ReasonCodeDatabaseError:                 {code: ReasonCodeDatabaseError, category: CategoryDatabasePersistence},
	ReasonCodeDatabaseUnavailable:           {code: ReasonCodeDatabaseUnavailable, category: CategoryDatabasePersistence},
	ReasonCodeDatabaseTimeout:               {code: ReasonCodeDatabaseTimeout, category: CategoryDatabasePersistence},
	ReasonCodeRecordNotFound:                {code: ReasonCodeRecordNotFound, category: CategoryDatabasePersistence},
	ReasonCodeRecordAlreadyExists:           {code: ReasonCodeRecordAlreadyExists, category: CategoryDatabasePersistence},
	ReasonCodeUniqueConstraintViolation:     {code: ReasonCodeUniqueConstraintViolation, category: CategoryDatabasePersistence},
	ReasonCodeForeignKeyConstraintViolation: {code: ReasonCodeForeignKeyConstraintViolation, category: CategoryDatabasePersistence},
	ReasonCodeConstraintViolation:           {code: ReasonCodeConstraintViolation, category: CategoryDatabasePersistence},
	ReasonCodeTransactionFailed:             {code: ReasonCodeTransactionFailed, category: CategoryDatabasePersistence},
	ReasonCodeTransactionTimeout:            {code: ReasonCodeTransactionTimeout, category: CategoryDatabasePersistence},
	ReasonCodeTransactionRollback:           {code: ReasonCodeTransactionRollback, category: CategoryDatabasePersistence},

	// CACHE
	// Reason codes for cache related errors
	ReasonCodeCacheError:       {code: ReasonCodeCacheError, category: CategoryCache},
	ReasonCodeCacheUnavailable: {code: ReasonCodeCacheUnavailable, category: CategoryCache},
	ReasonCodeCacheTimeout:     {code: ReasonCodeCacheTimeout, category: CategoryCache},
	ReasonCodeCacheMiss:        {code: ReasonCodeCacheMiss, category: CategoryCache},

	// SECURITY
	// Reason codes for security and cryptography related errors
	ReasonCodeSecurityViolation: {code: ReasonCodeSecurityViolation, category: CategorySecurity},
	ReasonCodeSuspiciousRequest: {code: ReasonCodeSuspiciousRequest, category: CategorySecurity},
	ReasonCodeInvalidSignature:  {code: ReasonCodeInvalidSignature, category: CategorySecurity},
	ReasonCodeSignatureExpired:  {code: ReasonCodeSignatureExpired, category: CategorySecurity},
	ReasonCodeSignatureInvalid:  {code: ReasonCodeSignatureInvalid, category: CategorySecurity},
	ReasonCodeCSRFTokenInvalid:  {code: ReasonCodeCSRFTokenInvalid, category: CategorySecurity},
	ReasonCodeCSRFTokenMissing:  {code: ReasonCodeCSRFTokenMissing, category: CategorySecurity},
	ReasonCodeEncryptionFailed:  {code: ReasonCodeEncryptionFailed, category: CategorySecurity},
	ReasonCodeDecryptionFailed:  {code: ReasonCodeDecryptionFailed, category: CategorySecurity},
	ReasonCodeHashingFailed:     {code: ReasonCodeHashingFailed, category: CategorySecurity},
}
