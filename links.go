package replify

import (
	"net/http"
	"strings"

	"github.com/polarixa/replify/pkg/slogger"
	"github.com/polarixa/replify/pkg/strchain"
	"github.com/polarixa/replify/pkg/strutil"
)

// L represents a wrapper around the [links] struct for HATEOAS support.
type L struct {
	*links
}

// link represents a single HATEOAS link following HAL specification.
type link struct {
	href        string // URI or URI template (required)
	method      string // HTTP method (optional, defaults to "GET")
	title       string // Human-readable identifier (optional)
	typez       string // Media type of resource representation (optional)
	templated   bool   // Whether href is a URI template (optional)
	name        string // Secondary key for disambiguation (optional)
	deprecation string // URL providing deprecation information (optional)
	profile     string // URI that hints about the profile (optional)
	hreflang    string // Language of the linked resource (optional)
}

// links represents a collection of HATEOAS links keyed by relation type.
type links struct {
	items map[string]*link
}

// Link creates a new HATEOAS link with the specified href and optional HTTP method.
//
// Parameters:
//   - href: The URI or URI template for the link.
//   - method: Optional HTTP method (defaults to "GET").
//
// Returns:
//   - A pointer to the newly created [link] instance.
func Link(href string, method ...string) *link {
	l := &link{
		href:   href,
		method: "GET",
	}
	if len(method) > 0 && strutil.IsNotEmpty(method[0]) {
		l.method = strings.ToUpper(method[0])
	}
	return l
}

// WithHref sets the href (URI or URI template) for the link.
//
// Parameters:
//   - href: The URI or URI template for the link.
//
// Returns:
//   - A pointer to the [link] instance with the updated href.
func (ls *link) WithHref(href string) *link {
	ls.href = href
	return ls
}

// WithMethod sets the HTTP method for the link.
//
// Parameters:
//   - method: The HTTP method for the link.
//
// Returns:
//   - A pointer to the [link] instance with the updated HTTP method.
func (ls *link) WithMethod(method string) *link {
	if strutil.IsNotEmpty(method) {
		ls.method = strings.ToUpper(method)
	}
	return ls
}

// WithTitle sets the human-readable identifier for the link.
//
// Parameters:
//   - title: The human-readable identifier for the link.
//
// Returns:
//   - A pointer to the [link] instance with the updated title.
func (ls *link) WithTitle(title string) *link {
	ls.title = title
	return ls
}

// WithType sets the media type of the resource representation for the link.
//
// Parameters:
//   - typez: The media type of the resource representation for the link.
//
// Returns:
//   - A pointer to the [link] instance with the updated media type.
func (ls *link) WithType(typez string) *link {
	ls.typez = typez
	return ls
}

// WithTemplated sets whether the href is a URI template for the link.
//
// Parameters:
//   - templated: A boolean indicating whether the href is a URI template.
//
// Returns:
//   - A pointer to the [link] instance with the updated templated flag.
func (ls *link) WithTemplated(templated bool) *link {
	ls.templated = templated
	return ls
}

// WithName sets the secondary key for disambiguation for the link.
//
// Parameters:
//   - name: The secondary key for disambiguation for the link.
//
// Returns:
//   - A pointer to the [link] instance with the updated name.
func (ls *link) WithName(name string) *link {
	ls.name = name
	return ls
}

// WithDeprecation sets the deprecation notice for the link.
//
// Parameters:
//   - deprecation: The deprecation notice for the link.
//
// Returns:
//   - A pointer to the [link] instance with the updated deprecation notice.
func (ls *link) WithDeprecation(deprecation string) *link {
	ls.deprecation = deprecation
	return ls
}

// WithProfile sets the profile link for the link.
//
// Parameters:
//   - profile: The profile link for the link.
//
// Returns:
//   - A pointer to the [link] instance with the updated profile link.
func (ls *link) WithProfile(profile string) *link {
	ls.profile = profile
	return ls
}

// WithHreflang sets the language of the resource representation for the link.
//
// Parameters:
//   - hreflang: The language of the resource representation for the link.
//
// Returns:
//   - A pointer to the [link] instance with the updated hreflang.
func (ls *link) WithHreflang(hreflang string) *link {
	ls.hreflang = hreflang
	return ls
}

// Available checks whether the link is available, i.e., it is not nil and has a non-empty href.
//
// Returns:
//   - A boolean indicating whether the link is available.
func (ls *link) Available() bool {
	return ls != nil && strutil.IsNotEmpty(ls.href)
}

// Href returns the href of the link if it is available.
//
// Returns:
//   - The href of the link, or an empty string if the link is not available.
func (ls *link) Href() string {
	if !ls.Available() {
		return ""
	}
	return ls.href
}

// Method returns the HTTP method of the link if it is available.
//
// Returns:
//   - The HTTP method of the link, or http.MethodGet if the link is not available.
func (ls *link) Method() string {
	if !ls.Available() {
		return http.MethodGet
	}
	return ls.method
}

// Title returns the title of the link if it is available.
//
// Returns:
//   - The title of the link, or an empty string if the link is not available.
func (ls *link) Title() string {
	if !ls.Available() {
		return ""
	}
	return ls.title
}

// Type returns the type of the link if it is available.
//
// Returns:
//   - The type of the link, or an empty string if the link is not available.
func (ls *link) Type() string {
	if !ls.Available() {
		return ""
	}
	return ls.typez
}

// IsTemplated checks whether the link is templated, i.e., it has a non-empty templated attribute.
//
// Returns:
//   - A boolean indicating whether the link is templated.
func (ls *link) IsTemplated() bool {
	if !ls.Available() {
		return false
	}
	return ls.templated
}

// Name returns the name of the link if it is available.
//
// Returns:
//   - The name of the link, or an empty string if the link is not available.
func (ls *link) Name() string {
	if !ls.Available() {
		return ""
	}
	return ls.name
}

// Deprecation returns the deprecation information of the link if it is available.
//
// Returns:
//   - The deprecation information of the link, or an empty string if the link is not available.
func (ls *link) Deprecation() string {
	if !ls.Available() {
		return ""
	}
	return ls.deprecation
}

// Profile returns the profile of the link if it is available.
//
// Returns:
//   - The profile of the link, or an empty string if the link is not available.
func (ls *link) Profile() string {
	if !ls.Available() {
		return ""
	}
	return ls.profile
}

// Hreflang returns the hreflang of the link if it is available.
//
// Returns:
//   - The hreflang of the link, or an empty string if the link is not available.
func (ls *link) Hreflang() string {
	if !ls.Available() {
		return ""
	}
	return ls.hreflang
}

// Respond returns a map representation of the link if it is available.
//
// Returns:
//   - A map containing the link's attributes, or an empty map if the link is not available.
func (ls *link) Respond() map[string]any {
	m := make(map[string]any)
	if ls == nil {
		return m
	}

	if strutil.IsNotEmpty(ls.href) {
		m["href"] = ls.href
	}
	if strutil.IsNotEmpty(ls.method) {
		m["method"] = ls.method
	}
	if strutil.IsNotEmpty(ls.title) {
		m["title"] = ls.title
	}
	if strutil.IsNotEmpty(ls.typez) {
		m["type"] = ls.typez
	}
	if ls.templated {
		m["templated"] = ls.templated
	}
	if strutil.IsNotEmpty(ls.name) {
		m["name"] = ls.name
	}
	if strutil.IsNotEmpty(ls.deprecation) {
		m["deprecation"] = ls.deprecation
	}
	if strutil.IsNotEmpty(ls.profile) {
		m["profile"] = ls.profile
	}
	if strutil.IsNotEmpty(ls.hreflang) {
		m["hreflang"] = ls.hreflang
	}
	return m
}

// JSON returns a JSON representation of the link if it is available.
//
// Returns:
//   - A JSON string containing the link's attributes, or an empty JSON object if the link is not available.
func (ls *link) JSON() string {
	return jsonpass(ls.Respond())
}

// JSONPretty returns a pretty-printed JSON representation of the link if it is available.
//
// Returns:
//   - A pretty-printed JSON string containing the link's attributes, or an empty JSON object if the link is not available.
func (ls *link) JSONPretty() string {
	return jsonpretty(ls.Respond())
}

// String returns a string representation of the link if it is available.
//
// Returns:
//   - A string containing the link's attributes, or an empty string if the link is not available.
func (ls *link) String() string {
	if ls == nil {
		return ""
	}
	sw := strchain.New()
	sw.AppendF("href=%s", ls.href)
	sw.Space()
	sw.AppendF("method=%s", ls.method)
	sw.Space()
	sw.AppendF("title=%s", ls.title)
	sw.Space()
	sw.AppendF("type=%s", ls.typez)
	sw.Space()
	sw.AppendF("name=%s", ls.name)
	sw.Space()
	sw.AppendF("deprecation=%s", ls.deprecation)
	sw.Space()
	sw.AppendF("profile=%s", ls.profile)
	sw.Space()
	sw.AppendF("hreflang=%s", ls.hreflang)
	return sw.String()
}

// Logging logs the [link]'s attributes using the provided logger or the default logger if none is provided.
//
// Parameters:
//   - logger: Optional. A pointer to a [slogger.Logger] instance to use for logging.
//
// Returns:
//   - The [link] itself, allowing for method chaining.
func (ls *link) Logging(logger ...*slogger.Logger) *link {
	if ls == nil {
		return ls
	}
	l := slogger.S()
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	}

	msg := "replify::link::logging"

	child := l.With()
	child.WithCaller(true).WithCallerSkip(3)

	logAtLevel(child, slogger.InfoLevel, msg, slogger.JSON("LINK", ls.Respond()))
	return ls
}

// Slogging logs the [link]'s attributes as a string using the provided logger or the default logger if none is provided.
//
// Parameters:
//   - logger: Optional. A pointer to a [slogger.Logger] instance to use for logging.
//
// Returns:
//   - The [link] itself, allowing for method chaining.
func (ls *link) Slogging(logger ...*slogger.Logger) *link {
	if ls == nil {
		return ls
	}
	l := slogger.S()
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	}

	child := l.With()
	child.WithCaller(true).WithCallerSkip(3)

	slogAtLevel(child, slogger.InfoLevel, ls.String())
	return ls
}

// Clone creates a deep copy of the [link] instance.
//
// Returns:
//   - A new [link] instance with the same attribute values as the original, or nil if the original link is nil.
func (l *link) Clone() *link {
	if l == nil {
		return nil
	}
	clone := &link{
		href:        l.href,
		method:      l.method,
		title:       l.title,
		typez:       l.typez,
		templated:   l.templated,
		name:        l.name,
		deprecation: l.deprecation,
		profile:     l.profile,
		hreflang:    l.hreflang,
	}
	return clone
}
