package conf

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/8i8/conf/types"
)

// The program options.
var c config
var pkg = "conf"

// config contains all the program configuration config and flags.
type config struct {
	// Mode is the running mode of the program, this package facilitates
	// the generation of different substates to help keep the use of
	// option flags simple.
	mode
	// The string that is displayed as the help output header for the entire program.
	help string
	// flagset is where the flags are places once parsed.
	flagSet *flag.FlagSet
	// options are where the complete data for each flag is stored, this
	// includes the value of the key its default value and help string
	// along with the actual data once the flag has been parsed.
	options map[string]*Option
}

// Parse sets the running mode from the command line arguments.
func Parse() {
	if len(os.Args) > 1 && os.Args[1][0] != '-' {
		if list.is(os.Args[1]) {
			loadMode(os.Args[1])
			return
		}
		fmt.Printf("unknown mode: %q\n", os.Args[1])
		return
	}
	loadMode("def")
}

// Help sets the basis for the programs help output, the 'help header'.
func Help(help string) {
	c.help = help
}

// toFlagSet generates a flag within the given flagset for the current option.
func (o *Option) toFlagSet(fs *flag.FlagSet) {
	switch o.Type {
	case types.Int:
		o.flag = fs.Int(o.Key, o.Default.(int), o.Help)
	case types.String:
		o.flag = fs.String(o.Key, o.Default.(string), o.Help)
	case types.Bool:
		o.flag = fs.Bool(o.Key, o.Default.(bool), o.Help)
	case types.Float:
		o.flag = fs.Float64(o.Key, o.Default.(float64), o.Help)
	case types.Duration:
		o.flag = fs.Duration(o.Key, o.Default.(time.Duration), o.Help)
	default:
		log.Fatalf("conf: internal error: flag type not recognised "+
			"%q (%v, %T)", o.Name, o.Type)
	}
}

// optionsToFlagSet defines flags within the flagset for all options that have
// been specified that are within the current working set.
func (c *config) optionsToFlagSet() {
	for k, opt := range c.options {
		if c.mode.id&opt.Modes > 0 {
			c.options[k].toFlagSet(c.flagSet)
		}
	}
}
