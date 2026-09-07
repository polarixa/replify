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

// Links creates and returns a new [links] instance with an initialized items map.
//
// Returns:
//   - A new [links] instance with an empty items map.
func Links() *links {
	return &links{
		items: make(map[string]*link),
	}
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

// WithLink adds a new link to the [links] instance with the specified relation, href, and optional method.
//
// Parameters:
//   - rel: The relation of the link.
//   - href: The href of the link.
//   - method: Optional. The HTTP method of the link.
//
// Returns:
//   - The [links] instance itself, allowing for method chaining.
func (ls *links) WithLink(rel string, href string, method ...string) *links {
	if ls == nil {
		return ls
	}
	if ls.items == nil {
		ls.items = make(map[string]*link)
	}
	ls.items[rel] = Link(href, method...)
	return ls
}

// WithLinkObject adds an existing [link] object to the [links] instance with the specified relation.
//
// Parameters:
//   - rel: The relation of the link.
//   - lnk: A pointer to the [link] object to add.
//
// Returns:
//   - The [links] instance itself, allowing for method chaining.
func (ls *links) WithLinkObject(rel string, lnk *link) *links {
	if ls == nil || lnk == nil {
		return ls
	}
	if ls.items == nil {
		ls.items = make(map[string]*link)
	}
	ls.items[rel] = lnk
	return ls
}

// WithSelf adds a "self" link to the [links] instance with the specified href.
//
// Parameters:
//   - href: The href of the "self" link.
//
// Returns:
//   - The [links] instance itself, allowing for method chaining.
func (ls *links) WithSelf(href string) *links {
	return ls.WithLink("self", href)
}

// WithNext adds a "next" link to the [links] instance with the specified href.
//
// Parameters:
//   - href: The href of the "next" link.
//
// Returns:
//   - The [links] instance itself, allowing for method chaining.
func (ls *links) WithNext(href string) *links {
	return ls.WithLink("next", href)
}

// WithPrev adds a "prev" link to the [links] instance with the specified href.
//
// Parameters:
//   - href: The href of the "prev" link.
//
// Returns:
//   - The [links] instance itself, allowing for method chaining.
func (ls *links) WithPrev(href string) *links {
	return ls.WithLink("prev", href)
}

// WithFirst adds a "first" link to the [links] instance with the specified href.
//
// Parameters:
//   - href: The href of the "first" link.
//
// Returns:
//   - The [links] instance itself, allowing for method chaining.
func (ls *links) WithFirst(href string) *links {
	return ls.WithLink("first", href)
}

// WithLast adds a "last" link to the [links] instance with the specified href.
//
// Parameters:
//   - href: The href of the "last" link.
//
// Returns:
//   - The [links] instance itself, allowing for method chaining.
func (ls *links) WithLast(href string) *links {
	return ls.WithLink("last", href)
}

// Remove deletes the link with the specified relation from the [links] instance.
//
// Parameters:
//   - rel: The relation of the link to remove.
//
// Returns:
//   - The [links] instance itself, allowing for method chaining.
func (ls *links) Remove(rel string) *links {
	if ls == nil || ls.items == nil {
		return ls
	}
	delete(ls.items, rel)
	return ls
}

// Link retrieves the link with the specified relation from the [links] instance.
//
// Parameters:
//   - rel: The relation of the link to retrieve.
//
// Returns:
//   - A pointer to the [link] object if it exists, or nil otherwise.
func (ls *links) Link(rel string) *link {
	if ls == nil || ls.items == nil {
		return nil
	}
	return ls.items[rel]
}

// Has checks if a link with the specified relation exists in the [links] instance.
//
// Parameters:
//   - rel: The relation of the link to check.
//
// Returns:
//   - true if the link exists, false otherwise.
func (ls *links) Has(rel string) bool {
	if ls == nil || ls.items == nil {
		return false
	}
	_, exists := ls.items[rel]
	return exists
}

// Count returns the number of links in the [links] instance.
//
// Returns:
//   - The number of links.
func (ls *links) Count() int {
	if ls == nil || ls.items == nil {
		return 0
	}
	return len(ls.items)
}

// Available checks if there are any links in the [links] instance.
//
// Returns:
//   - true if there is at least one link, false otherwise.
func (ls *links) Available() bool {
	return ls != nil && ls.items != nil && len(ls.items) > 0
}

// Items returns the underlying map of links in the [links] instance.
//
// Returns:
//   - A map where the keys are link relations and the values are pointers to the corresponding [link] objects.
func (ls *links) Items() map[string]*link {
	if ls == nil {
		return nil
	}
	return ls.items
}

// Respond generates a map representation of the [links] instance suitable for serialization.
//
// Returns:
//   - A map where the keys are link relations and the values are the corresponding serialized link objects.
func (ls *links) Respond() map[string]any {
	m := make(map[string]any)
	if ls == nil || ls.items == nil {
		return m
	}
	for rel, lnk := range ls.items {
		if lnk != nil && lnk.Available() {
			m[rel] = lnk.Respond()
		}
	}
	return m
}

// JSON generates a JSON string representation of the [links] instance suitable for serialization.
//
// Returns:
//   - A JSON string representing the [links] instance.
func (ls *links) JSON() string {
	return jsonpass(ls.Respond())
}

// JSONPretty generates a pretty-printed JSON string representation of the [links] instance suitable for serialization.
//
// Returns:
//   - A pretty-printed JSON string representing the [links] instance.
func (ls *links) JSONPretty() string {
	return jsonpretty(ls.Respond())
}

// String generates a string representation of the [links] instance.
//
// Returns:
//   - A string representing the [links] instance.
func (ls *links) String() string {
	sw := strchain.New()
	if ls == nil || ls.items == nil {
		return sw.String()
	}
	for rel, lnk := range ls.items {
		if lnk != nil && lnk.Available() {
			sw.AppendF("%s={%s}", rel, lnk.String()).Space()
		}
	}
	return sw.String()
}

// Logging logs the [links] instance using the provided logger or the default logger if none is provided.
//
// Parameters:
//   - logger: Optional. A pointer to a slogger.Logger instance to use for logging.
//
// Returns:
//   - The same [links] instance.
func (ls *links) Logging(logger ...*slogger.Logger) *links {
	if ls == nil {
		return ls
	}
	l := slogger.S()
	if len(logger) > 0 && logger[0] != nil {
		l = logger[0]
	}

	child := l.With()
	child.WithCaller(true).WithCallerSkip(3)
	logAtLevel(child, slogger.InfoLevel, "replify::links::logging", slogger.JSON("LINKS", ls.Respond()))
	return ls
}

// Slogging logs the string representation of the [links] instance using the provided logger or the default logger if none is provided.
//
// Parameters:
//   - logger: Optional. A pointer to a slogger.Logger instance to use for logging.
//
// Returns:
//   - The same [links] instance.
func (ls *links) Slogging(logger ...*slogger.Logger) *links {
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

// Reply creates a new [L] instance containing the current [links] instance.
//
// Returns:
//   - A new [L] instance containing the current [links] instance.
func (ls *links) Reply() L {
	return L{links: ls}
}

// ReplyPtr creates a new [L] instance containing the current [links] instance and returns a pointer to it.
//
// Returns:
//   - A pointer to a new [L] instance containing the current [links] instance.
func (ls *links) ReplyPtr() *L {
	return &L{links: ls}
}

// Clone creates a deep copy of the [links] instance.
//
// Returns:
//   - A new [links] instance that is a deep copy of the current instance.
func (ls *links) Clone() *links {
	if ls == nil {
		return nil
	}
	clone := Links()
	if ls.items != nil {
		for rel, lnk := range ls.items {
			if lnk != nil {
				clone.items[rel] = lnk.Clone()
			}
		}
	}
	return clone
}

// Equal compares the current [links] instance with another [links] instance for equality.
//
// Parameters:
//   - other: A pointer to the [links] instance to compare with.
//
// Returns:
//   - true if the current [links] instance is equal to the other instance, false otherwise.
func (ls *links) Equal(other *links) bool {
	if ls == other {
		return true
	}
	if ls == nil || other == nil {
		return false
	}
	if len(ls.items) != len(other.items) {
		return false
	}
	for rel, lnk := range ls.items {
		otherLink, exists := other.items[rel]
		if !exists {
			return false
		}
		if lnk == nil && otherLink != nil {
			return false
		}
		if lnk != nil && otherLink == nil {
			return false
		}
		if lnk != nil && otherLink != nil {
			if lnk.href != otherLink.href ||
				lnk.method != otherLink.method ||
				lnk.title != otherLink.title ||
				lnk.typez != otherLink.typez ||
				lnk.templated != otherLink.templated {
				return false
			}
		}
	}
	return true
}
