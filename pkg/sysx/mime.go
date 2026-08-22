package sysx

import (
	"mime"
	"strings"

	"github.com/polarixa/replify/pkg/strutil"
)

// MimeFromName returns a best-effort IANA media type derived from the
// extension of name. Multi-segment extensions like ".tar.gz" map to
// MimeGZIP; unknown extensions and names without an extension fall back
// to MimeOctetStream.
//
// Parameters:
//   - `name`: the filename whose extension drives the lookup.
//
// Returns:
//
// The matching IANA media type.
//
// Example:
//
//	sysx.MimeFromName("report.csv")  // "text/csv; charset=utf-8"
//	sysx.MimeFromName("dump.tar.gz") // "application/gzip"
//	sysx.MimeFromName("blob")        // "application/octet-stream"
func MimeFromName(name string) string {
	idx := strings.LastIndexByte(name, '.')
	if idx < 0 || idx == len(name)-1 {
		return MimeOctetStream
	}
	// Use the standard library's mime.TypeByExtension to get the MIME type for the extension.
	var ct = mime.TypeByExtension(name[idx:]) // This returns the MIME type with parameters, e.g., "text/html; charset=utf-8".
	if strutil.IsNotEmpty(ct) {
		return ct
	}

	switch strings.ToLower(name[idx+1:]) {
	case "txt", "log":
		return MimeText
	case "csv":
		return MimeCSV
	case "json":
		return MimeJSON
	case "xml":
		return MimeXML
	case "html", "htm":
		return MimeHTML
	case "pdf":
		return MimePDF
	case "zip":
		return MimeZIP
	case "gz", "gzip", "tgz":
		return MimeGZIP
	case "sql":
		return MimeSQL
	default:
		return MimeOctetStream
	}
}
