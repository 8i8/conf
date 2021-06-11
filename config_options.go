package conf

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"time"
)

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Options
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

// ckFunc defines a function to check an options input data, the function
// is passed into the option when it is created by the user. They are run
// after all valid options have generated a flagset, which has in turn
// been parsed.
type ckFunc func(interface{}) (interface{}, error)

// Option contains the data required to create a flag.
type Option struct {
	// Flag contains the flag as it appears on the command line.
	Flag string
	// The data type of the option.
	Type
	// Value is a flag.Value interface, used when passing user defined
	// flag types into a flagset, for further information on using
	// this type, refer to the go flag package.
	Value flag.Value
	// Var is used to pass data by reference into the 'Var' group of
	// flag types.
	Var interface{}
	// Usage is the usage text that is displayed in help output when
	// the -help -h flags are used or a flag parsing error occurs.
	Usage string
	// data stores either the input user data or the default value
	// for the option, from where it is retrieved by making the
	// approprite method call for the type, `Config.Value[T]`.
	data interface{}
	// Default data, is the default data to be used in the case that
	// the flag is not called.
	Default interface{}
	// Commands is a bitmask, constructed from the bit flags
	// assigned to the Option on its creation, defining which
	// command sets the Option should appear within.
	Commands CMD
	// err stores any error that the option may have triggered
	// during its setup, allowing for errors to be returned at a
	// later time than when the options are being created.
	err error
	// Check is a user defined function that may be used to place
	// constraints upon, or alter, the data in the data field that
	// is provided by the user.
	Check ckFunc
}

// loadOptions loads all of the defined commands into the option map,
// running tests on each as they are loaded. Errors are accumulated into
// Config.errs which is checked upon leaving the function.
func loadOptions(c *Config, opts ...Option) error {
	const fname = "loadOptions"

	for i, opt := range opts {
		opts[i] = errCheckOption(c, opt)
		for j, cmd := range c.commands {
			// If the command is in an options set, then
			// save a pointer to the option in that command.
			if cmd.flag&opt.Commands != 0 {
				c.commands[j].options = append(
					c.commands[j].options, &opts[i])
			}
		}
	}

	if c.errs != nil {
		return fmt.Errorf("%s: %s: %w", fname, c.errs, errConfig)
	}

	if v(2) {
		log.Printf("%s: completed\n", fname)
	}
	return nil
}

// errCheckOption verifies user supplied data within an option including
// duplicate name and key values; All errors are accumulated and stored
// in the c.Err field.
func errCheckOption(c *Config, cmd Option) Option {
	const fname = "errCheckOption"
	var temperr error
	if c.errs == nil {
		// Empty output so as to go unnoticed when wrapped.
		temperr = errors.New("")
	} else {
		temperr = fmt.Errorf("%s: ", c.errs)
	}
	if err := checkFlag(c, &cmd); err != nil {
		cmd.err = fmt.Errorf("%s: %s: %w", fname, cmd.Flag, err)
		c.errs = fmt.Errorf("%s%w", temperr, cmd.err)
	}
	if err := checkDefault(c, cmd); err != nil {
		cmd.err = fmt.Errorf("%s: %s: %w", fname, cmd.Flag, err)
		c.errs = fmt.Errorf("%s%w", temperr, cmd.err)
	}
	if err := checkVar(c, cmd); err != nil {
		cmd.err = fmt.Errorf("%s: %s: %w", fname, cmd.Flag, err)
		c.errs = fmt.Errorf("%s%w", temperr, cmd.err)
	}
	if err := checkCmd(c, cmd); err != nil {
		cmd.err = fmt.Errorf("%s: %s: %w", fname, cmd.Flag, err)
		c.errs = fmt.Errorf("%s%w", temperr, cmd.err)
	}

	if v(3) && c.errs == nil {
		log.Printf("%s: %s: no errors\n", fname, cmd.Flag)
	}

	return cmd
}

// checkFlag checks that the flag field is not empty and that it is not
// a duplicate value within any one set.
func checkFlag(c *Config, o *Option) error {
	const fname = "checkFlag"
	if len(o.Flag) == 0 {
		return fmt.Errorf("%q: %w", fname, errNoValue)
	}
	for i, set := range c.commands {
		// If the option flag has already been registered on the
		// current subcommand, we return an error. Duplicate flags
		// on a differing sub-commands are OK.
		if set.flag&o.Commands != 0 {
			if set.seen.find(o.Flag) {
				return fmt.Errorf("%s: %w", fname, errDuplicate)
			}
			c.commands[i].seen = append(c.commands[i].seen, o.Flag)
		}
	}
	if v(3) {
		log.Printf("%s: completed\n", fname)
	}
	return nil
}

// checkDefault checks that the options default value has the correct
// type.
func checkDefault(c *Config, o Option) error {
	const fname = "checkDefault"
	switch o.Type {
	case Int, IntVar:
		if _, ok := o.Default.(int); !ok {
			return fmt.Errorf("%s: %s: %w",
				fname, o.Type, errType)
		}
	case Int64, Int64Var:
		if _, ok := o.Default.(int64); !ok {
			return fmt.Errorf("%s: %s: %w",
				fname, o.Type, errType)
		}
	case Uint, UintVar:
		if _, ok := o.Default.(uint); !ok {
			return fmt.Errorf("%s: %s: %w",
				fname, o.Type, errType)
		}
	case Uint64, Uint64Var:
		if _, ok := o.Default.(uint64); !ok {
			return fmt.Errorf("%s: %s: %w",
				fname, o.Type, errType)
		}
	case String, StringVar:
		if _, ok := o.Default.(string); !ok {
			return fmt.Errorf("%s: %s: %w",
				fname, o.Type, errType)
		}
	case Bool, BoolVar:
		if _, ok := o.Default.(bool); !ok {
			return fmt.Errorf("%s: %s: %w",
				fname, o.Type, errType)
		}
	case Float64, Float64Var:
		if _, ok := o.Default.(float64); !ok {
			return fmt.Errorf("%s: %s: %w",
				fname, o.Type, errType)
		}
	case Duration, DurationVar:
		if _, ok := o.Default.(time.Duration); !ok {
			return fmt.Errorf("%s: %s: %w",
				fname, o.Type, errType)
		}
	case Var:
		if _, ok := o.Value.(flag.Value); !ok {
			return fmt.Errorf("%s: %s: %w",
				fname, o.Type, errType)
		}
	case Nil:
		return fmt.Errorf("%s: %s: %w",
			fname, o.Type, errType)
	default:
		return fmt.Errorf("%s: %s: %w",
			fname, o.Type, errType)
	}
	if v(3) {
		log.Printf("%s: completed\n", fname)
	}
	return nil
}

// checkVar checks that the options var value has the correct type if it
// is required.
func checkVar(c *Config, o Option) error {
	const fname = "checkVar"
	switch o.Type {
	case Int:
	case IntVar:
		if _, ok := o.Var.(*int); !ok {
			return fmt.Errorf("%s: %s: %w",
				fname, o.Type, errType)
		}
	case Int64:
	case Int64Var:
		if _, ok := o.Var.(*int64); !ok {
			return fmt.Errorf("%s: %s: %w",
				fname, o.Type, errType)
		}
	case Uint:
	case UintVar:
		if _, ok := o.Var.(*uint); !ok {
			return fmt.Errorf("%s: %s: %w",
				fname, o.Type, errType)
		}
	case Uint64:
	case Uint64Var:
		if _, ok := o.Var.(*uint64); !ok {
			return fmt.Errorf("%s: %s: %w",
				fname, o.Type, errType)
		}
	case String:
	case StringVar:
		if _, ok := o.Var.(*string); !ok {
			return fmt.Errorf("%s: %s: %w",
				fname, o.Type, errType)
		}
	case Bool:
	case BoolVar:
		if _, ok := o.Var.(*bool); !ok {
			return fmt.Errorf("%s: %s: %w",
				fname, o.Type, errType)
		}
	case Float64:
	case Float64Var:
		if _, ok := o.Var.(*float64); !ok {
			return fmt.Errorf("%s: %s: %w",
				fname, o.Type, errType)
		}
	case Duration:
	case DurationVar:
		if _, ok := o.Var.(*time.Duration); !ok {
			return fmt.Errorf("%s: %s: %w",
				fname, o.Type, errType)
		}
	case Var:
		if _, ok := o.Value.(flag.Value); !ok {
			return fmt.Errorf("%s: %s: %w",
				fname, o.Type, errType)
		}
	case Nil:
		return fmt.Errorf("%s: %s: %w",
			fname, o.Type, errTypeNil)
	default:
		return fmt.Errorf("%s: %s: %w",
			fname, o.Type, errType)
	}
	if v(3) {
		log.Printf("%s: completed\n", fname)
	}
	return nil
}

// checkCmd verifies that the default command has been set and that any
// other commands are registered as valid commands within the current
// command set.
func checkCmd(c *Config, o Option) error {
	const fname = "checkCmd"
	if !isInSet(c, o.Commands) {
		return fmt.Errorf("%s: %w", fname, errSubCmd)
	}
	if v(3) {
		log.Printf("%s: completed\n", fname)
	}
	return nil
}
