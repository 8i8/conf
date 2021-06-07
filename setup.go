package conf

import (
	"fmt"
	"log"
	"os"
)

func configPreconditions(c *Config, opts ...Option) error {
	const fname = "configPreconditions"

	if c.errs != nil {
		return fmt.Errorf("%s: previous error: %w", fname, c.errs)
	}
	if len(opts) == 0 {
		return fmt.Errorf("%s: no options set", fname)
	}
	if len(c.commands) == 0 {
		return fmt.Errorf("%s: no commands set", fname)
	}

	if v(3) {
		log.Printf("%s: completed\n", fname)
	}

	return nil
}

// setupConfig completes all the Config struct setup requirements.
func setupConfig(c *Config) error {
	const fname = "setupConfig"

	if err := ascertainCmdSet(c); err != nil {
		return fmt.Errorf("%s: %w", fname, err)
	}

	if v(2) {
		log.Printf("%s: completed\n", fname)
	}

	return nil
}

// ascertainCmdSet sets the program operating mode, either the default or that
// specified by the first argument if it is not a flag.
func ascertainCmdSet(c *Config) error {
	const fname = "ascertainCmdSet"
	if len(os.Args) > 1 && os.Args[1][0] != '-' {
		if err := setCommand(c, os.Args[1]); err != nil {
			return fmt.Errorf("%s: %w", fname, err)
		}
		if v(3) {
			log.Printf("%s: %s: set defined\n", fname, os.Args[1])
		}
		return nil
	}
	if len(c.commands) == 0 {
		const event = "empty command set"
		return fmt.Errorf("%s: %s", fname, event)
	}
	c.set = &c.commands[0]
	if v(3) {
		log.Printf("%s: default: set defined\n", fname)
	}
	return nil
}
