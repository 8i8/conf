package conf

import "log"

var verbose = true

func init() {
	log.SetFlags(log.Llongfile)
}
