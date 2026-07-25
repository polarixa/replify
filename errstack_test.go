package replify_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/polarixa/replify"
)

func TestStackTrace(t *testing.T) {
	callers := replify.Callers()
	trace := callers.StackTrace()
	if len(trace) == 0 {
		t.Errorf("Expected non-empty stack trace")
	}
}

func TestFrameMarshalText(t *testing.T) {
	callers := replify.Callers()
	trace := callers.StackTrace()

	for _, frame := range trace {
		text, err := frame.MarshalText()
		if err != nil {
			t.Errorf("MarshalText returned an error: %v", err)
		}
		if len(text) == 0 {
			t.Errorf("MarshalText returned empty text")
		}
	}
}

func TestFrameFormat(t *testing.T) {
	callers := replify.Callers()
	trace := callers.StackTrace()

	for _, frame := range trace {
		buffer := &strings.Builder{}
		fmt.Fprintf(buffer, "%+v", frame)
		if buffer.Len() == 0 {
			t.Errorf("Formatted frame is empty")
		}
	}
}

func TestStackTraceFormat(t *testing.T) {
	callers := replify.Callers()
	trace := callers.StackTrace()

	buffer := &strings.Builder{}
	fmt.Fprintf(buffer, "%+v", trace)
	if buffer.Len() == 0 {
		t.Errorf("Formatted stack trace is empty")
	}
}

func TestStackFormat(t *testing.T) {
	callers := replify.Callers()
	buffer := &strings.Builder{}
	fmt.Fprintf(buffer, "%+v", callers)

	if buffer.Len() == 0 {
		t.Errorf("Formatted stack is empty")
	}
}

func TestCallers(t *testing.T) {
	callers := replify.Callers()
	if callers == nil {
		t.Errorf("Callers returned nil")
	}
	if callers == nil || len(*callers) == 0 {
		t.Errorf("Callers returned an empty stack")
	}
}

// func TestFrameFile(t *testing.T) {
// 	callers := replify.Callers()
// 	trace := callers.StackTrace()

// 	for _, frame := range trace {
// 		file := frame.File()
// 		if file == replify.ErrUnknown {
// 			t.Errorf("Frame file returned ErrUnknown")
// 		}
// 		if len(file) == 0 {
// 			t.Errorf("Frame file is empty")
// 		}
// 	}
// }

// func TestFrameLine(t *testing.T) {
// 	callers := replify.Callers()
// 	trace := callers.StackTrace()

// 	for _, frame := range trace {
// 		line := frame.Line()
// 		if line == 0 {
// 			t.Errorf("Frame line returned 0")
// 		}
// 	}
// }

// func TestFrameName(t *testing.T) {
// 	callers := replify.Callers()
// 	trace := callers.StackTrace()

// 	for _, frame := range trace {
// 		name := frame.Name()
// 		if name == "unknown" {
// 			t.Errorf("Frame name returned unknown")
// 		}
// 		if len(name) == 0 {
// 			t.Errorf("Frame name is empty")
// 		}
// 	}
// }
