/*
Package conf helps you to organise and maintain a packages flags, their
options and their documention.

COMMANDS sub-commands can be created by using the conf.Command function
which returns a token set to designate the command as a target when
creating an option. The first command created is the default set which
requires no sub command be called to initiate it; The standard flagset,
for this set the first argument is the base header for the programs -h
flag, for all consiquent calls to Command the first argument is the
actual command token that is used on the command line to instigate the
sub command.

	c := conf.Config{}
	base = c.Command("MyHeader", MyUsageString)
	cmd1 = c.Command("greet", MyUsageString)

	app greet -s "Hello, World!"

The cmd token is then used when defining an option, instructing the
package that the option is to be assigned to the command. The option
will appear in all of the commands for which tokens are provided, the
tokes are separated by the | character, indicating that all the
delineated tokens are to be used.

	c.Options{
		{
			Type:     conf.String,
			Flag:     "s",
			Default:  "Some string",
			Usage:    stringUse,
			Commands: base | cmd1 | cmd2,
		},
	}

OPTIONS contain the data required to create a flag, which is done when
the option has been assigned to the active command.

The `Check:` field takes a function value that may be defined whilst
creating an option. This function has the signature `func(interface{})
(interface{}, error)` which can be used to either specify tests and
conditions upons the options value, or, to change the value as it is
passed.

	c.Options{
		{
			Type:     conf.String,
			Flag:     "s",
			Default:  "Some string",
			Usage:    stringUse,
			Commands: cmd | cmd1 | cmd2,
			Check: func(interface{})(interface{}, error) {
				str := *v.(*string)
				if len(str) == 0 {
					return v, fmt.Errorf("-s is empty)
				}
				return v, nil
			},
		},
	}

The following is an example of the conf package in use:

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
	cmd, err := c.Compose(opts...)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("The current running mode is %v\n", cmd)

	switch cmd {
	case def:
	case one:
	case two:
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

var stringUse = ` This is the default string dooing its thing, as to best
	exemplify the use of this package in its current state; I
	thought it best to write something particularly wordy here.`
*/
package conf
