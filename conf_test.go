package conf

import (
	"errors"
	"flag"
	"fmt"
	"testing"
	"time"
)

type testValue struct {
	str string
}

func (t testValue) String() string {
	return t.str
}

func (t *testValue) Set(str string) error {
	return nil
}

func TestTooManyCmds(t *testing.T) {
	const fname = "TestTooManyCmds"
	config := Config{}
	_ = config.Setup("", "")
	names := make([]string, 65)
	for i := 0; i <= 64; i++ {
		names[i] = fmt.Sprint(i + '0')
	}
	for i := 0; i <= 64; i++ {
		_ = config.Command(names[i], "")
	}
	err := config.Options()
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

func TestConfigOption(t *testing.T) {
	const fname = "TestConfigOption"
	var str = "string"
	var b bool
	var d time.Duration
	options := map[string]struct {
		typ   Type
		def   interface{}
		v     interface{}
		value flag.Value
		exp   string
	}{
		// ValueInt
		"ValueIntPass":        {typ: Int, def: int(1), exp: "pass"},
		"ValueIntFailDefault": {typ: Int, def: "wrongType", exp: "fail"},
		"ValueIntErrNotThere": {typ: Int, def: int(1), exp: "errNoKey"},
		"ValueIntErrStored":   {typ: Int, def: "wrongType", exp: "fail"},
		"ValueIntErrNoData":   {typ: Int, def: int(1), exp: "errNoData"},
		// Value
		"ValuePass":        {typ: Int, def: int(1), exp: "pass"},
		"ValueFailDefault": {typ: Int, def: "wrongType", exp: "fail"},
		"ValueNotThere":    {typ: Int, def: int(1), exp: "errNoKey"},
		"ValueStored":      {typ: Int, def: "wrongType", exp: "fail"},
		"ValueNoData":      {typ: Int, def: int(1), exp: "errNoData"},

		// IntVar
		"IntVarPass":        {typ: IntVar, v: new(int), def: int(1), exp: "pass"},
		"IntVarFailVar":     {typ: IntVar, v: new(string), def: int(1), exp: "fail"},
		"IntVarFailDefault": {typ: IntVar, v: new(int), def: "wrongType", exp: "fail"},

		// ValueInt64
		"ValueInt64Pass":        {typ: Int64, def: int64(1), exp: "pass"},
		"ValueInt64FailDefault": {typ: Int64, def: "wrongType", exp: "fail"},
		"ValueInt64ErrNotThere": {typ: Int64, def: int64(1), exp: "errNoKey"},
		"ValueInt64ErrStored":   {typ: Int64, def: "wrongType", exp: "fail"},
		"ValueInt64ErrNoData":   {typ: Int64, def: int64(1), exp: "errNoData"},
		// Int64Var
		"Int64VarPass":        {typ: Int64Var, v: new(int64), def: int64(1), exp: "pass"},
		"Int64VarFailVar":     {typ: Int64Var, v: new(string), def: int64(1), exp: "fail"},
		"Int64VarFailDefault": {typ: Int64Var, v: new(int64), def: "wrongType", exp: "fail"},

		// ValueUint
		"ValueUintPass":        {typ: Uint, def: uint(1), exp: "pass"},
		"ValueUintFailDefault": {typ: Uint, def: "wrongType", exp: "fail"},
		"ValueUintErrNotThere": {typ: Uint, def: uint(1), exp: "errNoKey"},
		"ValueUintErrStored":   {typ: Uint, def: "wrongType", exp: "fail"},
		"ValueUintErrNoData":   {typ: Uint, def: uint(1), exp: "errNoData"},
		// UintVar
		"UintVarPass":        {typ: UintVar, v: new(uint), def: uint(1), exp: "pass"},
		"UintVarFailVar":     {typ: UintVar, v: new(string), def: uint(1), exp: "fail"},
		"UintVarFailDefault": {typ: UintVar, v: new(uint), def: "wrongType", exp: "fail"},

		// ValueUint64
		"ValueUint64Pass":        {typ: Uint64, def: uint64(1), exp: "pass"},
		"ValueUint64FailDefault": {typ: Uint64, def: "wrongType", exp: "fail"},
		"ValueUint64ErrNotThere": {typ: Uint64, def: uint64(1), exp: "errNoKey"},
		"ValueUint64ErrStored":   {typ: Uint64, def: "wrongType", exp: "fail"},
		"ValueUint64ErrNoData":   {typ: Uint64, def: uint64(1), exp: "errNoData"},
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

	for name, opt := range options {
		config := Config{}
		cmd := config.Setup("Usage Heading", "Mode Heading")
		opts := []Option{
			{Name: "one",
				Type:     opt.typ,
				Flag:     "i",
				Usage:    "do it like this",
				Default:  opt.def,
				Var:      opt.v,
				Value:    opt.value,
				Commands: cmd,
			},
		}
		switch opt.exp {
		case "pass":
			err := config.Options(opts...)
			if err != nil {
				t.Errorf("%s: %s: error: %s", fname, name, err)
			}
			err = config.Parse()
			if err != nil {
				t.Errorf("%s: %s: error: %s", fname, name, err)
			}
			switch opt.typ {
			case Int:
				i, err := config.ValueInt("one")
				if err != nil {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
				if i != 1 {
					t.Errorf("%s: %s: received %d expected 1",
						fname, name, i)
				}
				// Value
				in, typ, err := config.Value("one")
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
				i, err := config.ValueInt64("one")
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
				i, err := config.ValueUint("one")
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
				i, err := config.ValueUint64("one")
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
				i, err := config.ValueFloat64("one")
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
				i, err := config.ValueString("one")
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
				i, err := config.ValueBool("one")
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
				i, err := config.ValueDuration("one")
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
			err := config.Options(opts...)
			if !errors.Is(err, errConfig) {
				t.Errorf("%s: %s: error: %s", fname, name, err)
			}
			err = config.Parse()
			if !errors.Is(err, errConfig) {
				t.Errorf("%s: %s: error: %s", fname, name, err)
			}
			switch opt.typ {
			case Int:
				_, err = config.ValueInt("one")
				if !errors.Is(err, errStored) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
				// Value
				_, _, err = config.Value("one")
				if !errors.Is(err, errStored) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Int64:
				_, err = config.ValueInt64("one")
				if !errors.Is(err, errStored) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Uint:
				_, err = config.ValueUint("one")
				if !errors.Is(err, errStored) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Uint64:
				_, err = config.ValueUint64("one")
				if !errors.Is(err, errStored) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Float64:
				_, err = config.ValueFloat64("one")
				if !errors.Is(err, errStored) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case String:
				_, err = config.ValueString("one")
				if !errors.Is(err, errStored) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Bool:
				_, err = config.ValueBool("one")
				if !errors.Is(err, errStored) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Duration:
				_, err = config.ValueDuration("one")
				if !errors.Is(err, errStored) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Nil:
				_, typ, err := config.Value("one")
				switch typ {
				case Nil:
					if !errors.Is(err, errStored) {
						t.Errorf("%s: %s: error: %s", fname, name, err)
					}
				}
			case Default:
				_, typ, err := config.Value("one")
				switch typ {
				case Default:
					if !errors.Is(err, errStored) {
						t.Errorf("%s: %s: error: %s", fname, name, err)
					}
				}
			case IntVar, Int64Var, UintVar, Uint64Var, Float64Var, StringVar, BoolVar, DurationVar, Var:
			default:
				t.Errorf("%s: %s: end of case stament reached", fname, name)
			}
		case "errNoKey":
			switch opt.typ {
			case Int:
				_, err := config.ValueInt("notThere")
				if !errors.Is(err, errNoKey) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
				// Value
				_, _, err = config.Value("notThere")
				if !errors.Is(err, errNoKey) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Int64:
				_, err := config.ValueInt64("notThere")
				if !errors.Is(err, errNoKey) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Uint:
				_, err := config.ValueUint("notThere")
				if !errors.Is(err, errNoKey) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Uint64:
				_, err := config.ValueUint64("notThere")
				if !errors.Is(err, errNoKey) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Float64:
				_, err := config.ValueFloat64("notThere")
				if !errors.Is(err, errNoKey) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case String:
				_, err := config.ValueString("notThere")
				if !errors.Is(err, errNoKey) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Bool:
				_, err := config.ValueBool("notThere")
				if !errors.Is(err, errNoKey) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Duration:
				_, err := config.ValueDuration("notThere")
				if !errors.Is(err, errNoKey) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			default:
				t.Errorf("%s: %s: end of case stament reached", fname, name)
			}
		case "errNoData":
			err := config.Options(opts...)
			if err != nil {
				t.Errorf("%s: %s: error: %s", fname, name, err)
			}
			err = config.Parse()
			if err != nil {
				t.Errorf("%s: %s: error: %s", fname, name, err)
			}
			config.options["one"].data = nil
			switch opt.typ {
			case Int:
				_, err = config.ValueInt("one")
				if !errors.Is(err, errNoData) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
				// Value
				_, _, err = config.Value("one")
				if !errors.Is(err, errNoData) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Int64:
				_, err = config.ValueInt64("one")
				if !errors.Is(err, errNoData) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Uint:
				_, err = config.ValueUint("one")
				if !errors.Is(err, errNoData) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Uint64:
				_, err = config.ValueUint64("one")
				if !errors.Is(err, errNoData) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Float64:
				_, err = config.ValueFloat64("one")
				if !errors.Is(err, errNoData) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case String:
				_, err = config.ValueString("one")
				if !errors.Is(err, errNoData) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Bool:
				_, err = config.ValueBool("one")
				if !errors.Is(err, errNoData) {
					t.Errorf("%s: %s: error: %s", fname, name, err)
				}
			case Duration:
				_, err = config.ValueDuration("one")
				if !errors.Is(err, errNoData) {
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
