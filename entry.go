package replify

import (
	"net/http"
	"time"

	"github.com/polarixa/replify/pkg/conv"
	"github.com/polarixa/replify/pkg/encoding"
	"github.com/polarixa/replify/pkg/fj"
	"github.com/polarixa/replify/pkg/strutil"
)

// UnwrapJSON parses a raw JSON string and maps it into a [wrapper] struct.
//
// The input is first normalised (comments stripped, whitespace compacted) and
// validated before unmarshaling. The following top-level JSON keys are
// recognised and mapped to the corresponding wrapper field:
//
//	JSON key       wrapper field   Notes
//	──────────────────────────────────────────────────────────────────────────
//	"status_code"  statusCode      float64 → int
//	"total"        total           float64 → int
//	"message"      message         string
//	"path"         path            string
//	"data"         data            string/[]byte → json.RawMessage when valid
//	                                JSON; any other type stored as-is
//	"debug"        debug           map[string]any
//	"header"       header          object → *header (code, text, type,
//	                                description)
//	"meta"         meta            object → *meta (api_version, locale,
//	                                request_id, requested_time,
//	                                custom_fields)
//	"pagination"   pagination      object → *pagination (page, per_page,
//	                                total_pages, total_items, is_last)
//
// Unknown top-level keys are silently ignored. Missing keys leave the
// corresponding field at its zero value—no error is returned.
//
// Parameters:
//   - `jsonStr`: the raw JSON string to parse; may contain JS-style comments
//     or trailing commas, which are stripped during normalisation.
//
// Returns:
//
// a non-nil *wrapper and a nil error on success.
// Returns nil, err when jsonStr is empty, fails normalisation, or is not
// valid JSON after normalisation.
//
// Example:
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
//	        "total_items": 42, "total_pages": 21,
//	        "is_last": false
//	    },
//	    "meta": {
//	        "request_id":     "req_abc123",
//	        "api_version":    "v1.0.0",
//	        "locale":         "en_US",
//	        "requested_time": "2026-03-09T07:00:00Z"
//	    }
//	}`
//
//	w, err := replify.UnwrapJSON(jsonStr)
//	if err != nil {
//	    log.Fatalf("parse error: %v", err)
//	}
//	fmt.Println(w.JSONBodyParser().Get("0").Get("username").String()) // "alice"
//	fmt.Println(w.StatusCode())                                       // 200
//	fmt.Println(w.Pagination().TotalItems())                          // 42
func UnwrapJSON(jsonStr string) (w *wrapper, err error) {
	if strutil.IsEmpty(jsonStr) {
		return nil, NewError("JSON string is required")
	}
	specJSON := encoding.Spec([]byte(jsonStr))
	nJSON, err := encoding.NormalizeJSON(string(specJSON))
	if err != nil {
		return nil, err
	}
	if !encoding.IsValidJSONString(nJSON) || !fj.IsValidJSONString(nJSON) {
		return nil, NewErrorf("invalid JSON string: %s", jsonStr) // keep original JSON string
	}

	var data map[string]any
	err = encoding.UnmarshalJSONString(nJSON, &data)
	if err != nil {
		return nil, NewErrorAck(err)
	}
	if len(data) == 0 {
		return nil, NewErrorf("an unexpected error occurred while unmarshaling JSON to map, json: %s", nJSON)
	}
	w = &wrapper{}
	if value, exists := data["status_code"].(float64); exists {
		w.statusCode = int(value)
	}
	if value, exists := data["total"].(float64); exists {
		w.total = int(value)
	}
	if value, exists := data["message"].(string); exists {
		w.message = value
	}
	if value, exists := data["path"].(string); exists {
		w.path = value
	}
	if value, exists := data["debug"].(map[string]any); exists {
		w.debug = value
	}
	if values, exists := data["meta"].(map[string]any); exists {
		meta := &meta{}
		if value, exists := values["api_version"].(string); exists {
			meta.apiVersion = value
		}
		if value, exists := values["locale"].(string); exists {
			meta.locale = value
		}
		if value, exists := values["request_id"].(string); exists {
			meta.requestID = value
		}
		if customFields, exists := values["custom_fields"].(map[string]any); exists {
			meta.customFields = customFields
		}
		if value, exists := values["requested_time"].(string); exists {
			if t, err := time.Parse(time.RFC3339, value); err == nil {
				meta.requestedTime = t
			} else {
				// fallback using converter
				meta.requestedTime = conv.TimeOrDefault(value, time.Time{})
			}
		}
		w.meta = meta
	}
	if values, exists := data["header"].(map[string]any); exists {
		header := &header{}
		if value, exists := values["code"].(float64); exists {
			header.code = int(value)
		}
		if value, exists := values["text"].(string); exists {
			header.text = value
		}
		if value, exists := values["type"].(string); exists {
			header.typez = value
		}
		if value, exists := values["description"].(string); exists {
			header.description = value
		}
		w.header = header
	}
	if values, exists := data["pagination"].(map[string]any); exists {
		pagination := &pagination{}
		if value, exists := values["page"].(float64); exists {
			pagination.page = int(value)
		}
		if value, exists := values["per_page"].(float64); exists {
			pagination.perPage = int(value)
		}
		if value, exists := values["total_pages"].(float64); exists {
			pagination.totalPages = int(value)
		}
		if value, exists := values["total_items"].(float64); exists {
			pagination.totalItems = int(value)
		}
		if value, exists := values["is_last"].(bool); exists {
			pagination.isLast = value
		}
		w.pagination = pagination
	}
	if values, exists := data["cursor"].(map[string]any); exists {
		cursor := &cursor{}
		if value, exists := values["next"].(string); exists {
			cursor.next = value
		}
		if value, exists := values["previous"].(string); exists {
			cursor.previous = value
		}
		if value, exists := values["has_next"].(bool); exists {
			cursor.hasNext = value
		}
		if value, exists := values["has_previous"].(bool); exists {
			cursor.hasPrevious = value
		}
		if value, exists := values["limit"].(float64); exists {
			cursor.limit = int(value)
		}
		w.cursor = cursor
	}
	// if the data is a string, check if it is a valid JSON string and convert it to a json.RawMessage
	// otherwise, keep it as a string.
	// if the data is a []byte, check if it is a valid JSON byte slice and convert it to a json.RawMessage
	// otherwise, keep it as a []byte.
	// if the data is a json.RawMessage, keep it as a json.RawMessage
	// otherwise, keep it as a json.RawMessage.
	if value, exists := data["data"]; exists {
		w.data = safeBody(value)
	}
	return w, nil
}

// WrapFrom converts a map containing API response data into a [wrapper] struct
// by serializing the map into JSON format and then parsing it.
//
// The function is a helper that bridges between raw map data (e.g., deserialized JSON
// or other dynamic input) and the strongly-typed [wrapper] struct used in the codebase.
// It first converts the input map into a JSON string using `encoding.Json`, then calls
// the `Parse` function to handle the deserialization and field mapping to the [wrapper].
//
// Parameters:
//   - data: A map[string]interface{} containing the API response data.
//     The map should include keys like "status_code", "message", "meta", etc.,
//     that conform to the expected structure of a [wrapper].
//
// Returns:
//   - A pointer to a [wrapper] struct populated with data from the map.
//   - An error if the map is empty or if the JSON serialization/parsing fails.
//
// Error Handling:
//   - If the input map is empty or nil, the function returns an error
//     indicating that the data is invalid.
//   - If serialization or parsing fails, the error from `Parse` or `encoding.Json`
//     is propagated, providing context about the failure.
//
// Usage:
// This function is particularly useful when working with raw data maps (e.g., from
// dynamic inputs or unmarshaled data) that need to be converted into the [wrapper]
// struct for further processing.
//
// Example:
//
//	rawData := map[string]interface{}{
//	    "status_code": 200,
//	    "message": "Success",
//	    "data": "response body",
//	}
//	wrapper, err := replify.WrapFrom(rawData)
//	if err != nil {
//	    log.Println("Error extracting wrapper:", err)
//	} else {
//	    log.Println("Wrapper:", wrapper)
//	}
func WrapFrom(data map[string]any) (w *wrapper, err error) {
	if len(data) == 0 {
		return nil, NewError("data map must not be nil or empty")
	}
	json := jsonpass(data)
	return UnwrapJSON(json)
}

// WrapOk creates a wrapper for a successful HTTP response (200 OK).
//
// This function sets the HTTP status code to 200 (OK) and includes a message and data payload
// in the response body.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapOk(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusOK).
		WithHeader(OK).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapCreated creates a wrapper for a resource creation response (201 WrapCreated).
//
// This function sets the HTTP status code to 201 (WrapCreated) and includes a message and data payload
// in the response body.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapCreated(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusCreated).
		WithHeader(Created).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapBadRequest creates a wrapper for a client error response (400 Bad Request).
//
// This function sets the HTTP status code to 400 (Bad Request) and includes a message and data payload
// in the response body.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapBadRequest(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusBadRequest).
		WithHeader(BadRequest).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapNotFound creates a wrapper for a resource not found response (404 Not Found).
//
// This function sets the HTTP status code to 404 (Not Found) and includes a message and data payload
// in the response body.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapNotFound(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusNotFound).
		WithHeader(NotFound).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapNotImplemented creates a wrapper for a response indicating unimplemented functionality (501 Not Implemented).
//
// This function sets the HTTP status code to 501 (Not Implemented) and includes a message and data payload
// in the response body.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapNotImplemented(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusNotImplemented).
		WithHeader(NotImplemented).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapTooManyRequest creates a wrapper for a rate-limiting response (429 Too Many Requests).
//
// This function sets the HTTP status code to 429 (Too Many Requests) and includes a message and data payload
// in the response body.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapTooManyRequest(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusTooManyRequests).
		WithHeader(TooManyRequests).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapLocked creates a wrapper for a locked resource response (423 WrapLocked).
//
// This function sets the HTTP status code to 423 (WrapLocked) and includes a message and data payload
// in the response body.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapLocked(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusLocked).
		WithHeader(Locked).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapNoContent creates a wrapper for a successful response without a body (204 No Content).
//
// This function sets the HTTP status code to 204 (No Content) and includes a message and data payload
// in the response body.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapNoContent(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusNoContent).
		WithHeader(NoContent).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapProcessing creates a wrapper for a response indicating ongoing processing (102 WrapProcessing).
//
// This function sets the HTTP status code to 102 (WrapProcessing) and includes a message and data payload
// in the response body.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapProcessing(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusProcessing).
		WithHeader(Processing).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapUpgradeRequired creates a wrapper for a response indicating an upgrade is required (426 Upgrade Required).
//
// This function sets the HTTP status code to 426 (Upgrade Required) and includes a message and data payload
// in the response body. It is typically used when the client must switch to a different protocol.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapUpgradeRequired(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusUpgradeRequired).
		WithHeader(UpgradeRequired).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapServiceUnavailable creates a wrapper for a response indicating the service is temporarily unavailable (503 Service Unavailable).
//
// This function sets the HTTP status code to 503 (Service Unavailable) and includes a message and data payload
// in the response body. It is typically used when the server is unable to handle the request due to temporary overload
// or maintenance.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapServiceUnavailable(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusServiceUnavailable).
		WithHeader(ServiceUnavailable).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapInternalServerError creates a wrapper for a server error response (500 Internal Server Error).
//
// This function sets the HTTP status code to 500 (Internal Server Error) and includes a message and data payload
// in the response body. It is typically used when the server encounters an unexpected condition that prevents it
// from fulfilling the request.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapInternalServerError(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusInternalServerError).
		WithHeader(InternalServerError).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapGatewayTimeout creates a wrapper for a response indicating a gateway timeout (504 Gateway Timeout).
//
// This function sets the HTTP status code to 504 (Gateway Timeout) and includes a message and data payload
// in the response body. It is typically used when the server did not receive a timely response from an upstream server.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapGatewayTimeout(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusGatewayTimeout).
		WithHeader(GatewayTimeout).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapMethodNotAllowed creates a wrapper for a response indicating the HTTP method is not allowed (405 Method Not Allowed).
//
// This function sets the HTTP status code to 405 (Method Not Allowed) and includes a message and data payload
// in the response body. It is typically used when the server knows the method is not supported for the target resource.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapMethodNotAllowed(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusMethodNotAllowed).
		WithHeader(MethodNotAllowed).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapUnauthorized creates a wrapper for a response indicating authentication is required (401 WrapUnauthorized).
//
// This function sets the HTTP status code to 401 (WrapUnauthorized) and includes a message and data payload
// in the response body. It is typically used when the request has not been applied because it lacks valid
// authentication credentials.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapUnauthorized(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusUnauthorized).
		WithHeader(Unauthorized).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapForbidden creates a wrapper for a response indicating access to the resource is forbidden (403 WrapForbidden).
//
// This function sets the HTTP status code to 403 (WrapForbidden) and includes a message and data payload
// in the response body. It is typically used when the server understands the request but refuses to authorize it.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapForbidden(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusForbidden).
		WithHeader(Forbidden).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapAccepted creates a wrapper for a response indicating the request has been accepted for processing (202 WrapAccepted).
//
// This function sets the HTTP status code to 202 (WrapAccepted) and includes a message and data payload
// in the response body. It is typically used when the request has been received but processing is not yet complete.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapAccepted(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusAccepted).
		WithHeader(Accepted).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapRequestTimeout creates a wrapper for a response indicating the client request has timed out (408 Request Timeout).
//
// This function sets the HTTP status code to 408 (Request Timeout) and includes a message and data payload
// in the response body. It is typically used when the server did not receive a complete request message within the time it was prepared to wait.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapRequestTimeout(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusRequestTimeout).
		WithHeader(RequestTimeout).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapRequestEntityTooLarge creates a wrapper for a response indicating the request entity is too large (413 Payload Too Large).
//
// This function sets the HTTP status code to 413 (Payload Too Large) and includes a message and data payload
// in the response body. It is typically used when the server refuses to process a request because the request entity is larger than the server is willing or able to process.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapRequestEntityTooLarge(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusRequestEntityTooLarge).
		WithHeader(RequestEntityTooLarge).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapUnsupportedMediaType creates a wrapper for a response indicating the media type is not supported (415 Unsupported Media Type).
//
// This function sets the HTTP status code to 415 (Unsupported Media Type) and includes a message and data payload
// in the response body. It is typically used when the server refuses to accept the request because the payload is in an unsupported format.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapUnsupportedMediaType(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusUnsupportedMediaType).
		WithHeader(UnsupportedMediaType).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapHTTPVersionNotSupported creates a wrapper for a response indicating the HTTP version is not supported (505 HTTP Version Not Supported).
//
// This function sets the HTTP status code to 505 (HTTP Version Not Supported) and includes a message and data payload
// in the response body. It is typically used when the server does not support the HTTP protocol version used in the request.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapHTTPVersionNotSupported(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusHTTPVersionNotSupported).
		WithHeader(HTTPVersionNotSupported).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapPaymentRequired creates a wrapper for a response indicating payment is required (402 Payment Required).
//
// This function sets the HTTP status code to 402 (Payment Required) and includes a message and data payload
// in the response body. It is typically used when access to the requested resource requires payment.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapPaymentRequired(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusPaymentRequired).
		WithHeader(PaymentRequired).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapConflict creates a wrapper for a response indicating a conflict (409 Conflict).
//
// This function sets the HTTP status code to 409 (Conflict) and includes a message and data payload
// in the response body. It is typically used when the request conflicts with the current state of the server.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapConflict(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusConflict).
		WithHeader(Conflict).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapGone creates a wrapper for a response indicating the resource is gone (410 Gone).
//
// This function sets the HTTP status code to 410 (Gone) and includes a message and data payload
// in the response body. It is typically used when the requested resource is no longer available.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapGone(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusGone).
		WithHeader(Gone).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapUnprocessableEntity creates a wrapper for a response indicating the request was well-formed but was unable to be followed due to semantic errors (422 Unprocessable Entity).
//
// This function sets the HTTP status code to 422 (Unprocessable Entity) and includes a message and data payload
// in the response body. It is typically used when the server understands the content type of the request entity,
// and the syntax of the request entity is correct, but it was unable to process the contained instructions.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapUnprocessableEntity(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusUnprocessableEntity).
		WithHeader(UnprocessableEntity).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapPreconditionFailed creates a wrapper for a response indicating the precondition failed (412 Precondition Failed).
//
// This function sets the HTTP status code to 412 (Precondition Failed) and includes a message and data payload
// in the response body. It is typically used when the request has not been applied because one or more conditions were not met.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapPreconditionFailed(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusPreconditionFailed).
		WithHeader(PreconditionFailed).
		WithMessage(message).
		WithBody(data)
	return w
}

// WrapBadGateway creates a wrapper for a response indicating a bad gateway (502 Bad Gateway).
//
// This function sets the HTTP status code to 502 (Bad Gateway) and includes a message and data payload
// in the response body. It is typically used when the server, while acting as a gateway or proxy,
// received an invalid response from an upstream server.
//
// Parameters:
//   - message: A string containing the response message.
//   - data: The data payload to include in the response.
//
// Returns:
//   - A pointer to a [wrapper] instance representing the response.
func WrapBadGateway(message string, data any) *wrapper {
	w := New().
		WithStatusCode(http.StatusBadGateway).
		WithHeader(BadGateway).
		WithMessage(message).
		WithBody(data)
	return w
}
