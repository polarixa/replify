package replify

import (
	"net/http"
	"time"

	"github.com/polarixa/replify/pkg/strutil"
)

// NormAll performs a comprehensive normalization of the wrapper instance.
//
// It sequentially calls the following normalization methods:
//   - NormHSC
//   - NormPaging
//   - NormMeta
//   - NormBody
//   - NormMessage
//
// Returns:
//   - A pointer to the updated [wrapper] instance.
func (w *wrapper) NormAll() *wrapper {
	return w.NormHSC().
		NormMeta().
		NormBody().
		NormPaging().
		NormMessage().
		NormDebug()
}

// NormPaging normalizes the pagination information in the wrapper.
//
// If the pagination object is not already initialized, it creates a new one
// using the `NewPagination` function. It then calls the `Normalize` method
// on the pagination instance to ensure its values are consistent.
//
// Returns:
//   - A pointer to the updated [wrapper] instance.
func (w *wrapper) NormPaging() *wrapper {
	if w.IsPagingPresent() {
		w.pagination.Norm()
		w.RandDeltaValue() // Indicate that a change has occurred
	}
	return w
}

// NormHSC normalizes the relationship between the header and status code.
//
// If the status code is not present but the header is, it sets the status code
// from the header's code. If the header is not present but the status code is,
// it creates a new header with the status code and its corresponding text.
//
// If both the status code and header are present, it ensures the status code
// matches the header's code.
//
// Returns:
//   - A pointer to the updated [wrapper] instance.
func (w *wrapper) NormHSC() *wrapper {
	hasChanges := false
	switch {
	case !w.IsStatusCodePresent() && w.IsHeaderPresent():
		w.statusCode = w.header.Code()
		hasChanges = true
	case !w.IsHeaderPresent() && w.IsStatusCodePresent():
		w.header = Header().WithCode(w.statusCode).WithText(http.StatusText(w.statusCode))
		hasChanges = true
	case w.IsStatusCodePresent() && w.IsHeaderPresent():
		if w.statusCode == w.header.Code() {
			break
		}
		w.statusCode = w.header.Code()
		hasChanges = true
	}

	if hasChanges {
		w.RandDeltaValue() // Indicate that a change has occurred
	}
	return w
}

// NormMeta normalizes the metadata in the wrapper.
//
// If the meta object is not already initialized, it creates a new one
// using the [meta] function. It then ensures that essential fields such as
// locale, API version, request ID, and requested time are set to default
// values if they are not already present.
//
// Returns:
//   - A pointer to the updated [wrapper] instance.
func (w *wrapper) NormMeta() *wrapper {
	hasChanges := false

	if !w.IsMetaPresent() {
		w.meta = Meta()
		hasChanges = true
	}

	// Set default locale if not present
	if !w.meta.IsLocalePresent() {
		w.meta.locale = string(LocaleEnUS)
		hasChanges = true
	}

	// Set default API version if not present
	if !w.meta.IsApiVersionPresent() {
		w.meta.apiVersion = "v0.0.1"
		hasChanges = true
	}

	// Generate request ID if not present
	if !w.meta.IsRequestIDPresent() {
		w.meta.RandRequestID()
		hasChanges = true
	}

	// Set requested time if not present
	if !w.meta.IsRequestedTimePresent() {
		w.meta.requestedTime = time.Now()
		hasChanges = true
	}

	if hasChanges {
		w.RandDeltaValue() // Indicate that a change has occurred
	}
	return w
}

// NormBody normalizes the data/body field in the wrapper.
//
// This method ensures that the data field is properly handled:
//   - If data is nil and status code indicates success with content, logs a warning (optional)
//   - Validates that data type is consistent with the response type
//   - For list/array responses, ensures total count is synchronized
//
// Returns:
//   - A pointer to the updated [wrapper] instance.
func (w *wrapper) NormBody() *wrapper {
	hasChanges := false

	// Sync total count if data is a slice/array
	if w.data != nil {
		switch v := w.data.(type) {
		case []any:
			if w.total == 0 && len(v) > 0 {
				w.total = len(v)
				hasChanges = true
			}
		case []map[string]any:
			if w.total == 0 && len(v) > 0 {
				w.total = len(v)
				hasChanges = true
			}
		case []string:
			if w.total == 0 && len(v) > 0 {
				w.total = len(v)
				hasChanges = true
			}
		case []int:
			if w.total == 0 && len(v) > 0 {
				w.total = len(v)
				hasChanges = true
			}
		case []int8:
			if w.total == 0 && len(v) > 0 {
				w.total = len(v)
				hasChanges = true
			}
		case []int16:
			if w.total == 0 && len(v) > 0 {
				w.total = len(v)
				hasChanges = true
			}
		case []int32:
			if w.total == 0 && len(v) > 0 {
				w.total = len(v)
				hasChanges = true
			}
		case []int64:
			if w.total == 0 && len(v) > 0 {
				w.total = len(v)
				hasChanges = true
			}
		case []uint:
			if w.total == 0 && len(v) > 0 {
				w.total = len(v)
				hasChanges = true
			}
		case []uint8:
			if w.total == 0 && len(v) > 0 {
				w.total = len(v)
				hasChanges = true
			}
		case []uint16:
			if w.total == 0 && len(v) > 0 {
				w.total = len(v)
				hasChanges = true
			}
		case []uint32:
			if w.total == 0 && len(v) > 0 {
				w.total = len(v)
				hasChanges = true
			}
		case []uint64:
			if w.total == 0 && len(v) > 0 {
				w.total = len(v)
				hasChanges = true
			}
		case []float32:
			if w.total == 0 && len(v) > 0 {
				w.total = len(v)
				hasChanges = true
			}
		case []float64:
			if w.total == 0 && len(v) > 0 {
				w.total = len(v)
				hasChanges = true
			}
		}
	}

	// If we have pagination with total items but no total set on wrapper
	if w.IsPagingPresent() && w.pagination.TotalItems() > 0 && w.total == 0 {
		w.total = w.pagination.TotalItems()
		hasChanges = true
	}

	// Sync total items in pagination if wrapper total is set,
	// but pagination total items is zero
	if w.IsPagingPresent() && w.pagination.TotalItems() == 0 && w.total > 0 {
		w.pagination.WithTotalItems(w.total)
		hasChanges = true
	}

	if hasChanges {
		w.RandDeltaValue()
	}
	return w
}

// NormMessage normalizes the message field in the wrapper.
//
// If the message is empty and a status code is present, it sets a default
// message based on the status code category (success, redirection, client error,
// server error).
//
// Returns:
//   - A pointer to the updated [wrapper] instance.
func (w *wrapper) NormMessage() *wrapper {
	hasChanges := false

	// Set default message based on status code if message is empty
	if strutil.IsEmpty(w.message) && w.IsStatusCodePresent() {
		switch {
		case w.IsInformational(): // 1xx
			w.message = Continue.Type()
			hasChanges = true
		case w.IsSuccess(): // 2xx
			w.message = OK.Type()
			hasChanges = true
		case w.IsRedirection(): // 3xx
			w.message = MultipleChoices.Type()
			hasChanges = true
		case w.IsClientError(): // 4xx
			w.message = BadRequest.Type()
			hasChanges = true
		case w.IsServerError(): // 5xx
			w.message = InternalServerError.Type()
			hasChanges = true
		}
	}

	if hasChanges {
		w.RandDeltaValue() // Indicate that a change has occurred
	}
	return w
}

// NormDebug normalizes the debug information in the wrapper.
//
// This method removes any debug entries that have nil values to ensure
// the debug map only contains meaningful information.
//
// Returns:
//   - A pointer to the updated [wrapper] instance.
func (w *wrapper) NormDebug() *wrapper {
	if w.debug == nil {
		return w
	}
	hasChanges := false

	// Remove debug entries with nil values
	for key, value := range w.debug {
		if value == nil {
			delete(w.debug, key)
			hasChanges = true
		}
	}

	if hasChanges {
		w.RandDeltaValue() // Indicate that a change has occurred
	}
	return w
}
