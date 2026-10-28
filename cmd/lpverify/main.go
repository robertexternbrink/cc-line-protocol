package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ClusterCockpit/cc-line-protocol/v2/lineprotocol"
)

func main() {
	dec := lineprotocol.NewDecoder(os.Stdin)
	if !verify(dec) {
		os.Exit(1)
	}
}

func verify(dec *lineprotocol.Decoder) (ok bool) {
	ok = true
	logErr := func(err error) {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		ok = false
	}
nextLine:
	for dec.Next() {
		_, err := dec.Measurement()
		if err != nil {
			logErr(err)
			continue nextLine
		}
		for {
			key, _, err := dec.NextTag()
			if err != nil {
				logErr(err)
				continue nextLine
			}
			if key == nil {
				break
			}
		}
		for {
			key, _, err := dec.NextField()
			if err != nil {
				logErr(err)
				continue nextLine
			}
			if key == nil {
				break
			}
		}
		// Precision-dependent time bounds remain separate from integer syntax.
		timestamp, err := dec.TimeBytes()
		if err != nil {
			logErr(err)
			continue nextLine
		}
		if timestamp != nil {
			if _, err := strconv.ParseInt(string(timestamp), 10, 64); err != nil {
				logErr(fmt.Errorf("invalid timestamp: %w", err))
			}
		}
	}
	return ok
}
