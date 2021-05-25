### conf

Package conf helps you to organise and maintain package sub-commands,
their options and their flags.

COMMANDS sub-commands can be created by using the conf.Command function
which returns a token set to designate the command as a target when
creating an option.

	cmd = conf.Command("doit", doitUsageString)

The cmd token is then used when defining an option, instructing the
package that the option is to be assigned to the command. The option will
appear in all of the commands for which tokens are provided, the tokes are
separated by the | character, indicating that all the delineated tokens
are to be used.

	conf.Option{
		Commands: cmd | cmd1 | cmd2
	}

OPTIONS contain the data required to create a flag, which is done when the
option is present within the active commands flagset, however options may
also be modified by other methods, such as the user programme.

The `Check:` field takes a function value that may be defined whilst
creating an option. This function has the signature `func(interface{})
(interface{}, error)` which can be used to either specify tests and
conditions for the options value or to change the value as it is passed.

The following is an example of the conf package in use:

#### Example use

```go
package main

import (
	"fmt"

	"github.com/8i8/conf"
)

var (
	def = conf.Setup(helpBase, helpDef)
	one = conf.Command("one", helpOne)
	two = conf.Command("two", helpTwo)
)

var opts = []conf.Option{
	{Name: "intie",
		Type:     conf.Int,
		Flag:     "n",
		Default:  12,
		Usage:    intie,
		Commands: def | one | two,
		Check: func(v interface{}) (interface{}, error) {
			i := *v.(*int)
			if i < 120 {
				return v, fmt.Errorf("-n must be greater than 120")
			}
			return v, nil
		},
	},
	{Name: "thing",
		Type:     conf.String,
		Flag:     "s",
		Default:  "Some thing",
		Usage:    thing,
		Commands: def | one | two,
		Check: func(v interface{}) (interface{}, error) {
			s := *v.(*string)
			if len(s) == 0 {
				return v, fmt.Errorf("What is this ... No text?")
			}
			return v, nil
		},
	},
	{Name: "none",
		Type:     conf.Int,
		Flag:     "i",
		Default:  16,
		Usage:    "the i is the none of all the ints",
		Commands: def | one | two,
	},
	{Name: "verbosity",
		Type:     conf.Int,
		Flag:     "v",
		Default:  0,
		Usage:    "The overall chattiness of it all",
		Commands: def | one | two,
	},
}

func main() {
	conf.Options(opts...)
	if err := conf.Parse(); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("The current running mode is %q\n", conf.GetCmd())
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

	Further details of the use of each mode can be found by running the
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
if another way of writing the messages might be better.`

var thing = `This is the default string thing, so as to best exemplify
the use of this package in its current state I thought it
best to write something very wordy here.`
```

GNU Lesser General Public License v3 (LGPL-3.0)
See lgpl-3.0.md for a full version of the licence.
