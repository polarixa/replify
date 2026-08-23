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

// binarySniffLen is the number of leading bytes inspected when performing
// content-based binary detection. This mirrors the sample size used by
// net/http.DetectContentType and common gitattributes-style heuristics.
const binarySniffLen = 512
