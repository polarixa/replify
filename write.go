package replify

import (
	"net/http"
)

// WriteJSON writes the current response as JSON to w, sets the Content-Type header
// to "application/json; charset=utf-8", and writes the configured HTTP status
// code.
//
// WriteJSON returns the same wrapper so the Replify fluent chain can continue.
// If the underlying write fails, the error is stored in the wrapper via [WithErrorAck]
// and can be inspected with IsError() / Cause().
//
// Note: once WriteJSON is called the HTTP response is committed. Subsequent mutations
// to the wrapper do not affect the already-sent HTTP response.
func (r *wrapper) WriteJSON(w http.ResponseWriter) *wrapper {
	if !r.Available() {
		return r
	}
	if w == nil {
		return r.WithErrorAck(NewError("Write called with nil http.ResponseWriter"))
	}

	payload := r.JSONBytes()

	w.Header().Set(HeaderContentType.String(), MediaTypeApplicationJSONUTF8.String())
	w.WriteHeader(r.StatusCode())
	if _, err := w.Write(payload); err != nil {
		return r.WithErrorAck(err)
	}
	return r
}
