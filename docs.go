package replify

import (
	"github.com/polarixa/replify/pkg/conv"
	"github.com/polarixa/replify/pkg/strchain"
	"github.com/polarixa/replify/pkg/strutil"
)

// SummaryDoc generates a summary document for the wrapper instance, providing a concise overview of its status and message. It returns a StringWeaver instance that can be used to build and format the summary content.
//
// The summary document includes the following information:
//   - A header indicating that it is a summary.
//   - A table with two columns: "Field" and "Value".
//   - If a message is present, it will be included in the summary.
//   - If a status code is present, both the status code and its corresponding HTTP status text will be included.
//
// The generated summary document can be used for logging, reporting, or any other purpose where a concise overview of the wrapper's state is needed.
func (w *wrapper) SummaryDoc() *strchain.StringWeaver {
	w.autoAdjust()
	sw := strchain.New()
	sw.Append("##").Space()
	sw.Append("Summary").NewLines(2)
	sw.Pipe().Space().Append("Field").Space().Pipe().Space().Append("Value").Space().Pipe().NewLine()
	sw.Pipe().Dashes(3).Pipe().Dashes(3).Pipe().NewLine()

	sw.Pipe().Space().Append("Message").Space().Pipe().Space().
		Underscore().
		Append(strutil.DefaultIfEmpty(escapeMarkdownPipe(w.Message()), "<empty>")).
		Underscore().Space().Pipe().NewLine()

	if w.IsStatusCodePresent() {
		sw.Pipe().Space().Append("Status Code").Space().Pipe().Space().AppendInt(w.StatusCode()).Space().Pipe().NewLine()
		sw.Pipe().Space().Append("HTTP Status").Space().Pipe().Space().Append(escapeMarkdownPipe(w.StatusText())).Space().Pipe().NewLine()
	}

	return sw
}

// DebugDoc generates a debug document for the wrapper instance, providing detailed information about its debugging state. It returns a StringWeaver instance that can be used to build and format the debug content.
//
// The debug document includes the following information:
//   - A header indicating that it is debug information.
//   - A table with two columns: "Key" and "Value".
//   - If debugging information is present, each key-value pair will be included in the table, except for the "error_stack_trace" key, which will be skipped.
//
// The generated debug document can be used for logging, troubleshooting, or any other purpose where detailed debugging information is needed.
func (w *wrapper) DebugDoc() *strchain.StringWeaver {
	sw := strchain.New()
	if !w.IsDebuggingPresent() {
		return sw
	}
	sw.Append("##").Space()
	sw.Append("Debug Information").NewLines(2)
	sw.Pipe().Space().Append("Key").Space().Pipe().Space().Append("Value").Space().Pipe().NewLine()
	sw.Pipe().Dashes(3).Pipe().Dashes(3).Pipe().NewLine()

	for k, v := range w.Debugging() {
		if k == "error_stack_trace" { // Skip the error stack trace in the debug document
			continue
		}
		sw.Pipe().Space().Append(escapeMarkdownPipe(k)).Space().Pipe().Space().Append(escapeMarkdownPipe(conv.StringOrEmpty(v))).Space().Pipe().NewLine()
	}
	return sw
}

// ErrorStackTraceDoc generates a document for the error stack trace present in the wrapper instance, providing a detailed view of the error context. It returns a StringWeaver instance that can be used to build and format the error stack trace content.
//
// The error stack trace document includes the following information:
//   - A header indicating that it is an error stack trace.
//   - If debugging information is present and contains the "error_stack_trace" key, the corresponding value will be included in a code block format.
//
// The generated error stack trace document can be used for logging, troubleshooting, or any other purpose where detailed error context is needed.
func (w *wrapper) ErrorStackTraceDoc() *strchain.StringWeaver {
	sw := strchain.New()
	if !w.IsError() {
		return sw
	}
	w.autoAdjust()
	w.InjectStackTrace() // Ensure the stack trace is injected before generating the document

	sw.Append("##").Space()
	sw.Append("Error Stack Trace").NewLines(2)

	if w.IsDebuggingPresent() {
		if trace, ok := w.Debugging()["error_stack_trace"]; ok {
			codeblock := strchain.New()
			lines, _ := conv.StringSlice(trace)

			for i, line := range lines {
				codeblock.AppendF("%d. %s", i+1, escapeMarkdownPipe(line)).NewLine()
			}
			sw.CodeBlock("go", codeblock).NewLines(1)
		}
	}
	return sw
}

// SafeErrorStackTraceDoc generates a document for the error stack trace present in the wrapper instance, providing a detailed view of the error context. It returns a StringWeaver instance that can be used to build and format the error stack trace content.
//
// The SafeErrorStackTraceDoc method ensures that the stack trace injection is disabled after generating the document to avoid any side effects on the wrapper's state. This allows for safe retrieval of the error stack trace without modifying the wrapper's internal state.
//
// The error stack trace document includes the following information:
//   - A header indicating that it is an error stack trace.
//   - If debugging information is present and contains the "error_stack_trace" key, the corresponding value will be included in a code block format.
//
// The generated error stack trace document can be used for logging, troubleshooting, or any other purpose where detailed error context is needed.
func (w *wrapper) SafeErrorStackTraceDoc() *strchain.StringWeaver {
	sw := w.ErrorStackTraceDoc()
	w.DisableInjectStackTrace() // Disable stack trace injection after generating the document to avoid side effects
	return sw
}

// HeaderDoc generates a document for the headers present in the wrapper instance, providing a detailed view of the header information. It returns a StringWeaver instance that can be used to build and format the header content.
//
// The header document includes the following information:
//   - A header indicating that it is header information.
//   - A table with two columns: "Field" and "Value".
//   - If headers are present, each header field and its corresponding value will be included in the table.
//
// The generated header document can be used for logging, troubleshooting, or any other purpose where detailed header information is needed.
func (w *wrapper) HeaderDoc() *strchain.StringWeaver {
	sw := strchain.New()
	if !w.IsHeaderPresent() {
		return sw
	}
	sw.Append("##").Space()
	sw.Append("Headers").NewLines(2)
	sw.Pipe().Space().Append("Field").Space().Pipe().Space().Append("Value").Space().Pipe().NewLine()
	sw.Pipe().Dashes(3).Pipe().Dashes(3).Pipe().NewLine()

	h := w.Header()
	sw.Pipe().Space().Append("Code").Space().Pipe().Space().AppendInt(h.Code()).Space().Pipe().NewLine()
	sw.Pipe().Space().Append("Text").Space().Pipe().Space().Append(escapeMarkdownPipe(h.Text())).Space().Pipe().NewLine()
	sw.Pipe().Space().Append("Type").Space().Pipe().Space().Append(escapeMarkdownPipe(h.Type())).Space().Pipe().NewLine()
	return sw
}

// PagingDoc generates a document for the pagination information present in the wrapper instance, providing a detailed view of the pagination details. It returns a StringWeaver instance that can be used to build and format the pagination content.
//
// The pagination document includes the following information:
//   - A header indicating that it is pagination information.
//   - A table with two columns: "Field" and "Value".
//   - If pagination information is present, each pagination field and its corresponding value will be included in the table.
//
// The generated pagination document can be used for logging, troubleshooting, or any other purpose where detailed pagination information is needed.
func (w *wrapper) PagingDoc() *strchain.StringWeaver {
	sw := strchain.New()
	if !w.IsPagingPresent() {
		return sw
	}
	sw.Append("##").Space()
	sw.Append("Pagination").NewLines(2)
	sw.Pipe().Space().Append("Field").Space().Pipe().Space().Append("Value").Space().Pipe().NewLine()
	sw.Pipe().Dashes(3).Pipe().Dashes(3).Pipe().NewLine()

	p := w.Pagination()
	sw.Pipe().Space().Append("Page").Space().Pipe().Space().AppendInt(p.Page()).Space().Pipe().NewLine()
	sw.Pipe().Space().Append("Per Page").Space().Pipe().Space().AppendInt(p.PerPage()).Space().Pipe().NewLine()
	sw.Pipe().Space().Append("Total Items").Space().Pipe().Space().AppendInt(p.TotalItems()).Space().Pipe().NewLine()
	sw.Pipe().Space().Append("Total Pages").Space().Pipe().Space().AppendInt(p.TotalPages()).Space().Pipe().NewLine()
	sw.Pipe().Space().Append("Is Last").Space().Pipe().Space().AppendBool(p.IsLast()).Space().Pipe().NewLine()
	return sw
}

// CursorDoc generates a document for the cursor information present in the wrapper instance, providing a detailed view of the cursor details. It returns a StringWeaver instance that can be used to build and format the cursor content.
//
// The cursor document includes the following information:
//   - A header indicating that it is cursor information.
//   - A table with two columns: "Field" and "Value".
//   - If cursor information is present, each cursor field and its corresponding value will be included in the table.
//
// The generated cursor document can be used for logging, troubleshooting, or any other purpose where detailed cursor information is needed.
func (w *wrapper) CursorDoc() *strchain.StringWeaver {
	sw := strchain.New()
	if !w.IsCursorPresent() {
		return sw
	}
	sw.Append("##").Space()
	sw.Append("Cursor").NewLines(2)
	sw.Pipe().Space().Append("Field").Space().Pipe().Space().Append("Value").Space().Pipe().NewLine()
	sw.Pipe().Dashes(3).Pipe().Dashes(3).Pipe().NewLine()

	c := w.Cursor()
	sw.Pipe().Space().Append("Next").Space().Pipe().Space().Append(escapeMarkdownPipe(c.Next())).Space().Pipe().NewLine()
	sw.Pipe().Space().Append("Previous").Space().Pipe().Space().Append(escapeMarkdownPipe(c.Previous())).Space().Pipe().NewLine()
	sw.Pipe().Space().Append("Has Next").Space().Pipe().Space().AppendBool(c.HasNext()).Space().Pipe().NewLine()
	sw.Pipe().Space().Append("Has Previous").Space().Pipe().Space().AppendBool(c.HasPrevious()).Space().Pipe().NewLine()
	sw.Pipe().Space().Append("Limit").Space().Pipe().Space().AppendInt(c.Limit()).Space().Pipe().NewLine()
	return sw
}
