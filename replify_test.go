package replify_test

import (
	"sync"
	"testing"

	"github.com/polarixa/replify"
)

// TestRespond_ConcurrentSafety verifies that calling Respond from multiple
// goroutines does not trigger the race detector or produce inconsistent
// results. This exercises the double-checked locking pattern in Respond().
func TestRespond_ConcurrentSafety(t *testing.T) {
	t.Parallel()

	w := replify.New().
		WithHeader(replify.OK).
		WithMessage("concurrent test").
		WithBody(map[string]string{"key": "value"})

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			resp := w.Respond()
			if resp == nil {
				t.Error("Respond() returned nil")
			}
		}()
	}
	wg.Wait()
}
