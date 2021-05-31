package conf

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"
)

var (
	// The package name, used in help output.
	pkg = "conf"
	// limit ensures that no more than 64 command sets are possible.
	limit = CMD(math.MaxInt64>>1 + 1)
	// test is used by the test package to stop the flagset from
	// being parsed when the function Parse is called.
	test bool
)

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Config
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

// Config is the main package struct, all of the API functionality is
// centered about it.
type Config struct {
	// input is the raw string of arguments that were entered on the
	// command line.
	input string
	// cmds stores the list of available sub-commands.
	cmds cmdlist
	// index holds the next value that is to be use as a bitfield
	// for the next cmdlist command.
	index CMD
	// subcmd is the current running command mode.
	subcmd
	// The help output header for the program.
	help string
	// options are where the data for each option is stored, this
	// includes the flag with its default value and usage string
	// along with any data collected once the flag or config option
	// has been parsed, and the program run, it also contains a
	// function with which any value that is set may be verified.
	options map[string]*Option
	// opnames makes certain that no option name duplicates exist.
	opnames map[string]bool
	// flagset is the programs constructed flagset, the result of
	// running the Options command.
	flagSet *flag.FlagSet
	// Err stores any errors triggered on either generating or
	// parsing the flagset, returned to the user when either Options
	// or Parse are run, else when a flag is accessed from the
	// calling program.
	Err []error
}

// defaultSet defines the foundation for the programs flags and help,
// setting the heading and an initial default flagset.
func (c *Config) defaultSet(heading string, usage string) (bitfield CMD) {
	c.index++
	c.help = heading
	bitfield = c.FlagSet("default", usage)
	return
}

// FlagSet creates a new sub-command, returning a bitfield token which
// is used to assign an option and its flags to use within that mode.
func (c *Config) FlagSet(name, usage string) (bitfield CMD) {
	if c.index == 0 {
		bitfield = c.defaultSet(name, usage)
		return
	}
	if c.index >= limit {
		err := errors.New("index overflow, too many program modes")
		c.Err = append(c.Err, err)
	}
	m := subcmd{id: c.index, name: name, usage: usage}
	c.cmds = append(c.cmds, m)
	bitfield = c.index
	c.index = c.index << 1
	return
}

// GetCmd return the current running sub-commands name.
func (c Config) GetCmd() string {
	return c.subcmd.name
}

// Options initialises the programs options.
func (c *Config) Options(opts ...Option) error {
	const fname = "Options"
	// Record the original input string.
	if err := c.saveArgs(); err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}
	// Generate flags and usage data.
	if err := c.loadOptions(opts...); err != nil {
		return fmt.Errorf("%s: %s: %w", pkg, fname, err)
	}
	// TODO write a standard config file addition that records to a
	// config file when in mode 'config' and that reads in any
	// settings that have been previously recorded.
	//c.loadConfig()
	return nil
}

// Parse sets the current running mode from the command line arguments
// and then calls parse on them, so as to generate its required flagset.
func (c *Config) Parse() error {
	const fname = "Parse"
	offset := 1
	if len(os.Args) > 1 && os.Args[1][0] != '-' {
		if err := c.loadCmd(os.Args[1]); err != nil {
			return fmt.Errorf("%s: %s: %w", pkg,
				fname, err)
		}
		offset++ // We have used another argument.
		return parse(c, offset, fname)
	}
	// If no sub-command has been specified, load the default cmd.
	if err := c.loadCmd("default"); err != nil {
		return fmt.Errorf("%s: %s: %w", pkg, fname, err)
	}
	return parse(c, 1, fname)
}

// parse runs the parse command on the configs flagset.
func parse(c *Config, offset int, fname string) error {
	if !test {
		err := c.flagSet.Parse(os.Args[offset:])
		if err != nil {
			return fmt.Errorf("%s: %w", fname, err)
		}
	}
	// Run all user specified conditions against the parsed data.
	if err := c.runCheckFn(); err != nil {
		return fmt.Errorf("%s: %s: %w", pkg, fname, err)
	}
	return nil
}

// ArgString returns a command line arguments string, as input.
func (c Config) ArgString() string {
	return c.input
}

// saveArgs records the literal input arguments as a string.
func (c *Config) saveArgs() error {
	const fname = "saveArgs"
	if len(os.Args) == 0 {
		return fmt.Errorf("%s: %w", fname, errConfig)
	}
	var str strings.Builder
	_, err := str.WriteString(os.Args[0])
	if err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}
	for _, arg := range os.Args[1:] {
		if err := str.WriteByte(' '); err != nil {
			return fmt.Errorf("%s: %w", fname, err)
		}
		_, err = str.WriteString(arg)
		if err != nil {
			return fmt.Errorf("%s: %w", fname, err)
		}
	}
	c.input = str.String()
	return nil
}

// optionsToFsErrAccum defines the flagset for all options that have
// been specified within the current working set; All errors are
// accumulated in the Config.Err field.
func (c *Config) optionsToFsErrAccum() {
	const msg = "Option: flagSet"
	if len(c.options) == 0 {
		c.Err = append(c.Err,
			fmt.Errorf("%s: %w", msg, errNoData))
		return
	}
	for name, o := range c.options {
		if c.subcmd.id&o.Commands > 0 {
			if c.subcmd.flags[o.Flag] > 1 {
				continue
			}
			err := c.options[o.Name].toFlagSet(c.flagSet)
			if err != nil {
				c.options[name].Err = fmt.Errorf(
					"%s: %s: %w", msg, name, err)
				c.Err = append(
					c.Err, c.options[name].Err)
			}
		}
	}
}

// loadOptions loads all of the given options into the option map,
// running tests on each as they are loaded.  On leaving the function
// the Config.Err field is checked and any errors reported, it is then
// emptied.
func (c *Config) loadOptions(opts ...Option) error {
	const fname = "loadOptions"
	if c.options == nil {
		c.options = make(map[string]*Option)
	}
	if c.opnames == nil {
		c.opnames = make(map[string]bool)
	}
	// Make a duplicate verification map of the flags for each
	// sub-command, flags may not be duplicated in a sub-command,
	// however the same flag name can be used again for different
	// sub-commands for different operations, if that flag has not
	// been reused within the context of the same command.
	for i := range c.cmds {
		if c.cmds[i].flags == nil {
			c.cmds[i].flags = make(map[string]int)
		}
	}
	for i, opt := range opts {
		opts[i] = c.checkOptionErrAccum(opt)
		c.options[opt.Name] = &opts[i]
	}
	return c.Error("checkOptionErrAccum", errConfig)
}

// Error returns an error if any errors have occurred and been recorded
// within the slice of errors in Config.Err. Any errors are concatenated
// into one string and wrap the given error.
func (c *Config) Error(msg string, err error) error {
	if len(c.Err) == 0 {
		return nil
	}
	str := strings.Builder{}
	str.WriteString(c.Err[0].Error())
	if len(c.Err) > 1 {
		for _, err := range c.Err[1:] {
			str.WriteString(" | ")
			str.WriteString(err.Error())
		}
	}
	c.Err = c.Err[:0]
	return fmt.Errorf("%s: %s: %w", msg, str.String(), err)
}

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Pre Parse Option Checks
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

var (
	// ErrCheck is the error that is returned when a user defined
	// check function fails.
	ErrCheck     = errors.New("user defined error")
	errConfig    = errors.New("configuration error")
	errType      = errors.New("type error")
	errTypeNil   = errors.New("the type is not defined")
	errNoValue   = errors.New("value required")
	errNotFound  = errors.New("not found")
	errNotValid  = errors.New("not valid")
	errDuplicate = errors.New("duplicate value")
	errSubCmd    = errors.New("sub-command error")
)

// checkOptionErrAccum verifies user supplied data within an option
// including duplicate name and key values; All errors are accumulated
// and stored in the c.Err field.
func (c *Config) checkOptionErrAccum(o Option) Option {
	const msg = "Option: check"
	if err := c.checkName(o); err != nil {
		o.Err = fmt.Errorf("%s: %s: %w", msg, o.Name, err)
		c.Err = append(c.Err, o.Err)
		return o
	}
	o, err := c.checkFlag(o)
	if err != nil {
		o.Err = fmt.Errorf("%s: %s: %w", msg, o.Name, err)
		c.Err = append(c.Err, o.Err)
		return o
	}
	if err := c.checkDefault(o); err != nil {
		o.Err = fmt.Errorf("%s: %s: %w", msg, o.Name, err)
		c.Err = append(c.Err, o.Err)
	}
	if err := c.checkVar(o); err != nil {
		o.Err = fmt.Errorf("%s: %s: %w", msg, o.Name, err)
		c.Err = append(c.Err, o.Err)
	}
	if err := c.checkCmd(o); err != nil {
		o.Err = fmt.Errorf("%s: %s: %w", msg, o.Name, err)
		c.Err = append(c.Err, o.Err)
	}
	return o
}

// chekcName checks that the name is not empty and that it is not a
// duplicate value.
func (c *Config) checkName(o Option) error {
	const msg = "Option.Name"
	if len(o.Name) == 0 {
		return fmt.Errorf("%s: %w", msg, errNoValue)
	}
	if c.opnames[o.Name] {
		return fmt.Errorf("%s: %w", msg, errDuplicate)
	}
	c.opnames[o.Name] = true
	return nil
}

// checkFlag checks that the flag field is not empty and that it is not
// a duplicate value.
func (c *Config) checkFlag(o Option) (Option, error) {
	const msg = "Option.Flag"
	if len(o.Flag) == 0 {
		return o, fmt.Errorf("%q: %w", msg, errNoValue)
	}
	for i := range c.cmds {
		// If the option flag has already been registered on the
		// current subcommand, we return an error. Duplicate flags
		// on a differing sub-commands is OK.
		if c.cmds[i].id&o.Commands > 0 {
			if c.cmds[i].flags[o.Flag] > 0 {
				c.cmds[i].flags[o.Flag]++
				err := fmt.Errorf("%s: %s: %w",
					msg, o.Flag, errDuplicate)
				c.Err = append(c.Err, err)
				o.Err = err
			}
			c.cmds[i].flags[o.Flag]++
		}
	}
	return o, nil
}

// checkDefault checks that the options default value has the correct
// type.
func (c *Config) checkDefault(o Option) error {
	const msg = "Option.Default"
	switch o.Type {
	case Int, IntVar:
		if _, ok := o.Default.(int); !ok {
			return fmt.Errorf("%s: %s: %w",
				msg, o.Type, errType)
		}
	case Int64, Int64Var:
		if _, ok := o.Default.(int64); !ok {
			return fmt.Errorf("%s: %s: %w",
				msg, o.Type, errType)
		}
	case Uint, UintVar:
		if _, ok := o.Default.(uint); !ok {
			return fmt.Errorf("%s: %s: %w",
				msg, o.Type, errType)
		}
	case Uint64, Uint64Var:
		if _, ok := o.Default.(uint64); !ok {
			return fmt.Errorf("%s: %s: %w",
				msg, o.Type, errType)
		}
	case String, StringVar:
		if _, ok := o.Default.(string); !ok {
			return fmt.Errorf("%s: %s: %w",
				msg, o.Type, errType)
		}
	case Bool, BoolVar:
		if _, ok := o.Default.(bool); !ok {
			return fmt.Errorf("%s: %s: %w",
				msg, o.Type, errType)
		}
	case Float64, Float64Var:
		if _, ok := o.Default.(float64); !ok {
			return fmt.Errorf("%s: %s: %w",
				msg, o.Type, errType)
		}
	case Duration, DurationVar:
		if _, ok := o.Default.(time.Duration); !ok {
			return fmt.Errorf("%s: %s: %w",
				msg, o.Type, errType)
		}
	case Var:
		if _, ok := o.Value.(flag.Value); !ok {
			return fmt.Errorf("%s: %s: %w",
				msg, o.Type, errType)
		}
	case Nil:
		return fmt.Errorf("%s: %s: %w",
			msg, o.Type, errType)
	default:
		return fmt.Errorf("%s: %s: %s: %w",
			pkg, msg, o.Type, errType)
	}
	return nil
}

// checkVar checks that the options var value has the correct type if it
// is required.
func (c *Config) checkVar(o Option) error {
	const msg = "Var"
	switch o.Type {
	case Int:
	case IntVar:
		if _, ok := o.Var.(*int); !ok {
			return fmt.Errorf("%s: %s: %w",
				msg, o.Type, errType)
		}
	case Int64:
	case Int64Var:
		if _, ok := o.Var.(*int64); !ok {
			return fmt.Errorf("%s: %s: %w",
				msg, o.Type, errType)
		}
	case Uint:
	case UintVar:
		if _, ok := o.Var.(*uint); !ok {
			return fmt.Errorf("%s: %s: %w",
				msg, o.Type, errType)
		}
	case Uint64:
	case Uint64Var:
		if _, ok := o.Var.(*uint64); !ok {
			return fmt.Errorf("%s: %s: %w",
				msg, o.Type, errType)
		}
	case String:
	case StringVar:
		if _, ok := o.Var.(*string); !ok {
			return fmt.Errorf("%s: %s: %w",
				msg, o.Type, errType)
		}
	case Bool:
	case BoolVar:
		if _, ok := o.Var.(*bool); !ok {
			return fmt.Errorf("%s: %s: %w",
				msg, o.Type, errType)
		}
	case Float64:
	case Float64Var:
		if _, ok := o.Var.(*float64); !ok {
			return fmt.Errorf("%s: %s: %w",
				msg, o.Type, errType)
		}
	case Duration:
	case DurationVar:
		if _, ok := o.Var.(*time.Duration); !ok {
			return fmt.Errorf("%s: %s: %w",
				msg, o.Type, errType)
		}
	case Var:
		if _, ok := o.Value.(flag.Value); !ok {
			return fmt.Errorf("%s: %s: %w",
				msg, o.Type, errType)
		}
	case Nil:
		return fmt.Errorf("%s: %s: %w",
			msg, o.Type, errTypeNil)
	default:
		return fmt.Errorf("%s: %s: %s: %w",
			pkg, msg, o.Type, errType)
	}
	return nil
}

// checkCmd verifies that the default command has been set and that any
// other commands are registered as valid commands within the current
// command set.
func (c *Config) checkCmd(o Option) error {
	const msg = "Commands"
	if !c.cmdTokenIs(o.Commands) {
		return fmt.Errorf("%s: %w", msg, errSubCmd)
	}
	return nil
}

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Post Parse Option Checks
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

// runCheckFn runs all user given ckFunc functions within the current
// command set, called after having first parsed all valid options.
func (c *Config) runCheckFn() error {
	const msg = "Check"
	for key, o := range c.options {
		if o.Check == nil || o.data == nil {
			continue
		}
		var err error
		c.options[key].data, err = o.Check(o.data)
		if err != nil {
			c.options[key].Err = fmt.Errorf("%s, %w",
				err, ErrCheck)
			c.Err = append(c.Err, fmt.Errorf("%s, %w",
				msg, err))
		}
	}
	return c.Error("runCheckFn", ErrCheck)
}

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Sub-Commands
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

// subcmd contains the required data to create a program sub-command and
// its flags.
type subcmd struct {
	// id is the bitfield of the command.
	id CMD
	// The options name.
	name string
	// The usage output for the command displayed when -h is called or
	// an error raised on parsing.
	usage string
	// flags makes certain that no flag duplicates exist.
	flags map[string]int
}

// The cmdlist type is used to store all of the commands inside of a
// Config struct.
type cmdlist []subcmd

// CMD is the bitfield that defines which commands an option is to be
// applied to.
type CMD int

// cmdTokenIs returns true if a sub-command token exists within the
// configuration set of sub-commands, false if it does not.
func (c Config) cmdTokenIs(bitfield CMD) bool {
	if bitfield == 0 {
		return false
	}
	if bitfield == (c.index-1)&bitfield {
		return true
	}
	return false
}

// setCmd defines the programs current running state.
func (c *Config) setCmd(name string) error {
	const fname = "setCmd"
	if name == "default" {
		c.subcmd = c.cmds[0]
		return nil
	}
	for _, m := range c.cmds {
		if strings.Compare(name, m.name) == 0 {
			c.subcmd = m
			return nil
		}
	}
	return fmt.Errorf("%s: %w", fname, errNotFound)
}

// loadCmd sets the programs operating mode and loads all required options
// along with their usage data into the relevant flagset.
func (c *Config) loadCmd(cmd string) error {
	const fname = "loadCmd"
	if err := c.setCmd(cmd); err != nil {
		return fmt.Errorf("%s: %q: %w", fname, cmd, err)
	}
	c.flagSet = flag.NewFlagSet(c.subcmd.name, flag.ExitOnError)
	c.optionsToFsErrAccum()
	c.flagSet.Usage = c.setUsageFn(os.Stdout)
	return c.Error("optionsToFsErrAccum", errConfig)
}

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Option
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

// ckFunc defines a function to check an options input data, the function
// is passed into the option when it is created by the user. They are run
// after all valid options have generated a flagset, which has in turn
// been parsed.
type ckFunc func(interface{}) (interface{}, error)

// Option contains all of the data required for creating a command option
// and defining its flag.
type Option struct {
	// The name of the option, also used as a key in the options map.
	Name string
	// Type defined the data type of the option.
	Type
	// Value is a flag.Value interface, used when passing user defined
	// flag types into a flagset.
	Value flag.Value
	// Var is used to pass values by reference into the 'Var' group of
	// flag types.
	Var interface{}
	// Flag defines the string or character required to either set or
	// modify the option.
	Flag string
	// Usage string defines the usage text that is displayed in help
	// output.
	Usage string
	// data is where the data of the option is stored.
	data interface{}
	// Default data, is the default data used in the case that the
	// flag is not called.
	Default interface{}
	// Commands contains the set of program commands for which the
	// option should be included.
	Commands CMD
	// Err stores any errors that the option may have triggered whilst
	// being set up and parsed.
	Err error
	// Check is a user defined function that may be used to either
	// constrain or alter the data in the data field.
	Check ckFunc
}

// toFlagSet generates a flag within the given flagset for the current
// option.
func (o *Option) toFlagSet(fls *flag.FlagSet) error {
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
		o.data = fls.Int(o.Flag, i, o.Usage)
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
		fls.IntVar(v, o.Flag, i, o.Usage)
	case Int64:
		i, ok := o.Default.(int64)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		o.data = fls.Int64(o.Flag, i, o.Usage)
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
		fls.Int64Var(v, o.Flag, i, o.Usage)
	case Uint:
		i, ok := o.Default.(uint)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		o.data = fls.Uint(o.Flag, i, o.Usage)
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
		fls.UintVar(v, o.Flag, i, o.Usage)
	case Uint64:
		i, ok := o.Default.(uint64)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		o.data = fls.Uint64(o.Flag, i, o.Usage)
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
		fls.Uint64Var(v, o.Flag, i, o.Usage)
	case Float64:
		f, ok := o.Default.(float64)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		o.data = fls.Float64(o.Flag, f, o.Usage)
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
		fls.Float64Var(v, o.Flag, f, o.Usage)
	case String:
		s, ok := o.Default.(string)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		o.data = fls.String(o.Flag, s, o.Usage)
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
		fls.StringVar(v, o.Flag, s, o.Usage)
	case Bool:
		b, ok := o.Default.(bool)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		o.data = fls.Bool(o.Flag, b, o.Usage)
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
		fls.BoolVar(v, o.Flag, b, o.Usage)
	case Duration:
		d, ok := o.Default.(time.Duration)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		o.data = fls.Duration(o.Flag, d, o.Usage)
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
		fls.DurationVar(v, o.Flag, d, o.Usage)
	case Var:
		if o.Value == nil {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errTypeNil)
		}
		fls.Var(o.Value, o.Flag, o.Usage)
	case Nil:
		return fmt.Errorf("%s: %q: %w", o.Type, def,
			errTypeNil)
	default:
		return fmt.Errorf("%s: %s: internal error: (%q, %s) %w",
			pkg, fname, o.Name, o.Type, errType)
	}
	return nil
}

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Usage display output
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

// setUsageFn is set as flag.FlagSet.Usage, generating the usage output.
func (c Config) setUsageFn(w io.Writer) func() {
	c.flagSet.SetOutput(w)
	if w == nil {
		w = os.Stderr
	}
	return func() {
		io.WriteString(w, c.help)
		io.WriteString(w, c.subcmd.usage)
		c.flagSet.VisitAll(flagUsage)
	}
}

// space sets a space after the flag name in the help output, aligning the
// flags description correctly for output.
func space(b []byte, l int) string {
	w := len(b)
	l = w - l
	for i := 0; i < l; i++ {
		w--
		b[w] = ' '
	}
	return string(b[w:])
}

// flagUsage writes the usage message for each individual flag.
func flagUsage(f *flag.Flag) {
	l := len(f.Name) + 1 // for the '-' char.
	var buf [8]byte
	sp := space(buf[:], l)
	s := fmt.Sprintf("        -%s%s", f.Name, sp)
	_, usage := flag.UnquoteUsage(f)
	if l > 6 {
		s += "\n        \t"
	}
	s += strings.ReplaceAll(usage, "\n", "\n            \t")
	fmt.Fprint(os.Stdout, s, "\n\n")
}

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Values and Types
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

// Type are the supported types that can be used as config flags.
type Type uint64

const (
	// Nil is not a type.
	Nil Type = iota
	// Int are the native int type.
	Int
	// IntVar are the native &int type.
	IntVar
	// Int64 are the native int64 type.
	Int64
	// Int64Var are the native &int64 type.
	Int64Var
	// Uint are the native uint type.
	Uint
	// UintVar are the native &uint type.
	UintVar
	// Uint64 are the native &uint64 type.
	Uint64
	// Uint64Var are the native &uint64 type.
	Uint64Var
	// Float64 are the native float64 type.
	Float64
	// Float64Var are the native &float64 type.
	Float64Var
	// String are the native string type.
	String
	// StringVar are the native &string type.
	StringVar
	// Bool are the native bool type.
	Bool
	// BoolVar are the native &bool type.
	BoolVar
	// Duration are the samaya.Duration type.
	Duration
	// DurationVar are the &samaya.Duration type.
	DurationVar
	// Var are the interface{} type.
	Var
	// Default are an unknown type.
	Default
)

var (
	errNoData = errors.New("no data")
	errNoKey  = errors.New("key not found")
	errStored = errors.New("stored")
)

func (t Type) String() string {
	switch t {
	case Nil:
		return "nil"
	case Int:
		return "int"
	case IntVar:
		return "*int"
	case Int64:
		return "int64"
	case Int64Var:
		return "*int64"
	case Uint:
		return "uint"
	case UintVar:
		return "*uint"
	case Uint64:
		return "uint64"
	case Uint64Var:
		return "*uint64"
	case Float64:
		return "float64"
	case Float64Var:
		return "*float64"
	case String:
		return "string"
	case StringVar:
		return "*string"
	case Bool:
		return "bool"
	case BoolVar:
		return "*bool"
	case Duration:
		return "time.Duration"
	case DurationVar:
		return "*time.Duration"
	case Var:
		return "flag.Value"
	default:
		return "error: unknown type"
	}
}

// Value returns the content of an option along with its type, else an
// error, if one has been raised during the options creation.
func (c Config) Value(key string) (interface{}, Type, error) {
	const fname = "Value"
	o, ok := c.options[key]
	if !ok {
		return nil, Nil, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoKey)
	}
	if o.Err != nil {
		return o.data, o.Type, fmt.Errorf("%s: %w", fname, o.Err)
	}
	if o.data == nil {
		return nil, Nil, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoData)
	}
	return o.data, o.Type, nil
}

// ValueInt returns the value of an int option, else an error if one has
// been raised during the options creation.
func (c Config) ValueInt(key string) (int, error) {
	const fname = "ValueInt"
	o, ok := c.options[key]
	if !ok {
		return 0, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoKey)
	}
	if o.Err != nil {
		return 0, fmt.Errorf("%s: %w", fname, o.Err)
	}
	if o.data == nil {
		return 0, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoData)
	}
	return *o.data.(*int), nil
}

// ValueInt64 returns the value of an int64 option, else an error if one
// has been raised during the options creation.
func (c Config) ValueInt64(key string) (int64, error) {
	const fname = "ValueInt64"
	o, ok := c.options[key]
	if !ok {
		return 0, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoKey)
	}
	if o.Err != nil {
		return 0, fmt.Errorf("%s: %w", fname, o.Err)
	}
	if o.data == nil {
		return 0, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoData)
	}
	return *o.data.(*int64), nil
}

// ValueUint returns the value of an uint option, else an error if one has
// been raised during the options creation.
func (c Config) ValueUint(key string) (uint, error) {
	const fname = "ValueUint"
	o, ok := c.options[key]
	if !ok {
		return 0, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoKey)
	}
	if o.Err != nil {
		return 0, fmt.Errorf("%s: %w", fname, o.Err)
	}
	if o.data == nil {
		return 0, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoData)
	}
	return *o.data.(*uint), nil
}

// ValueUint64 returns the value of an uint64 option, else an error if one
// has been raised during the options creation.
func (c Config) ValueUint64(key string) (uint64, error) {
	const fname = "ValueUint64"
	o, ok := c.options[key]
	if !ok {
		return 0, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoKey)
	}
	if o.Err != nil {
		return 0, fmt.Errorf("%s: %w", fname, o.Err)
	}
	if o.data == nil {
		return 0, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoData)
	}
	return *o.data.(*uint64), nil
}

// ValueFloat64 returns the value of an float64 option, else an error if
// one has been raised during the options creation.
func (c Config) ValueFloat64(key string) (float64, error) {
	const fname = "ValueFloat64"
	o, ok := c.options[key]
	if !ok {
		return 0, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoKey)
	}
	if o.Err != nil {
		return 0, fmt.Errorf("%s: %w", fname, o.Err)
	}
	if o.data == nil {
		return 0, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoData)
	}
	return *o.data.(*float64), nil
}

// ValueString returns the value of an string option, else an error if one
// has been raised during the options creation.
func (c Config) ValueString(key string) (string, error) {
	const fname = "ValueString"
	o, ok := c.options[key]
	if !ok {
		return "", fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoKey)
	}
	if o.Err != nil {
		return "", fmt.Errorf("%s: %w", fname, o.Err)
	}
	if o.data == nil {
		return "", fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoData)
	}
	return *o.data.(*string), nil
}

// ValueBool returns the value of an string option, else an error if one
// has been raised during the options creation.
func (c Config) ValueBool(key string) (bool, error) {
	const fname = "ValueBool"
	o, ok := c.options[key]
	if !ok {
		return false, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoKey)
	}
	if o.Err != nil {
		return false, fmt.Errorf("%s: %w", fname, o.Err)
	}
	if o.data == nil {
		return false, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoData)
	}
	return *o.data.(*bool), nil
}

// ValueDuration returns the value of an time.Duration option, else an
// error if one has been raised during the options creation.
func (c Config) ValueDuration(key string) (time.Duration, error) {
	const fname = "ValueDuration"
	o, ok := c.options[key]
	if !ok {
		return time.Duration(0), fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoKey)
	}
	if o.Err != nil {
		return 0, fmt.Errorf("%s: %w", fname, o.Err)
	}
	if o.data == nil {
		return 0, fmt.Errorf("%s: %s: %w",
			pkg, fname, errNoData)
	}
	return *o.data.(*time.Duration), nil
}
