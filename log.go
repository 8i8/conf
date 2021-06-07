package conf

import "log"

var verbose = false

func init() {
	log.SetFlags(log.Llongfile)
}
