package conf

import (
	"flag"
	"log"

	"github.com/8i8/cmd/conf/types"
)

// The program options.
var c config

// config contains all the program configuration config and flags.
type config struct {
	Mode
	help    string
	flagSet *flag.FlagSet
	options map[string]*Option
}

// loadFlags sets the options into the flagset.
func loadFlags(fs *flag.FlagSet, opt *Option) {
	switch opt.Type {
	case types.Int:
		opt.flag = fs.Int(opt.Key, opt.Default.(int), opt.Help)
	case types.String:
		opt.flag = fs.String(opt.Key, opt.Default.(string), opt.Help)
	case types.Bool:
		opt.flag = fs.Bool(opt.Key, opt.Default.(bool), opt.Help)
	default:
		log.Fatal("flag type not recognised")
	}
}

// setOptionsToFlags defines flags within the flagset for all options that have
// been specified that are within the current working modes option set.
func (c *config) setOptionsToFlags() {
	for k, opt := range c.options {
		if c.Mode.id&opt.Modes > 0 {
			loadFlags(c.flagSet, c.options[k])
		}
	}
}
