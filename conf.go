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
	// Global package name functnion used in help output.
	pkg = "conf"
	// limit ensures that no more than 64 base modes are possible.
	limit = math.MaxInt64>>1 + 1
	// config contains the program data for the default settings
	// struct used when not running on an exported struct.
	c Config
	// list contains the modes created by the user.
	list modelist
	// index contains the value of the previously created bitflag used
	// to maintain an incremental value that is agmented every time
	// that a new program mode is created.
	index = 1
)

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Main package functions
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

// Help sets the basis for the programs help output, the 'help header'.
func Help(help string) {
	c.help = help
}

// Mode creates a new mode, returning the bitflag requred to set that
// mode.
func Mode(name, help string) (bitflag int) {
	if index >= limit {
		log.Fatal("index overflow, to many program modes")
	}
	m := mode{id: index, name: name, help: help}
	list = append(list, m)
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
		if list.is(os.Args[1]) {
			if err := c.load(os.Args[1]); err != nil {
				return fmt.Errorf("%s: %w", fname, err)
			}
			// Check all set verifications against parsed data.
			if err := c.checkOptions(); err != nil {
				return fmt.Errorf("%s: %w", fname, err)
			}
			return nil
		}
		return fmt.Errorf("unknown mode: %q\n", os.Args[1])
	}
	// Load the default arguments.
	if err := c.load("default"); err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}
	// Check all set verifications against parsed data.
	if err := c.checkOptions(); err != nil {
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
	// options are where the data for each flag or option is stored,
	// this includes the value of the key its default value and help
	// string along with the actual data once the flag or config
	// option has been parsed, it also contains a function by which
	// the value that has been set may be checked.
	options map[string]*Option
	// names makes certain that no option name duplicates exist.
	names map[string]bool
	// keys makes certain that no key duplicates exist.
	keys map[string]bool
}

// NewConfig returns a new confiuration struct.
func NewConfig() Config {
	return Config{}
}

// Help sets the basis for the programs help output, the 'help header'.
func (c *Config) Help(help string) {
	c.help = help
}

// Mode creates a new mode, returning the bitflag requred to set that
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
			if err := c.checkOptions(); err != nil {
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
	if err := c.checkOptions(); err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}
	return nil
}

// ArgList returns the full command line argument list as a string as it
// was input.
func (c Config) ArgList() string {
	return c.input
}

// saveArgs records the input argumeants as a string for display output.
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
	c.options = make(map[string]*Option)
	c.names = make(map[string]bool)
	c.keys = make(map[string]bool)
	if c.names == nil {
		c.names = make(map[string]bool)
	}
	for i, opt := range opts {
		if c.names[opt.Name] {
			log.Fatal("conf: loadOptions: %q: duplicate "+
				"option name", opt.Name)
		}
		if c.keys[opt.Key] {
			log.Fatal("conf: loadOptions: %q: duplicate "+
				"option key", opt.Key)
		}
		c.options[opt.Name] = &opts[i]
		c.names[opt.Name] = true
		c.keys[opt.Key] = true
	}
}

// checkOptions runs all user given check functions for the option set,
// called after having parsed all option data.
func (c *Config) checkOptions() error {
	const fname = "checkOptions"
	if c.options == nil {
		return fmt.Errorf("%s: %s: option set empty", pkg, fname)
	}
	var err error
	for _, o := range c.options {
		if o.Check != nil {
			if o.data, err = o.Check(o.data); err != nil {
				fmt.Println(err)
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

// setMode defines the programs current running mode.
func (o *Config) setMode(mode string) error {
	if mode == "default" {
		o.mode = list[0]
	}
	for _, m := range list {
		if strings.Compare(mode, m.name) == 0 {
			o.mode = m
			return nil
		}
	}
	return fmt.Errorf("setMode: mode not found")
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
	if mode == "default" {
		err := c.flagSet.Parse(os.Args[1:])
		if err != nil {
			return fmt.Errorf("%s: %w", fname, err)
		}
		return nil
	}
	err := c.flagSet.Parse(os.Args[2:])
	if err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}
	return nil
}

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  Option
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

// ckFunc defines a function to check an options input data, the funciton
// is passed into the option when it is created by the user and run when
// the intput flags and any configuration options are parsed.
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
	// Keypress required to activate the flag.
	Key string
	// Help string.
	Help string
	// Data.
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

var errType = errors.New("type error")

// toFlagSet generates a flag within the given flagset for the current
// option.
func (o *Option) toFlagSet(fs *flag.FlagSet) error {
	const fname = "toFlagSet"
	switch o.Type {
	case Int:
		i, ok := o.Default.(int)
		if !ok {
			return fmt.Errorf("%s: int: %w", fname, errType)
		}
		o.data = fs.Int(o.Key, i, o.Help)
	case String:
		s, ok := o.Default.(string)
		if !ok {
			return fmt.Errorf("%s: string: %w", fname, errType)
		}
		o.data = fs.String(o.Key, s, o.Help)
	case Bool:
		b, ok := o.Default.(bool)
		if !ok {
			return fmt.Errorf("%s: bool: %w", fname, errType)
		}
		o.data = fs.Bool(o.Key, b, o.Help)
	case Float:
		f, ok := o.Default.(float64)
		if !ok {
			return fmt.Errorf("%s: float64 %w", fname, errType)
		}
		o.data = fs.Float64(o.Key, f, o.Help)
	case Duration:
		d, ok := o.Default.(time.Duration)
		if !ok {
			return fmt.Errorf("%s: time.Duration %w",
				fname, errType)
		}
		o.data = fs.Duration(o.Key, d, o.Help)
	case Var:
		if o.Value == nil {
			return fmt.Errorf("%s: Value %w", fname, errType)
		}
		fs.Var(o.Value, o.Key, o.Help)
	default:
		log.Fatalf("conf: internal error: flag type not "+
			"recognised (%q, %s)", o.Name, o.Type)
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

// Value returns the content of an option flag its type and also a boolean
// that expresses whether or not the flag has been found.
func Value(key string) (interface{}, Type, error) {
	const fname = "Value"
	o, ok := c.options[key]
	if !ok {
		return nil, Nil, fmt.Errorf("%s: %s: %q key not found",
			pkg, fname, key)
	}
	if o.data == nil {
		return nil, Nil, fmt.Errorf("%s: %s: nil pointer",
			pkg, fname)
	}
	return o.data, o.Type, nil
}

// ValueInt returns the value of an int option.
func ValueInt(key string) (int, error) {
	const fname = "ValueInt"
	o, ok := c.options[key]
	if !ok {
		return 0, fmt.Errorf("%s: %s: %q key not found",
			pkg, fname, key)
	}
	if o.data == nil {
		return 0, fmt.Errorf("%s: %s: %q: nil pointer",
			pkg, fname, key)
	}
	return *o.data.(*int), nil
}

// ValueFloat64 returns the value of int options.
func ValueFloat(key string) (float64, error) {
	const fname = "ValueFloat"
	o, ok := c.options[key]
	if !ok {
		return 0, fmt.Errorf("%s: %s: %q key not found",
			pkg, fname, key)
	}
	if o.data == nil {
		return 0, fmt.Errorf("%s: %s: nil pointer",
			pkg, fname)
	}
	return *o.data.(*float64), nil
}

// ValueString returns the value of a string options.
func ValueString(key string) (string, error) {
	const fname = "ValueString"
	o, ok := c.options[key]
	if !ok {
		return "", fmt.Errorf("%s: %s: %q key not found",
			pkg, fname, key)
	}
	if o.data == nil {
		return "", fmt.Errorf("%s: %s: nil pointer",
			pkg, fname)
	}
	return *o.data.(*string), nil
}

// ValueBool returns the value of a boolean options.
func ValueBool(key string) (bool, error) {
	const fname = "ValueBool"
	o, ok := c.options[key]
	if !ok {
		return false, fmt.Errorf("%s: %s: %q key not found",
			pkg, fname, key)
	}
	if o.data == nil {
		return false, fmt.Errorf("%s: %s: nil pointer",
			pkg, fname)
	}
	return *o.data.(*bool), nil
}

// ValueDuration returs the value of a time.Duration option.
func ValueDuration(key string) (time.Duration, error) {
	const fname = "ValueDuration"
	o, ok := c.options[key]
	if !ok {
		return time.Duration(0), fmt.Errorf("%s: %s: %q "+
			"key not found", pkg, fname, key)
	}
	if o.data == nil {
		return 0, fmt.Errorf("%s: nil pointer", fname)
	}
	return *o.data.(*time.Duration), nil
}
