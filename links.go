package replify

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
