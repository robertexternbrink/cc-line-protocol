package main

import "testing"

func TestVerifyLineBytesRejectsInvalidNumericTimestamps(t *testing.T) {
	for _, tc := range []struct {
		timestamp string
		want      int
	}{
		{"-", 1},
		{"--1", 1},
		{"1-2", 1},
		{"9223372036854775808", 1},
		{"-9223372036854775809", 1},
		{"-1", 0},
		{"0", 0},
		{"9223372036854775807", 0},
		{"-9223372036854775808", 0},
		{"", 0},
	} {
		t.Run(tc.timestamp, func(t *testing.T) {
			if got := verifyLineBytes([]byte("audit value=1 " + tc.timestamp + "\n")); got != tc.want {
				t.Fatalf("timestamp %q: result %d, want %d", tc.timestamp, got, tc.want)
			}
		})
	}
}
