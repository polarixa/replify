package replify

import "github.com/polarixa/replify/pkg/strchain"

func (w *wrapper) newSummaryDoc() *strchain.StringWeaver {
	sw := strchain.New()

	sw.Append("##").Space()
	sw.Append("Summary").NewLines(2)

	return sw
}
