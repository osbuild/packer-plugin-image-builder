package main

import (
	"io"
	"testing"
)

func TestTailWriterStringWritethough(t *testing.T) {
	tests := []struct {
		size int
		data []string
		want string
	}{
		{
			size: 1,
			data: []string{"abc"},
			want: "c",
		},
		{
			size: 2,
			data: []string{"abc"},
			want: "bc",
		},
		{
			size: 3,
			data: []string{"abcde"},
			want: "cde",
		},
		{
			size: 2,
			data: []string{"abc\n"},
			want: "c\n",
		},
		{
			size: 5,
			data: []string{"hello", "world"},
			want: "world",
		},
	}

	for _, tt := range tests {
		tw := NewTailWriter(tt.size)
		for _, d := range tt.data {
			tw.Write([]byte(d))
		}
		if got := tw.String(); got != tt.want {
			t.Errorf("unexpected tail: got %q, want %q", got, tt.want)
		}

		tw = NewTailWriterThrough(tt.size, io.Discard, nil)
		for _, d := range tt.data {
			tw.Write([]byte(d))
		}
		if got := tw.String(); got != tt.want {
			t.Errorf("unexpected tail: got %q, want %q", got, tt.want)
		}
	}
}

func TestTailWriterLastLines(t *testing.T) {
	tests := []struct {
		size  int
		data  []string
		lines int
		want  []string
	}{
		{
			size:  1,
			data:  []string{"a"},
			lines: 1,
			want:  []string{"a"},
		},
		{
			size:  2,
			data:  []string{"abc\n"},
			lines: 1,
			want:  []string{"c"},
		},
		{
			size:  1024,
			data:  []string{"one\n", "two\n", "three\n"},
			lines: 1,
			want:  []string{"three"},
		},
		{
			size:  1024,
			data:  []string{"one\n", "two\n", "three\n"},
			lines: 5,
			want:  []string{"one", "two", "three"},
		},
		{
			size:  1024,
			data:  []string{"one\n", "two\n", "three\n"},
			lines: 2,
			want:  []string{"two", "three"},
		},
	}

	for _, tt := range tests {
		tw := NewTailWriter(tt.size)
		for _, d := range tt.data {
			tw.Write([]byte(d))
		}
		if got := tw.LastLines(tt.lines); !equal(got, tt.want) {
			t.Errorf("unexpected last lines: got %q, want %q", got, tt.want)
		}
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
