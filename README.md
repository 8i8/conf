### conf

#### Example use

```go
package main

import (
	"fmt"

	"github.com/8i8/conf"
)

var (
	c   = &conf.Config{}
	def = c.Command(helpBase, helpDef)
	one = c.Command("one", helpOne)
	two = c.Command("two", helpTwo)
)

var opts = []conf.Option{
	{
		Type:     conf.Int,
		Flag:     "n",
		Default:  12,
		Usage:    intUse,
		Commands: def | one | two,
		Check: func(v interface{}) (interface{}, error) {
			i := *v.(*int)
			if i != 12 {
				return v, fmt.Errorf("-n must be 12")
			}
			return v, nil
		},
	},
	{
		Type:     conf.String,
		Flag:     "s",
		Default:  "Some thing in the way she smiles",
		Usage:    stringUse,
		Commands: def | one | two,
		Check: func(v interface{}) (interface{}, error) {
			s := *v.(*string)
			if len(s) == 0 {
				return v, fmt.Errorf("What ... No text?")
			}
			return v, nil
		},
	},
	{
		Type:     conf.Int,
		Flag:     "i",
		Default:  16,
		Usage:    "the i is the none of all the ints",
		Commands: def | one | two,
	},
	{
		Type:     conf.Int,
		Flag:     "v",
		Default:  0,
		Usage:    "Oh! The overall chattiness of it all",
		Commands: def | one | two,
	},
}

func main() {
	_, err := c.Compose(opts...)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("The current running mode is %v\n", c.Running())
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

	one     one does all things that one should do.

	two     two, despite appearances is second to none, achieving
		things in no less than a second.

	Further details of the use of each mode can be found by running
	the following command.

	conf [mode] -help  or conf [mode] -h

FLAGS`

var helpOne = `MODE
	conf one
		Best used like this ...

FLAGS`

var helpTwo = `MODE
	conf two
		Advisable to use like this ...

FLAGS`

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
//  Flags
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var intUse = `This is the very default value in the most simplest of modes, to
	test if another way of writing the messages might could possibly
	be better.`

var stringUse = `This is the default string doodling its thing, as to best
	exemplify the use of this package in its current state; I
	thought it best to write something particularly wordy here.`
``
