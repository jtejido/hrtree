package hrtree

import (
	"fmt"
)

func assert(ok bool) {
	assert2(ok, "assertion failed!")
}

func assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		panic(fmt.Sprintf(msg, args...))
	}
}
