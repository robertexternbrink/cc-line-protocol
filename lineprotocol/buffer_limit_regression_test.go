package lineprotocol

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestReaderBufferLimitAcceptsCompleteLineAtLimit(t *testing.T) {
	for _, text := range []string{"audit value=1 0\n", "audit value=1 0"} {
		for _, bufferSlack := range []int{0, 1} {
			t.Run(fmt.Sprintf("newline=%t/slack=%d", strings.HasSuffix(text, "\n"), bufferSlack), func(t *testing.T) {
				decoder := NewDecoder(strings.NewReader(text))
				decoder.SetMaxBufferSize(len(text) + bufferSlack)
				if !decoder.Next() {
					t.Fatalf("entry unavailable: %v", decoder.Err())
				}
				if _, err := decoder.Measurement(); err != nil {
					t.Fatal(err)
				}
				for {
					key, _, err := decoder.NextTag()
					if err != nil {
						t.Fatal(err)
					}
					if key == nil {
						break
					}
				}
				for {
					key, _, err := decoder.NextField()
					if err != nil {
						t.Fatal(err)
					}
					if key == nil {
						break
					}
				}
				if _, err := decoder.TimeBytes(); err != nil {
					t.Fatal(err)
				}
				if decoder.Next() {
					t.Fatal("unexpected extra entry")
				}
				if err := decoder.Err(); err != nil && err != io.EOF {
					t.Fatalf("valid %d-byte line within %d-byte limit rejected: %v", len(text), len(text)+bufferSlack, err)
				}
			})
		}
	}
}

type limitedChunkReader struct {
	data     *strings.Reader
	chunk    int
	terminal error
}

func (r *limitedChunkReader) Read(p []byte) (int, error) {
	if len(p) > r.chunk {
		p = p[:r.chunk]
	}
	n, err := r.data.Read(p)
	if err == io.EOF || r.data.Len() == 0 {
		return n, r.terminal
	}
	return n, err
}

func TestReaderBufferLimitStreamAndTerminalErrors(t *testing.T) {
	const line = "audit value=1 0\n"
	fault := errors.New("synthetic bounded decoder reader failure")
	for _, tt := range []struct {
		name, input string
		limit       int
		chunk       int
		terminal    error
		entries     int
		wantErr     bool
	}{
		{"multiple maximum lines", line + line + line, len(line), 4096, io.EOF, 3, false},
		{"multiple maximum lines in short reads", line + line + line, len(line), 1, io.EOF, 3, false},
		{"EOF combined with last byte", strings.TrimSuffix(line, "\n"), len(line) - 1, 4096, io.EOF, 1, false},
		{"reader error combined with last bytes", line, len(line), 4096, fault, 1, true},
		{"reader error after exact non-newline limit", strings.TrimSuffix(line, "\n"), len(line) - 1, 1, fault, 1, true},
		{"one byte beyond limit", line, len(line) - 1, 4096, io.EOF, 1, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			decoder := NewDecoder(&limitedChunkReader{strings.NewReader(tt.input), tt.chunk, tt.terminal})
			decoder.SetMaxBufferSize(tt.limit)
			entries := 0
			for decoder.Next() {
				if _, err := decoder.Measurement(); err != nil {
					t.Fatal(err)
				}
				for {
					key, _, err := decoder.NextTag()
					if err != nil {
						t.Fatal(err)
					}
					if key == nil {
						break
					}
				}
				for {
					key, _, err := decoder.NextField()
					if err != nil {
						t.Fatal(err)
					}
					if key == nil {
						break
					}
				}
				if _, err := decoder.TimeBytes(); err != nil {
					t.Fatal(err)
				}
				entries++
				if cap(decoder.buf) > tt.limit {
					t.Fatalf("internal capacity%d exceeded limit%d", cap(decoder.buf), tt.limit)
				}
			}
			if entries != tt.entries || (decoder.Err() != nil) != tt.wantErr {
				t.Fatalf("entries%d want%d error%v", entries, tt.entries, decoder.Err())
			}
			if tt.terminal == fault && !errors.Is(decoder.Err(), fault) {
				t.Fatalf("reader error not preserved: %v", decoder.Err())
			}
		})
	}
}
