package conf

import (
	"errors"
	"fmt"
	"testing"
)

func TestCommandGetCmd(t *testing.T) {
	const fname = "TestCommandGetCmd"
	c = &Config{}
	cmd := c.Command("one", "like that it is")
	cmd2 := c.Command("two", "or like that")
	var opts = []Option{
		{
			Type:     Int,
			Flag:     "a",
			Usage:    "like this",
			Default:  1,
			Commands: cmd,
		},
	}
	_, err := c.Compose(opts...)
	if err != nil {
		t.Errorf("%s: should not raise an error: %s",
			fname, err)
	}
	mode := c.Cmd()
	if mode != cmd {
		t.Errorf("%s: expected \"*\" received %q", fname, mode)
	}
	if !isInSet(c, cmd2) {
		t.Errorf("%s: not a valid Command token", fname)
	}
}

func TestCommandDuplicateKeys(t *testing.T) {
	const fname = "TestCommandDuplicateKeys"
	config := Config{}
	m1 := config.Command("one", "like that it is")
	m2 := config.Command("modetwo", "alternatively so")
	var opts = []Option{
		{
			Type:     Int,
			Flag:     "a",
			Usage:    "like this",
			Default:  1,
			Commands: m1,
		},
		{
			Type:     Int,
			Flag:     "a",
			Usage:    "like that",
			Default:  1,
			Commands: m2,
		},
	}
	_, err := config.Compose(opts...)
	if err != nil {
		t.Errorf("%s: should not raise an error: %s",
			fname, err)
	}
}

func TestCommandTooMany(t *testing.T) {
	const fname = "TestCommandTooMany"
	config := Config{}
	_ = config.Command("one", "like this")
	names := make([]string, 65)
	for i := 0; i <= 64; i++ {
		names[i] = fmt.Sprint(i + '0')
	}
	for i := 0; i <= 64; i++ {
		_ = config.Command(names[i], "")
	}
	_, err := config.Compose()
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: %s", fname, err)
	}
}

func TestCommandNotThere(t *testing.T) {
	const fname = "TestCommandNotThere"
	config := Config{}
	_ = config.Command("", "")
	m := CMD(2)
	var opts = []Option{
		{
			Type:     Int,
			Flag:     "a",
			Usage:    "like this",
			Default:  1,
			Commands: m,
		},
	}
	_, err := config.Compose(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: %s", fname, err)
	}
}

func TestCommandTokens(t *testing.T) {
	const fname = "TestCommandTokenIs"
	c := &Config{}
	cmd1 := c.Command("one", "the first way")
	cmd2 := c.Command("two", "the second way")
	cmd3 := c.Command("three", "the third way")
	var opts = []Option{
		{
			Type:     Int,
			Flag:     "a",
			Usage:    "like this",
			Default:  1,
			Commands: cmd1,
		},
	}
	_, err := c.Compose(opts...)
	if err != nil {
		t.Errorf("%s: %s", fname, err)
	}
	v := isInSet(c, 0)
	if v {
		t.Errorf("%s: received true expected false", fname)
	}
	v = isInSet(c, cmd1)
	if !v {
		t.Errorf("%s: received false expected true", fname)
	}
	v = isInSet(c, cmd2)
	if !v {
		t.Errorf("%s: received false expected true", fname)
	}
	v = isInSet(c, cmd3)
	if !v {
		t.Errorf("%s: received false expected true", fname)
	}
	v = isInSet(c, cmd1|cmd3)
	if !v {
		t.Errorf("%s: received false expected true", fname)
	}
	v = isInSet(c, cmd1|cmd2|cmd3)
	if !v {
		t.Errorf("%s: received false expected true", fname)
	}
	v = isInSet(c, c.position)
	if !v {
		t.Errorf("%s: received false expected true", fname)
	}
}
