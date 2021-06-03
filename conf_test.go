package conf

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

var c Config

func init() {
	test = true
}

type testValue struct {
	str string
}

func (t testValue) String() string {
	return t.str
}

func (t *testValue) Set(str string) error {
	return nil
}

func TestConfigValues(t *testing.T) {
	const fname = "TestConfig"
	var (
		str = "string"
		b   bool
		d   time.Duration
	)
	options := map[string]struct {
		typ   Type
		def   interface{}
		v     interface{}
		value flag.Value
		exp   string
	}{
		// ValueInt
		"ValueInt_Pass":        {typ: Int, def: int(1), exp: "pass"},
		"ValueInt_FailDefault": {typ: Int, def: "wrongType", exp: "fail"},
		"ValueInt_ErrNotThere": {typ: Int, def: int(1), exp: "errNoKey"},
		"ValueInt_ErrStored":   {typ: Int, def: "wrongType", exp: "fail"},
		"ValueInt_ErrNoData":   {typ: Int, def: int(1), exp: "errNoData"},
		// Value
		"Value_Pass":        {typ: Int, def: int(1), exp: "pass"},
		"Value_FailDefault": {typ: Int, def: "wrongType", exp: "fail"},
		"Value_NotThere":    {typ: Int, def: int(1), exp: "errNoKey"},
		"Value_Stored":      {typ: Int, def: "wrongType", exp: "fail"},
		"Value_NoData":      {typ: Int, def: int(1), exp: "errNoData"},
		// IntVar
		"IntVar_Pass":        {typ: IntVar, v: new(int), def: int(1), exp: "pass"},
		"IntVar_FailVar":     {typ: IntVar, v: new(string), def: int(1), exp: "fail"},
		"IntVar_FailDefault": {typ: IntVar, v: new(int), def: "wrongType", exp: "fail"},
		// ValueInt64
		"ValueInt64_Pass":        {typ: Int64, def: int64(1), exp: "pass"},
		"ValueInt64_FailDefault": {typ: Int64, def: "wrongType", exp: "fail"},
		"ValueInt64_ErrNotThere": {typ: Int64, def: int64(1), exp: "errNoKey"},
		"ValueInt64_ErrStored":   {typ: Int64, def: "wrongType", exp: "fail"},
		"ValueInt64_ErrNoData":   {typ: Int64, def: int64(1), exp: "errNoData"},
		// Int64Var
		"Int64Var_Pass":        {typ: Int64Var, v: new(int64), def: int64(1), exp: "pass"},
		"Int64Var_FailVar":     {typ: Int64Var, v: new(string), def: int64(1), exp: "fail"},
		"Int64Var_FailDefault": {typ: Int64Var, v: new(int64), def: "wrongType", exp: "fail"},
		// ValueUint
		"ValueUint_Pass":        {typ: Uint, def: uint(1), exp: "pass"},
		"ValueUint_FailDefault": {typ: Uint, def: "wrongType", exp: "fail"},
		"ValueUint_ErrNotThere": {typ: Uint, def: uint(1), exp: "errNoKey"},
		"ValueUint_ErrStored":   {typ: Uint, def: "wrongType", exp: "fail"},
		"ValueUint_ErrNoData":   {typ: Uint, def: uint(1), exp: "errNoData"},
		// UintVar
		"UintVar_Pass":        {typ: UintVar, v: new(uint), def: uint(1), exp: "pass"},
		"UintVar_FailVar":     {typ: UintVar, v: new(string), def: uint(1), exp: "fail"},
		"UintVar_FailDefault": {typ: UintVar, v: new(uint), def: "wrongType", exp: "fail"},
		// ValueUint64
		"ValueUint64_Pass":        {typ: Uint64, def: uint64(1), exp: "pass"},
		"ValueUint64_FailDefault": {typ: Uint64, def: "wrongType", exp: "fail"},
		"ValueUint64_ErrNotThere": {typ: Uint64, def: uint64(1), exp: "errNoKey"},
		"ValueUint64_ErrStored":   {typ: Uint64, def: "wrongType", exp: "fail"},
		"ValueUint64_ErrNoData":   {typ: Uint64, def: uint64(1), exp: "errNoData"},
		// Uint64Var
		"Uint64VarPass":        {typ: Uint64Var, v: new(uint64), def: uint64(1), exp: "pass"},
		"Uint64VarFailVar":     {typ: Uint64Var, v: new(string), def: uint64(1), exp: "fail"},
		"Uint64VarFailDefault": {typ: Uint64Var, v: new(uint64), def: "wrongType", exp: "fail"},
		// ValueFloat64
		"ValueFloat64Pass":        {typ: Float64, def: float64(1), exp: "pass"},
		"ValueFloat64FailDefault": {typ: Float64, def: "wrongType", exp: "fail"},
		"ValueFloat64ErrNotThere": {typ: Float64, def: float64(1), exp: "errNoKey"},
		"ValueFloat64ErrStored":   {typ: Float64, def: "wrongType", exp: "fail"},
		"ValueFloat64ErrNoData":   {typ: Float64, def: float64(1), exp: "errNoData"},
		// Float64Var
		"Float64VarPass":        {typ: Float64Var, v: new(float64), def: float64(1), exp: "pass"},
		"Float64VarFailVar":     {typ: Float64Var, v: new(string), def: float64(1), exp: "fail"},
		"Float64VarFailDefault": {typ: Float64Var, v: new(float64), def: "wrongType", exp: "fail"},
		// ValueString
		"ValueStringPass":        {typ: String, def: "string", exp: "pass"},
		"ValueStringFailDefault": {typ: String, def: 1, exp: "fail"},
		"ValueStringErrNotThere": {typ: String, def: "string", exp: "errNoKey"},
		"ValueStringErrStored":   {typ: String, def: 1, exp: "fail"},
		"ValueStringErrNoData":   {typ: String, def: "string", exp: "errNoData"},
		// StringVar
		"StringVarPass":        {typ: StringVar, v: &str, def: "string", exp: "pass"},
		"StringVarFailVar":     {typ: StringVar, v: new(int), def: "string", exp: "fail"},
		"StringVarFailDefault": {typ: StringVar, v: &str, def: 1, exp: "fail"},
		// ValueBool
		"ValueBoolPass":        {typ: Bool, def: true, exp: "pass"},
		"ValueBoolFailDefault": {typ: Bool, def: "wrongType", exp: "fail"},
		"ValueBoolErrNotThere": {typ: Bool, def: true, exp: "errNoKey"},
		"ValueBoolErrStored":   {typ: Bool, def: "wrongType", exp: "fail"},
		"ValueBoolErrNoData":   {typ: Bool, def: true, exp: "errNoData"},
		// BoolVar
		"BoolVarPass":        {typ: BoolVar, v: &b, def: true, exp: "pass"},
		"BoolVarFailVar":     {typ: BoolVar, v: new(string), def: true, exp: "fail"},
		"BoolVarFailDefault": {typ: BoolVar, v: &b, def: 1, exp: "fail"},
		// ValueDuration
		"ValueDurationPass":        {typ: Duration, def: time.Duration(0), exp: "pass"},
		"ValueDurationFailDefault": {typ: Duration, def: "wrongType", exp: "fail"},
		"ValueDurationErrNotThere": {typ: Duration, def: time.Duration(0), exp: "errNoKey"},
		"ValueDurationErrStored":   {typ: Duration, def: "wrongType", exp: "fail"},
		"ValueDurationErrNoData":   {typ: Duration, def: time.Duration(0), exp: "errNoData"},
		// DurationVar
		"DurationVarPass":        {typ: DurationVar, v: &d, def: d, exp: "pass"},
		"DurationVarFailVar":     {typ: DurationVar, v: new(string), def: d, exp: "fail"},
		"DurationVarFailDefault": {typ: DurationVar, v: &d, def: 1, exp: "fail"},
		// Var
		"VarPass":        {typ: Var, def: testValue{}, value: &testValue{str: "string"}, exp: "pass"},
		"VarFail":        {typ: Var, def: testValue{}, value: nil, exp: "fail"},
		"VarFailDefault": {typ: Var, def: "wrongType", exp: "fail"},
		// Nil
		"NilFail": {typ: Nil, def: nil, value: nil, exp: "fail"},
		// Default
		"DefaultFail": {typ: Default, def: nil, value: nil, exp: "fail"},
	}

	// TODO check whether this test is required as it may be a
	// duplicate test now that the global Config value has been
	// removed.
	for name, opt := range options {
		c = Config{}
		cmd := c.defaultSet("Usage heading", "cmd's heading")
		opts := []Option{
			{
				Type:     opt.typ,
				Flag:     "one",
				Usage:    "do it like this",
				Default:  opt.def,
				Var:      opt.v,
				Value:    opt.value,
				Commands: cmd,
			},
		}
		switch opt.exp {
		case "pass":
			err := c.Compose(opts...)
			if err != nil {
				t.Errorf("%s: %s: error: %s", fname, name, err)
			}
			err = c.Parse()
			if err != nil {
				t.Errorf("%s: %s: error: %s", fname, name, err)
			}
			switch opt.typ {
			case Int:
				i, err := c.ValueInt("one")
				if err != nil {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
				if i != 1 {
					t.Errorf("%s: %s: received %d expected 1",
						fname, name, i)
				}
				// Value
				in, typ, err := c.Value("one")
				if err != nil {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
				switch typ {
				case Int:
					if *in.(*int) != 1 {
						t.Errorf("%s: %s: error: %s", fname, name, err)
					}
				default:
					t.Errorf("%s: %s: end of case stament reached", fname, name)
				}
			case IntVar:
				v := *opt.v.(*int)
				if v != 1 {
					t.Errorf("%s: %s: received %d expected 1",
						fname, name, v)
				}
			case Int64:
				i, err := c.ValueInt64("one")
				if err != nil {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
				if i != 1 {
					t.Errorf("%s: %s: received %d expected 1",
						fname, name, i)
				}
			case Int64Var:
				v := *opt.v.(*int64)
				if v != 1 {
					t.Errorf("%s: %s: received %d expected 1",
						fname, name, v)
				}
			case Uint:
				i, err := c.ValueUint("one")
				if err != nil {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
				if i != 1 {
					t.Errorf("%s: %s: received %d expected 1",
						fname, name, i)
				}
			case UintVar:
				v := *opt.v.(*uint)
				if v != 1 {
					t.Errorf("%s: %s: received %d expected 1",
						fname, name, v)
				}
			case Uint64:
				i, err := c.ValueUint64("one")
				if err != nil {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
				if i != 1 {
					t.Errorf("%s: %s: received %d expected 1",
						fname, name, i)
				}
			case Uint64Var:
				v := *opt.v.(*uint64)
				if v != 1 {
					t.Errorf("%s: %s: received %d expected 1",
						fname, name, v)
				}
			case Float64:
				i, err := c.ValueFloat64("one")
				if err != nil {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
				if i != 1 {
					t.Errorf("%s: %s: received %f expected 1",
						fname, name, i)
				}
			case Float64Var:
				v := *opt.v.(*float64)
				if v != 1 {
					t.Errorf("%s: %s: received %f expected 1",
						fname, name, v)
				}
			case String:
				i, err := c.ValueString("one")
				if err != nil {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
				if i != "string" {
					t.Errorf("%s: %s: received %s expected \"strint\"",
						fname, name, i)
				}
			case StringVar:
				v := *opt.v.(*string)
				if v != "string" {
					t.Errorf("%s: %s: received %s expected \"string\"",
						fname, name, v)
				}
			case Bool:
				i, err := c.ValueBool("one")
				if err != nil {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
				if i != true {
					t.Errorf("%s: %s: received %t expected \"true\"",
						fname, name, i)
				}
			case BoolVar:
				v := *opt.v.(*bool)
				if v != true {
					t.Errorf("%s: %s: received %t expected \"true\"",
						fname, name, v)
				}
			case Duration:
				i, err := c.ValueDuration("one")
				if err != nil {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
				if i != time.Duration(0) {
					t.Errorf("%s: %s: received %s expected \"0s\"",
						fname, name, i)
				}
			case DurationVar:
				v := *opt.v.(*time.Duration)
				if v != time.Duration(0) {
					t.Errorf("%s: %s: received %s expected \"0s\"",
						fname, name, v)
				}
			case Var:
				v := *opt.value.(*testValue)
				if v.str != "string" {
					t.Errorf("%s: %s: received %s expected \"srting\"",
						fname, name, v.str)
				}
			default:
				t.Errorf("%s: %s: end of case stament reached", fname, name)
			}
		case "fail":
			// Both Options and Parse return an errConfig.
			err := c.Compose(opts...)
			if !errors.Is(err, errConfig) {
				t.Errorf("%s: %s: error: %s", fname, name, err)
			}
			err = c.Parse()
			if !errors.Is(err, errConfig) {
				t.Errorf("%s: %s: error: %s", fname, name, err)
			}
			// The errors raised in Options and Parse are put
			// into option.Err and wrapped with errStored
			// which is returned here.
			switch opt.typ {
			case Int:
				_, err = c.ValueInt("one")
				if !errors.Is(err, errType) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
				// Value
				_, _, err = c.Value("one")
				if !errors.Is(err, errType) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Int64:
				_, err = c.ValueInt64("one")
				if !errors.Is(err, errType) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Uint:
				_, err = c.ValueUint("one")
				if !errors.Is(err, errType) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Uint64:
				_, err = c.ValueUint64("one")
				if !errors.Is(err, errType) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Float64:
				_, err = c.ValueFloat64("one")
				if !errors.Is(err, errType) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case String:
				_, err = c.ValueString("one")
				if !errors.Is(err, errType) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Bool:
				_, err = c.ValueBool("one")
				if !errors.Is(err, errType) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Duration:
				_, err = c.ValueDuration("one")
				if !errors.Is(err, errType) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Nil:
				_, typ, err := c.Value("one")
				switch typ {
				case Nil:
					if !errors.Is(err, errTypeNil) {
						t.Errorf("%s: %s: error: %s", fname, name, err)
					}
				default:
					t.Errorf("%s: %s: end of case stament reached", fname, name)
				}
			case Default:
				_, typ, err := c.Value("one")
				switch typ {
				case Default:
					if !errors.Is(err, errType) {
						t.Errorf("%s: %s: error: %s", fname, name, err)
					}
				default:
					t.Errorf("%s: %s: end of case stament reached", fname, name)
				}
			case IntVar, Int64Var, UintVar, Uint64Var,
				Float64Var, StringVar, BoolVar, DurationVar, Var:
			default:
				t.Errorf("%s: %s: end of case stament reached", fname, name)
			}
		case "errNoData", "errNoKey":
			// errNoKey requirs that Options and Parse not be
			// called, whereas errNoData requires that they
			// are called and then the option data removed.
			var err error
			if opt.exp == "errNoData" {
				err = c.Compose(opts...)
				if err != nil {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
				err = c.Parse()
				if err != nil {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
				c.options["one"].data = nil
			}
			switch opt.typ {
			case Int:
				_, err = c.ValueInt("one")
				if !(errors.Is(err, errNoData) || errors.Is(err, errNoKey)) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
				// Value
				_, _, err = c.Value("one")
				if !(errors.Is(err, errNoData) || errors.Is(err, errNoKey)) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Int64:
				_, err = c.ValueInt64("one")
				if !(errors.Is(err, errNoData) || errors.Is(err, errNoKey)) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Uint:
				_, err := c.ValueUint("one")
				if !(errors.Is(err, errNoData) || errors.Is(err, errNoKey)) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Uint64:
				_, err = c.ValueUint64("one")
				if !(errors.Is(err, errNoData) || errors.Is(err, errNoKey)) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Float64:
				_, err = c.ValueFloat64("one")
				if !(errors.Is(err, errNoData) || errors.Is(err, errNoKey)) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case String:
				_, err = c.ValueString("one")
				if !(errors.Is(err, errNoData) || errors.Is(err, errNoKey)) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Bool:
				_, err = c.ValueBool("one")
				if !(errors.Is(err, errNoData) || errors.Is(err, errNoKey)) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Duration:
				_, err = c.ValueDuration("one")
				if !(errors.Is(err, errNoData) || errors.Is(err, errNoKey)) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			default:
				t.Errorf("%s: %s: end of case stament reached", fname, name)
			}
		default:
			t.Errorf("%s: %s: end of case stament reached", fname, name)
		}
	}
}

func TestConfigArgString(t *testing.T) {
	const fname = "TestConfigArgString"
	temp := os.Args
	os.Args = []string{"one", "two", "three"}
	c = Config{}
	c.defaultSet("", "")
	err := c.Compose()
	if errors.Is(err, errConfig) {
		t.Errorf("%s: %w", fname, err)
	}
	str := c.Args()
	if str != "one two three" {
		t.Errorf("%s: recieved %q expected \"one two three\"",
			fname, str)
	}
	os.Args = temp
}

func TestOptionsCheckUserFn(t *testing.T) {
	const fname = "TestOptionsCheckUserFn"
	config := Config{}
	m := config.defaultSet("", "")
	var opts = []Option{
		{
			Type:     Int,
			Flag:     "one",
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
		t.Errorf("%s: error: %s", fname, err)
	}
	err = config.Parse()
	if err != nil {
		t.Errorf("%s: error: %s", fname, err)
	}
	i, err := config.ValueInt("one")
	if err != nil {
		t.Errorf("%s: error: %s", fname, err)
	}
	if i != 2 {
		t.Errorf("%s: recieved %d expected 2", fname, i)
	}
}

func TestOptionsCheckUserFnError(t *testing.T) {
	const fname = "TestOptionsCheckUserFnError"
	config := Config{}
	m := config.defaultSet("", "")
	var opts = []Option{
		{
			Type:     Int,
			Flag:     "one",
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
	if err != nil {
		t.Errorf("%s: error: %s", fname, err)
	}
	err = config.Parse()
	if !errors.Is(err, ErrCheck) {
		t.Errorf("%s: error: %s", fname, err)
	}
	_, err = config.ValueInt("one")
	if !errors.Is(err, ErrCheck) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

func TestOptionsCheckName(t *testing.T) {
	const fname = "TestOptionsCheckName"
	config := Config{}
	m := config.defaultSet("", "")
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
		t.Errorf("%s: error: %s", fname, err)
	}
}

func TestOptionsCheckFlagPresent(t *testing.T) {
	const fname = "TestOptionsCheckFlagPresent"
	config := Config{}
	m := config.defaultSet("", "")
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
		t.Errorf("%s: error: %s", fname, err)
	}
	err = config.Parse()
	if err != nil {
		t.Errorf("%s: error: %s", fname, err)
	}
}

func TestOptionsCheckFlagDuplicate(t *testing.T) {
	const fname = "TestOptionsCheckFlagDuplicate"
	config := Config{}
	m := config.defaultSet("", "")
	//m2 := config.Command("modetwo", "")
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
		t.Errorf("%s: error: %s", fname, err)
	}
	err = config.Parse()
	if err != nil {
		t.Errorf("%s: error: %s", fname, err)
	}
}

func TestOptionsEdgeCaseNoArgs(t *testing.T) {
	const fname = "TestOptionsEdgeCaseNoArgs"
	temp := os.Args
	os.Args = os.Args[:0]
	config := Config{}
	m := config.defaultSet("", "")
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
		t.Errorf("%s: error: %s", fname, err)
	}
	err = config.Parse()
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	os.Args = temp
}

func TestCommandGetCmd(t *testing.T) {
	const fname = "TestCommandGetCmd"
	c = Config{}
	cmd := c.defaultSet("", "")
	cmd2 := c.Command("", "")
	var opts = []Option{
		{
			Type:     Int,
			Flag:     "a",
			Usage:    "like this",
			Default:  1,
			Commands: cmd,
		},
	}
	err := c.Compose(opts...)
	if err != nil {
		t.Errorf("%s: this case should not raise an error: %s",
			fname, err)
	}
	err = c.Parse()
	if err != nil {
		t.Errorf("%s: this case should not raise an error: %s",
			fname, err)
	}
	mode, _ := c.Is()
	if mode != "default" {
		t.Errorf("%s: expected \"default\" received %q",
			fname, mode)
	}
	if !c.isInSet(cmd2) {
		t.Errorf("%s: not a valid Command token", fname)
	}
}

func TestCommandDuplicateKeys(t *testing.T) {
	const fname = "TestCommandDuplicateKeys"
	config := Config{}
	m1 := config.defaultSet("", "")
	m2 := config.Command("modetwo", "")
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
			Usage:    "like this",
			Default:  1,
			Commands: m2,
		},
	}
	err := config.Compose(opts...)
	if err != nil {
		t.Errorf("%s: this case should not raise an error: %s",
			fname, err)
	}
	err = config.Parse()
	if err != nil {
		t.Errorf("%s: this case should not raise an error: %s",
			fname, err)
	}
}

func TestCommandTooMany(t *testing.T) {
	const fname = "TestCommandTooMany"
	config := Config{}
	_ = config.defaultSet("", "")
	names := make([]string, 65)
	for i := 0; i <= 64; i++ {
		names[i] = fmt.Sprint(i + '0')
	}
	for i := 0; i <= 64; i++ {
		_ = config.Command(names[i], "")
	}
	err := config.Compose()
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

func TestCommandNotThere(t *testing.T) {
	const fname = "TestCommandNotThere"
	config := Config{}
	_ = config.defaultSet("", "")
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
	err := config.Compose(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

func TestCommandTokens(t *testing.T) {
	const fname = "TestCommandTokenIs"
	config := Config{}
	cmd1 := config.defaultSet("", "")
	cmd2 := config.Command("", "")
	cmd3 := config.Command("", "")
	var opts = []Option{
		{
			Type:     Int,
			Flag:     "a",
			Usage:    "like this",
			Default:  1,
			Commands: cmd1,
		},
	}
	err := config.Compose(opts...)
	if err != nil {
		t.Errorf("%s: %s", fname, err)
	}
	err = config.Parse()
	if err != nil {
		t.Errorf("%s: %s", fname, err)
	}
	v := config.isInSet(0)
	if v {
		t.Errorf("%s: received true expected false", fname)
	}
	v = config.isInSet(cmd1)
	if !v {
		t.Errorf("%s: received false expected true", fname)
	}
	v = config.isInSet(cmd2)
	if !v {
		t.Errorf("%s: received false expected true", fname)
	}
	v = config.isInSet(cmd3)
	if !v {
		t.Errorf("%s: received false expected true", fname)
	}
	v = config.isInSet(cmd1 | cmd3)
	if !v {
		t.Errorf("%s: received false expected true", fname)
	}
	v = config.isInSet(cmd1 | cmd2 | cmd3)
	if !v {
		t.Errorf("%s: received false expected true", fname)
	}
	v = config.isInSet(config.position)
	if v {
		t.Errorf("%s: received true expected false", fname)
	}
}

func TestParse(t *testing.T) {
	const fname = "TestParse"
	c = Config{}
	cmd := c.defaultSet("", "")
	cmd2 := c.Command("cmd2", "")
	temp := os.Args[1]
	os.Args[1] = "cmd2"
	var opts = []Option{
		{
			Type:     Int,
			Flag:     "a",
			Usage:    "like this",
			Default:  1,
			Commands: cmd2,
		},
	}
	err := c.Compose(opts...)
	if err != nil {
		t.Errorf("%s: this case should not raise an error: %s",
			fname, err)
	}
	err = c.Parse()
	os.Args[1] = temp
	if err != nil {
		t.Errorf("%s: this case should not raise an error: %s",
			fname, err)
	}
	mode, _ := c.Is()
	if mode != "cmd2" {
		t.Errorf("%s: expected \"default\" received %q",
			fname, mode)
	}
	if !c.isInSet(cmd) {
		t.Errorf("%s: not a valid Command token", fname)
	}
}

func TestParseInvalidCmd(t *testing.T) {
	const fname = "TestParseInvalidCmd"
	c = Config{}
	cmd := c.defaultSet("", "")
	cmd2 := c.Command("cmd2", "")
	temp := os.Args[1]
	os.Args[1] = "unknownCmd"
	var opts = []Option{
		{
			Type:     Int,
			Flag:     "a",
			Usage:    "like this",
			Default:  1,
			Commands: cmd2,
		},
	}
	err := c.Compose(opts...)
	if err != nil {
		t.Errorf("%s: error: %s", fname, err)
	}
	err = c.Parse()
	os.Args[1] = temp
	if !errors.Is(err, errNotFound) {
		t.Errorf("%s: error: %s", fname, err)
	}
	mode, _ := c.Is()
	if mode != "" {
		t.Errorf("%s: expected \"cmd2\" received %q",
			fname, mode)
	}
	if !c.isInSet(cmd) {
		t.Errorf("%s: %s", fname, errNotValid)
	}
}

func TestFlagSetUsageFn(t *testing.T) {
	const fname = "TestFlagSetUsageFn"
	config := Config{}
	cmd := config.defaultSet("Usage Heading", "Mode Heading")
	opts := []Option{
		{
			Type:     Int,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  1,
			Commands: cmd,
		},
		{
			Type:     Int,
			Flag:     "flagWithAVeryLongName",
			Usage:    "do it like this",
			Default:  1,
			Commands: cmd,
		},
	}
	err := config.Compose(opts...)
	if err != nil {
		t.Errorf("%s: error: %s", fname, err)
	}
	err = config.Parse()
	if err != nil {
		t.Errorf("%s: error: %s", fname, err)
	}
	fn := config.setUsageFn(nil)
	fn = config.setUsageFn(ioutil.Discard)
	fn()
}
