/*
package conf example
package main

import (
	"log"

	"github.com/8i8/conf/conf"
	"github.com/8i8/conf/types"
)

func main() {
	conf.Help(helpBase)
	if err := conf.Modes(&def, &one, &two); err != nil {
		log.Fatal("Config: ", err)
	}
	conf.Options(options()...)
	conf.Parse()
}

var (
	def = conf.Mode{Name: "def", Help: helpDef}
	one = conf.Mode{Name: "one", Help: helpOne}
	two = conf.Mode{Name: "two", Help: helpTwo}
)

func options() []conf.Option {
	return []conf.Option{
		{Name: "intie",
			Type:    types.Int,
			Key:     "n",
			Default: 12,
			Help:    intie,
			Modes:   conf.SetBits(def, one, two),
		},
		{Name: "thing",
			Type:    types.String,
			Key:     "s",
			Default: "Some thing",
			Help:    thing,
			Modes:   conf.SetBits(def, one, two),
		},
		{Name: "none",
			Type:    types.Int,
			Key:     "i",
			Default: 16,
			Help:    "the i is the none of all the ints",
			Modes:   conf.SetBits(def),
		},
		{Name: "verbosity",
			Type:    types.Int,
			Key:     "v",
			Default: 0,
			Help:    "The overall chattiness of it all",
			Modes:   conf.SetBits(def, one, two),
		},
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
//  Base
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var helpBase = `NAME
        conf

SYNOPSIS
        conf | [mode] | -[flag] | -[flag] <value> | -[flag] <'value,value,value'>

EXAMPLE
	conf one -n 36 -s "Hello, World!"`

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
//  Modes
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var helpDef = `
MODES
        conf [mode] -[flag]

	one     one does all things in the oneiest way.

	two     two, despite appearances is second to none, doing things in an
	        agreeable two like fashion.

	Further detatils of the use of each mode can be found by running the
	following command.

	conf [mode] -help  or conf [mode] -h

FLAGS`

var helpOne = `MODE
	conf one

FLAGS`

var helpTwo = `MODE
	conf two

FLAGS`

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
//  Flags
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var intie = `This is the very default value in the most simple mode, to test
if another way of writing the messages might be better.
`

var thing = `This is the default string thing, so as to best exemplify
the use of this package in its current state I thought it
best to write something very wordy here.
`
*/
package conf
