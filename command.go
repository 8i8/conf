package conf

import (
	"fmt"
	"log"
	"strings"
)

// Command defines a set of flags for a command line applications.
// Upon its first call, Command creates a base flagset, that which
// acts upon the programs execution upon its command line arguments as
// flags and their options.
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
		return 0 // no bits set
	}

	// Is this the first command set.
	if c.position == 0 {
		cmd := setDefaultCommand(c, cmd, usage)
		if v(1) {
			log.Printf("%s: completed\n", fname)
		}
		return cmd
	}

	if err := checkDuplicate(c, cmd); err != nil {
		c.errs = fmt.Errorf("%s: %w", fname, err)
		return 0
	}

	// OK, set the position flag and define the command set.
	c.position = c.position << 1
	m := command{flag: c.position, header: cmd, usage: usage}
	c.commands = append(c.commands, m)

	if v(1) {
		log.Printf("%s: completed\n", fname)
	}

	return c.position
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

	if v(2) {
		log.Printf("%s: completed\n", fname)
	}

	return nil
}

const defCmdSet = "***"

// This is the first command, we need to set the Config header and
// then to define this as the default command set.
func setDefaultCommand(c *Config, cmd, usage string) CMD {
	const fname = "setDefaultCommand"

	c.header = cmd
	c.position = 1  // 1 is the first flag, 0 will not do here.
	cmd = defCmdSet // default cmd place holder.
	m := command{flag: c.position, header: cmd, usage: usage}
	c.commands = append(c.commands, m)

	if v(2) {
		log.Printf("%s: completed\n", fname)
	}

	return c.position
}

// checkDuplicate returns an error if the given command name is already
// in use.
func checkDuplicate(c *Config, cmd string) error {
	const fname = "checkDuplicate"
	for _, c := range c.commands {
		if strings.Compare(c.header, cmd) == 0 {
			const event = "duplicate command"
			return fmt.Errorf("%s: %s: %s",
				fname, cmd, event)
		}
	}

	if v(2) {
		log.Printf("%s: completed\n", fname)
	}

	return nil
}
