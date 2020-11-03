package conf

import (
	"github.com/8i8/cmd/conf/types"
)

// Option contains all of the data required for setting a default flag and
// receiving subsequent option settings.
type Option struct {
	Name    string      // The name of the option, also used as a key in the options map.
	Type    types.T     // Type of flag
	Key     string      // Keypress required to activate the flag.
	Help    string      // Help string
	flag    interface{} // Data
	Default interface{} // Default data
	set     bool        // Can the value be overridden?
	Modes   int         // Which program modes should the flag be included in?
}

// loadOptions loads all of the given options into the option map.
func (c *config) loadOptions(opts ...Option) {
	c.options = make(map[string]*Option)
	for i, opt := range opts {
		c.options[opt.Name] = &opts[i]
	}
}

// Options initialises the programs options.
func Options(opts ...Option) {

	// Load all default options.
	c.loadOptions(opts...)
	//c.loadConfig()
}
