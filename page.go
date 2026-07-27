package replify

import (
	"github.com/polarixa/replify/pkg/slogger"
	"github.com/polarixa/replify/pkg/strchain"
)

// WithPage sets the page number for the [pagination] instance.
//
// This function updates the `page` field of the [pagination] and
// returns the modified [pagination] instance to allow method chaining.
//
// Parameters:
//   - `v`: An integer representing the page number to set.
//
// Returns:
//   - A pointer to the modified [pagination] instance (enabling method chaining).
func (p *pagination) WithPage(v int) *pagination {
	if v < 1 {
		v = 1
	}
	p.page = v
	p.calculate()
	return p
}

// WithPerPage sets the number of items per page for the [pagination] instance.
//
// This function updates the `perPage` field of the [pagination] and
// returns the modified [pagination] instance to allow method chaining.
// Validates that perPage is >= 1, defaults to 10 if invalid value.
//
// Parameters:
//   - `v`: An integer representing the number of items per page to set.
//
// Returns:
//   - A pointer to the modified [pagination] instance (enabling method chaining).
func (p *pagination) WithPerPage(v int) *pagination {
	if v < 1 {
		v = 10
	}
	p.perPage = v
	p.calculate()
	return p
}

// WithTotalPages sets the total number of pages for the [pagination] instance.
// Ensure that totalPages is >= 0, defaults to 0 if invalid value.
//
// This function updates the `totalPages` field of the [pagination] and
// returns the modified [pagination] instance to allow method chaining.
//
// Parameters:
//   - `v`: An integer representing the total number of pages to set.
//
// Returns:
//   - A pointer to the modified [pagination] instance (enabling method chaining).
func (p *pagination) WithTotalPages(v int) *pagination {
	if v < 0 {
		v = 0
	}
	p.totalPages = v
	return p
}

// WithTotalItems sets the total number of items for the [pagination] instance.
// Ensure that totalItems is >= 0, defaults to 0 if invalid value.
//
// This function updates the `totalItems` field of the [pagination] and
// returns the modified [pagination] instance to allow method chaining.
//
// Parameters:
//   - `v`: An integer representing the total number of items to set.
//
// Returns:
//   - A pointer to the modified [pagination] instance (enabling method chaining).
func (p *pagination) WithTotalItems(v int) *pagination {
	if v < 0 {
		v = 0
	}
	p.totalItems = v
	p.calculate()
	return p
}

// WithIsLast sets whether this is the last page in the [pagination] instance.
//
// This function updates the `isLast` field of the [pagination] and
// returns the modified [pagination] instance to allow method chaining.
//
// Parameters:
//   - `v`: A boolean value indicating whether this is the last page.
//
// Returns:
//   - A pointer to the modified [pagination] instance (enabling method chaining).
func (p *pagination) WithIsLast(v bool) *pagination {
	p.isLast = v
	return p
}

// Available checks whether the [pagination] instance is non-nil.
//
// This function ensures that the [pagination] object exists and is not nil.
// It serves as a safety check to avoid null pointer dereferences when accessing the instance's fields or methods.
//
// Returns:
//   - A boolean value indicating whether the [pagination] instance is non-nil:
//   - `true` if the [pagination] instance is non-nil.
//   - `false` if the [pagination] instance is nil.
func (p *pagination) Available() bool {
	return p != nil
}

// Page retrieves the current page number from the [pagination] instance.
//
// This function checks if the [pagination] instance is available (non-nil) before
// returning the value of the `page` field. If the instance is not available, it
// returns a default value of `0`.
//
// Returns:
//   - An integer representing the current page number.
//   - `0` if the [pagination] instance is not available.
func (p *pagination) Page() int {
	if !p.Available() {
		return 0
	}
	return p.page
}

// PerPage retrieves the number of items per page from the [pagination] instance.
//
// This function checks if the [pagination] instance is available (non-nil) before
// returning the value of the `perPage` field. If the instance is not available, it
// returns a default value of `0`.
//
// Returns:
//   - An integer representing the number of items per page.
//   - `0` if the [pagination] instance is not available.
func (p *pagination) PerPage() int {
	if !p.Available() {
		return 0
	}
	return p.perPage
}

// TotalPages retrieves the total number of pages from the [pagination] instance.
//
// This function checks if the [pagination] instance is available (non-nil) before
// returning the value of the `totalPages` field. If the instance is not available, it
// returns a default value of `0`.
//
// Returns:
//   - An integer representing the total number of pages.
//   - `0` if the [pagination] instance is not available.
func (p *pagination) TotalPages() int {
	if !p.Available() {
		return 0
	}
	return p.totalPages
}

// TotalItems retrieves the total number of items from the [pagination] instance.
//
// This function checks if the [pagination] instance is available (non-nil) before
// returning the value of the `totalItems` field. If the instance is not available, it
// returns a default value of `0`.
//
// Returns:
//   - An integer representing the total number of items.
//   - `0` if the [pagination] instance is not available.
func (p *pagination) TotalItems() int {
	if !p.Available() {
		return 0
	}
	return p.totalItems
}

// IsLast checks whether the current pagination represents the last page.
//
// This function checks the `isLast` field of the [pagination] instance to determine if the current page is the last one.
// The `isLast` field is typically set to `true` when there are no more pages of data available.
//
// Returns:
//   - A boolean value indicating whether the current page is the last:
//   - `true` if `isLast` is `true`, indicating this is the last page of results.
//   - `false` if `isLast` is `false`, indicating more pages are available.
func (p *pagination) IsLast() bool {
	if !p.Available() {
		return true
	}
	return p.isLast
}

// Norm adjusts the pagination fields to ensure consistency.
//
// This method recalculates the `totalPages` based on `totalItems` and `perPage`,
// ensuring that the pagination state is coherent. It also adjusts the `page` field
// to ensure it does not exceed `totalPages`, and sets the `isLast` field appropriately
// based on the current page position.
func (p *pagination) Norm() *pagination {
	// Calculate total pages only if perPage is valid to avoid division by zero.
	if p.perPage > 0 {
		// Calculate total pages (ceiling division)
		p.totalPages = (p.totalItems + p.perPage - 1) / p.perPage
	} else {
		// If perPage is zero or negative, we cannot calculate totalPages
		p.totalPages = 0
	}

	// If there are no items, set totalPages to 0 and page to 1.
	if p.totalItems == 0 {
		p.totalPages = 0
		p.page = 1
		p.isLast = true
		return p
	}

	// Adjust the current page if it exceeds totalPages.
	if p.page > p.totalPages && p.totalPages > 0 {
		p.page = p.totalPages
	}

	// Determine if we are on the last page.
	p.isLast = p.page >= p.totalPages
	return p
}

// Respond generates a map representation of the [pagination] instance.
//
// This method collects various fields related to pagination (e.g., `page`, `per_page`, etc.)
// and organizes them into a key-value map. It ensures that only valid pagination details
// are included in the response.
//
// The following fields are included in the pagination response:
//   - `page`: The current page number.
//   - `per_page`: The number of items per page.
//   - `total_pages`: The total number of pages available.
//   - `total_items`: The total number of items available across all pages.
//   - `is_last`: A boolean indicating if this is the last page.
//
// Returns:
//   - A `map[string]interface{}` containing the structured pagination data.
func (p *pagination) Respond() map[string]any {
	m := make(map[string]any)
	if !p.Available() {
		return m
	}
	m["page"] = p.page
	m["per_page"] = p.perPage
	m["total_pages"] = p.totalPages
	m["total_items"] = p.totalItems
	m["is_last"] = p.isLast
	return m
}

// JSON serializes the [pagination] instance into a compact JSON string.
//
// This function uses the `encoding.JSON` utility to generate a JSON representation
// of the [pagination] instance. The output is a compact JSON string with no additional
// whitespace or formatting, providing a minimalistic view of the pagination data.
//
// Returns:
//   - A compact JSON string representation of the [pagination] instance.
func (p *pagination) JSON() string {
	return jsonpass(p.Respond())
}

// JSONPretty serializes the [pagination] instance into a prettified JSON string.
//
// This function uses the `encoding.JSONPretty` utility to generate a JSON representation
// of the [pagination] instance. The output is a human-readable JSON string with
// proper indentation and formatting for better readability, which is helpful for
// inspecting pagination data during development or debugging.
//
// Returns:
//   - A prettified JSON string representation of the [pagination] instance.
func (p *pagination) JSONPretty() string {
	return jsonpretty(p.Respond())
}

// Equal compares the current [pagination] instance with another [pagination] instance for equality.
//
// This method checks if both [pagination] instances are non-nil and then compares their
// fields (`page`, `perPage`, `totalPages`, `totalItems`, and `isLast`) for equality.
// It returns true if all corresponding fields are equal, indicating that the two
// pagination instances represent the same pagination state.
//
// Parameters:
//   - `other`: A pointer to another [pagination] instance to compare against.
//
// Returns:
//   - A boolean value indicating whether the two [pagination] instances are equal.
func (p *pagination) Equal(other *pagination) bool {
	if p == nil && other == nil {
		return true
	}
	if p == nil || other == nil {
		return false
	}
	return p.page == other.page &&
		p.perPage == other.perPage &&
		p.totalPages == other.totalPages &&
		p.totalItems == other.totalItems &&
		p.isLast == other.isLast
}

// String returns a string representation of the [pagination] instance.
//
// This method constructs a string that summarizes the key fields of the [pagination] instance,
// such as `page`, `per_page`, `total_pages`, `total_items`, and `is_last`. The resulting string
// provides a concise overview of the pagination state, which can be useful for logging or debugging purposes.
//
// Returns:
//   - A string representation of the [pagination] instance, summarizing its key fields.
func (p *pagination) String() string {
	sw := strchain.New()
	if p == nil {
		return sw.String()
	}
	sw.AppendF("page=%d", p.page).Space()
	sw.AppendF("per_page=%d", p.perPage).Space()
	sw.AppendF("total_pages=%d", p.totalPages).Space()
	sw.AppendF("total_items=%d", p.totalItems).Space()
	sw.AppendF("is_last=%t", p.isLast)
	return sw.String()
}

// Logging dispatches a structured log entry for this response using [slogger] with the log message set to a default or custom message.
// The log level is automatically selected based on the HTTP status code range:
//
//   - 1xx → Debug  (informational)
//   - 2xx → Info   (success)
//   - 3xx → Warn   (redirection)
//   - 4xx → Error  (client error)
//   - 5xx → Error  (server error; [slogger.Logger.Fatal] is intentionally avoided because it calls os.Exit(1))
//   - other → Trace (no status code set)
//
// The log field key is "PAGINATION" and its value is the structured map returned
// by [pagination.Respond], serialized as JSON by the active formatter.
//
// # Thread-safety
//
// Logging is safe for concurrent use. The supplied logger is never mutated: a
// goroutine-local child is derived via [slogger.Logger.With] on every call so
// that the caller-skip adjustment and caller-enable flag stay local to the
// current goroutine. Concurrent callers sharing the same *[slogger.Logger]
// will not race. The pagination fields (e.g., page, pageSize) are read exactly once per call to give a consistent snapshot; pagination fields are expected to be immutable after construction via [Pagination] / [With*] options.
//
// # Caller reporting
//
// Caller information is always enabled for this call. callerSkip is set to 3
// to skip Logging, the slogger trampoline (Trace/Debug/Info/Warn/Error), and the caller of Logging, so the reported file and line resolve to the actual call site of Logging.
//
// Parameters:
//   - `logger`: optional *[slogger.Logger] to use. When omitted or nil, the
//     package-level global logger ([slogger.GlobalLogger]) is used.
//
// Returns:
//
// the receiver *pagination unchanged, enabling method chaining.
//
// Example:
//
//	replify.Pagination().
//	    WithPage(2).
//	    WithPageSize(20).
//	    Logging()
func (p *pagination) Logging(logger ...*slogger.Logger) *pagination {
	if p == nil {
		return p
	}
	l := slogger.S()
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	}

	msg := "replify::pagination::logging"

	child := l.With()
	child.WithCaller(true).WithCallerSkip(3)

	logAtLevel(child, slogger.InfoLevel, msg, slogger.JSON("PAGINATION", p.Respond()))
	return p
}

// Slogging dispatches a structured log entry for this response using [slogger] with the log message set to the pagination's string representation.
// The log level is automatically selected based on the HTTP status code range:
//
//   - 1xx → Debug  (informational)
//   - 2xx → Info   (success)
//   - 3xx → Warn   (redirection)
//   - 4xx → Error  (client error)
//   - 5xx → Error  (server error; [slogger.Logger.Fatal] is intentionally avoided because it calls os.Exit(1))
//   - other → Trace (no status code set)
//
// The log field key is "PAGINATION" and its value is the structured map returned
// by [pagination.Respond], serialized as Text by the active formatter.
// # Thread-safety
//
// Slogging is safe for concurrent use. The supplied logger is never mutated: a
// goroutine-local child is derived via [slogger.Logger.With] on every call so
// that the caller-skip adjustment and caller-enable flag stay local to the
// current goroutine. Concurrent callers sharing the same *[slogger.Logger]
// will not race. The pagination fields (e.g., page, pageSize) are read exactly once per call to give a consistent snapshot; pagination fields are expected to be immutable after construction via [Pagination] / [With*] options.
//
// # Caller reporting
//
// Caller information is always enabled for this call. callerSkip is set to 3
// to skip Slogging, the slogger trampoline (Trace/Debug/Info/Warn/Error), and the caller of Slogging, so the reported file and line resolve to the actual call site of Slogging.
//
// Parameters:
//   - `logger`: optional *[slogger.Logger] to use. When omitted or nil, the
//     package-level global logger ([slogger.GlobalLogger]) is used.
//
// Returns:
//
//	the receiver *pagination unchanged, enabling method chaining.
//
// Example:
//
//	replify.Pagination().
//	    WithPage(2).
//	    WithPageSize(20).
//	    Slogging()
func (p *pagination) Slogging(logger ...*slogger.Logger) *pagination {
	if p == nil {
		return p
	}
	l := slogger.S()
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	}

	child := l.With()
	child.WithCaller(true).WithCallerSkip(3)

	slogAtLevel(child, slogger.InfoLevel, p.String())
	return p
}

// Reply returns a [P] instance that wraps the current [pagination] instance.
//
// This method is a convenience function that allows the current [pagination] instance
// to be wrapped in a [P] type, which may provide additional functionality or context
// for handling pagination in responses. It enables seamless integration of pagination
// data into response structures.
//
// Returns:
//   - A [P] instance containing the current [pagination] instance.
func (p *pagination) Reply() P {
	return P{pagination: p}
}

// ReplyPtr returns a pointer to a [P] instance that wraps the current [pagination] instance.
//
// This method is a convenience function that allows the current [pagination] instance
// to be wrapped in a pointer to a [P] type, which may provide additional functionality
// or context for handling pagination in responses. It enables seamless integration of
// pagination data into response structures while allowing for pointer semantics.
//
// Returns:
//   - A pointer to a [P] instance containing the current [pagination] instance.
func (p *pagination) ReplyPtr() *P {
	return &P{pagination: p}
}

// calculate computes the total pages and determines if the current page is the last one.
//
// This method performs calculations based on the `totalItems` and `perPage` fields
// to derive the `totalPages`. It uses ceiling division to ensure that any remaining
// items that don't fill a complete page are still counted as an additional page.
// Additionally, it checks if the current `page` is the last page by comparing it
// to `totalPages`, setting the `isLast` field accordingly.
func (p *pagination) calculate() {
	// Ensure page is at least 1.
	if p.page <= 0 {
		p.page = 1
	}

	// Only calculate if perPage is valid to avoid division by zero.
	if p.totalItems > 0 && p.perPage > 0 {
		// Calculate total pages (ceiling division)
		p.totalPages = (p.totalItems + p.perPage - 1) / p.perPage
	}

	// If there are no items, there is 1 page (page 1) which is empty,
	// or 0 pages. Usually, for API consistency, we treat 0 items as being on the last page.
	if p.totalItems == 0 {
		p.totalPages = 0
		p.isLast = true
		return
	}

	// Determine if we are on the last page.
	// We are on the last page if the current page is greater than or equal to totalPages.
	if p.totalPages > 0 {
		p.isLast = p.page >= p.totalPages
	}
}
