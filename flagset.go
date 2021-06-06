package conf

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// createFlagSet defines the flagset for all options that have
// been specified within the current working set; All errors are
// accumulated in the Config.errs field and checked at the end of the
// function.
func setupFlagSet(c *Config) error {
	const fname = "createFlagSet"

	if c.set == nil {
		const event = "Config.set is nil"
		return fmt.Errorf("%s: %s", fname, event)
	}

	createFlagSet(c)

	if err := optionsToFlagSet(c); err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}

	flagSetUsage(c, os.Stdout)

	if err := parseFlagSet(c); err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}

	if verbose {
		log.Printf("%s: completed\n", fname)
	}

	return nil
}

func createFlagSet(c *Config) {
	const fname = "createFlagSet"

	if c.set == nil {
		panic("")
	}

	// Create our custom flagset.
	c.flagSet = flag.NewFlagSet(c.set.header, flag.ExitOnError)

	if verbose {
		log.Printf("%s: completed\n", fname)
	}
}

func flagSetUsage(c *Config, w io.Writer) {
	const fname = "flagSetUsage"

	// Define help or usage output function, overriding the default
	// flag package help function.
	if w == nil {
		w = os.Stderr
	}
	c.flagSet.SetOutput(w)
	c.flagSet.Usage = func() {
		io.WriteString(w, c.header)
		io.WriteString(w, c.set.usage)
		c.flagSet.VisitAll(flagUsage)
	}

	if verbose {
		log.Printf("%s: completed\n", fname)
	}
}

// TODO now the flag is not being set into the data in toFlagSet
func optionsToFlagSet(c *Config) error {
	const fname = "optionsToFlagSet"
	for _, o := range c.options {
		if c.set.flag&o.Commands > 0 {
			err := flagsToFlagSet(c, c.options[o.Flag])
			if err != nil {
				c.options[o.Flag].err = fmt.Errorf(
					"%s: %s: %w", fname, o.Flag, err)
				c.errs = fmt.Errorf("%s|%w", c.errs.Error(),
					c.options[o.Flag].err)
			}
			if verbose {
				log.Printf("%s: %s: option added\n",
					fname, c.options[o.Flag].Flag)
			}
		}
	}

	if err := checkError(c, errConfig); err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}

	if verbose {
		log.Printf("%s: completed\n", fname)
	}

	return nil
}

// flagsToFlagSet generates a flag within the given set for the current
// option.
func flagsToFlagSet(c *Config, o *Option) error {
	const fname = "toFlagSet"
	const def = "Default"
	const va = "Var"
	switch o.Type {
	case Int:
		i, ok := o.Default.(int)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		o.data = c.flagSet.Int(o.Flag, i, o.Usage)
	case IntVar:
		i, ok := o.Default.(int)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		v, ok := o.Var.(*int)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, va,
				errType)
		}
		c.flagSet.IntVar(v, o.Flag, i, o.Usage)
	case Int64:
		i, ok := o.Default.(int64)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		o.data = c.flagSet.Int64(o.Flag, i, o.Usage)
	case Int64Var:
		i, ok := o.Default.(int64)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		v, ok := o.Var.(*int64)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, va,
				errType)
		}
		c.flagSet.Int64Var(v, o.Flag, i, o.Usage)
	case Uint:
		i, ok := o.Default.(uint)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		o.data = c.flagSet.Uint(o.Flag, i, o.Usage)
	case UintVar:
		i, ok := o.Default.(uint)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		v, ok := o.Var.(*uint)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, va,
				errType)
		}
		c.flagSet.UintVar(v, o.Flag, i, o.Usage)
	case Uint64:
		i, ok := o.Default.(uint64)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		o.data = c.flagSet.Uint64(o.Flag, i, o.Usage)
	case Uint64Var:
		i, ok := o.Default.(uint64)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		v, ok := o.Var.(*uint64)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, va,
				errType)
		}
		c.flagSet.Uint64Var(v, o.Flag, i, o.Usage)
	case Float64:
		f, ok := o.Default.(float64)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		o.data = c.flagSet.Float64(o.Flag, f, o.Usage)
	case Float64Var:
		f, ok := o.Default.(float64)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		v, ok := o.Var.(*float64)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, va,
				errType)
		}
		c.flagSet.Float64Var(v, o.Flag, f, o.Usage)
	case String:
		s, ok := o.Default.(string)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		o.data = c.flagSet.String(o.Flag, s, o.Usage)
	case StringVar:
		s, ok := o.Default.(string)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		v, ok := o.Var.(*string)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, va,
				errType)
		}
		c.flagSet.StringVar(v, o.Flag, s, o.Usage)
	case Bool:
		b, ok := o.Default.(bool)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		o.data = c.flagSet.Bool(o.Flag, b, o.Usage)
	case BoolVar:
		b, ok := o.Default.(bool)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		v, ok := o.Var.(*bool)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, va,
				errType)
		}
		c.flagSet.BoolVar(v, o.Flag, b, o.Usage)
	case Duration:
		d, ok := o.Default.(time.Duration)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		o.data = c.flagSet.Duration(o.Flag, d, o.Usage)
	case DurationVar:
		d, ok := o.Default.(time.Duration)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		v, ok := o.Var.(*time.Duration)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, va,
				errType)
		}
		c.flagSet.DurationVar(v, o.Flag, d, o.Usage)
	case Var:
		if o.Value == nil {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errTypeNil)
		}
		c.flagSet.Var(o.Value, o.Flag, o.Usage)
	case Nil:
		return fmt.Errorf("%s: %q: %w", o.Type, def,
			errTypeNil)
	default:
		return fmt.Errorf("%s: internal error: (%q, %s) %w",
			fname, o.Flag, o.Type, errType)
	}
	return nil
}

// parseFlagSet runs the parse command on the configs flagset.
func parseFlagSet(c *Config) error {
	const fname = "parseFlagSet"

	// When tests are being run, we do not parse the flagset here.
	if test {
		return nil
	}
	// If not "*" then a command has been used and we need to offset
	// the args by one.
	var offset int
	if c.set.header != "*" {
		offset++
	}
	err := c.flagSet.Parse(os.Args[offset:])
	if err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}

	if verbose {
		log.Printf("%s: completed\n", fname)
	}

	return nil
}
