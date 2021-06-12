package conf

import (
	"log"
	"sync/atomic"
)

const (
	none uint32 = 1 << iota
	one
	two
	three
)

var level = none

func init() {
	log.SetFlags(log.Llongfile)
}

func v1() bool { return level >= one }
func v2() bool { return level >= two }
func v3() bool { return level >= three }

func v(l uint32) uint32 {
	prev := level
	atomic.StoreUint32(&level, l)
	return prev
}
