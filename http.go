package replify

// Continue sets the HTTP status code of the [wrapper] to 100 Continue.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the initial part of a request has been received and has not yet been rejected by the server.
// It internally calls the `WithHeader` method with the `Continue` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) Continue() *wrapper {
	return w.WithHeader(Continue)
}

// SwitchingProtocols sets the HTTP status code of the [wrapper] to 101 Switching Protocols.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server understands and is willing to comply with the client's request to switch protocols.
// It internally calls the `WithHeader` method with the `SwitchingProtocols` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) SwitchingProtocols() *wrapper {
	return w.WithHeader(SwitchingProtocols)
}

// Processing sets the HTTP status code of the [wrapper] to 102 Processing.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server has received and is processing the request, but no response is available yet.
// It internally calls the `WithHeader` method with the `Processing` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) Processing() *wrapper {
	return w.WithHeader(Processing)
}

// OK sets the HTTP status code of the [wrapper] to 200 OK.
//
// This method is a convenience function that allows for quick setting of the status code to indicate a successful response.
// It internally calls the `WithHeader` method with the `OK` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) OK() *wrapper {
	return w.WithHeader(OK)
}

// Created sets the HTTP status code of the [wrapper] to 201 Created.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the request has been fulfilled and has resulted in one or more new resources being created.
// It internally calls the `WithHeader` method with the `Created` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) Created() *wrapper {
	return w.WithHeader(Created)
}

// Accepted sets the HTTP status code of the [wrapper] to 202 Accepted.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the request has been accepted for processing, but the processing has not been completed.
// It internally calls the `WithHeader` method with the `Accepted` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) Accepted() *wrapper {
	return w.WithHeader(Accepted)
}

// NonAuthoritativeInformation sets the HTTP status code of the [wrapper] to 203 Non-Authoritative Information.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the request was successful but the enclosed payload has been modified from that of the origin server's 200 OK response by a transforming proxy.
// It internally calls the `WithHeader` method with the `NonAuthoritativeInformation` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) NonAuthoritativeInformation() *wrapper {
	return w.WithHeader(NonAuthoritativeInformation)
}

// NoContent sets the HTTP status code of the [wrapper] to 204 No Content.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server has successfully fulfilled the request and that there is no additional content to send in the response payload body.
// It internally calls the `WithHeader` method with the `NoContent` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) NoContent() *wrapper {
	return w.WithHeader(NoContent)
}

// ResetContent sets the HTTP status code of the [wrapper] to 205 Reset Content.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server has fulfilled the request and desires that the user agent reset the "document view" which caused the request to be sent.
// It internally calls the `WithHeader` method with the `ResetContent` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) ResetContent() *wrapper {
	return w.WithHeader(ResetContent)
}

// PartialContent sets the HTTP status code of the [wrapper] to 206 Partial Content.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server is delivering only part of the resource due to a range header sent by the client.
// It internally calls the `WithHeader` method with the `PartialContent` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) PartialContent() *wrapper {
	return w.WithHeader(PartialContent)
}

// MultiStatus sets the HTTP status code of the [wrapper] to 207 Multi-Status.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the message body that follows is an XML message and can contain a number of separate response codes, depending on how many sub-requests were made.
// It internally calls the `WithHeader` method with the `MultiStatus` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) MultiStatus() *wrapper {
	return w.WithHeader(MultiStatus)
}

// AlreadyReported sets the HTTP status code of the [wrapper] to 208 Already Reported.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the members of a DAV binding have already been enumerated in a previous reply to this request, and are not being included again.
// It internally calls the `WithHeader` method with the `AlreadyReported` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) AlreadyReported() *wrapper {
	return w.WithHeader(AlreadyReported)
}

// IMUsed sets the HTTP status code of the [wrapper] to 226 IM Used.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server has fulfilled a GET request for the resource, and the response is a representation of the result of one or more instance-manipulations applied to the current instance.
// It internally calls the `WithHeader` method with the `IMUsed` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) IMUsed() *wrapper {
	return w.WithHeader(IMUsed)
}

// MultipleChoices sets the HTTP status code of the [wrapper] to 300 Multiple Choices.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the request has more than one possible response.
// It internally calls the `WithHeader` method with the `MultipleChoices` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) MultipleChoices() *wrapper {
	return w.WithHeader(MultipleChoices)
}

// MovedPermanently sets the HTTP status code of the [wrapper] to 301 Moved Permanently.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the requested resource has been assigned a new permanent URI and any future references to this resource should use one of the returned URIs.
// It internally calls the `WithHeader` method with the `MovedPermanently` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) MovedPermanently() *wrapper {
	return w.WithHeader(MovedPermanently)
}

// Found sets the HTTP status code of the [wrapper] to 302 Found.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the requested resource resides temporarily under a different URI.
// It internally calls the `WithHeader` method with the `Found` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) Found() *wrapper {
	return w.WithHeader(Found)
}

// SeeOther sets the HTTP status code of the [wrapper] to 303 See Other.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the response to the request can be found under a different URI and should be retrieved using a GET method on that resource.
// It internally calls the `WithHeader` method with the `SeeOther` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) SeeOther() *wrapper {
	return w.WithHeader(SeeOther)
}

// NotModified sets the HTTP status code of the [wrapper] to 304 Not Modified.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the resource has not been modified since the version specified by the request headers.
// It internally calls the `WithHeader` method with the `NotModified` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) NotModified() *wrapper {
	return w.WithHeader(NotModified)
}

// UseProxy sets the HTTP status code of the [wrapper] to 305 Use Proxy.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the requested resource must be accessed through the proxy given by the Location field.
// It internally calls the `WithHeader` method with the `UseProxy` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) UseProxy() *wrapper {
	return w.WithHeader(UseProxy)
}

// Reserved sets the HTTP status code of the [wrapper] to 306 Reserved.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the requested resource is no longer available at the server and no forwarding address is known.
// It internally calls the `WithHeader` method with the `Reserved` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) Reserved() *wrapper {
	return w.WithHeader(Reserved)
}

// TemporaryRedirect sets the HTTP status code of the [wrapper] to 307 Temporary Redirect.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the requested resource resides temporarily under a different URI and the user agent must not change the request method if it performs an automatic redirection to that URI.
// It internally calls the `WithHeader` method with the `TemporaryRedirect` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) TemporaryRedirect() *wrapper {
	return w.WithHeader(TemporaryRedirect)
}

// PermanentRedirect sets the HTTP status code of the [wrapper] to 308 Permanent Redirect.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the requested resource has been assigned a new permanent URI and any future references to this resource should use one of the returned URIs.
// It internally calls the `WithHeader` method with the `PermanentRedirect` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) PermanentRedirect() *wrapper {
	return w.WithHeader(PermanentRedirect)
}

// BadRequest sets the HTTP status code of the [wrapper] to 400 Bad Request.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server cannot or will not process the request due to something that is perceived to be a client error (e.g., malformed request syntax, invalid request message framing, or deceptive request routing).
// It internally calls the `WithHeader` method with the `BadRequest` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) BadRequest() *wrapper {
	return w.WithHeader(BadRequest)
}

// Unauthorized sets the HTTP status code of the [wrapper] to 401 Unauthorized.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the request has not been applied because it lacks valid authentication credentials for the target resource.
// It internally calls the `WithHeader` method with the `Unauthorized` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) Unauthorized() *wrapper {
	return w.WithHeader(Unauthorized)
}

// PaymentRequired sets the HTTP status code of the [wrapper] to 402 Payment Required.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the request cannot be fulfilled until the client makes a payment.
// It internally calls the `WithHeader` method with the `PaymentRequired` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) PaymentRequired() *wrapper {
	return w.WithHeader(PaymentRequired)
}

// Forbidden sets the HTTP status code of the [wrapper] to 403 Forbidden.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server understood the request but refuses to authorize it.
// It internally calls the `WithHeader` method with the `Forbidden` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) Forbidden() *wrapper {
	return w.WithHeader(Forbidden)
}

// NotFound sets the HTTP status code of the [wrapper] to 404 Not Found.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server can't find the requested resource.
// It internally calls the `WithHeader` method with the `NotFound` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) NotFound() *wrapper {
	return w.WithHeader(NotFound)
}

// MethodNotAllowed sets the HTTP status code of the [wrapper] to 405 Method Not Allowed.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the request method is known by the server but is not supported by the target resource.
// It internally calls the `WithHeader` method with the `MethodNotAllowed` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) MethodNotAllowed() *wrapper {
	return w.WithHeader(MethodNotAllowed)
}

// NotAcceptable sets the HTTP status code of the [wrapper] to 406 Not Acceptable.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server cannot produce a response matching the list of acceptable values defined in the request's proactive content negotiation headers.
// It internally calls the `WithHeader` method with the `NotAcceptable` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) NotAcceptable() *wrapper {
	return w.WithHeader(NotAcceptable)
}

// ProxyAuthenticationRequired sets the HTTP status code of the [wrapper] to 407 Proxy Authentication Required.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the client must first authenticate itself with the proxy.
// It internally calls the `WithHeader` method with the `ProxyAuthenticationRequired` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) ProxyAuthenticationRequired() *wrapper {
	return w.WithHeader(ProxyAuthenticationRequired)
}

// RequestTimeout sets the HTTP status code of the [wrapper] to 408 Request Timeout.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server timed out waiting for the request.
// It internally calls the `WithHeader` method with the `RequestTimeout` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) RequestTimeout() *wrapper {
	return w.WithHeader(RequestTimeout)
}

// Conflict sets the HTTP status code of the [wrapper] to 409 Conflict.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the request could not be completed due to a conflict with the current state of the target resource.
// It internally calls the `WithHeader` method with the `Conflict` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) Conflict() *wrapper {
	return w.WithHeader(Conflict)
}

// Gone sets the HTTP status code of the [wrapper] to 410 Gone.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the target resource is no longer available at the server and no forwarding address is known.
// It internally calls the `WithHeader` method with the `Gone` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) Gone() *wrapper {
	return w.WithHeader(Gone)
}

// LengthRequired sets the HTTP status code of the [wrapper] to 411 Length Required.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server refuses to accept the request without a defined Content-Length.
// It internally calls the `WithHeader` method with the `LengthRequired` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) LengthRequired() *wrapper {
	return w.WithHeader(LengthRequired)
}

// PreconditionFailed sets the HTTP status code of the [wrapper] to 412 Precondition Failed.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that one or more conditions given in the request header fields evaluated to false when tested on the server.
// It internally calls the `WithHeader` method with the `PreconditionFailed` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) PreconditionFailed() *wrapper {
	return w.WithHeader(PreconditionFailed)
}

// RequestEntityTooLarge sets the HTTP status code of the [wrapper] to 413 Request Entity Too Large.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server is refusing to process a request because the request payload is larger than the server is willing or able to process.
// It internally calls the `WithHeader` method with the `RequestEntityTooLarge` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) RequestEntityTooLarge() *wrapper {
	return w.WithHeader(RequestEntityTooLarge)
}

// RequestURITooLong sets the HTTP status code of the [wrapper] to 414 Request-URI Too Long.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server is refusing to service the request because the request-target is longer than the server is willing to interpret.
// It internally calls the `WithHeader` method with the `RequestURITooLong` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) RequestURITooLong() *wrapper {
	return w.WithHeader(RequestURITooLong)
}

// UnsupportedMediaType sets the HTTP status code of the [wrapper] to 415 Unsupported Media Type.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server is refusing to service the request because the payload is in a format not supported by the target resource for this method.
// It internally calls the `WithHeader` method with the `UnsupportedMediaType` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) UnsupportedMediaType() *wrapper {
	return w.WithHeader(UnsupportedMediaType)
}

// RequestedRangeNotSatisfiable sets the HTTP status code of the [wrapper] to 416 Requested Range Not Satisfiable.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server cannot serve the requested range.
// It internally calls the `WithHeader` method with the `RequestedRangeNotSatisfiable` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) RequestedRangeNotSatisfiable() *wrapper {
	return w.WithHeader(RequestedRangeNotSatisfiable)
}

// ExpectationFailed sets the HTTP status code of the [wrapper] to 417 Expectation Failed.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the expectation given in the request's Expect header field could not be met by at least one of the inbound servers.
// It internally calls the `WithHeader` method with the `ExpectationFailed` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) ExpectationFailed() *wrapper {
	return w.WithHeader(ExpectationFailed)
}

// I'mATeapot sets the HTTP status code of the [wrapper] to 418 I'm a teapot.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server refuses to brew coffee because it is, permanently, a teapot.
// It internally calls the `WithHeader` method with the `ImATeapot` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) ImATeapot() *wrapper {
	return w.WithHeader(ImATeapot)
}

// EnhanceYourCalm sets the HTTP status code of the [wrapper] to 420 Enhance Your Calm.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server is overwhelmed and the client should slow down its request rate.
// It internally calls the `WithHeader` method with the `EnhanceYourCalm` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) EnhanceYourCalm() *wrapper {
	return w.WithHeader(EnhanceYourCalm)
}

// UnprocessableEntity sets the HTTP status code of the [wrapper] to 422 Unprocessable Entity.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server understands the content type of the request entity, and the syntax of the request entity is correct, but it was unable to process the contained instructions.
// It internally calls the `WithHeader` method with the `UnprocessableEntity` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) UnprocessableEntity() *wrapper {
	return w.WithHeader(UnprocessableEntity)
}

// Locked sets the HTTP status code of the [wrapper] to 423 Locked.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the resource that is being accessed is locked.
// It internally calls the `WithHeader` method with the `Locked` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) Locked() *wrapper {
	return w.WithHeader(Locked)
}

// FailedDependency sets the HTTP status code of the [wrapper] to 424 Failed Dependency.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the method could not be performed on the resource because the requested action depended on another action and that action failed.
// It internally calls the `WithHeader` method with the `FailedDependency` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) FailedDependency() *wrapper {
	return w.WithHeader(FailedDependency)
}

// UnorderedCollection sets the HTTP status code of the [wrapper] to 425 Unordered Collection.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server has received the request, but the order of the collection is not as expected.
// It internally calls the `WithHeader` method with the `UnorderedCollection` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) UnorderedCollection() *wrapper {
	return w.WithHeader(UnorderedCollection)
}

// UpgradeRequired sets the HTTP status code of the [wrapper] to 426 Upgrade Required.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the client should switch to a different protocol.
// It internally calls the `WithHeader` method with the `UpgradeRequired` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) UpgradeRequired() *wrapper {
	return w.WithHeader(UpgradeRequired)
}

// PreconditionRequired sets the HTTP status code of the [wrapper] to 428 Precondition Required.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the origin server requires the request to be conditional.
// It internally calls the `WithHeader` method with the `PreconditionRequired` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) PreconditionRequired() *wrapper {
	return w.WithHeader(PreconditionRequired)
}

// TooManyRequests sets the HTTP status code of the [wrapper] to 429 Too Many Requests.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the user has sent too many requests in a given amount of time.
// It internally calls the `WithHeader` method with the `TooManyRequests` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) TooManyRequests() *wrapper {
	return w.WithHeader(TooManyRequests)
}

// RequestHeaderFieldsTooLarge sets the HTTP status code of the [wrapper] to 431 Request Header Fields Too Large.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server is unwilling to process the request because its header fields are too large.
// It internally calls the `WithHeader` method with the `RequestHeaderFieldsTooLarge` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) RequestHeaderFieldsTooLarge() *wrapper {
	return w.WithHeader(RequestHeaderFieldsTooLarge)
}

// NoResponse sets the HTTP status code of the [wrapper] to 444 No Response.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server has returned no response.
// It internally calls the `WithHeader` method with the `NoResponse` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) NoResponse() *wrapper {
	return w.WithHeader(NoResponse)
}

// RetryWith sets the HTTP status code of the [wrapper] to 449 Retry With.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the request should be retried after performing the appropriate action.
// It internally calls the `WithHeader` method with the `RetryWith` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) RetryWith() *wrapper {
	return w.WithHeader(RetryWith)
}

// BlockedByWindowsParentalControls sets the HTTP status code of the [wrapper] to 450 Blocked by Windows Parental Controls.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the request was blocked by Windows Parental Controls.
// It internally calls the `WithHeader` method with the `BlockedByWindowsParentalControls` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) BlockedByWindowsParentalControls() *wrapper {
	return w.WithHeader(BlockedByWindowsParentalControls)
}

// UnavailableForLegalReasons sets the HTTP status code of the [wrapper] to 451 Unavailable For Legal Reasons.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server is unavailable for legal reasons.
// It internally calls the `WithHeader` method with the `UnavailableForLegalReasons` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) UnavailableForLegalReasons() *wrapper {
	return w.WithHeader(UnavailableForLegalReasons)
}

// ClientClosedRequest sets the HTTP status code of the [wrapper] to 499 Client Closed Request.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the client closed the request before the server could send a response.
// It internally calls the `WithHeader` method with the `ClientClosedRequest` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) ClientClosedRequest() *wrapper {
	return w.WithHeader(ClientClosedRequest)
}

// InternalServerError sets the HTTP status code of the [wrapper] to 500 Internal Server Error.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server encountered an unexpected condition that prevented it from fulfilling the request.
// It internally calls the `WithHeader` method with the `InternalServerError` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) InternalServerError() *wrapper {
	return w.WithHeader(InternalServerError)
}

// NotImplemented sets the HTTP status code of the [wrapper] to 501 Not Implemented.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server does not support the functionality required to fulfill the request.
// It internally calls the `WithHeader` method with the `NotImplemented` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) NotImplemented() *wrapper {
	return w.WithHeader(NotImplemented)
}

// BadGateway sets the HTTP status code of the [wrapper] to 502 Bad Gateway.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server, while acting as a gateway or proxy, received an invalid response from the upstream server it accessed in attempting to fulfill the request.
// It internally calls the `WithHeader` method with the `BadGateway` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) BadGateway() *wrapper {
	return w.WithHeader(BadGateway)
}

// ServiceUnavailable sets the HTTP status code of the [wrapper] to 503 Service Unavailable.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server is currently unable to handle the request due to temporary overloading or maintenance of the server.
// It internally calls the `WithHeader` method with the `ServiceUnavailable` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) ServiceUnavailable() *wrapper {
	return w.WithHeader(ServiceUnavailable)
}

// GatewayTimeout sets the HTTP status code of the [wrapper] to 504 Gateway Timeout.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server, while acting as a gateway or proxy, did not receive a timely response from the upstream server specified by the URI (e.g., HTTP, FTP, LDAP) or some other auxiliary server (e.g., DNS) it needed to access in attempting to complete the request.
// It internally calls the `WithHeader` method with the `GatewayTimeout` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) GatewayTimeout() *wrapper {
	return w.WithHeader(GatewayTimeout)
}

// HTTPVersionNotSupported sets the HTTP status code of the [wrapper] to 505 HTTP Version Not Supported.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server does not support, or refuses to support, the HTTP protocol version that was used in the request message.
// It internally calls the `WithHeader` method with the `HTTPVersionNotSupported` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) HTTPVersionNotSupported() *wrapper {
	return w.WithHeader(HTTPVersionNotSupported)
}

// VariantAlsoNegotiates sets the HTTP status code of the [wrapper] to 506 Variant Also Negotiates.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server has an internal configuration error: the chosen variant resource is configured to engage in transparent content negotiation itself, and is therefore not a proper end point in the negotiation process.
// It internally calls the `WithHeader` method with the `VariantAlsoNegotiates` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) VariantAlsoNegotiates() *wrapper {
	return w.WithHeader(VariantAlsoNegotiates)
}

// InsufficientStorage sets the HTTP status code of the [wrapper] to 507 Insufficient Storage.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server is unable to store the representation needed to complete the request.
// It internally calls the `WithHeader` method with the `InsufficientStorage` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) InsufficientStorage() *wrapper {
	return w.WithHeader(InsufficientStorage)
}

// LoopDetected sets the HTTP status code of the [wrapper] to 508 Loop Detected.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server detected an infinite loop while processing a request with "Depth: infinity".
// It internally calls the `WithHeader` method with the `LoopDetected` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) LoopDetected() *wrapper {
	return w.WithHeader(LoopDetected)
}

// BandwidthLimitExceeded sets the HTTP status code of the [wrapper] to 509 Bandwidth Limit Exceeded.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the server has exceeded the bandwidth specified by the server administrator; this is often used by shared hosting providers to limit the bandwidth of customers.
// It internally calls the `WithHeader` method with the `BandwidthLimitExceeded` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) BandwidthLimitExceeded() *wrapper {
	return w.WithHeader(BandwidthLimitExceeded)
}

// NotExtended sets the HTTP status code of the [wrapper] to 510 Not Extended.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the policy for accessing the resource has not been met in the request.
// It internally calls the `WithHeader` method with the `NotExtended` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) NotExtended() *wrapper {
	return w.WithHeader(NotExtended)
}

// NetworkAuthenticationRequired sets the HTTP status code of the [wrapper] to 511 Network Authentication Required.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the client needs to authenticate to gain network access.
// It internally calls the `WithHeader` method with the `NetworkAuthenticationRequired` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) NetworkAuthenticationRequired() *wrapper {
	return w.WithHeader(NetworkAuthenticationRequired)
}

// NetworkReadTimeoutError sets the HTTP status code of the [wrapper] to 598 Network Read Timeout Error.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the network read operation timed out.
// It internally calls the `WithHeader` method with the `NetworkReadTimeoutError` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) NetworkReadTimeoutError() *wrapper {
	return w.WithHeader(NetworkReadTimeoutError)
}

// NetworkConnectTimeoutError sets the HTTP status code of the [wrapper] to 599 Network Connect Timeout Error.
//
// This method is a convenience function that allows for quick setting of the status code to indicate that the network connection operation timed out.
// It internally calls the `WithHeader` method with the `NetworkConnectTimeoutError` status code.
//
// Returns:
//   - A pointer to the modified [wrapper] instance (enabling method chaining).
func (w *wrapper) NetworkConnectTimeoutError() *wrapper {
	return w.WithHeader(NetworkConnectTimeoutError)
}
