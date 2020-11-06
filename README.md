### conf

package conf helps to organise and maintain package options and flags including
program operating modes that may be set from the command line.

MODES operating modes can be created by using the conf.Mode function, the
function returns a bit flag with the appropriate bit set to enable the mode
when creating an option.

```go
	newmode = conf.Mode("name", helpData)
```

The newmode flag is then used when defining an option in the Modes field, the
option will appear in all of the modes that are specified in this declaration.

```go
conf.Option{
	Modes: (newmode | mode1 | mode2)
}
```

OPTIONS contain the data required to create a flag when included within the
current flag set, however they may also be set from configuration files or
other methods, an option also contains a user definable function that may be
set to verify the data when it is set.

The following is an example of the conf package in use:


#### Example use

```go
package main

import (
	"fmt"

	"github.com/8i8/conf"
)

var (
	def = conf.Mode("def", helpDef)
	one = conf.Mode("one", helpOne)
	two = conf.Mode("two", helpTwo)
)

var opts = []conf.Option{
	{Name: "intie",
		Type:    conf.Int,
		Key:     "n",
		Default: 12,
		Help:    intie,
		Modes:   (def | one | two),
		Check: func(v interface{}) (interface{}, error) {
			i := *v.(*int)
			if i != 12 {
				return v, fmt.Errorf("-n must be 12")
			}
			return v, nil
		},
	},
	{Name: "thing",
		Type:    conf.String,
		Key:     "s",
		Default: "Some thing",
		Help:    thing,
		Modes:   (def | one | two),
		Check: func(v interface{}) (interface{}, error) {
			s := *v.(*string)
			if len(s) == 0 {
				return v, fmt.Errorf("What ... No text?")
			}
			return v, nil
		},
	},
	{Name: "none",
		Type:    conf.Int,
		Key:     "i",
		Default: 16,
		Help:    "the i is the none of all the ints",
		Modes:   (def | one | two),
	},
	{Name: "verbosity",
		Type:    conf.Int,
		Key:     "v",
		Default: 0,
		Help:    "The overall chattiness of it all",
		Modes:   (def | one | two),
	},
}

func main() {
	conf.Help(helpBase)
	conf.Options(opts...)
	if err := conf.Parse(); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("The current running mode is %q\n", conf.GetMode())
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
if another way of writing the messages might be better.`

var thing = `This is the default string thing, so as to best exemplify
the use of this package in its current state I thought it
best to write something very wordy here.`
```

GNU Lesser General Public License v3 (LGPL-3.0)
See lgpl-3.0.md for a full version of the licence.
