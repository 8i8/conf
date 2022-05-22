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

	if v3() {
		log.Printf("%s: completed\n", fname)
	}

	return nil
}

// ascertainCmdSet sets the program operating mode, either the default or that
// specified by the first argument if it is not a flag.
func ascertainCmdSet(c *Config) (set CMD, err error) {
	const fname = "ascertainCmdSet"
	if len(os.Args) > 1 && os.Args[1][0] != '-' {
		if set, err = setCommand(c, os.Args[1]); err != nil {
			// Avoid an error in the case when a argument is
			// required and no flags nor operating commands
			// have been given, this should not raise an
			// error.
			if v2() {
				log.Printf("%s: default: set defined, file: %s\n",
					fname, os.Args[1])
			}
			c.set = &c.commands[0]
			set = 1
			return set, nil
		}
		if v2() {
			log.Printf("%s: %s: set defined\n", fname, os.Args[1])
		}
		return
	}
	if len(c.commands) == 0 {
		const event = "empty command set"
		err = fmt.Errorf("%s: %s", fname, event)
		return
	}
	c.set = &c.commands[0]
	set = 1
	if v2() {
		log.Printf("%s: default: set defined\n", fname)
	}
	return
}
