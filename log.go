package conf

import "log"

var verbose = 0

func init() {
	log.SetFlags(log.Llongfile)
}

func v(i int) bool {
	if verbose >= i {
		return true
	}
	return false
}
