package main

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"sync"
)

// TailWriter implements a byte ring buffer that keeps last N bytes of the
// written data. It also supports writethrough writer and regexp callback.
type TailWriter struct {
	buf    []byte
	length int
	r      int
	w      int

	mu sync.Mutex

	writethrough io.Writer
	rc           *RegexpCallback
}

type RegexpCallback struct {
	Regexp   *regexp.Regexp
	Prefix   string
	Callback func(string)
	m        *bytes.Buffer
}

// NewTailWriter creates a new TailWriter with the given size.
func NewTailWriter(size int) *TailWriter {
	if size <= 0 {
		panic("size must be positive")
	}

	return &TailWriter{
		buf: make([]byte, size),
	}
}

// NewTailWriterThrough creates a new TailWriter with the given size, the writethrough
// writer and the regexp callback.
func NewTailWriterThrough(size int, writethrough io.Writer, rc *RegexpCallback) *TailWriter {
	tw := NewTailWriter(size)
	tw.writethrough = writethrough
	if rc != nil {
		tw.rc = rc
		tw.rc.m = &bytes.Buffer{}
	}
	return tw
}

// Write writes p to the buffer.
func (tw *TailWriter) Write(p []byte) (n int, err error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	for _, b := range p {
		tw.buf[tw.w] = b
		tw.w = (tw.w + 1) % len(tw.buf)

		if tw.length >= len(tw.buf) {
			tw.r = tw.w
		} else {
			tw.length++
		}
	}

	if tw.rc != nil && tw.rc.Regexp != nil && tw.rc.Callback != nil {
		match := tw.rc.Regexp.Find(p)
		if match != nil && !bytes.Equal(match, tw.rc.m.Bytes()) {
			tw.rc.Callback(tw.rc.Prefix + string(match))
			tw.rc.m.Reset()
			tw.rc.m.Write(match)
		}
	}

	if tw.writethrough != nil {
		return tw.writethrough.Write(p)
	}

	return len(p), nil
}

// String returns the written data as a string.
func (tw *TailWriter) String() string {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if tw.r < tw.w {
		return string(tw.buf[tw.r:tw.w])
	}
	return string(tw.buf[tw.r:]) + string(tw.buf[:tw.w])
}

// LastLines returns last n non-empty lines.
func (tw *TailWriter) LastLines(n int) []string {
	lines := make([]string, 0, n)

	for _, line := range strings.Split(tw.String(), "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}

	if len(lines) < n {
		return lines
	}

	return lines[len(lines)-n:]
}
