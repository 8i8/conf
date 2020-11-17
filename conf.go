package conf

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"
)

var (
	// Global package name function used in help output.
	pkg = "conf"
	// limit ensures that no more than 64 base modes are possible.
	limit = math.MaxInt64>>1 + 1
	// Config contains the program data for the default settings
	// struct used when not running on an exported struct.
	c Config
	// test is used by the test package to stop the flagset from being
	// parsed when the function Parse is called, the test code itself
	// when called calls flags.Parse, as such it is not necessary to
	// call it again. However the function contains other
	// functionality that we need to run.
	test bool
)

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Main package functions
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

// Setup sets the basis for the programs help output, the 'help header'
// and help 'Sub header', it also returns the bitfield for the modes field
// for use in the creation of Options, consequent calls to c.Mode will
// create and return further more flags.
func Setup(heading string, subheading string) (mode int) {
	mode = c.Setup(heading, subheading)
	return
}

// Mode creates a new mode, returning the bitfield required to set that
// mode.
func Mode(name, help string) (bitfield int) {
	bitfield = c.Mode(name, help)
	return
}

// GetMode return the current running modes name.
func GetMode() string {
	return c.mode.name
}

// Options initialises the programs options.
func Options(opts ...Option) error {
	return c.Options(opts...)
}

// Parse sets the running mode from the command line arguments and then
// parses the flagset.
func Parse() error {
	return c.Parse()
}

// ArgList returns the full command line argument list as a string as it
// was input.
func ArgList() string {
	return c.input
}

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Config
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

// Config contains all the program configuration Config and flags.
type Config struct {
	// input is the string of arguments that was entered on the
	// command line.
	input string
	// list is a list of the possible configuration submodes.
	list modelist
	// index holds the next value to use as a bitfield for the next
	// modelist mode.
	index int
	// Mode is the running mode of the program, this package
	// facilitates the generation of sub states.
	mode
	// The help output header for the program.
	help string
	// flagset is the programs flagset.
	flagSet *flag.FlagSet
	// names makes certain that no option name duplicates exist.
	names map[string]bool
	// options are where the data for each flag or option is stored,
	// this includes the value of the key its default value and help
	// string along with the actual data once the flag or config
	// option has been parsed, it also contains a function by which
	// the value that has been set may be checked.
	options map[string]*Option
	// Err stores any error triggered on parsing the flagset, it is
	// then return to the used when any flag is accesssed.
	Err []error
}

// Setup sets the basis for the programs help output, the 'help header'
// and help 'Sub header', it also returns the bitfield for the modes field
// for use in the creation of Options, consequent calls to c.Mode will
// create and return further more flags.
func (c *Config) Setup(heading string, subheading string) (mode int) {
	c.help = heading
	mode = c.Mode("default", subheading)
	return
}

// Mode creates a new mode, returning the bitfield required to set that
// mode.
func (c *Config) Mode(name, help string) (bitfield int) {
	// Make sure that we start at 1.
	if c.index == 0 {
		c.index++
	}
	if c.index >= limit {
		log.Fatal("index overflow, too many program modes")
	}
	m := mode{id: c.index, name: name, help: help}
	c.list = append(c.list, m)
	bitfield = c.index
	c.index = c.index << 1
	return
}

// GetMode return the current running modes name.
func (c Config) GetMode() string {
	return c.mode.name
}

// Options initialises the programs options.
func (c *Config) Options(opts ...Option) error {
	const fname = "Options"
	// Record the original input string.
	if err := c.saveArgs(); err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}
	// Generate flags and help data.
	if err := c.loadOptions(opts...); err != nil {
		return fmt.Errorf("%s: %s: %w", pkg, fname, err)
	}
	// TODO write a standard config file addition that records config
	// when in mode 'config' and that reads in any settings that have
	// been previously recorded or written.
	//c.loadConfig()
	return nil
}

// Parse sets the running mode from the command line arguments and then
// parses the flagset.
func (c *Config) Parse() error {
	const fname = "Parse"
	offset := 1
	if len(os.Args) > 1 && os.Args[1][0] != '-' {
		if c.list.is(os.Args[1]) {
			if err := c.load(os.Args[1]); err != nil {
				return fmt.Errorf("%s: %s: %w", pkg,
					fname, err)
			}
			offset++ // We have used another argument.
			goto parse
		}
		return fmt.Errorf("unknown mode: %q\n", os.Args[1])
	}
	// If no other mode has been specified, load the default.
	if err := c.load("default"); err != nil {
		return fmt.Errorf("%s: %s: %w", pkg, fname, err)
	}
parse:
	if !test {
		if err := c.parse(offset); err != nil {
			return fmt.Errorf("%s: %s: %w", pkg, fname, err)
		}
	}
	// Check all set verifications against parsed data.
	if err := c.runCheckFn(); err != nil {
		return fmt.Errorf("%s: %s: %w", pkg, fname, err)
	}
	return nil
}

// ArgList returns the full command line argument list as a string as it
// was input.
func (c Config) ArgList() string {
	return c.input
}

// saveArgs records the input arguments as a string for display output.
func (c *Config) saveArgs() error {
	const fname = "saveArgs"
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

// optionsToFsErrAccum defines flags within the flagset for all options that
// have been specified that are within the current working set. Errors are
// accumulated into the c.Err field.
func (c *Config) optionsToFsErrAccum() {
	const msg = "Option"
	for key, value := range c.options {
		if c.mode.id&value.Modes > 0 {
			err := c.options[key].toFlagSet(c.flagSet)
			if err != nil {
				c.Err = append(c.Err,
					fmt.Errorf("%s: %q: %w",
						msg, key, err))
			}
		}
	}
}

// loadOptions loads all of the given options into the option map,
// running tests on each as they are loaded, errors are accumulated into
// the c.Err field.
func (c *Config) loadOptions(opts ...Option) error {
	const fname = "loadOptions"
	c.options = make(map[string]*Option)
	if c.names == nil {
		c.names = make(map[string]bool)
	}
	// Make a duplicate verification map of the keys within the
	// operation mode struct, keys may not be duplicated within one
	// program operating mode, however the same key may be used again
	// for different options in the context of a different operating
	// modes.
	for i := range c.list {
		if c.list[i].keys == nil {
			c.list[i].keys = make(map[string]bool)
		}
	}
	for i, opt := range opts {
		c.checkOptionErrAccum(opt)
		c.options[opt.Name] = &opts[i]
	}
	return c.Error("checkOptionErrAccum", errOption)
}

// Error returns an error if any have occurred, concatenated into one
// string and wrapping the given error.
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

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Pre Parse Option Checks
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

var (
	errOption     = errors.New("option error")
	errType       = errors.New("type error")
	errTypeNil    = errors.New("the type is not defined")
	errTypeUnkown = errors.New("unknown type, please file a bug report")
	errNoValue    = errors.New("value required")
	errDuplicate  = errors.New("duplicate value")
	errNotInSet   = errors.New(
		"mode bitfield is not entirely within the current set")
)

// checkOptionErrAccum verifies the user supplied data contained within an
// option including duplicate name and key values, errors are accumulated
// and stored in the c.Err field.
func (c *Config) checkOptionErrAccum(o Option) {
	const fname = "Option"
	if err := c.checkName(o); err != nil {
		c.Err = append(c.Err, fmt.Errorf("%s: %w", fname, err))
	}
	if err := c.checkKey(o); err != nil {
		c.Err = append(c.Err, fmt.Errorf("%s: %w", fname, err))
	}
	if err := c.checkDefault(o); err != nil {
		c.Err = append(c.Err, fmt.Errorf("%s: %w", fname, err))
	}
	if err := c.checkMode(o); err != nil {
		c.Err = append(c.Err, fmt.Errorf("%s: %w", fname, err))
	}
}

// chekcName checks that the name is not empty and that it is not a
// duplicate value.
func (c *Config) checkName(o Option) error {
	const msg = "check name"
	if len(o.Name) == 0 {
		return fmt.Errorf("%s: %q: %s", msg, o.Name, errNoValue)
	}
	if c.names[o.Name] {
		return fmt.Errorf("%s: %q: %s", msg, o.Name, errDuplicate)
	}
	c.names[o.Name] = true
	return nil
}

// chekcName checks that the key is not empty and that it is not a
// duplicate value.
func (c *Config) checkKey(o Option) error {
	const msg = "check key"
	if len(o.Key) == 0 {
		return fmt.Errorf("%s: %q: %s", msg, o.Name, errNoValue)
	}
	for i := range c.list {
		// If the options key has already been registered for the
		// current mode, return an error; Using the same key in
		// different modes is fine.
		if c.list[i].id&o.Modes > 0 {
			if c.list[i].keys[o.Key] {
				return fmt.Errorf("%s: %q: %q: %s",
					msg, o.Name, o.Key, errDuplicate)
			}
			c.list[i].keys[o.Key] = true
		}
	}
	return nil
}

// checkDefault checks that the options default value has the correct
// type.
func (c *Config) checkDefault(o Option) error {
	const msg = "check Default"
	// We only check for the ecistance of the Default value on flags
	// that are being passed by reference, errors are returned when
	// passing flags by value.
	switch o.Type {
	case Int, IntVar:
		if _, ok := o.Default.(int); !ok {
			return fmt.Errorf("%s: %q: %+v: %s",
				msg, o.Name, o.Type, errType)
		}
	case Int64, Int64Var:
		if _, ok := o.Default.(int64); !ok {
			return fmt.Errorf("%s: %q: %+v: %s",
				msg, o.Name, o.Type, errType)
		}
	case Uint, UintVar:
		if _, ok := o.Default.(uint); !ok {
			return fmt.Errorf("%s: %q: %+v: %s",
				msg, o.Name, o.Type, errType)
		}
	case Uint64, Uint64Var:
		if _, ok := o.Default.(uint64); !ok {
			return fmt.Errorf("%s: %q: %+v: %s",
				msg, o.Name, o.Type, errType)
		}
	case String, StringVar:
		if _, ok := o.Default.(string); !ok {
			return fmt.Errorf("%s: %q: %+v: %s",
				msg, o.Name, o.Type, errType)
		}
	case Bool, BoolVar:
		if _, ok := o.Default.(bool); !ok {
			return fmt.Errorf("%s: %q: %+v: %s",
				msg, o.Name, o.Type, errType)
		}
	case Float64, Float64Var:
		if _, ok := o.Default.(float64); !ok {
			return fmt.Errorf("%s: %q: %+v: %s",
				msg, o.Name, o.Type, errType)
		}
	case Duration, DurationVar:
		if _, ok := o.Default.(time.Duration); !ok {
			return fmt.Errorf("%s: %q: %+v: %s",
				msg, o.Name, o.Type, errType)
		}
	case Var:
		// Both Var and Default are interfaces as such we can not
		// test this here, so we let Var pass with no check.
	case Nil:
		return fmt.Errorf("%s: %q: %+v: %s",
			msg, o.Name, o.Type, errTypeNil)
	default:
		return fmt.Errorf("%s: %s: %q: %+v: %s",
			pkg, msg, o.Name, o.Type, errTypeUnkown)
	}
	return nil
}

// checkMode verifies that an option has modes set and that those modes
// are contained within the the current mode set.
func (c *Config) checkMode(o Option) error {
	const msg = "check mode"
	if !c.flagIs(o.Modes) {
		return fmt.Errorf("%s: %q: %s", msg, o.Name, errNotInSet)
	}
	return nil
}

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Post Parse Option Checks
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

// runCheckFn runs all user given check functions for the option set,
// called after having parsed all option data.
func (c *Config) runCheckFn() error {
	const msg = "Option: Check"
	if c.options == nil {
		return fmt.Errorf("%s: %s: option set empty", pkg, msg)
	}
	var err error
	for _, o := range c.options {
		if o.Check != nil {
			if o.data, err = o.Check(o.data); err != nil {
				fmt.Printf("%s, %s, %s", pkg, msg, err)
				os.Exit(0)
			}
		}
	}
	return nil
}

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Mode
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

// mode contains the required data to create a programs operating mode and
// its flag, the sub heading of a program run mode for its specific
// operating flags.
type mode struct {
	id   int
	name string
	// The help output for the particular mode displayed when -h is
	// used.
	help string
	// keys makes certain that no key duplicates exist.
	keys map[string]bool
}
type modelist []mode

// is returns true if the given mode exists and false if it does not.
func (l modelist) is(mode string) bool {
	for _, m := range l {
		if strings.Compare(mode, m.name) == 0 {
			return true
		}
	}
	return false
}

// flagIs returns true if a bitfield is entirely contained within the
// configurations full set of modes, returning false if it is not.
func (c Config) flagIs(bitfield int) bool {
	if bitfield == 0 {
		return false
	}
	if bitfield == (c.index-1)&bitfield {
		return true
	}
	return false
}

// setMode defines the programs current running mode.
func (c *Config) setMode(mode string) error {
	const fname = "setMode"
	if mode == "default" {
		c.mode = c.list[0]
	}
	for _, m := range c.list {
		if strings.Compare(mode, m.name) == 0 {
			c.mode = m
			return nil
		}
	}
	return fmt.Errorf("%s: mode not found", fname)
}

// load sets the programs operating mode and loads all required options
// along with their help data into the relevant flagset.
func (c *Config) load(mode string) error {
	const fname = "load"
	if err := c.setMode(mode); err != nil {
		return fmt.Errorf("%s: %q: %w", fname, mode, err)
	}
	c.flagSet = flag.NewFlagSet(c.mode.name, flag.ExitOnError)
	// Errors created in this function are for users and need to be
	// dealt with differently.
	c.optionsToFsErrAccum()
	c.flagSet.Usage = func() {
		fmt.Println(c.help)
		fmt.Println(c.mode.help)
		c.flagSet.VisitAll(flagHelpMsg)
	}
	return c.Error("optionsToFsErrAccum", errOption)
}

// parse parses the flagset using the given offest to counter the
// arguments already used.
func (c *Config) parse(offset int) error {
	const fname = "parse"
	err := c.flagSet.Parse(os.Args[offset:])
	if err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}
	return nil
}

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Option
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

// ckFunc defines a function to check an options input data, the function
// is passed into the option when it is created by the user and run when
// the input flags and any configuration options are parsed.
type ckFunc func(interface{}) (interface{}, error)

// Option contains all of the data required for setting a default flag and
// receiving subsequent option settings.
type Option struct {
	// The name of the option, also used as a key in the options map.
	Name string
	// Type of flag.
	Type
	// Value is used when passing user defined flag types into a
	// flagset
	Value flag.Value
	// Var is used to pass values by reference into the Var flag
	// types.
	Var interface{}
	// Key defines the flag required to modify the option.
	Key string
	// Help string.
	Help string
	// data.
	data interface{}
	// Default data.
	Default interface{}
	// Can the value be overridden?
	set bool
	// Which program modes should the flag be included in?
	Modes int
	// Function to verify option data.
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
		o.data = fls.Int(o.Key, i, o.Help)
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
		fls.IntVar(v, o.Key, i, o.Help)
	case Int64:
		i, ok := o.Default.(int64)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		o.data = fls.Int64(o.Key, i, o.Help)
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
		fls.Int64Var(v, o.Key, i, o.Help)
	case Uint:
		i, ok := o.Default.(uint)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		o.data = fls.Uint(o.Key, i, o.Help)
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
		fls.UintVar(v, o.Key, i, o.Help)
	case Uint64:
		i, ok := o.Default.(uint64)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		o.data = fls.Uint64(o.Key, i, o.Help)
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
		fls.Uint64Var(v, o.Key, i, o.Help)
	case String:
		s, ok := o.Default.(string)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		o.data = fls.String(o.Key, s, o.Help)
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
		fls.StringVar(v, o.Key, s, o.Help)
	case Bool:
		b, ok := o.Default.(bool)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		o.data = fls.Bool(o.Key, b, o.Help)
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
		fls.BoolVar(v, o.Key, b, o.Help)
	case Float64:
		f, ok := o.Default.(float64)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		o.data = fls.Float64(o.Key, f, o.Help)
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
		fls.Float64Var(v, o.Key, f, o.Help)
	case Duration:
		d, ok := o.Default.(time.Duration)
		if !ok {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errType)
		}
		o.data = fls.Duration(o.Key, d, o.Help)
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
		fls.DurationVar(v, o.Key, d, o.Help)
	case Var:
		if o.Value == nil {
			return fmt.Errorf("%s: %q: %w", o.Type, def,
				errNoValue)
		}
		fls.Var(o.Value, o.Key, o.Help)
	default:
		return fmt.Errorf("%s: %s: internal error: (%q, %s) %s",
			pkg, fname, o.Name, o.Type, errTypeUnkown)
	}
	return nil
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

// flagHelpMsg writes the help message for each individual flag.
func flagHelpMsg(f *flag.Flag) {
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

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Values and Types
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

type Type uint64

const (
	Nil Type = iota
	Int
	IntVar
	Int64
	Int64Var
	Uint
	UintVar
	Uint64
	Uint64Var
	Float64
	Float64Var
	String
	StringVar
	Bool
	BoolVar
	Duration
	DurationVar
	Var
)

var (
	errNoData = errors.New("no data")
	errNoKey  = errors.New("key not found")
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
		return "interface{}"
	default:
		return "error: unknown type"
	}
}

// Value returns the content of an option flag its type and also a boolean
// that expresses whether or not the flag has been found.
func (c Config) Value(key string) (interface{}, Type, error) {
	const fname = "Value"
	o, ok := c.options[key]
	if !ok {
		return nil, Nil, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoKey)
	}
	if o.data == nil {
		return nil, Nil, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoData)
	}
	return o.data, o.Type, nil
}

// Value returns the content of an option flag its type and also a boolean
// that expresses whether or not the flag has been found.
func Value(key string) (interface{}, Type, error) {
	return c.Value(key)
}

// ValueInt returns the value of an int option.
func (c Config) ValueInt(key string) (int, error) {
	const fname = "ValueInt"
	o, ok := c.options[key]
	if !ok {
		return 0, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoKey)
	}
	if o.data == nil {
		return 0, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoData)
	}
	return *o.data.(*int), nil
}

// ValueInt returns the value of an int option.
func ValueInt(key string) (int, error) {
	return c.ValueInt(key)
}

// ValueInt64 returns the value of an int64 option.
func (c Config) ValueInt64(key string) (int64, error) {
	const fname = "ValueInt64"
	o, ok := c.options[key]
	if !ok {
		return 0, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoKey)
	}
	if o.data == nil {
		return 0, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoData)
	}
	return *o.data.(*int64), nil
}

// ValueInt64 returns the value of an int64 option.
func ValueInt64(key string) (int64, error) {
	return c.ValueInt64(key)
}

// ValueUint returns the value of an uint option.
func (c Config) ValueUint(key string) (uint, error) {
	const fname = "ValueUint"
	o, ok := c.options[key]
	if !ok {
		return 0, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoKey)
	}
	if o.data == nil {
		return 0, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoData)
	}
	return *o.data.(*uint), nil
}

// ValueUint returns the value of an uint option.
func ValueUint(key string) (uint, error) {
	return c.ValueUint(key)
}

// ValueUint64 returns the value of an uint64 option.
func (c Config) ValueUint64(key string) (uint64, error) {
	const fname = "ValueUint64"
	o, ok := c.options[key]
	if !ok {
		return 0, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoKey)
	}
	if o.data == nil {
		return 0, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoData)
	}
	return *o.data.(*uint64), nil
}

// ValueUint64 returns the value of an uint64 option.
func ValueUint64(key string) (uint64, error) {
	return c.ValueUint64(key)
}

// ValueFloat64 returns the value a float64 options.
func (c Config) ValueFloat64(key string) (float64, error) {
	const fname = "ValueFloat64"
	o, ok := c.options[key]
	if !ok {
		return 0, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoKey)
	}
	if o.data == nil {
		return 0, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoData)
	}
	return *o.data.(*float64), nil
}

// ValueFloat64 returns the value a float64 options.
func ValueFloat64(key string) (float64, error) {
	return c.ValueFloat64(key)
}

// ValueString returns the value of a string options.
func (c Config) ValueString(key string) (string, error) {
	const fname = "ValueString"
	o, ok := c.options[key]
	if !ok {
		return "", fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoKey)
	}
	if o.data == nil {
		return "", fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoData)
	}
	return *o.data.(*string), nil
}

// ValueString returns the value of a string options.
func ValueString(key string) (string, error) {
	return c.ValueString(key)
}

// ValueBool returns the value of a boolean options.
func (c Config) ValueBool(key string) (bool, error) {
	const fname = "ValueBool"
	o, ok := c.options[key]
	if !ok {
		return false, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoKey)
	}
	if o.data == nil {
		return false, fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoData)
	}
	return *o.data.(*bool), nil
}

// ValueBool returns the value of a boolean options.
func ValueBool(key string) (bool, error) {
	return c.ValueBool(key)
}

// ValueDuration returs the value of a time.Duration option.
func (c Config) ValueDuration(key string) (time.Duration, error) {
	const fname = "ValueDuration"
	o, ok := c.options[key]
	if !ok {
		return time.Duration(0), fmt.Errorf("%s: %s: %q: %w",
			pkg, fname, key, errNoKey)
	}
	if o.data == nil {
		return 0, fmt.Errorf("%s: %s: %w",
			pkg, fname, errNoData)
	}
	return *o.data.(*time.Duration), nil
}

// ValueDuration returs the value of a time.Duration option.
func ValueDuration(key string) (time.Duration, error) {
	return c.ValueDuration(key)
}
