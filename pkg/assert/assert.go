package assert

import "log"

func Assert(truth bool, msg string) {
	if !truth {
		log.Fatal(msg)
	}
}
