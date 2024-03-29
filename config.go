package conf

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
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

// Config contains an array of commands the user can call when starting
// the command line application, the selected command then loads its
// corresponding flagset and operating mode, parsing any following
// arguments as flags and their parameters.
//
// The 'header' is best used as a formatted `string`, so that what you
// see is what you get, an example of which might typically be:
//
// `NAME
//
//	app
//
// SYNOPSIS
//
//	app | [cmd] | -[flag] | -[flag] [opt] | -[flag] ['opt,opt,opt']
//
// EXAMPLE
//
//	app write -n 36 -s "Hello, World!"`
type Config struct {

	// Command line help flags output header.
	header string

	// All user commands created at start up, essentially bit masks
	// their header strings and nomenclature.
	commands []command
	// The next available bit for use as a command bit mask
	position CMD

	// All possible flags.
	all options
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
func (c *Config) Compose(opts ...Option) (set CMD, err error) {
	const fname = "Config.Compose"

	if err = configPreconditions(c, opts...); err != nil {
		err = fmt.Errorf("%s: %w", fname, err)
		return
	}
	if set, err = ascertainCmdSet(c); err != nil {
		err = fmt.Errorf("%s: %w", fname, err)
		return
	}
	if err = loadOptions(c, opts...); err != nil {
		err = fmt.Errorf("%s: %w", fname, err)
		return
	}
	if err = setupFlagSet(c); err != nil {
		err = fmt.Errorf("%s: %w", fname, err)
		return
	}
	if err = runUserCheckFuncs(c); err != nil {
		err = fmt.Errorf("%s: %w", fname, err)
		return
	}

	if v1() {
		log.Printf("%s: completed\n", fname)
	}

	// TODO write a standard config file addition that records to a
	// config file when in mode 'config' and that reads in any
	// settings that have been previously recorded.
	//c.loadConfig()
	return
}

// Is returns true of the option flag is in any set.
func (c *Config) Is(flag string) bool {
	for _, o := range c.all {
		if strings.Compare(o.Flag, flag) == 0 {
			return true
		}
	}
	return false
}

// Cmd returns the current running commands bitflag as a token, directly
// comparable with the tokens returned when registering a Command() with
// the conf package.
func (c Config) Cmd() CMD {
	if c.set == nil {
		panic("Config.set is nil, have you run Config.Compose?")
	}
	return c.set.flag
}

// IsSet returns true if the flag is active in the current running mode.
func (c Config) IsSet(flag CMD) bool {
	if c.set == nil {
		panic("Config.set is nil, have you run Config.Compose?")
	}
	return c.set.flag&flag != 0
}

// NArg returns the number of arguments remaining after the flags have
// been processed.
func (c Config) NArg() int {
	if c.flagSet == nil {
		panic("Config.flagSet is nil, have you run Config.Compose?")
	}
	return c.flagSet.NArg()
}

// Args returns the non-flag arguments.
func (c Config) Args() []string {
	if c.flagSet == nil {
		panic("Config.flagSet is nil, have you run Config.Compose?")
	}
	return c.flagSet.Args()
}

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Errors
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

var (
	// ErrCheck is the error that is returned when a user defined
	// check function fails.
	ErrCheck           = errors.New("user defined error")
	errConfig          = errors.New("configuration error")
	errType            = errors.New("type error")
	errTypeNil         = errors.New("the type is not defined")
	errNoValue         = errors.New("value required")
	errNotFound        = errors.New("not found")
	errNotValid        = errors.New("not valid")
	errDuplicate       = errors.New("duplicate value")
	errSubCmd          = errors.New("sub-command error")
	errNoData          = errors.New("no data")
	errNoFlag          = errors.New("flag not found")
	ErrNotInCurrentSet = errors.New("flag not available in this set")
	errCommands        = errors.New("commands not set")
	ErrUnknownCMD      = errors.New("unknown CMD token")
)

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

func want(want, got any) error {
	return fmt.Errorf("want %T got %T: %w", want, got, errType)
}

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
	fail := func(err error) (interface{}, Type, error) {
		return nil, Nil, fmt.Errorf("%s: %w", fname, err)
	}
	if c.set == nil {
		return fail(errCommands)
	}
	o := c.set.options.find(key)
	if o == nil && c.Is(key) {
		return fail(ErrNotInCurrentSet)
	}
	if o == nil {
		return fail(errNoFlag)
	}
	if o.err != nil {
		return fail(o.err)
	}
	if o.data == nil {
		return fail(errNoData)
	}
	return o.data, o.Type, nil
}

// ValueInt returns the value of an int option, else an error if one has
// been raised during the options creation.
func (c Config) ValueInt(key string) (int, error) {
	const fname = "ValueInt"
	var out int
	fail := func(err error) (int, error) {
		return out, fmt.Errorf("%s: %w", fname, err)
	}
	if c.set == nil {
		return fail(errCommands)
	}
	o := c.set.options.find(key)
	if o == nil && c.Is(key) {
		return fail(ErrNotInCurrentSet)
	}
	if o == nil {
		return fail(errNoFlag)
	}
	if o.err != nil {
		return fail(o.err)
	}
	switch t := o.data.(type) {
	case nil:
		return fail(errNoData)
	case *int:
		out = *t
	default:
		return fail(want(out, t))
	}
	return out, nil
}

// ValueInt64 returns the value of an int64 option, else an error if one
// has been raised during the options creation.
func (c Config) ValueInt64(key string) (int64, error) {
	const fname = "ValueInt64"
	var out int64
	fail := func(err error) (int64, error) {
		return out, fmt.Errorf("%s: %w", fname, err)
	}
	if c.set == nil {
		return fail(errCommands)
	}
	o := c.set.options.find(key)
	if o == nil && c.Is(key) {
		return fail(ErrNotInCurrentSet)
	}
	if o == nil {
		return fail(errNoFlag)
	}
	if o.err != nil {
		return fail(o.err)
	}
	switch t := o.data.(type) {
	case nil:
		return fail(errNoData)
	case *int64:
		out = *t
	default:
		return fail(want(out, t))
	}
	return out, nil
}

// ValueUint returns the value of an uint option, else an error if one has
// been raised during the options creation.
func (c Config) ValueUint(key string) (uint, error) {
	const fname = "ValueUint"
	var out uint
	fail := func(err error) (uint, error) {
		return out, fmt.Errorf("%s: %w", fname, err)
	}
	if c.set == nil {
		return fail(errCommands)
	}
	o := c.set.options.find(key)
	if o == nil && c.Is(key) {
		return fail(ErrNotInCurrentSet)
	}
	if o == nil {
		return fail(errNoFlag)
	}
	if o.err != nil {
		return fail(o.err)
	}
	switch t := o.data.(type) {
	case nil:
		return fail(errNoData)
	case *uint:
		out = *t
	default:
		return fail(want(out, t))
	}
	return out, nil
}

// ValueUint64 returns the value of an uint64 option, else an error if one
// has been raised during the options creation.
func (c Config) ValueUint64(key string) (uint64, error) {
	const fname = "ValueUint64"
	var out uint64
	fail := func(err error) (uint64, error) {
		return out, fmt.Errorf("%s: %w", fname, err)
	}
	if c.set == nil {
		return fail(errCommands)
	}
	o := c.set.options.find(key)
	if o == nil && c.Is(key) {
		return fail(ErrNotInCurrentSet)
	}
	if o == nil {
		return fail(errNoFlag)
	}
	if o.err != nil {
		return fail(o.err)
	}
	switch t := o.data.(type) {
	case nil:
		return fail(errNoData)
	case *uint64:
		out = *t
	default:
		return fail(want(out, t))
	}
	return out, nil
}

// ValueFloat64 returns the value of an float64 option, else an error if
// one has been raised during the options creation.
func (c Config) ValueFloat64(key string) (float64, error) {
	const fname = "ValueFloat64"
	var out float64
	fail := func(err error) (float64, error) {
		return out, fmt.Errorf("%s: %w", fname, err)
	}
	if c.set == nil {
		return fail(errCommands)
	}
	o := c.set.options.find(key)
	if o == nil && c.Is(key) {
		return fail(ErrNotInCurrentSet)
	}
	if o == nil {
		return fail(errNoFlag)
	}
	if o.err != nil {
		return fail(o.err)
	}
	switch t := o.data.(type) {
	case nil:
		return fail(errNoData)
	case *float64:
		out = *t
	default:
		return fail(want(out, t))
	}
	return out, nil
}

// ValueString returns the value of an string option, else an error if one
// has been raised during the options creation.
func (c Config) ValueString(key string) (string, error) {
	const fname = "ValueString"
	var out string
	fail := func(err error) (string, error) {
		return out, fmt.Errorf("%s: %w", fname, err)
	}
	if c.set == nil {
		return fail(errCommands)
	}
	o := c.set.options.find(key)
	if o == nil && c.Is(key) {
		return fail(ErrNotInCurrentSet)
	}
	if o == nil {
		return fail(errNoFlag)
	}
	if o.err != nil {
		return fail(o.err)
	}
	switch t := o.data.(type) {
	case nil:
		return fail(errNoData)
	case *string:
		out = *t
	default:
		return fail(want(out, t))
	}
	return out, nil
}

// ValueBool returns the value of an string option, else an error if one
// has been raised during the options creation.
func (c Config) ValueBool(key string) (bool, error) {
	const fname = "ValueBool"
	var out bool
	fail := func(err error) (bool, error) {
		return out, fmt.Errorf("%s: %w", fname, err)
	}
	if c.set == nil {
		return fail(errCommands)
	}
	o := c.set.options.find(key)
	if o == nil && c.Is(key) {
		return fail(ErrNotInCurrentSet)
	}
	if o == nil {
		return fail(errNoFlag)
	}
	if o.err != nil {
		return fail(o.err)
	}
	switch t := o.data.(type) {
	case nil:
		return fail(errNoData)
	case *bool:
		out = *t
	default:
		return fail(want(out, t))
	}
	return out, nil
}

// ValueDuration returns the value of an time.Duration option, else an
// error if one has been raised during the options creation.
func (c Config) ValueDuration(key string) (time.Duration, error) {
	const fname = "ValueDuration"
	var out time.Duration
	fail := func(err error) (time.Duration, error) {
		return out, fmt.Errorf("%s: %w:", fname, err)
	}
	if c.set == nil {
		return fail(errCommands)
	}
	o := c.set.options.find(key)
	if o == nil && c.Is(key) {
		return fail(ErrNotInCurrentSet)
	}
	if o == nil {
		return fail(errNoFlag)
	}
	if o.err != nil {
		return fail(o.err)
	}
	switch t := o.data.(type) {
	case nil:
		return fail(errNoData)
	case *time.Duration:
		out = *t
	default:
		return fail(want(out, t))
	}
	return out, nil
}

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Usage display
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
