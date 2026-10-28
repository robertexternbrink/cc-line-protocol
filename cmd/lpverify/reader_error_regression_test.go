package main

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/ClusterCockpit/cc-line-protocol/v2/lineprotocol"
)

type verifierFaultReader struct {
	prefix *strings.Reader
	fault  error
}

func (r *verifierFaultReader) Read(p []byte) (int, error) {
	n, err := r.prefix.Read(p)
	if err == io.EOF {
		return n, r.fault
	}
	return n, err
}

func TestVerifyReportsReaderFailures(t *testing.T) {
	for _, prefix := range []string{"", "audit value=1 0\n"} {
		decoder := lineprotocol.NewDecoder(&verifierFaultReader{strings.NewReader(prefix), errors.New("synthetic verifier reader failure")})
		if verify(decoder) || decoder.Err() == nil {
			t.Fatal("verifier must report the decoder's non-EOF reader failure")
		}
		if !verify(lineprotocol.NewDecoder(strings.NewReader(prefix))) {
			t.Fatal("normal EOF must remain successful")
		}
	}
}
