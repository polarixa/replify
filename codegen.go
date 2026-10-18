package replify

import "github.com/polarixa/replify/pkg/codegen"

// WithMetaRequestIDGenerated generates a cryptographically secure random string
// using [codegen.Generator] and sets it as the Request ID in the response metadata.
//
// If no options are provided, it defaults to a 16-character alphanumeric string.
// This is highly recommended for ensuring every API response has a unique, traceable
// request identifier without manual boilerplate.
//
// Example:
//
//	w := replify.New().
//	    WithMetaRequestIDGenerated(
//	        codegen.WithLength(24),
//	        codegen.WithCharset(codegen.CharsetAlphanumericUpper),
//	        codegen.WithPrefix("req_"),
//	    )
//	fmt.Println(w.Meta().RequestID()) // "req_A3BF9KP2XQ17..."
func (w *wrapper) WithMetaRequestIDGenerated(opts ...codegen.Option) *wrapper {
	if !w.Available() {
		return w
	}

	g, err := codegen.New(opts...)
	if err != nil {
		return w.WithErrorAck(err).
			WithDebuggingKV("invalid_codegen_options", true)
	}

	id, err := g.Generate()
	if err != nil {
		return w.WithErrorAck(err).WithDebuggingKV("invalid_codegen_generated_random", true)
	}

	if w.meta == nil {
		w.meta = Meta()
	}
	w.meta.WithRequestID(id)
	return w
}
