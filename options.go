package conf

import (
	"errors"
	"flag"
	"fmt"
	"time"
)

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Load options
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

// loadOptions loads all of the defined commands into the option map,
// running tests on each as they are loaded. Errors are accumulated into
// Config.errs which is checked upon leaving the function.
func loadOptions(c *Config, opts ...Option) error {
	const fname = "loadOptions"

	for i, opt := range opts {
		opts[i] = errCheckOption(c, opt)
		c.options[opt.Flag] = &opts[i]
	}
	if err := checkError(c, errConfig); err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}
	if verbose {
		fmt.Printf("%s: completed\n", fname)
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
		temperr = errors.New("")
	} else {
		temperr = fmt.Errorf("%s: ", c.errs)
	}
	if err := checkName(c, cmd); err != nil {
		cmd.err = fmt.Errorf("%s: %s: %w", fname, cmd.Flag, err)
		c.errs = fmt.Errorf("%s%w", temperr, cmd.err)
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

	if verbose && c.errs == nil {
		fmt.Printf("%s: %s: no errors\n", fname, cmd.Flag)
	}

	return cmd
}

// checkName checks that the name is not empty and that it is not a
// duplicate value.
// TODO now this function is filling the seen[o.Flag] map
func checkName(c *Config, o Option) error {
	const fname = "checkName"
	if len(o.Flag) == 0 {
		return fmt.Errorf("%s: %w", fname, errNoValue)
	}
	for i, set := range c.commands {
		if set.seen[o.Flag] > 2 {
			return fmt.Errorf("%s: %w", fname, errDuplicate)
		}
		c.commands[i].seen[o.Flag]++
	}
	c.seen[o.Flag] = true
	return nil
}

// checkFlag checks that the flag field is not empty and that it is not
// a duplicate value.
// TODO now this function is filling the c.commands[n].seen map
func checkFlag(c *Config, o *Option) error {
	const fname = "checkFlag"
	if len(o.Flag) == 0 {
		return fmt.Errorf("%q: %w", fname, errNoValue)
	}
	for i := range c.commands {
		// If the option flag has already been registered on the
		// current subcommand, we return an error. Duplicate flags
		// on a differing sub-commands is OK.
		if c.commands[i].flag&o.Commands > 0 {
			if c.commands[i].seen[o.Flag] > 2 {
				o.err = fmt.Errorf("%s: %s: %w",
					fname, o.Flag, errDuplicate)
				c.errs = fmt.Errorf("%s: %w", o.err, errDuplicate)
			}
			c.commands[i].seen[o.Flag]++
		}
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
	return nil
}

// checkCmd verifies that the default command has been set and that any
// other commands are registered as valid commands within the current
// command set.
// TODO now this set verification looks very dubious.
func checkCmd(c *Config, o Option) error {
	const fname = "checkCmd"
	if !isInSet(c, o.Commands) {
		return fmt.Errorf("%s: %w", fname, errSubCmd)
	}
	return nil
}
