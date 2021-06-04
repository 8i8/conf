package conf

import (
	"fmt"
	"os"
)

// setupConfig completes all the Config struct setup requirements.
func setupConfig(c *Config) error {
	const fname = "setupConfig"

	if err := setupMaps(c); err != nil {
		if verbose {
			fmt.Printf("%s: failed\n", fname)
		}
		return fmt.Errorf("%s: %w", fname, err)
	}
	if err := ascertainCmdSet(c); err != nil {
		if verbose {
			fmt.Printf("%s: failed\n", fname)
		}
		return fmt.Errorf("%s: %w", fname, err)
	}

	if verbose {
		fmt.Printf("%s: completed\n", fname)
	}

	return nil
}

// setupMaps creates all the map types that the Config struct
// requires to function.
func setupMaps(c *Config) error {
	const fname = "setupMaps"

	c.options = make(map[string]*Option)
	c.seen = make(map[string]bool)

	if verbose {
		fmt.Printf("%s: completed\n", fname)
	}

	return nil
}

// ascertainCmdSet sets the program operating mode, either the default or that
// specified by the first agument if it is not a flag.
func ascertainCmdSet(c *Config) error {
	const fname = "ascertainCmdSet"
	if len(os.Args) > 1 && os.Args[1][0] != '-' {
		if err := setCommand(c, os.Args[1]); err != nil {
			if verbose {
				fmt.Printf("%s: failed\n", fname)
			}
			return fmt.Errorf("%s: %w", fname, err)
		}
		if verbose {
			fmt.Printf("%s: %s: set defined\n", fname, os.Args[1])
		}
		return nil
	}
	if len(c.commands) == 0 {
		if verbose {
			fmt.Printf("%s: failed\n", fname)
		}
		const event = "empty command set"
		return fmt.Errorf("%s: %s", fname, event)
	}
	c.set = &c.commands[0]
	if verbose {
		fmt.Printf("%s: default: set defined\n", fname)
	}
	return nil
}
