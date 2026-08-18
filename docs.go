package replify

import (
	"github.com/polarixa/replify/pkg/conv"
	"github.com/polarixa/replify/pkg/strchain"
	"github.com/polarixa/replify/pkg/strutil"
)

// SummaryDoc generates a summary document for the wrapper instance, providing a concise overview of its status and message. It returns a StringWeaver instance that can be used to build and format the summary content.
//
// The summary document includes the following information:
// - A header indicating that it is a summary.
// - A table with two columns: "Field" and "Value".
// - If a message is present, it will be included in the summary.
// - If a status code is present, both the status code and its corresponding HTTP status text will be included.
//
// The generated summary document can be used for logging, reporting, or any other purpose where a concise overview of the wrapper's state is needed.
func (w *wrapper) SummaryDoc() *strchain.StringWeaver {
	sw := strchain.New()
	sw.Append("##").Space()
	sw.Append("Summary").NewLines(2)
	sw.Pipe().Space().Append("Field").Space().Pipe().Space().Append("Value").Space().Pipe().NewLine()
	sw.Pipe().Dashes(3).Pipe().Dashes(3).Pipe().NewLine()

	if strutil.IsNotEmpty(w.Message()) {
		sw.Pipe().Space().Append("Message").Space().Pipe().Space().Underscore().Append(escapeMarkdownPipe(w.Message())).Underscore().Space().Pipe().NewLine()
	}
	if w.IsStatusCodePresent() {
		sw.Pipe().Space().Append("Status Code").Space().Pipe().Space().AppendInt(w.StatusCode()).Space().Pipe().NewLine()
		sw.Pipe().Space().Append("HTTP Status").Space().Pipe().Space().Append(escapeMarkdownPipe(w.StatusText())).Space().Pipe().NewLine()
	}

	return sw
}

// DebugDoc generates a debug document for the wrapper instance, providing detailed information about its debugging state. It returns a StringWeaver instance that can be used to build and format the debug content.
//
// The debug document includes the following information:
// - A header indicating that it is debug information.
// - A table with two columns: "Key" and "Value".
// - If debugging information is present, each key-value pair will be included in the table, except for the "error_stack_trace" key, which will be skipped.
//
// The generated debug document can be used for logging, troubleshooting, or any other purpose where detailed debugging information is needed.
func (w *wrapper) DebugDoc() *strchain.StringWeaver {
	sw := strchain.New()
	sw.Append("##").Space()
	sw.Append("Debug Information").NewLines(2)
	sw.Pipe().Space().Append("Key").Space().Pipe().Space().Append("Value").Space().Pipe().NewLine()
	sw.Pipe().Dashes(3).Pipe().Dashes(3).Pipe().NewLine()

	if w.IsDebuggingPresent() {
		for k, v := range w.Debugging() {
			if k == "error_stack_trace" {
				continue
			}
			sw.Pipe().Space().Append(escapeMarkdownPipe(k)).Space().Pipe().Space().Append(escapeMarkdownPipe(conv.StringOrEmpty(v))).Space().Pipe().NewLine()
		}
	}
	return sw
}
