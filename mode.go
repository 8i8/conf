package conf

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
)

// mode contains the required data to create a program operating mode flag, the
// sub heading of a program run mode for its specific operating flags.
type mode struct {
	id   int
	name string
	// The help output for the particular mode displayed when -h is used.
	help string
}
type modelist []mode

var (
	// list contains the modes created by the user.
	list modelist
	// index contains the value of the previously created bitflag used to
	// maintain an incremental value that is agmented every time that a new
	// program mode is created.
	index = 1
	// limit ensures that no more than 64 base modes are possible.
	limit = math.MaxInt64>>1 + 1
)

// Mode creates a new mode, returning the bitflag requred to set that mode.
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

// ID returns the current bitfield that describes the Mode.
func (m mode) ID() int {
	return m.id
}

// SetBits returns a bitfield with all the given flags set.
func SetBits(modes ...mode) int {
	m := 0
	for _, mode := range modes {
		m = m | mode.id
	}
	return m
}

// isMode returns true if the given mode exists and false if it does not.
func (l modelist) is(mode string) bool {
	for _, m := range l {
		if strings.Compare(mode, m.name) == 0 {
			return true
		}
	}
	return false
}

// setMode defines the programs current running mode.
func (o *config) setMode(mode string) error {
	for _, m := range list {
		if strings.Compare(mode, m.name) == 0 {
			o.mode = m
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

// Modes adds a name to the list of possible modes.
func Modes(modes ...*mode) error {
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

func (m mode) String() string {
	return m.name
}

// GetMode returns the current set program mode.
func (c *config) GetMode() string {
	return c.mode.name
}

// loadMode sets the programs operating mode.
func loadMode(mode string) {
	const fname = "Mode"

	if err := c.setMode(mode); err != nil {
		log.Fatal(fname, ": ", err)
	}
	c.flagSet = flag.NewFlagSet(c.mode.String(), flag.ExitOnError)
	c.optionsToFlagSet()
	c.flagSet.Usage = func() {
		fmt.Println(c.help)
		fmt.Println(c.mode.help)
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

func space(b []byte, l int) string {
	w := len(b)
	l = w - l
	for i := 0; i < l; i++ {
		w--
		b[w] = ' '
	}
	return string(b[w:])
}

// loadFlagHelpMsg writes the help message for each individual flag.
func loadFlagHelpMsg(f *flag.Flag) {
	l := len(f.Name) + 1 // for the '-' char.
	var buf [8]byte
	sp := space(buf[:], l)
	s := fmt.Sprintf("        -%s%s", f.Name, sp)
	_, usage := flag.UnquoteUsage(f)
	if l > 6 {
		s += "\n        \t"
	}
	s += strings.ReplaceAll(usage, "\n", "\n            \t")
	fmt.Fprint(os.Stdout, "FLAGS\n")
	fmt.Fprint(os.Stdout, s, "\n\n")
}
