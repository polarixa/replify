package replify

import (
	"github.com/polarixa/replify/pkg/strchain"
	"github.com/polarixa/replify/pkg/strutil"
)

func (w *wrapper) NewSummaryDoc() *strchain.StringWeaver {
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
