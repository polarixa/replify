package replify

import (
	"fmt"
	"time"
)

// Wrap creates a new [wrapper] instance and applies each provided [ROption]
// in order. It is the functional-options entry-point alternative to the
// fluent builder chain started by [New].
//
// Zero options are valid; Wrap() returns a freshly initialized wrapper
// identical to what [New] returns.
//
// Example:
//
//	w := replify.Wrap(
//	    replify.WithStatusCode(replify.OK),
//	    replify.WithMessage("users retrieved"),
//	    replify.WithBody(users),
//	    replify.WithPagination(replify.FromPages(120, 10).WithPage(1)),
//	    replify.WithMeta(
//	        replify.Meta().
//	            WithApiVersion("v1.0.0").
//	            WithLocale("en_US"),
//	    ),
//	)
func Wrap(opts ...ROption) *wrapper {
	w := New()
	for _, opt := range opts {
		if opt != nil {
			opt(w)
		}
	}
	return w
}

// WithStatusCode returns an [ROption] that sets the HTTP status code of the response.
// Use the redeclared Status* constants (e.g. [OK], [BadRequest]) rather than
func WithStatusCode(h *header) ROption {
	return func(w *wrapper) {
		w.WithHeader(h)
	}
}

// WithTotal returns an [ROption] that sets the total item count.
func WithTotal(total int) ROption {
	return func(w *wrapper) {
		w.WithTotal(total)
	}
}

// WithMessage returns an [ROption] that sets the response message.
func WithMessage(message string) ROption {
	return func(w *wrapper) {
		w.WithMessage(message)
	}
}

// WithMessagef returns an [ROption] that sets the response message using
// [fmt.Sprintf] formatting.
func WithMessagef(format string, args ...any) ROption {
	return func(w *wrapper) {
		w.WithMessage(fmt.Sprintf(format, args...))
	}
}

// WithBody returns an [ROption] that sets the response body payload.
func WithBody(v any) ROption {
	return func(w *wrapper) {
		w.WithBody(v)
	}
}

// WithSkipBody returns an [ROption] that controls whether the body payload
// is omitted from [String], [build], and [Slogging] output.
// Pass true to suppress the body (useful when the payload is too large or
// sensitive to include in logs); false restores default behavior.
func WithSkipBody(skip bool) ROption {
	return func(w *wrapper) {
		w.WithSkipBody(skip)
	}
}

// WithPath returns an [ROption] that sets the request path.
func WithPath(v string) ROption {
	return func(w *wrapper) {
		w.WithPath(v)
	}
}

// WithPathf returns an [ROption] that sets the request path using
// [fmt.Sprintf] formatting.
func WithPathf(format string, args ...any) ROption {
	return func(w *wrapper) {
		w.WithPath(fmt.Sprintf(format, args...))
	}
}

// WithMeta returns an [ROption] that sets the response metadata.
func WithMeta(v *meta) ROption {
	return func(w *wrapper) {
		w.WithMeta(v)
	}
}

// WithPagination returns an [ROption] that attaches pagination details.
func WithPagination(v *pagination) ROption {
	return func(w *wrapper) {
		w.WithPagination(v)
	}
}

// WithCursor returns an [ROption] that attaches a pagination cursor.
func WithCursor(v *cursor) ROption {
	return func(w *wrapper) {
		w.WithCursor(v)
	}
}

// WithDebugging returns an [ROption] that replaces the debug map wholesale.
func WithDebugging(v map[string]any) ROption {
	return func(w *wrapper) {
		w.WithDebugging(v)
	}
}

// WithDebuggingKV returns an [ROption] that adds a single key/value pair to
// the debug map.
func WithDebuggingKV(key string, value any) ROption {
	return func(w *wrapper) {
		w.WithDebuggingKV(key, value)
	}
}

// WithDebuggingKVf returns an [ROption] that adds a formatted string value
// under key to the debug map.
func WithDebuggingKVf(key string, format string, args ...any) ROption {
	return func(w *wrapper) {
		w.WithDebuggingKVf(key, format, args...)
	}
}

// WithError returns an [ROption] that attaches an error by message string.
func WithError(message string) ROption {
	return func(w *wrapper) {
		w.WithError(message)
	}
}

// WithErrorf returns an [ROption] that attaches a formatted error message.
func WithErrorf(format string, args ...any) ROption {
	return func(w *wrapper) {
		w.WithErrorf(format, args...)
	}
}

// WithErrorAck returns an [ROption] that wraps an existing error.
func WithErrorAck(err error) ROption {
	return func(w *wrapper) {
		w.WithErrorAck(err)
	}
}

// WithErrorAckf returns an [ROption] that wraps an existing error with
// additional formatted context.
func WithErrorAckf(err error, format string, args ...any) ROption {
	return func(w *wrapper) {
		w.WithErrorAckf(err, format, args...)
	}
}

// WithApiVersion returns an [ROption] that sets the API version on the
// embedded meta. A meta is created automatically if not yet present.
func WithApiVersion(v string) ROption {
	return func(w *wrapper) {
		w.WithApiVersion(v)
	}
}

// WithApiVersionf returns an [ROption] that sets the API version using
// [fmt.Sprintf] formatting.
func WithApiVersionf(format string, args ...any) ROption {
	return func(w *wrapper) {
		w.WithApiVersionf(format, args...)
	}
}

// WithRequestID returns an [ROption] that sets the request ID on the meta.
func WithRequestID(v string) ROption {
	return func(w *wrapper) {
		w.WithRequestID(v)
	}
}

// WithRequestIDf returns an [ROption] that sets the request ID using
// [fmt.Sprintf] formatting.
func WithRequestIDf(format string, args ...any) ROption {
	return func(w *wrapper) {
		w.WithRequestIDf(format, args...)
	}
}

// WithLocale returns an [ROption] that sets the locale on the meta.
func WithLocale(v string) ROption {
	return func(w *wrapper) {
		w.WithLocale(v)
	}
}

// WithLocaleValue returns an [ROption] that sets the locale on the embedded
// meta using a [Locale] typed value.
func WithLocaleValue(locale Locale) ROption {
	return func(w *wrapper) {
		if !w.IsMetaPresent() {
			w.meta = Meta()
		}
		w.meta.WithLocaleValue(locale)
	}
}

// WithRequestedTime returns an [ROption] that sets the requested time on the meta.
func WithRequestedTime(v time.Time) ROption {
	return func(w *wrapper) {
		w.WithRequestedTime(v)
	}
}

// WithCustomFields returns an [ROption] that replaces the custom fields map
// on the meta.
func WithCustomFields(values map[string]any) ROption {
	return func(w *wrapper) {
		w.WithCustomFields(values)
	}
}

// WithCustomFieldKV returns an [ROption] that adds a single custom field
// key/value pair to the meta.
func WithCustomFieldKV(key string, value any) ROption {
	return func(w *wrapper) {
		w.WithCustomFieldKV(key, value)
	}
}

// WithCustomFieldKVf returns an [ROption] that adds a formatted custom field
// value under key to the meta.
func WithCustomFieldKVf(key string, format string, args ...any) ROption {
	return func(w *wrapper) {
		w.WithCustomFieldKVf(key, format, args...)
	}
}

// WithPage returns an [ROption] that sets the current page number on the
// embedded pagination. A pagination object is created automatically if not
// yet present.
func WithPage(v int) ROption {
	return func(w *wrapper) {
		w.WithPage(v)
	}
}

// WithPerPage returns an [ROption] that sets the items-per-page count on
// the embedded pagination.
func WithPerPage(v int) ROption {
	return func(w *wrapper) {
		w.WithPerPage(v)
	}
}

// WithTotalPages returns an [ROption] that sets the total number of pages on
// the embedded pagination.
func WithTotalPages(v int) ROption {
	return func(w *wrapper) {
		w.WithTotalPages(v)
	}
}

// WithTotalItems returns an [ROption] that sets the total item count on the
// embedded pagination.
func WithTotalItems(v int) ROption {
	return func(w *wrapper) {
		w.WithTotalItems(v)
	}
}

// WithIsLast returns an [ROption] that marks whether the current page is the
// last page on the embedded pagination.
func WithIsLast(v bool) ROption {
	return func(w *wrapper) {
		w.WithIsLast(v)
	}
}

// WithAppendErrorAck returns an [ROption] that wraps err with an additional
// message and appends it to the wrapper's error chain.
func WithAppendErrorAck(err error, message string) ROption {
	return func(w *wrapper) {
		w.AppendErrorAck(err, message)
	}
}

// WithAppendError returns an [ROption] that wraps err with a plain contextual
// message and appends it to the wrapper's error chain.
func WithAppendError(err error, message string) ROption {
	return func(w *wrapper) {
		w.AppendError(err, message)
	}
}

// WithAppendErrorf returns an [ROption] that wraps err with a formatted
// contextual message and appends it to the wrapper's error chain.
func WithAppendErrorf(err error, format string, args ...any) ROption {
	return func(w *wrapper) {
		w.AppendErrorf(err, format, args...)
	}
}

// WithAppendErrors returns an [ROption] that folds a slice of errors into the
// wrapper's error chain, skipping nil entries.
func WithAppendErrors(errs ...error) ROption {
	return func(w *wrapper) {
		w.AppendErrors(errs)
	}
}

// WithStackTrace returns an [ROption] that captures the call stack at the
// point [Wrap] is called and stores the frames in the debug map under the
// key "stack_trace". Each frame is serialized using [Frame.MarshalText].
//
// This is the functional-option equivalent of [wrapper.WithStackTrace].
func WithStackTrace() ROption {
	return func(w *wrapper) {
		w.WithStackTrace()
	}
}

// WithInjectStackTrace returns an [ROption] that extracts the [StackTrace]
// from the wrapper's stored error and writes the formatted frames into the
// debug map under the key "error_stack_trace". It is a no-op when no error
// is set.
//
// This is the functional-option equivalent of [wrapper.InjectStackTrace].
func WithInjectStackTrace() ROption {
	return func(w *wrapper) {
		w.InjectStackTrace()
	}
}

// WithCursorNext returns an [ROption] that sets the next cursor value for pagination.
func WithCursorNext(next string) ROption {
	return func(w *wrapper) {
		w.WithCursorNext(next)
	}
}

// WithCursorPrevious returns an [ROption] that sets the previous cursor value for pagination.
func WithCursorPrevious(previous string) ROption {
	return func(w *wrapper) {
		w.WithCursorPrev(previous)
	}
}

// WithCursorHasNext returns an [ROption] that sets whether there is a next cursor available for pagination.
func WithCursorHasNext(hasNext bool) ROption {
	return func(w *wrapper) {
		w.WithCursorHasNext(hasNext)
	}
}

// WithCursorHasPrevious returns an [ROption] that sets whether there is a previous cursor available for pagination.
func WithCursorHasPrevious(hasPrevious bool) ROption {
	return func(w *wrapper) {
		w.WithCursorHasPrev(hasPrevious)
	}
}

// WithCursorNextF returns an [ROption] that sets the next cursor value for pagination using a formatted string.
func WithCursorNextF(format string, args ...any) ROption {
	return func(w *wrapper) {
		w.WithCursorNextF(format, args...)
	}
}

// WithCursorPreviousF returns an [ROption] that sets the previous cursor value for pagination using a formatted string.
func WithCursorPreviousF(format string, args ...any) ROption {
	return func(w *wrapper) {
		w.WithCursorPrevF(format, args...)
	}
}

// WithCursorLimit returns an [ROption] that sets the limit for the number of items in the cursor.
func WithCursorLimit(limit int) ROption {
	return func(w *wrapper) {
		w.WithCursorLimit(limit)
	}
}
