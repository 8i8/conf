package conf

import "log"

const (
	nop = 1 << iota
	one
	two
	three
)

var v = nop

func init() {
	log.SetFlags(log.Llongfile)
}

func v1() bool { return one&v != 0 }
func v2() bool { return two&v != 0 }
func v3() bool { return three&v != 0 }
