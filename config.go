package conf

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	// limit ensures that no more than 64 command sets are possible.
	limit = CMD(math.MaxInt64>>1 + 1)
	// test is used by the test package to stop the flagset from
	// being parsed when the function Parse is called.
	test bool
)

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Config
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

// Config contains an array of commands the user can call when starting
// the command line application, the selected command then loads its
// corresponding flagset and operating mode, parsing any following
// arguments as flags and their parameters.
//
// The 'header' is best used as a formatted `string`, so that what you
// see is what you get, and example of which might typically be:
//
// `NAME
//	app
//
// SYNOPSIS
//	app | [cmd] | -[flag] | -[flag] [opt] | -[flag] ['opt,opt,opt']
//
// EXAMPLE
//	app write -n 36 -s "Hello, World!"`
//
type Config struct {

	// Command line help flags output header.
	header string

	// All user commands created at start up, essentially bit masks
	// their header strings and nomenclature.
	commands []command
	// The next available bit for use as a command bit mask
	position CMD
	// The current running command set.
	set *command

	// The flagset that is composed at startup according to the
	// predefined command line commands and their options.
	flagSet *flag.FlagSet

	// errs stores any errors triggered on either generating or
	// parsing the flagset, returned to the user when either Options
	// or Parse are run, else when a flag is accessed by the program
	// runtime.
	errs error
}

// Compose initialises the programs options.
func (c *Config) Compose(opts ...Option) error {
	const fname = "Config.Compose"

	if err := configPreconditions(c, opts...); err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}
	if err := ascertainCmdSet(c); err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}
	if err := loadOptions(c, opts...); err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}
	if err := setupFlagSet(c); err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}
	if err := runUserCheckFuncs(c); err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}

	if v(1) {
		log.Printf("%s: completed\n", fname)
	}

	// TODO write a standard config file addition that records to a
	// config file when in mode 'config' and that reads in any
	// settings that have been previously recorded.
	//c.loadConfig()
	return nil
}

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Post Parse Option Checks
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

// runUserCheckFuncs runs all user given ckFunc functions in the command
// set if data has been provided.
func runUserCheckFuncs(c *Config) error {
	const fname = "runUserCheckFuncs"
	for _, o := range c.set.options {
		if o.Check == nil || o.data == nil {
			continue
		}
		var err error
		opt := c.set.options.find(o.Flag)
		opt.data, err = o.Check(o.data)
		if err != nil {
			err := fmt.Errorf("%s, %w",
				err, ErrCheck)
			opt.err = err
			if c.errs != nil {
				c.errs = fmt.Errorf("%s|%w", c.errs, err)
			} else {
				c.errs = err
			}
		}
	}

	if err := checkError(c); err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}

	if v(2) {
		log.Printf("%s: completed\n", fname)
	}

	return nil
}

// Error will return an error if any have occurred from the stored error
// in Config.errs. Where all errors have been concatenated into one.
func checkError(c *Config, err ...error) error {
	if c.errs == nil {
		return nil
	}
	if len(err) > 0 {
		return fmt.Errorf("%s: %w", c.errs.Error(), err[0])
	}
	return c.errs
}

// Is returns the current running sub-commands name and state.
func (c Config) Running() CMD {
	if c.set == nil {
		panic("Config.set is nil")
	}
	return c.set.flag
}

// Is returns true if the flag is set withing the current running mode.
func (c Config) Is(flag CMD) bool {
	if c.set == nil {
		panic("Config.set is nil")
	}
	return c.set.flag&flag > 0
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

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Commands
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

// options is an array of option pointers that are used to maintain
// which commands contain which options.
type options []*Option

func (o options) find(flag string) *Option {
	for i, opt := range o {
		if strings.Compare(opt.Flag, flag) == 0 {
			return o[i]
		}
	}
	return nil
}

type flags []string

func (f flags) find(flag string) bool {
	for _, f := range f {
		if strings.Compare(f, flag) == 0 {
			return true
		}
	}
	return false
}

// command contains the required data to create a program sub-command and
// its flags.
type command struct {
	// flag is the set bit that represents the command.
	flag CMD
	// The options header.
	header string
	// The usage output for the command displayed when -h is called or
	// an error raised on parsing.
	usage string
	// seen makes certain that no flag duplicates exist.
	seen flags
	// options is a slice that contains pointers to all of the
	// options that have been assigned to this command set.
	options options
}

// CMD is a bitmask that defines which command a FlagSet is to be
// applied to.
type CMD int

func (c CMD) String() string {
	return strconv.Itoa(int(c))
}

// isInSet returns true if a command token exists within the
// configured set of commands, false if it does not.
func isInSet(c *Config, bitfield CMD) bool {
	if bitfield == 0 {
		return false
	}
	fullset := (c.position << 1) - 1
	if bitfield == (fullset)&bitfield {
		return true
	}
	return false
}

// setCommand defines the programs current running state.
func setCommand(c *Config, name string) error {
	const fname = "setCommand"
	for i, m := range c.commands {
		if strings.Compare(name, m.header) == 0 {
			c.set = &c.commands[i]
			return nil
		}
	}
	return fmt.Errorf("%s: %w", fname, errNotFound)
}

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Command sequences
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
	// Type is the data type of the option.
	Type
	// Value is a flag.Value interface, used when passing user defined
	// flag types into a flagset.
	Value flag.Value
	// Var is used to pass values by reference into the 'Var' group of
	// flag types.
	Var interface{}
	// Usage string defines the usage text that is displayed in help
	// output.
	Usage string
	// data store the input user data of the option when required.
	data interface{}
	// Default data, is the default data used in the case that the
	// flag is not called.
	Default interface{}
	// Commands is a set of flags that represent which commands the
	// option should be included within.
	Commands CMD
	// err stores any errors that the option may have triggered whilst
	// being set up and parsed.
	err error
	// Check is a user defined function that may be used to either
	// constrain or alter the data in the data field.
	Check ckFunc
}

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Usage display output
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

// setUsageFn is set as flag.FlagSet.Usage, generating the usage output.
func setUsageFn(w io.Writer, c *Config) {
	c.flagSet.SetOutput(w)
	if w == nil {
		w = os.Stderr
	}
	c.flagSet.Usage = func() {
		io.WriteString(w, c.header)
		io.WriteString(w, c.set.usage)
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
	errNoData   = errors.New("no data")
	errNoKey    = errors.New("key not found")
	errStored   = errors.New("stored")
	errCommands = errors.New("commands not set")
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
	if c.set == nil {
		return nil, Nil, fmt.Errorf("%s: %q: %w",
			fname, key, errCommands)
	}
	o := c.set.options.find(key)
	if o == nil {
		return nil, Nil, fmt.Errorf("%s: %q: %w",
			fname, key, errNoKey)
	}
	if o.err != nil {
		return o.data, o.Type, fmt.Errorf("%s: %w", fname, o.err)
	}
	if o.data == nil {
		return nil, Nil, fmt.Errorf("%s: %q: %w",
			fname, key, errNoData)
	}
	return o.data, o.Type, nil
}

// ValueInt returns the value of an int option, else an error if one has
// been raised during the options creation.
func (c Config) ValueInt(key string) (int, error) {
	const fname = "ValueInt"
	if c.set == nil {
		return 0, fmt.Errorf("%s: %q: %w",
			fname, key, errCommands)
	}
	o := c.set.options.find(key)
	if o == nil {
		return 0, fmt.Errorf("%s: %q: %w",
			fname, key, errNoKey)
	}
	if o.err != nil {
		return 0, fmt.Errorf("%s: %w", fname, o.err)
	}
	if o.data == nil {
		return 0, fmt.Errorf("%s: %q: %w",
			fname, key, errNoData)
	}
	return *o.data.(*int), nil
}

// ValueInt64 returns the value of an int64 option, else an error if one
// has been raised during the options creation.
func (c Config) ValueInt64(key string) (int64, error) {
	const fname = "ValueInt64"
	if c.set == nil {
		return 0, fmt.Errorf("%s: %q: %w",
			fname, key, errCommands)
	}
	o := c.set.options.find(key)
	if o == nil {
		return 0, fmt.Errorf("%s: %q: %w",
			fname, key, errNoKey)
	}
	if o.err != nil {
		return 0, fmt.Errorf("%s: %w", fname, o.err)
	}
	if o.data == nil {
		return 0, fmt.Errorf("%s: %q: %w",
			fname, key, errNoData)
	}
	return *o.data.(*int64), nil
}

// ValueUint returns the value of an uint option, else an error if one has
// been raised during the options creation.
func (c Config) ValueUint(key string) (uint, error) {
	const fname = "ValueUint"
	if c.set == nil {
		return 0, fmt.Errorf("%s: %q: %w",
			fname, key, errCommands)
	}
	o := c.set.options.find(key)
	if o == nil {
		return 0, fmt.Errorf("%s: %q: %w",
			fname, key, errNoKey)
	}
	if o.err != nil {
		return 0, fmt.Errorf("%s: %w", fname, o.err)
	}
	if o.data == nil {
		return 0, fmt.Errorf("%s: %q: %w",
			fname, key, errNoData)
	}
	return *o.data.(*uint), nil
}

// ValueUint64 returns the value of an uint64 option, else an error if one
// has been raised during the options creation.
func (c Config) ValueUint64(key string) (uint64, error) {
	const fname = "ValueUint64"
	if c.set == nil {
		return 0, fmt.Errorf("%s: %q: %w",
			fname, key, errCommands)
	}
	o := c.set.options.find(key)
	if o == nil {
		return 0, fmt.Errorf("%s: %q: %w",
			fname, key, errNoKey)
	}
	if o.err != nil {
		return 0, fmt.Errorf("%s: %w", fname, o.err)
	}
	if o.data == nil {
		return 0, fmt.Errorf("%s: %q: %w",
			fname, key, errNoData)
	}
	return *o.data.(*uint64), nil
}

// ValueFloat64 returns the value of an float64 option, else an error if
// one has been raised during the options creation.
func (c Config) ValueFloat64(key string) (float64, error) {
	const fname = "ValueFloat64"
	if c.set == nil {
		return 0, fmt.Errorf("%s: %q: %w",
			fname, key, errCommands)
	}
	o := c.set.options.find(key)
	if o == nil {
		return 0, fmt.Errorf("%s: %q: %w",
			fname, key, errNoKey)
	}
	if o.err != nil {
		return 0, fmt.Errorf("%s: %w", fname, o.err)
	}
	if o.data == nil {
		return 0, fmt.Errorf("%s: %q: %w",
			fname, key, errNoData)
	}
	return *o.data.(*float64), nil
}

// ValueString returns the value of an string option, else an error if one
// has been raised during the options creation.
func (c Config) ValueString(key string) (string, error) {
	const fname = "ValueString"
	if c.set == nil {
		return "", fmt.Errorf("%s: %q: %w",
			fname, key, errCommands)
	}
	o := c.set.options.find(key)
	if o == nil {
		return "", fmt.Errorf("%s: %q: %w",
			fname, key, errNoKey)
	}
	if o.err != nil {
		return "", fmt.Errorf("%s: %w", fname, o.err)
	}
	if o.data == nil {
		return "", fmt.Errorf("%s: %q: %w",
			fname, key, errNoData)
	}
	return *o.data.(*string), nil
}

// ValueBool returns the value of an string option, else an error if one
// has been raised during the options creation.
func (c Config) ValueBool(key string) (bool, error) {
	const fname = "ValueBool"
	if c.set == nil {
		return false, fmt.Errorf("%s: %q: %w",
			fname, key, errCommands)
	}
	o := c.set.options.find(key)
	if o == nil {
		return false, fmt.Errorf("%s: %q: %w",
			fname, key, errNoKey)
	}
	if o.err != nil {
		return false, fmt.Errorf("%s: %w", fname, o.err)
	}
	if o.data == nil {
		return false, fmt.Errorf("%s: %q: %w",
			fname, key, errNoData)
	}
	return *o.data.(*bool), nil
}

// ValueDuration returns the value of an time.Duration option, else an
// error if one has been raised during the options creation.
func (c Config) ValueDuration(key string) (time.Duration, error) {
	const fname = "ValueDuration"
	if c.set == nil {
		return 0, fmt.Errorf("%s: %q: %w",
			fname, key, errCommands)
	}
	o := c.set.options.find(key)
	if o == nil {
		return 0, fmt.Errorf("%s: %q: %w",
			fname, key, errNoKey)
	}
	if o.err != nil {
		return 0, fmt.Errorf("%s: %w", fname, o.err)
	}
	if o.data == nil {
		return 0, fmt.Errorf("%s: %w", fname, errNoData)
	}
	return *o.data.(*time.Duration), nil
}
