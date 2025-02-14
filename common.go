package ibk

import (
	"bytes"
	"context"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// SyncedBuffer is a concurrent buffer writer. Can be used for combined
// stdout/stderr output and ensures that lines do not overlap.
type SyncedBuffer struct {
	b  bytes.Buffer
	mu sync.Mutex
}

// Write writes to the buffer
func (w *SyncedBuffer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.b.Write(p)
}

// String returns the buffer as a string
func (w *SyncedBuffer) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.b.String()
}

// FirstLine returns the first line of the output trimmed of leading and trailing whitespace
func (w *SyncedBuffer) FirstLine() string {
	w.mu.Lock()
	defer w.mu.Unlock()

	return strings.TrimSpace(strings.SplitN(w.b.String(), "\n", 1)[0])
}

// Reset clears the buffer
func (w *SyncedBuffer) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.b.Reset()
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

var randSource = rand.New(rand.NewSource(time.Now().UnixNano()))

// RandomString generates a random string of length n that is not cryptographically safe
func RandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[randSource.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

// Wait waits for the context to be done or for the function to return
func Wait(ctx context.Context, f func() error) error {
	done := make(chan error, 1)
	go func() {
		err := f()
		done <- err
	}()

	select {
	case err := <-done:
		return err

	case <-ctx.Done():
		return ctx.Err()
	}
}
