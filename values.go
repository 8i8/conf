package conf

import (
	"fmt"
	"time"

	"github.com/8i8/conf/types"
)

// Value returns the content of a an option flag its type and also a boolen
// that expresse whether or not the flag has been found.
func Value(flag string) (interface{}, types.T, bool) {
	o, ok := c.options[flag]
	if !ok {
		return nil, types.Nul, false
	}
	return o.flag, o.Type, true
}

// ValueString returns the value of a string options.
func ValueString(flag string) (string, error) {
	const fname = "ValueString"
	o, ok := c.options[flag]
	if !ok {
		return "", fmt.Errorf("%s: %s: %q flag not found",
			pkg, fname, flag)
	}
	v, ok := o.flag.(string)
	if !ok {
		return "", fmt.Errorf("%s: %s: %q flag type error "+
			"(%v, %T)", pkg, fname, flag, o.Type)
	}
	return v, nil
}

// ValueInt returns the value of an int option.
func ValueInt(flag string) (int, error) {
	const fname = "ValueInt"
	o, ok := c.options[flag]
	if !ok {
		return 0, fmt.Errorf("%s: %s: %q flag not found",
			pkg, fname, flag)
	}
	v, ok := o.flag.(int)
	if !ok {
		return 0, fmt.Errorf("%s: %s: %q flag type error "+
			"(%v, %T)", pkg, fname, flag, o.Type)
	}
	return v, nil
}

// ValueFloat64 returns the value of int options.
func ValueFloat(flag string) (float64, error) {
	const fname = "ValueFloat"
	o, ok := c.options[flag]
	if !ok {
		return 0, fmt.Errorf("%s: %s: %q flag not found",
			pkg, fname, flag)
	}
	v, ok := o.flag.(float64)
	if !ok {
		return 0, fmt.Errorf("%s: %s: %q flag type error "+
			"(%v, %T)", pkg, fname, flag, o.Type)
	}
	return v, nil
}

// ValueBool returns the value of a boolean options.
func ValueBool(flag string) (bool, error) {
	const fname = "ValueBool"
	o, ok := c.options[flag]
	if !ok {
		return false, fmt.Errorf("%s: %s: %q flag not found",
			pkg, fname, flag)
	}
	v, ok := o.flag.(bool)
	if !ok {
		return false, fmt.Errorf("%s: %s: %q flag type error "+
			"(%v, %T)", pkg, fname, flag, o.Type)
	}
	return v, nil
}

// ValueDuration returs the value of a time.Duration option.
func ValueDuration(flag string) (time.Duration, error) {
	const fname = "ValueDuration"
	o, ok := c.options[flag]
	if !ok {
		return time.Duration(0), fmt.Errorf("%s: %s: %q flag not found",
			pkg, fname, flag)
	}
	v, ok := o.flag.(time.Duration)
	if !ok {
		return time.Duration(0), fmt.Errorf("%s: %s: %q flag type error",
			"(%v, %T)", pkg, fname, flag, o.Type)
	}
	return v, nil
}
