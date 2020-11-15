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
	// index contains the value of the previously created bitflag used
	// to maintain an incremental value that is augmented every time
	// that a new program mode is created.
	index = 1
)

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Main package functions
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

// Setup sets the basis for the programs help output, the 'help header'
// and help 'Sub header', it also returns the bitflag for the modes field
// for use in the creation of Options, consequent calls to c.Mode will
// create and return further more flags.
func Setup(heading string, subheading string) (mode int) {
	c.help = heading
	mode = Mode("default", subheading)
	return
}

// Mode creates a new mode, returning the bitflag required to set that
// mode.
func Mode(name, help string) (bitflag int) {
	if index >= limit {
		log.Fatal("index overflow, to many program modes")
	}
	m := mode{id: index, name: name, help: help}
	c.list = append(c.list, m)
	bitflag = index
	index = index << 1
	return
}

// GetMode return the current running modes name.
func GetMode() string {
	return c.mode.name
}

// Options initialises the programs options.
func Options(opts ...Option) {
	c.saveArgs()
	c.loadOptions(opts...)
	//c.loadConfig()
}

// Parse sets the running mode from the command line arguments and then
// parses the flagset.
func Parse() error {
	const fname = "Parse"
	if len(os.Args) > 1 && os.Args[1][0] != '-' {
		// Use specified operating mode.
		if c.list.is(os.Args[1]) {
			if err := c.load(os.Args[1]); err != nil {
				return fmt.Errorf("%s: %w", fname, err)
			}
			goto checkFn
		}
		return fmt.Errorf("unknown mode: %q\n", os.Args[1])
	}
	// Load the default mode.
	if err := c.load("default"); err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}
checkFn:
	// Run all user given check functions.
	if err := c.runCheckFn(); err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}
	return nil
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
	// index holds the next value to use as a bitflag for the next
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
}

// Setup sets the basis for the programs help output, the 'help header'
// and help 'Sub header', it also returns the bitflag for the modes field
// for use in the creation of Options, consequent calls to c.Mode will
// create and return further more flags.
func (c *Config) Setup(heading string, subheading string) (mode int) {
	c.help = heading
	mode = c.Mode("default", subheading)
	return
}

// Mode creates a new mode, returning the bitflag required to set that
// mode.
func (c *Config) Mode(name, help string) (bitflag int) {
	if c.index >= limit {
		log.Fatal("index overflow, to many program modes")
	}
	m := mode{id: index, name: name, help: help}
	c.list = append(c.list, m)
	bitflag = c.index
	c.index = c.index << 1
	return
}

// GetMode return the current running modes name.
func (c Config) GetMode() string {
	return c.mode.name
}

// Options initialises the programs options.
func (c *Config) Options(opts ...Option) {
	// Record the original input string.
	c.saveArgs()
	// Generate flags and help data.
	c.loadOptions(opts...)
	// TODO write a standard config file addition that records config
	// when in mode 'config' and that reads in any settings that have
	// been previously recorded or written.
	//c.loadConfig()
}

// Parse sets the running mode from the command line arguments and then
// parses the flagset.
func (c *Config) Parse() error {
	const fname = "Parse"
	if len(os.Args) > 1 && os.Args[1][0] != '-' {
		if c.list.is(os.Args[1]) {
			if err := c.load(os.Args[1]); err != nil {
				return fmt.Errorf("%s: %w", fname, err)
			}
			// Check all set verifications against parsed data.
			if err := c.runCheckFn(); err != nil {
				return fmt.Errorf("%s: %w", fname, err)
			}
			return nil
		}
		return fmt.Errorf("unknown mode: %q\n", os.Args[1])
	}
	// If no other mode has been specified, load the default.
	if err := c.load("default"); err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}
	// Check all set verifications against parsed data.
	if err := c.runCheckFn(); err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}
	return nil
}

// ArgList returns the full command line argument list as a string as it
// was input.
func (c Config) ArgList() string {
	return c.input
}

// saveArgs records the input arguments as a string for display output.
func (c *Config) saveArgs() {
	var str strings.Builder
	str.WriteString(os.Args[0])
	for _, arg := range os.Args[1:] {
		str.WriteByte(' ')
		str.WriteString(arg)
	}
	c.input = str.String()
}

// optionsToFlagSet defines flags within the flagset for all options that
// have been specified that are within the current working set.
func (c *Config) optionsToFlagSet() error {
	const fname = "optionsToFlagSet"
	for k, opt := range c.options {
		if c.mode.id&opt.Modes > 0 {
			err := c.options[k].toFlagSet(c.flagSet)
			if err != nil {
				return fmt.Errorf("%s: key %q: %w",
					fname, k, err)
			}
		}
	}
	return nil
}

// loadOptions loads all of the given options into the option map.
func (c *Config) loadOptions(opts ...Option) {
	const fname = "Option"
	c.options = make(map[string]*Option)
	if c.names == nil {
		c.names = make(map[string]bool)
	}
	for i := range c.list {
		if c.list[i].keys == nil {
			c.list[i].keys = make(map[string]bool)
		}
	}
	for i, opt := range opts {
		c.checkOption(opt)
		c.options[opt.Name] = &opts[i]
	}
}

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Pre Parse Option Checks
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

// checkOption checks the user supplied data within an option including
// duplicate name and key values.
func (c Config) checkOption(o Option) {
	c.checkName(o)
	c.checkType(o)
	c.checkKey(o)
	c.checkDefault(o)
	c.checkMode(o)
}

var (
	errType       = errors.New("type error")
	errTypeNil    = errors.New("the type is not defined")
	errTypeUnkown = errors.New("unknown type")
	errNoValue    = errors.New("value required")
	errDuplicate  = errors.New("duplicate value")
	errNotInSet   = errors.New(
		"specified mode not within the current set")
)

// chekcName checks that the name is not empty and that it is not a
// duplicate value.
func (c *Config) checkName(o Option) {
	const msg = "Option"
	if len(o.Name) == 0 {
		log.Fatalf("%s: %s: %q: %s", pkg, msg,
			o.Name, errNoValue)
	}
	if c.names[o.Name] {
		log.Fatalf("%s: %s: %q: %s", pkg, msg,
			o.Name, errDuplicate)
	}
	c.names[o.Name] = true
}

// chekcName checks that the key is not empty and that it is not a
// duplicate value.
func (c *Config) checkKey(o Option) {
	const msg = "Option"
	if len(o.Key) == 0 {
		log.Fatalf("%s: %s: %q: %s", pkg, msg, o.Name,
			errNoValue)
	}
	for i := range c.list {
		// If the option key is already registered for this
		// mode fail it; This limit is specific to each mode, the
		// same key value may be used for different options if
		// they are not in the same set.
		if c.list[i].id&o.Modes > 0 {
			if c.list[i].keys[o.Key] {
				log.Fatalf("%s: %s: %q: %q: %s", pkg,
					msg, o.Name, o.Key, errDuplicate)
			}
			c.list[i].keys[o.Key] = true
		}
	}
}

// checkType checks that the option has a type set.
func (c *Config) checkType(o Option) {
	const msg = "Option"
	if o.Type == Nil {
		log.Fatalf("%s: %s: %q: %q: %s",
			pkg, msg, o.Name, o.Type, errTypeNil)
	}
}

// checkDefault checks that the options default value has the correct
// type.
// TODO add types.
func (c *Config) checkDefault(o Option) {
	const msg = "Option"
	switch o.Type {
	case Int:
		if _, ok := o.Default.(int); !ok {
			log.Fatalf("%s: %s: %q: Default: %+v: %s",
				pkg, msg, o.Name, o.Type, errType)
		}
	case String:
		if _, ok := o.Default.(string); !ok {
			log.Fatalf("%s: %s: %q: Default: %+v: %s",
				pkg, msg, o.Name, o.Type, errType)
		}
	case Bool:
		if _, ok := o.Default.(bool); !ok {
			log.Fatalf("%s: %s: %q: Default: %+v: %s",
				pkg, msg, o.Name, o.Type, errType)
		}
	case Float:
		if _, ok := o.Default.(float64); !ok {
			log.Fatalf("%s: %s: %q: Default: %+v: %s",
				pkg, msg, o.Name, o.Type, errType)
		}
	case Duration:
		if _, ok := o.Default.(time.Duration); !ok {
			log.Fatalf("%s: %s: %q: Default: %+v: %s",
				pkg, msg, o.Name, o.Type, errType)
		}
	case Var:
		// The default is usually nil but could be anything as var
		// represents and interface.
	case Nil:
		log.Fatalf("%s: %s: %q: Default: %+v: %s",
			pkg, msg, o.Name, o.Type, errTypeNil)
	default:
		log.Fatalf("%s: %s: %q: Default: %+v: %s",
			pkg, msg, o.Name, o.Type, errTypeUnkown)
	}
}

// checkMode verifies that an option has a mode set and if it does that
// that mode is contained within the existing sets.
func (c Config) checkMode(o Option) {
	const msg = "Option"
	if !c.list.flagIs(o.Modes) {
		log.Fatalf("%s: %s: %q: Modes: %s", pkg, msg, o.Name,
			errNotInSet)
	}
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

// flagIs returns true if a bitfield is entirely contained in the mode
// list full set, returning false if it is not.
func (l modelist) flagIs(bitflag int) bool {
	if bitflag == 0 {
		return false
	}
	for _, f := range l {
		if bitflag&f.id > 0 {
			bitflag = bitflag ^ f.id
		}
	}
	if bitflag > 0 {
		return false
	}
	return true
}

// setMode defines the programs current running mode.
func (o *Config) setMode(mode string) error {
	const fname = "setMode"
	if mode == "default" {
		o.mode = c.list[0]
	}
	for _, m := range c.list {
		if strings.Compare(mode, m.name) == 0 {
			o.mode = m
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
	if err := c.optionsToFlagSet(); err != nil {
		return fmt.Errorf("%s: mode %q: %w", fname, mode, err)
	}
	c.flagSet.Usage = func() {
		fmt.Println(c.help)
		fmt.Println(c.mode.help)
		c.flagSet.VisitAll(flagHelpMsg)
	}
	// Offset the cmd line arguments and parse the remaining flags.
	offset := 2
	if mode == "default" {
		offset-- // one fewer args on the command line.
	}
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
// TODO add types.
func (o *Option) toFlagSet(fs *flag.FlagSet) error {
	const fname = "toFlagSet"
	const msg = "Option"
	switch o.Type {
	case Int:
		i, ok := o.Default.(int)
		if !ok {
			return fmt.Errorf("%s: %s: %s: %w",
				pkg, msg, o.Type, errType)
		}
		o.data = fs.Int(o.Key, i, o.Help)
	case String:
		s, ok := o.Default.(string)
		if !ok {
			return fmt.Errorf("%s: %s: %s: %w",
				pkg, msg, o.Type, errType)
		}
		o.data = fs.String(o.Key, s, o.Help)
	case Bool:
		b, ok := o.Default.(bool)
		if !ok {
			return fmt.Errorf("%s: %s: %s: %w",
				pkg, msg, o.Type, errType)
		}
		o.data = fs.Bool(o.Key, b, o.Help)
	case Float:
		f, ok := o.Default.(float64)
		if !ok {
			return fmt.Errorf("%s: %s: %s: %w",
				pkg, msg, o.Type, errType)
		}
		o.data = fs.Float64(o.Key, f, o.Help)
	case Duration:
		d, ok := o.Default.(time.Duration)
		if !ok {
			return fmt.Errorf("%s: %s: %s: %w",
				pkg, msg, o.Type, errType)
		}
		o.data = fs.Duration(o.Key, d, o.Help)
	case Var:
		if o.Value == nil {
			return fmt.Errorf("%s: %s: %s: %w",
				pkg, msg, o.Type, errNoValue)
		}
		fs.Var(o.Value, o.Key, o.Help)
	default:
		log.Fatalf("%s: %s: internal error: (%q, %s) %s",
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

// Type defines the type of the configuration option, essential when
// setting flags, converting from interfaces.
// TODO add types.
// uint8       the set of all unsigned  8-bit integers (0 to 255)
// uint16      the set of all unsigned 16-bit integers (0 to 65535)
// uint32      the set of all unsigned 32-bit integers (0 to 4294967295)
// uint64      the set of all unsigned 64-bit integers (0 to 18446744073709551615)

// int8        the set of all signed  8-bit integers (-128 to 127)
// int16       the set of all signed 16-bit integers (-32768 to 32767)
// int32       the set of all signed 32-bit integers (-2147483648 to 2147483647)
// int64       the set of all signed 64-bit integers (-9223372036854775808 to 9223372036854775807)

// float32     the set of all IEEE-754 32-bit floating-point numbers
// float64     the set of all IEEE-754 64-bit floating-point numbers

// complex64   the set of all complex numbers with float32 real and imaginary parts
// complex128  the set of all complex numbers with float64 real and imaginary parts
type Type uint64

const (
	Nil Type = iota
	Int
	Float
	String
	Bool
	Duration
	Var
)

var (
	errNoData = errors.New("no data")
	errNoKey  = errors.New("key not found")
)

// TODO add types.
func (t Type) String() string {
	switch t {
	case Nil:
		return "nil"
	case Int:
		return "int"
	case Float:
		return "float64"
	case String:
		return "string"
	case Bool:
		return "bool"
	case Duration:
		return "time.Duration"
	case Var:
		return "interface{}"
	default:
		return "type not listed in Stringer"
	}
}

// Value returns the content of an option flag its type and also a boolean
// that expresses whether or not the flag has been found.
func Value(key string) (interface{}, Type, error) {
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

// ValueInt returns the value of an int option.
func ValueInt(key string) (int, error) {
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

// ValueFloat64 returns the value of int options.
func ValueFloat(key string) (float64, error) {
	const fname = "ValueFloat"
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

// ValueString returns the value of a string options.
func ValueString(key string) (string, error) {
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

// ValueBool returns the value of a boolean options.
func ValueBool(key string) (bool, error) {
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

// ValueDuration returs the value of a time.Duration option.
func ValueDuration(key string) (time.Duration, error) {
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
