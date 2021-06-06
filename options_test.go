package conf

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

func TestOptionsCheckUserFn(t *testing.T) {
	const fname = "TestOptionsCheckUserFn"
	config := Config{}
	m := config.Command("one", "like this")
	var opts = []Option{
		{
			Type:     Int,
			Flag:     "i",
			Usage:    "like this",
			Default:  1,
			Commands: m,
			Check: func(v interface{}) (interface{}, error) {
				i := v.(*int)
				*i++
				return i, nil
			},
		},
	}
	err := config.Compose(opts...)
	if err != nil {
		t.Errorf("%s: %s", fname, err)
	}
	i, err := config.ValueInt("i")
	if err != nil {
		t.Errorf("%s: %s", fname, err)
	}
	if i != 2 {
		t.Errorf("%s: recieved %d expected 2", fname, i)
	}
}

func TestOptionsCheckUserFnError(t *testing.T) {
	const fname = "TestOptionsCheckUserFnError"
	config := Config{}
	m := config.Command("one", "like that")
	var opts = []Option{
		{
			Type:     Int,
			Flag:     "i",
			Usage:    "like this",
			Default:  1,
			Commands: m,
			Check: func(v interface{}) (interface{}, error) {
				i := *v.(*int)
				if i != 101 {
					return v, fmt.Errorf("%s: that does not count", fname)
				}
				return v, nil
			},
		},
	}
	err := config.Compose(opts...)
	if !errors.Is(err, ErrCheck) {
		t.Errorf("%s: %s", fname, err)
	}
	_, err = config.ValueInt("i")
	if !errors.Is(err, ErrCheck) {
		t.Errorf("%s: %s", fname, err)
	}
}

func TestOptionsCheckName(t *testing.T) {
	const fname = "TestOptionsCheckName"
	config := Config{}
	m := config.Command("", "like that it is")
	var opts = []Option{
		{
			Type:     Int,
			Flag:     "a",
			Usage:    "like this",
			Default:  1,
			Commands: m,
		},
		{
			Type:     Int,
			Flag:     "b",
			Usage:    "like this",
			Default:  1,
			Commands: m,
		},
		{
			Type:     Int,
			Flag:     "c",
			Usage:    "like this",
			Default:  1,
			Commands: m,
		},
	}
	err := config.Compose(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: %s", fname, err)
	}
}

func TestOptionsCheckFlagPresent(t *testing.T) {
	const fname = "TestOptionsCheckFlagPresent"
	config := Config{}
	m := config.Command("one", "like that it is")
	var opts = []Option{
		{
			Type:     Int,
			Usage:    "like this",
			Default:  1,
			Commands: m,
		},
	}
	err := config.Compose(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: %s", fname, err)
	}
}

func TestOptionsCheckFlagDuplicate(t *testing.T) {
	const fname = "TestOptionsCheckFlagDuplicate"
	config := Config{}
	m := config.Command("one", "like that it is")
	var opts = []Option{
		{
			Type:     Int,
			Flag:     "a",
			Usage:    "like this",
			Default:  1,
			Commands: m,
		},
		{
			Type:     Int,
			Flag:     "a",
			Usage:    "like this",
			Default:  1,
			Commands: m,
		},
	}
	err := config.Compose(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: %s", fname, err)
	}
}

func TestOptionsEdgeCaseNoArgs(t *testing.T) {
	const fname = "TestOptionsEdgeCaseNoArgs"
	temp := os.Args
	os.Args = os.Args[:0]
	config := Config{}
	m := config.Command("", "ops! I forgot myself")
	var opts = []Option{
		{
			Type:     Int,
			Flag:     "a",
			Usage:    "like this",
			Default:  1,
			Commands: m,
		},
	}
	err := config.Compose(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: %s", fname, err)
	}
	os.Args = temp
}
