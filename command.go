package conf

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

// Command defines a set of flags for a command line applications. With
// its first call, Command creates a base flagset, that which will act on
// standard program execution, reading basic command line arguments, flags
// and their options.
//
// app [-flag] [-flag] [opt] [-flag] [opt] [-flag] ...
//
// Subsequent calls to Command each define a sub command, used to call
// sub routines within the main program, each providing its own specific
// flagset of flags and their options for that sub routine.
//
// app [cmd] [-flag] [-flag] [opt] [-flag] [opt] [-flag] ...
//
// Any errors are accumulated into the Config.errs value and dealt with
// when Compose, if ignored then returned when a value from the command
// set is accessed.
func (c *Config) Command(cmd, usage string) CMD {
	const fname = "Config.Command"

	// If not OK store the error and leave.
	if err := cmdPreconditions(c, cmd, usage); err != nil {
		c.errs = fmt.Errorf("%s: %w", fname, err)
		return 0
	}

	// First run
	if c.position == 0 {
		c.position = 1
		c.header = cmd
		// As we know the max length we can make the backing array now.
		c.commands = make([]command, 0, 64)
	}

	// We do not need to check for duplicates on either the default
	// command set nor the first sub command.
	if c.position > 2 {
		if err := checkDuplicate(c, cmd); err != nil {
			c.errs = fmt.Errorf("%s: %w", fname, err)
			return 0
		}
	}

	// Set the new command.
	m := command{flag: c.position, cmd: cmd, usage: usage}
	c.commands = append(c.commands, m)
	set := c.position
	c.position = c.position << 1

	if v1() {
		log.Printf("%s: completed\n", fname)
	}

	return set
}

func cmdPreconditions(c *Config, cmd, usage string) error {
	const fname = "cmdPreconditions"

	if cmd == "" {
		const event = "empty cmd string not permitted"
		return fmt.Errorf("%s: %s: %w", fname, event, errConfig)
	}
	if c.position >= limit {
		const event = "64 sub command limit reached"
		return fmt.Errorf("%s|%s: %s: %w",
			c.errs, fname, event, errConfig)
	}

	if v2() {
		log.Printf("%s: completed\n", fname)
	}

	return nil
}

// checkDuplicate returns an error if the given command name is already
// in use.
func checkDuplicate(c *Config, cmd string) error {
	const fname = "checkDuplicate"
	for _, c := range c.commands {
		if strings.Compare(c.cmd, cmd) == 0 {
			const event = "duplicate command"
			return fmt.Errorf("%s: %s: %s",
				fname, cmd, event)
		}
	}

	if v2() {
		log.Printf("%s: completed\n", fname)
	}

	return nil
}

/* ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *  command
 * ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ */

// options is an array of option pointers are used to keep lists of
// which commands contain which options.
type options []*Option

// find makes a search of the unerlying slice for the given value.
func (o options) find(flag string) *Option {
	for i, opt := range o {
		if strings.Compare(opt.Flag, flag) == 0 {
			return o[i]
		}
	}
	return nil
}

// flags is a slice of flag names that is used to insure that no
// duplicate flags can be created within the same commans set.
type flags []string

// find makes a search of the unerlying slice for the given value.
func (f flags) find(flag string) bool {
	for _, f := range f {
		if strings.Compare(f, flag) == 0 {
			return true
		}
	}
	return false
}

// command contains the data required to create flagSet for a program
// sub command and its flags.
type command struct {
	// flag is the set bit that represents the command.
	flag CMD
	// cmd is the token used to instiage the running mode, in the
	// case of the default set, cmd contains the defCmdSet place
	// holder.
	cmd string
	// The usage output for the command displayed when -h is called or
	// an error raised upon parsing the flagset.
	usage string
	// seen makes certain that no flag duplicates exist within the
	// set.
	seen flags
	// options contains pointers to all of the options that have
	// been assigned to this command set.
	options options
}

// CMD is a bitfield that records which Options have been registered
// with a command set.
type CMD uint64

func (c CMD) String() string {
	var count int
	if c < 0 {
		panic("illegal value")
	}
	for c > 0 {
		c = c >> 1
		count++
	}
	return strconv.Itoa(count)
}

// isInSet returns true if a command token exists within the
// configured set of commands, false if it does not.
func isInSet(c *Config, bitfield CMD) bool {
	fullset := (c.position << 1) - 1
	if bitfield == (fullset)&bitfield {
		return true
	}
	return false
}

// setCommand sets the requested command set into the Config struct as
// its current running state, returning an error if the named command
// set does not exist.
func setCommand(c *Config, name string) (set CMD, err error) {
	const fname = "setCommand"
	for i, m := range c.commands {
		if strings.Compare(name, m.cmd) == 0 {
			c.set = &c.commands[i]
			set = c.set.flag
			return
		}
	}
	err = fmt.Errorf("%s: %w", fname, ErrUnknownCMD)
	return
}
