package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"sort"

	"github.com/polarixa/replify/pkg/slogger"
	"github.com/polarixa/replify/pkg/strchain"
	"github.com/polarixa/replify/pkg/translit/table"
)

func main() {
	sections := make([]int, 0, len(table.Tables))
	for sec := range table.Tables {
		sections = append(sections, int(sec))
	}
	sort.Ints(sections)

	maxSection := 0
	for _, s := range sections {
		if s > maxSection {
			maxSection = s
		}
	}

	sectionStart := make([]int32, maxSection+1)
	sectionLen := make([]uint16, maxSection+1)
	for i := range sectionStart {
		sectionStart[i] = -1
	}

	var data bytes.Buffer
	var entryOff []uint32
	var entryLen []uint8

	for _, sec := range sections {
		tb := table.Tables[rune(sec)]
		sectionStart[sec] = int32(len(entryOff))
		sectionLen[sec] = uint16(len(tb))
		for _, s := range tb {
			off := data.Len()
			if off > 1<<32-1 {
				slogger.Fatalf("data offset overflow")
			}
			if len(s) > 255 {
				slogger.Fatalf("replacement %q exceeds 255 bytes", s)
			}
			entryOff = append(entryOff, uint32(off))
			entryLen = append(entryLen, uint8(len(s)))
			data.WriteString(s)
		}
	}

	c := strchain.New()
	emitHeader(c)
	c.AppendF("const maxSection = %d", maxSection).NewLines(2)
	c.Append("// data holds every replacement string back to back, in the").NewLine()
	c.Append("// order the (section, position) pairs below were emitted.").NewLine()
	c.AppendF("// %d bytes.", data.Len()).NewLine()
	c.AppendF("const data = %q", data.String()).NewLines(2)
	c.Append("// sectionStart[s] is the index into entryOff/entryLen where").NewLine()
	c.Append("// section s's positions begin, or -1 if section s has no data.").NewLine()
	emitInt32(c, "sectionStart", sectionStart)
	c.Append("// sectionLen[s] is the number of valid positions (0..255) in section s.").NewLine()
	emitUint16(c, "sectionLen", sectionLen)
	c.Append("// entryOff[i] is the byte offset into data of entry i.").NewLine()
	emitUint32(c, "entryOff", entryOff)
	c.Append("// entryLen[i] is the byte length within data of entry i.").NewLine()
	emitUint8(c, "entryLen", entryLen)

	out, err := format.Source(c.Bytes())
	if err != nil {
		os.WriteFile("tables_gen.go.raw", c.Bytes(), 0644)
		slogger.Fatalf("gofmt: %v", err)
	}

	if err := os.WriteFile("../tables_gen.go", out, 0644); err != nil {
		slogger.Fatalf("write tables_gen.go: %v", err)
	}

	slogger.Infof("[translit/gen] tables_gen.go: sections=%d maxSection=0x%04x entries=%d dataBytes=%d",
		len(sections), maxSection, len(entryOff), data.Len())
}

// emitRows writes n comma-separated cells into c, inserting a newline every
// perRow cells. The text for each cell at index i is produced by calling
// cell(i). A final newline is always appended after the last element.
//
// emitRows is the shared row-formatting helper for all array-emission
// functions; centralising the layout logic ensures consistent column width
// across every generated array, regardless of element type.
func emitRows(c *strchain.StringWeaver, n, perRow int, cell func(int) string) {
	for i := 0; i < n; i++ {
		c.Append(cell(i)).AppendByte(',')
		if (i+1)%perRow == 0 {
			c.NewLine()
		}
	}
	c.NewLine()
}

// emitInt32 writes a Go var declaration for a fixed-size [N]int32
// array named name into c, where N is len(v). Elements are arranged
// 16 per row by emitRows. A blank line is appended after the closing brace
// to separate successive declarations in the generated file.
func emitInt32(c *strchain.StringWeaver, name string, v []int32) {
	c.AppendF("var %s = [%d]int32{", name, len(v)).NewLine()
	emitRows(c, len(v), 16, func(i int) string {
		return fmt.Sprintf("%d", v[i])
	})
	c.Append("}").NewLines(2)
}

// emitUint16 writes a Go var declaration for a fixed-size [N]uint16
// array named name into c, where N is len(v). Elements are arranged
// 16 per row by emitRows. A blank line is appended after the closing brace
// to separate successive declarations in the generated file.
func emitUint16(c *strchain.StringWeaver, name string, v []uint16) {
	c.AppendF("var %s = [%d]uint16{", name, len(v)).NewLine()
	emitRows(c, len(v), 16, func(i int) string {
		return fmt.Sprintf("%d", v[i])
	})
	c.Append("}").NewLines(2)
}

// emitUint32 writes a Go var declaration for a fixed-size [N]uint32
// array named name into c, where N is len(v). Elements are arranged
// 12 per row by emitRows (narrower than 16 because uint32 decimal values
// are wider on average). A blank line is appended after the closing brace.
func emitUint32(c *strchain.StringWeaver, name string, v []uint32) {
	c.AppendF("var %s = [%d]uint32{", name, len(v)).NewLine()
	emitRows(c, len(v), 12, func(i int) string {
		return fmt.Sprintf("%d", v[i])
	})
	c.Append("}").NewLines(2)
}

// emitUint8 writes a Go var declaration for a fixed-size [N]uint8
// array named name into c, where N is len(v). Elements are arranged
// 20 per row by emitRows (more per row than other types because uint8
// values are at most 3 decimal digits wide). A blank line is appended
// after the closing brace.
func emitUint8(c *strchain.StringWeaver, name string, v []uint8) {
	c.AppendF("var %s = [%d]uint8{", name, len(v)).NewLine()
	emitRows(c, len(v), 20, func(i int) string {
		return fmt.Sprintf("%d", v[i])
	})
	c.Append("}").NewLines(2)
}

// emitHeader writes the standard generated-file preamble into c:
// the machine-readable "DO NOT EDIT" banner recognised by tools such as
// go generate and gopls, the regeneration hint, and the package declaration.
// It must be called before any other content is written to c.
func emitHeader(c *strchain.StringWeaver) {
	c.Append("// Code generated by gen/main.go; DO NOT EDIT.").NewLines(2)
	c.Append("// Regenerate with: go generate ./...").NewLines(2)
	c.Append("package translit").NewLines(2)
}
