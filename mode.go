package conf

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
)

type Mode struct {
	id   int
	Name string
	Help string
}
type modelist []Mode

var (
	list  modelist               // List of modes set by the user.
	index = 1                    // The flag to set in the mode bitfield.
	limit = math.MaxInt64>>1 + 1 // no more than 64 base modes are possible.
)

// ID returns the current bitfield that describes the Mode.
func (m Mode) ID() int {
	return m.id
}

// SetModes returns the bitfield containing of all the given bit flags.
func SetModes(modes ...Mode) int {
	m := 0
	for _, mode := range modes {
		m = m | mode.id
	}
	return m
}

// isMode returns true if the given mode exists and false if it does not.
func (l modelist) is(mode string) bool {
	for _, m := range l {
		if strings.Compare(mode, m.Name) == 0 {
			return true
		}
	}
	return false
}

// setMode defines the programs current running mode.
func (o *config) setMode(mode string) error {
	for _, m := range list {
		if strings.Compare(mode, m.Name) == 0 {
			o.Mode = m
			return nil
		}
	}
	return fmt.Errorf("setMode: mode not found")
}

// Parse sets the running mode from the command line arguments.
func Parse() {
	if len(os.Args) > 1 && os.Args[1][0] != '-' {
		if list.is(os.Args[1]) {
			loadMode(os.Args[1])
			return
		}
		fmt.Printf("unknown mode: %q\n", os.Args[1])
		return
	}
	loadMode("def")
}

// Help sets the base for the programs help output.
func Help(help string) {
	c.help = help
}

// Modes adds a name to the list of possible modes.
func Modes(modes ...*Mode) error {
	for i := range modes {
		if index >= limit {
			return fmt.Errorf("index overflow, to many program modes")
		}
		modes[i].id = index
		list = append(list, *modes[i])
		index = index << 1
	}
	return nil
}

func (m Mode) String() string {
	return m.Name
}

// GetMode returns the current set program mode.
func (c *config) GetMode() string {
	return c.Mode.Name
}

// loadMode sets the programs operating mode.
func loadMode(mode string) {
	const fname = "Mode"

	if err := c.setMode(mode); err != nil {
		log.Fatal(fname, ": ", err)
	}
	c.flagSet = flag.NewFlagSet(c.Mode.String(), flag.ExitOnError)
	c.setOptionsToFlags()
	c.flagSet.Usage = func() {
		fmt.Println(c.help)
		fmt.Println(c.Mode.Help)
		c.flagSet.VisitAll(loadFlagHelpMsg)
	}
	if mode == "def" {
		err := c.flagSet.Parse(os.Args[1:])
		if err != nil {
			fmt.Println(fname, ": ", err)
		}
		return
	}
	err := c.flagSet.Parse(os.Args[2:])
	if err != nil {
		fmt.Println(fname, ": ", err)
	}
}

func loadFlagHelpMsg(f *flag.Flag) {
	s := fmt.Sprintf("        -%s", f.Name)
	name, usage := flag.UnquoteUsage(f)
	if len(name) > 0 {
		s += " " + name
	}
	s += "\n        \t"
	s += strings.ReplaceAll(usage, "\n", "\n            \t")
	fmt.Fprint(os.Stdout, s, "\n")
}
