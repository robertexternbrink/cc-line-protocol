package main

import (
	"strings"
	"testing"

	"github.com/ClusterCockpit/cc-line-protocol/v2/lineprotocol"
)

func TestVerifyRejectsInvalidNumericTimestamps(t *testing.T) {
	for _, tt := range []struct {
		timestamp string
		valid     bool
	}{
		{"-", false},
		{"--1", false},
		{"1-2", false},
		{"9223372036854775808", false},
		{"-9223372036854775809", false},
		{"-1", true},
		{"0", true},
		{"9223372036854775807", true},
		{"-9223372036854775808", true},
		{"", true},
	} {
		t.Run(tt.timestamp, func(t *testing.T) {
			input := "audit value=1" + " " + tt.timestamp + "\n"
			if actual := verify(lineprotocol.NewDecoder(strings.NewReader(input))); actual != tt.valid {
				t.Fatalf("verification accepted%t want%t", actual, tt.valid)
			}
		})
	}
}
