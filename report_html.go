package replify

import (
	"bytes"
	_ "embed"
	"html/template"
	"regexp"
	"strings"
	"time"
)

//go:embed templates/report.html
var htmlTemplateContent string

var htmlTemplate *template.Template

func init() {
	htmlTemplate = template.Must(template.New("report.html").Parse(htmlTemplateContent))
}

type htmlDumpData struct {
	Title        string
	IsEST        bool
	Payload      string
	ErrorMessage string
	Frames       []estFrame
	GeneratedAt  string
}

type estFrame struct {
	DisplayName string
	Location    string
	Annotation  string
	IsRuntime   bool
	IsRoot      bool
	IsEntry     bool
}

type estData struct {
	ErrorMessage string
	Frames       []estFrame
}

func parseESTMarkdown(md string) estData {
	var res estData

	// Extract error message from ```go ... ```
	reCode := regexp.MustCompile("(?s)`{3}go\n(.*?)\n`{3}")
	if match := reCode.FindStringSubmatch(md); len(match) > 1 {
		res.ErrorMessage = strings.TrimSpace(match[1])
	}

	// Extract frames
	// Format: 1. [ ] `displayName` — `location` **(...)** OR *(...)*
	reFrame := regexp.MustCompile(`(?m)^\d+\.\s\[\s\]\s` + "`" + `([^` + "`" + `]+)` + "`" + `(?:\s—\s` + "`" + `([^` + "`" + `]*)` + "`" + `)?\s*(?:\*\*(.*?)\*\*|\*(.*?)\*)?`)
	matches := reFrame.FindAllStringSubmatch(md, -1)
	for _, m := range matches {
		f := estFrame{
			DisplayName: strings.TrimSpace(m[1]),
		}
		if len(m) > 2 && m[2] != "" {
			f.Location = strings.TrimSpace(m[2])
		}

		if len(m) > 3 && m[3] != "" {
			f.Annotation = m[3]
			if strings.Contains(f.Annotation, "root cause") {
				f.IsRoot = true
			}
			if strings.Contains(f.Annotation, "entry point") {
				f.IsEntry = true
			}
		} else if len(m) > 4 && m[4] != "" {
			f.Annotation = m[4]
			if strings.Contains(f.Annotation, "runtime") {
				f.IsRuntime = true
			}
		}
		res.Frames = append(res.Frames, f)
	}

	return res
}

// DumpHTML serializes the full [wrapper] response as an interactive HTML document and
// writes it into a self-cleaning temporary file. The returned [Dump] owns the file.
//
// The generated HTML includes an interactive JSON tree with search, expand/collapse
// controls, and syntax highlighting. It requires no external network dependencies.
//
// Both return values are always non-nil:
//   - (*Dump, *wrapper) — Dump holds the file; wrapper carries the outcome.
//   - On error: Dump is nil, wrapper has InternalServerError + error detail.
//
// Parameters:
//   - ignoringJSONfields: top-level JSON fields to ignore in the output.
func (w *wrapper) DumpHTML(ignoringJSONfields ...string) (*Dump, *wrapper) {
	if !w.Available() {
		return nil, New().
			WithHeader(InternalServerError).
			WithMessage("DumpHTML: wrapper is required")
	}

	payload := w.JSONBytesIgnoring(ignoringJSONfields...)

	var buf bytes.Buffer
	err := htmlTemplate.Execute(&buf, htmlDumpData{
		Title:       "Diagnostic Report",
		IsEST:       false,
		Payload:     string(payload),
		GeneratedAt: time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return nil, New().
			WithHeader(InternalServerError).
			WithErrorAck(err).
			WithMessage("DumpHTML: failed to execute HTML template")
	}

	d, err := dumpHTML(buf.Bytes())
	if err != nil {
		return nil, New().
			WithHeader(InternalServerError).
			WithErrorAck(err).
			WithMessage("DumpHTML: failed to create temp file")
	}

	return &Dump{syr: d}, New().
		WithHeader(OK).
		WithMessagef("DumpHTML: succeeded and written to temp file %s", d.Name())
}

// DumpResolveESTHTMLDoc serializes the [wrapper]'s diagnostic report as an interactive
// HTML document and writes it into a self-cleaning temporary file.
//
// The generated HTML visualizes the step-by-step resolution guide for the error
// present in the wrapper, including a root cause code block and an interactive
// checklist of stack frames.
//
// Both return values are always non-nil:
//   - (*Dump, *wrapper) — Dump holds the file; wrapper carries the outcome.
//   - On error: Dump is nil, wrapper has InternalServerError + error detail.
func (w *wrapper) DumpResolveESTHTMLDoc() (*Dump, *wrapper) {
	if !w.Available() {
		return nil, New().
			InternalServerError().
			WithMessage("DumpResolveESTHTMLDoc: wrapper is required")
	}

	mdStr := w.ResolveESTOrderedDoc().String()
	est := parseESTMarkdown(mdStr)

	var buf bytes.Buffer
	err := htmlTemplate.Execute(&buf, htmlDumpData{
		Title:        "Resolve Error (Ordered)",
		IsEST:        true,
		ErrorMessage: est.ErrorMessage,
		Frames:       est.Frames,
		GeneratedAt:  time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return nil, New().
			InternalServerError().
			WithErrorAck(err).
			WithMessage("DumpResolveESTHTMLDoc: failed to execute HTML template")
	}

	d, err := dumpHTML(buf.Bytes())
	if err != nil {
		return nil, New().
			InternalServerError().
			WithErrorAck(err).
			WithMessage("DumpResolveESTHTMLDoc: failed to create temp file")
	}

	return &Dump{syr: d}, New().
		OK().
		WithMessagef("DumpResolveESTHTMLDoc: succeeded and written to temp file %s", d.Name())
}
