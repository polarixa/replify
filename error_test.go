package replify_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/polarixa/replify"
)

func TestWithError(t *testing.T) {
	err := replify.NewError("Test error")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if err.Error() != "Test error" {
		t.Errorf("Expected 'Test error', got %s", err.Error())
	}
}

func TestWithErrorf(t *testing.T) {
	err := replify.NewErrorf("Failed to load file %s", "test.txt")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if err.Error() != "Failed to load file test.txt" {
		t.Errorf("Expected 'Failed to load file test.txt', got %s", err.Error())
	}
}

func TestWithErrStack(t *testing.T) {
	originalErr := errors.New("original error")
	errWithStack := replify.NewErrorAck(originalErr)
	if errWithStack == nil {
		t.Errorf("Expected error with stack, got nil")
	}
	if errWithStack.Error() != "original error" {
		t.Errorf("Expected 'original error', got %s", errWithStack.Error())
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("file not found")
	wrappedErr := replify.AppendErrorAck(originalErr, "Failed to read the file")
	if wrappedErr == nil {
		t.Errorf("Expected wrapped error, got nil")
	}
	if wrappedErr.Error() != "Failed to read the file: file not found" {
		t.Errorf("Expected 'Failed to read the file: file not found', got %s", wrappedErr.Error())
	}
}

func TestWrapf(t *testing.T) {
	originalErr := errors.New("file not found")
	wrappedErr := replify.NewErrorAckf(originalErr, "Failed to open file %s", "data.txt")
	if wrappedErr == nil {
		t.Errorf("Expected wrapped error, got nil")
	}
	if wrappedErr.Error() != "Failed to open file data.txt: file not found" {
		t.Errorf("Expected 'Failed to open file data.txt: file not found', got %s", wrappedErr.Error())
	}
}

func TestWithMessage(t *testing.T) {
	originalErr := errors.New("original error")
	errWithMessage := replify.AppendError(originalErr, "Additional context")
	if errWithMessage == nil {
		t.Errorf("Expected error with message, got nil")
	}
	if errWithMessage.Error() != "Additional context: original error" {
		t.Errorf("Expected 'Additional context: original error', got %s", errWithMessage.Error())
	}
}

func TestWithMessagef(t *testing.T) {
	originalErr := errors.New("original error")
	errWithMessagef := replify.AppendErrorf(originalErr, "Context: %s", "something went wrong")
	if errWithMessagef == nil {
		t.Errorf("Expected error with message, got nil")
	}
	if errWithMessagef.Error() != "Context: something went wrong: original error" {
		t.Errorf("Expected 'Context: something went wrong: original error', got %s", errWithMessagef.Error())
	}
}

func TestCause(t *testing.T) {
	originalErr := errors.New("file not found")
	wrappedErr := replify.AppendErrorAck(originalErr, "Failed to open file")
	causeErr := replify.Cause(wrappedErr)
	if causeErr == nil {
		t.Errorf("Expected cause error, got nil")
	}
	if causeErr.Error() != "file not found" {
		t.Errorf("Expected 'file not found', got %s", causeErr.Error())
	}
}

func TestErrorFormatting(t *testing.T) {
	err := replify.NewError("Test error")
	if err.Error() != "Test error" {
		t.Errorf("Expected 'Test error', got %s", err.Error())
	}

	// Ensure that the format includes the stack trace if the '+' flag is set
	if errFmt := fmt.Sprintf("%+v", err); errFmt == "Test error" {
		t.Errorf("Expected formatted string with stack trace, got %s", errFmt)
	}
}

// func TestUnwrap(t *testing.T) {
// 	originalErr := errors.New("file not found")
// 	wrappedErr := replify.AppendErrorAck(originalErr, "Failed to open file")
// 	unwrappedErr := wrappedErr.(*replify.underlyingStack).Unwrap()
// 	if unwrappedErr == nil {
// 		t.Errorf("Expected unwrapped error, got nil")
// 	}
// 	if unwrappedErr.Error() != "file not found" {
// 		t.Errorf("Expected 'file not found', got %s", unwrappedErr.Error())
// 	}
// }
