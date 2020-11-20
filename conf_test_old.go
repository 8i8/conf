package conf

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
)

func init() {
	test = true
}

var (
	m1     = Setup("Usage Heading", "Mode Heading")
	m2     = Command("one", "one's heading")
	m3     = Command("two", "two's heading")
	global Config
	cm1    = global.Setup("Usage Heading", "Mode Heading")
	cm2    = global.Command("cone", "cone's heading")
	cm3    = global.Command("ctwo", "ctwo's heading")
	ptInt  int
)

var opts = []Option{
	{Name: "int",
		Type:     Int,
		Flag:     "a",
		Usage:    "like this",
		Default:  1,
		Commands: m1 | m2 | m3,
	},
	{Name: "intVar",
		Type:     IntVar,
		Flag:     "b",
		Usage:    "do it like this",
		Default:  2345,
		Var:      &ptInt,
		Commands: m1,
	},
	{Name: "intNoDefault",
		Type:     Int,
		Flag:     "c",
		Usage:    "like that",
		Commands: m1,
	},
	{Name: "int64",
		Type:     Int64,
		Flag:     "d",
		Usage:    "like this",
		Default:  int64(1),
		Commands: m1 | m2 | m3,
	},
	{Name: "uint",
		Type:     Uint,
		Flag:     "e",
		Usage:    "like this",
		Default:  uint(1),
		Commands: m1 | m2 | m3,
	},
	{Name: "uint64",
		Type:     Uint64,
		Flag:     "f",
		Usage:    "like this",
		Default:  uint64(1),
		Commands: m1 | m2 | m3,
	},
	{Name: "float64",
		Type:     Float64,
		Flag:     "g",
		Usage:    "like this",
		Default:  float64(1),
		Commands: m1 | m2 | m3,
	},
	{Name: "string",
		Type:     String,
		Flag:     "h",
		Usage:    "like this",
		Default:  "string",
		Commands: m1 | m2 | m3,
	},
	{Name: "bool",
		Type:     Bool,
		Flag:     "i",
		Usage:    "like this",
		Default:  true,
		Commands: m1 | m2 | m3,
	},
	{Name: "duration",
		Type:     Duration,
		Flag:     "j",
		Usage:    "like this",
		Default:  time.Duration(0),
		Commands: m1 | m2 | m3,
	},
}

var optsmap = make(map[string][]Option)

func init() {
}

func TestDoubleNameError(t *testing.T) {
	const fname = "TestDoubleNameError"
	config := Config{}
	m := config.Setup("", "")
	var opts = []Option{
		{Name: "errors",
			Type:     Int,
			Flag:     "d",
			Usage:    "like this",
			Default:  1,
			Commands: m,
		},
		{Name: "errors",
			Type:     Int,
			Flag:     "d",
			Usage:    "like this",
			Default:  1,
			Commands: m,
		},
	}

	err := config.Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

func TestEmptyNameError(t *testing.T) {
	const fname = "TestEmptyNameError"
	config := Config{}
	m := config.Setup("", "")
	var opts = []Option{
		{Name: "",
			Type:     Int,
			Flag:     "d",
			Usage:    "like this",
			Default:  1,
			Commands: m,
		},
		{Name: "errors",
			Type:     Int,
			Flag:     "d",
			Usage:    "like this",
			Default:  1,
			Commands: m,
		},
		{Name: "errors",
			Type:     Int,
			Flag:     "d",
			Usage:    "like this",
			Default:  1,
			Commands: m,
		},
	}
	err := config.Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

func TestMultipleErrora(t *testing.T) {
	const fname = "TestMultipleErrora"
	config := Config{}
	m := config.Setup("", "")
	var opts = []Option{
		{Name: "",
			Type:     Int,
			Flag:     "d",
			Usage:    "like this",
			Default:  1,
			Commands: m,
		},
	}
	err := config.Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

func TestSaveArgs(t *testing.T) {
	const fname = "TestSaveArgs"
	temp := os.Args
	os.Args = os.Args[:0]
	config := Config{}
	m := config.Setup("", "")
	var opts = []Option{
		{Name: "int",
			Type:     Int,
			Flag:     "a",
			Usage:    "like this",
			Default:  1,
			Commands: m,
		},
	}
	err := config.Options(opts...)
	if err == nil {
		t.Errorf("%s: error: %s", fname, err)
	}
	os.Args = temp
}

// TestNames test that no two options can have the same name.
func TestNames(t *testing.T) {
	const fname = "TestNames"
	config := Config{}
	m := config.Setup("", "")
	var similarNames = []Option{
		{Name: "int",
			Type:     Int,
			Flag:     "a",
			Usage:    "like this",
			Default:  1,
			Commands: m,
		},
		{Name: "int",
			Type:     Int,
			Flag:     "b",
			Usage:    "like this",
			Default:  1,
			Commands: m,
		},
	}
	err := config.Options(similarNames...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

func TestParseSubCommand(t *testing.T) {
	const fname = "TestParseSubCommand"
	config := Config{}
	_ = config.Setup("Usage Heading", "Mode Heading")
	_ = config.Command("cmd", "")
	temp := os.Args
	os.Args = []string{"test", "cmd"}
	err := config.Parse()
	if err != nil {
		t.Errorf("%s: error: %s", fname, err)
	}
	os.Args = []string{"test", "notThere"}
	err = config.Parse()
	if err == nil {
		t.Errorf("%s: expected and error", fname)
	}
	os.Args = temp
}

func TestFlagSetUsageFn(t *testing.T) {
	const fname = "TestFlagSetUsageFn"
	config := Config{}
	cmd := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     Int,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  1,
			Commands: cmd,
		},
		{Name: "two",
			Type:     Int,
			Flag:     "flagWithAVeryLongName",
			Usage:    "do it like this",
			Default:  1,
			Commands: cmd,
		},
	}
	err := config.Options(opts...)
	if err != nil {
		t.Errorf("%s: error: %s", fname, err)
	}
	err = config.Parse()
	if err != nil {
		t.Errorf("%s: error: %s", fname, err)
	}
	b := bytes.Buffer{}
	buf := bufio.NewWriter(&b)
	fn := config.setUsageFn(buf)
	fn()
}

func TestFlagDuplicateNoPanic(t *testing.T) {
	const fname = "TestFlagDuplicateNoPanic"
	config := Config{}
	cmd := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     Int,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  1,
			Commands: cmd,
		},
		{Name: "two",
			Type:     Int,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  1,
			Commands: cmd,
		},
	}
	err := config.Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	err = config.Parse()
	if err != nil {
		t.Errorf("%s: error: %s", fname, err)
	}
}

// TestNamesModeNoError test that no two options can have the same name,
// even when in different modes.
func TestNamesModeNoError(t *testing.T) {
	const fname = "TestNamesModeNoError"
	config := Config{}
	m1 := config.Setup("", "")
	m2 := config.Command("modetwo", "")
	var opts = []Option{
		{Name: "int",
			Type:     Int,
			Flag:     "a",
			Usage:    "like this",
			Default:  1,
			Commands: m1,
		},
		{Name: "int",
			Type:     Int,
			Flag:     "b",
			Usage:    "like this",
			Default:  1,
			Commands: m2,
		},
	}
	err := config.Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

func TestCheckFn(t *testing.T) {
	const fname = "TestCheckFn"
	config := Config{}
	m := config.Setup("", "")
	var opts = []Option{
		{Name: "int1",
			Type:     Int,
			Flag:     "a",
			Usage:    "like this",
			Default:  1,
			Commands: m,
			Check: func(v interface{}) (interface{}, error) {
				i := *v.(*int)
				i++
				return &i, nil
			},
		},
		{Name: "int2",
			Type:     Int,
			Flag:     "b",
			Usage:    "like this",
			Default:  1,
			Commands: m,
			Check: func(v interface{}) (interface{}, error) {
				i := *v.(*int)
				if i == 1 {
					return v, fmt.Errorf("to much 1")
				}
				return v, nil
			},
		},
	}
	err := config.Options(opts...)
	if err != nil {
		t.Errorf("%s: error: %s", fname, err)
	}
	err = config.Parse()
	if !errors.Is(err, errCheck) {
		t.Errorf("%s: error: %s", fname, err)
	}

	i, err := config.ValueInt("int1")
	if err != nil {
		t.Errorf("%s: error: %s", fname, err)
	}
	if i != 2 {
		t.Errorf("%s: received %d expected 2", fname, i)
	}

	_, err = config.ValueInt("int2")
	if !errors.Is(err, errStored) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

// TestKeys tests that duplicate keys return an error from the Options
// function.
func TestKeys(t *testing.T) {
	const fname = "TestKeys"
	config := Config{}
	m := config.Setup("", "")
	var similarKeys = []Option{
		{Name: "int",
			Type:     Int,
			Flag:     "a",
			Usage:    "like this",
			Default:  1,
			Commands: m,
		},
		{Name: "similarKeys",
			Type:     Int,
			Flag:     "a",
			Usage:    "like this",
			Default:  1,
			Commands: m,
		},
	}
	err := config.Options(similarKeys...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

func TestKeyNoValue(t *testing.T) {
	const fname = "TestKeyNoValue"
	config := Config{}
	m := config.Setup("", "")
	var opts = []Option{
		{Name: "int",
			Type:     Int,
			Usage:    "like this",
			Default:  1,
			Commands: m,
		},
	}
	err := config.Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

func TestModesNotThere(t *testing.T) {
	const fname = "TestModesNotThere"
	config := Config{}
	_ = config.Setup("", "")
	m := cmd(2)
	var opts = []Option{
		{Name: "int",
			Type:     Int,
			Flag:     "a",
			Usage:    "like this",
			Default:  1,
			Commands: m,
		},
	}
	err := config.Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

func TestSetCmdNotThere(t *testing.T) {
	const fname = "TestSetCmdNotThere"
	config := Config{}
	_ = config.Setup("", "")
	err := config.setCmd("notThere")
	if err == nil {
		t.Errorf("%s: expected an error", fname)
	}
}

func TestLoadCmd(t *testing.T) {
	const fname = "TestLoadCmd"
	config := Config{}
	_ = config.Setup("", "")
	err := config.loadCmd("notThere")
	if err == nil {
		t.Errorf("%s: expected an error", fname)
	}
}

// TestKeysModesSimilarKeys tests that similar keys can be used within
// differeing modes.
func TestKeysModesSimilarKeys(t *testing.T) {
	const fname = "TestKeysModesSimilarKeys"
	config := Config{}
	m1 := config.Setup("", "")
	m2 := config.Command("modetwo", "")
	var similarKeys = []Option{
		{Name: "int",
			Type:     Int,
			Flag:     "a",
			Usage:    "like this",
			Default:  1,
			Commands: m1,
		},
		{Name: "similarKeys",
			Type:     Int,
			Flag:     "a",
			Usage:    "like this",
			Default:  1,
			Commands: m2,
		},
	}
	err := config.Options(similarKeys...)
	if err != nil {
		t.Errorf("%s: error: %s", fname, err)
	}
}

func TestTypes(t *testing.T) {
	const fname = "TestTypes"
	err := Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	err = Parse()
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	err = global.Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	err = global.Parse()
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	t.Run("ArgList", testArgList)
	t.Run("c.ArgList", testArgListConfig)
	t.Run("Modes", testModes)
	t.Run("c.Modes", testModesConfig)
	t.Run("flagIs", testFlagIs)
	t.Run("Int", testInt)
	t.Run("IntVar", testIntVar)
}

func testArgList(t *testing.T) {
	const fname = "TestArgList"
	l := ArgString()
	if l == "" {
		t.Errorf("%s: received an empty string", fname)
	}
}

func testArgListConfig(t *testing.T) {
	const fname = "TestArgListConf"
	l := global.ArgString()
	if l == "" {
		t.Errorf("%s: received an empty string", fname)
	}
}

func testFlagIs(t *testing.T) {
	const fname = "TestFlagIs"
	v := c.cmdTokenIs(0)
	if v {
		t.Errorf("%s: received true expected false", fname)
	}
	v = c.cmdTokenIs(m1)
	if !v {
		t.Errorf("%s: received false expected true", fname)
	}
	v = c.cmdTokenIs(m2)
	if !v {
		t.Errorf("%s: received false expected true", fname)
	}
	v = c.cmdTokenIs(m3)
	if !v {
		t.Errorf("%s: received false expected true", fname)
	}
	v = c.cmdTokenIs(m1 | m3)
	if !v {
		t.Errorf("%s: received false expected true", fname)
	}
	v = c.cmdTokenIs(m1 | m2 | m3)
	if !v {
		t.Errorf("%s: received false expected true", fname)
	}
	v = c.cmdTokenIs(c.index)
	if v {
		t.Errorf("%s: received true expected false", fname)
	}

	v = global.cmdTokenIs(0)
	if v {
		t.Errorf("%s: received true expected false", fname)
	}
	v = global.cmdTokenIs(cm1)
	if !v {
		t.Errorf("%s: received false expected true", fname)
	}
	v = global.cmdTokenIs(cm2)
	if !v {
		t.Errorf("%s: received false expected true", fname)
	}
	v = global.cmdTokenIs(cm3)
	if !v {
		t.Errorf("%s: received false expected true", fname)
	}
	v = global.cmdTokenIs(cm1 | cm3)
	if !v {
		t.Errorf("%s: received false expected true", fname)
	}
	v = global.cmdTokenIs(cm1 | cm2 | cm3)
	if !v {
		t.Errorf("%s: received false expected true", fname)
	}
	v = global.cmdTokenIs(global.index)
	if v {
		t.Errorf("%s: received true expected false", fname)
	}
}

func testModes(t *testing.T) {
	const fname = "TestModes"
	m := GetCmd()
	if m != "default" {
		t.Errorf("%s: expected \"default\" received %q", fname, m)
	}
	if !c.cmdTokenIs(m1) {
		t.Errorf("%s: recived false expected true", fname)
	}
}

func testModesConfig(t *testing.T) {
	const fname = "TestModesConfig"
	m := c.GetCmd()
	if m != "default" {
		t.Errorf("%s: expected default received %s", fname, m)
	}
	if !global.cmdTokenIs(cm1) {
		t.Errorf("%s: recived false expected true", fname)
	}
}

func testInt(t *testing.T) {
	const fname = "testInt"
	i, err := ValueInt("int")
	if err != nil {
		t.Errorf("%s: error: %s", fname, err)
	}
	if i != 1 {
		t.Errorf("%s: received %d expected 1", fname, i)
	}
	i, err = ValueInt("intNoDefault")
	if err == nil {
		t.Errorf("%s: expected an error: %s", fname, err)
	}
	i, err = ValueInt("twoSimilarKeys")
	if err == nil {
		t.Errorf("%s: expected an error", fname)
	}
}

func testIntVar(t *testing.T) {
	const fname = "testIntVar"
	if ptInt != 2345 {
		t.Errorf("%s: received %d expected 2345", fname, ptInt)
	}
}

// Int

// func TestConfInt(t *testing.T) {
// 	const fname = "TestConfInt"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Int,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  1,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	i, err := config.ValueInt("one")
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if i != 1 {
// 		t.Errorf("%s: received %d expected 1", fname, i)
// 	}
// }

// func TestConfIntDefaultTypeError(t *testing.T) {
// 	const fname = "TestConfIntDefaultTypeError"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Int,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  "wrongType",
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	_, err = config.ValueInt("one")
// 	if !errors.Is(err, errType) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestValueIntNotThere(t *testing.T) {
// 	const fname = "TestValueIntNotThere"
// 	_, err := global.ValueInt("notThere")
// 	if !errors.Is(err, errNoKey) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestConfIntNoData(t *testing.T) {
// 	const fname = "TestConfIntNoData"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Int,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  int(1),
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	config.options["one"].data = nil
// 	_, err = config.ValueInt("one")
// 	if !errors.Is(err, errNoData) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// IntVar

// func TestConfIntVar(t *testing.T) {
// 	const fname = "TestConfIntVar"
// 	var i int
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     IntVar,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  1,
// 			Var:      &i,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if i != 1 {
// 		t.Errorf("%s: received %d expected 1", fname, i)
// 	}
// }

// func TestConfIntVarTypeError(t *testing.T) {
// 	const fname = "TestConfIntVarTypeError"
// 	var i string
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     IntVar,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  2,
// 			Var:      &i,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestConfIntVarDefaultTypeError(t *testing.T) {
// 	const fname = "TestConfIntVarDefaultTypeError"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     IntVar,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  "wrongType",
// 			Var:      2,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// Int64

// func TestConfInt64(t *testing.T) {
// 	const fname = "TestConfInt64"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Int64,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  int64(1),
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	i, err := config.ValueInt64("one")
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if i != 1 {
// 		t.Errorf("%s: received %d expected 1", fname, i)
// 	}
// }

// func TestConfInt64DefaultTypeError(t *testing.T) {
// 	const fname = "TestConfInt64DefaultTypeError"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Int64,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  2,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	_, err = config.ValueInt64("one")
// 	if !errors.Is(err, errType) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestValueInt64NotThere(t *testing.T) {
// 	const fname = "TestValueInt64NotThere"
// 	_, err := global.ValueInt64("notThere")
// 	if !errors.Is(err, errNoKey) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestConfInt64NoData(t *testing.T) {
// 	const fname = "TestConfInt64NoData"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Int64,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  int64(1),
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	config.options["one"].data = nil
// 	_, err = config.ValueInt64("one")
// 	if !errors.Is(err, errNoData) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestValueInt64(t *testing.T) {
// 	const fname = "TestValueInt64"
// 	i, err := ValueInt64("int64")
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if i != 1 {
// 		t.Errorf("%s: received %d expected 1", fname, i)
// 	}
// }

// Int64Var

// func TestConfInt64Var(t *testing.T) {
// 	const fname = "TestConfInt64Var"
// 	var i int64
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Int64Var,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  int64(1),
// 			Var:      &i,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if i != 1 {
// 		t.Errorf("%s: received %d expected 1", fname, i)
// 	}
// }

// func TestConfInt64VarTypeError(t *testing.T) {
// 	const fname = "TestConfInt64VarTypeError"
// 	var i string
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Int64Var,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  int64(2),
// 			Var:      &i,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestConfInt64VarDefaultTypeError(t *testing.T) {
// 	const fname = "TestConfInt64VarDefaultTypeError"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Int64Var,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  "wrongType",
// 			Var:      int64(2),
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// Uint

// func TestConfUint(t *testing.T) {
// 	const fname = "TestConfUint"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Uint,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  uint(1),
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	i, err := config.ValueUint("one")
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if i != 1 {
// 		t.Errorf("%s: received %d expected 1", fname, i)
// 	}
// }

// func TestConfUintDefaultTypeError(t *testing.T) {
// 	const fname = "TestConfUintDefaultTypeError"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Uint,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  "wrongType",
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	_, err = config.ValueUint("one")
// 	if !errors.Is(err, errType) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestValueUintNotThere(t *testing.T) {
// 	const fname = "TestValueUintNotThere"
// 	_, err := global.ValueUint("notThere")
// 	if !errors.Is(err, errNoKey) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestConfUintNoData(t *testing.T) {
// 	const fname = "TestConfUintNoData"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Uint,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  uint(1),
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	config.options["one"].data = nil
// 	_, err = config.ValueUint("one")
// 	if !errors.Is(err, errNoData) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// UintVar

// func TestConfUintVar(t *testing.T) {
// 	const fname = "TestConfUintVar"
// 	var i uint
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     UintVar,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  uint(1),
// 			Var:      &i,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if i != 1 {
// 		t.Errorf("%s: received %d expected 1", fname, i)
// 	}
// }

// func TestConfUintVarTypeError(t *testing.T) {
// 	const fname = "TestConfUintVarTypeError"
// 	var i string
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     UintVar,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  uint(2),
// 			Var:      &i,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestConfUintVarDefaultTypeError(t *testing.T) {
// 	const fname = "TestConfUintVarDefaultTypeError"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     UintVar,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  "wrongType",
// 			Var:      uint(2),
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestValueUint(t *testing.T) {
// 	const fname = "TestValueUint"
// 	i, err := ValueUint("uint")
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if i != 1 {
// 		t.Errorf("%s: received %d expected 1", fname, i)
// 	}
// }

// Uint64

// func TestConfUint64(t *testing.T) {
// 	const fname = "TestConfUint64"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Uint64,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  uint64(1),
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	i, err := config.ValueUint64("one")
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if i != 1 {
// 		t.Errorf("%s: received %d expected 1", fname, i)
// 	}
// }

// func TestConfUint64DefaultTypeError(t *testing.T) {
// 	const fname = "TestConfUint64DefaultTypeError"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Uint64,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  "wrongType",
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	_, err = config.ValueUint64("one")
// 	if !errors.Is(err, errStored) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestValueUint64NotThere(t *testing.T) {
// 	const fname = "TestValueUint64NotThere"
// 	_, err := global.ValueUint64("notThere")
// 	if !errors.Is(err, errNoKey) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestConfUint64NoData(t *testing.T) {
// 	const fname = "TestConfUint64NoData"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Uint64,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  uint64(1),
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	config.options["one"].data = nil
// 	_, err = config.ValueUint64("one")
// 	if !errors.Is(err, errNoData) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestValueUint64(t *testing.T) {
// 	const fname = "TestValueUint64"
// 	i, err := ValueUint64("uint64")
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if i != 1 {
// 		t.Errorf("%s: received %d expected 1", fname, i)
// 	}
// }

// Uint64Var

// func TestConfUint64Var(t *testing.T) {
// 	const fname = "TestConfUint64Var"
// 	var i uint64
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Uint64Var,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  uint64(1),
// 			Var:      &i,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if i != 1 {
// 		t.Errorf("%s: received %d expected 1", fname, i)
// 	}
// }

// func TestConfUint64VarTypeError(t *testing.T) {
// 	const fname = "TestConfUint64VarTypeError"
// 	var i string
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Uint64Var,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  uint64(2),
// 			Var:      &i,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestConfUint64VarDefaultTypeError(t *testing.T) {
// 	const fname = "TestConfUint64VarDefaultTypeError"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Uint64Var,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  "wrongType",
// 			Var:      uint64(2),
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// Float64

// func TestConfFloat64(t *testing.T) {
// 	const fname = "TestConfFloat64"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Float64,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  float64(1),
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	i, err := config.ValueFloat64("one")
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if i != 1 {
// 		t.Errorf("%s: received %f expected 1", fname, i)
// 	}
// }

// func TestConfFloat64DefaultTypeError(t *testing.T) {
// 	const fname = "TestConfFloat64DefaultTypeError"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Float64,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  "wrongType",
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	_, err = config.ValueFloat64("one")
// 	if !errors.Is(err, errStored) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestValueFloat64NotThere(t *testing.T) {
// 	const fname = "TestValueFloat64NotThere"
// 	_, err := global.ValueFloat64("notThere")
// 	if !errors.Is(err, errNoKey) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestConfFloat64NoData(t *testing.T) {
// 	const fname = "TestConfFloat64NoData"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Float64,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  float64(1),
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	config.options["one"].data = nil
// 	_, err = config.ValueFloat64("one")
// 	if !errors.Is(err, errNoData) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestValueFloat64(t *testing.T) {
// 	const fname = "TestValueFloat64"
// 	i, err := ValueFloat64("float64")
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if i != 1 {
// 		t.Errorf("%s: received %f expected 1", fname, i)
// 	}
// }

// Float64Var

// func TestConfFloat64Var(t *testing.T) {
// 	const fname = "TestConfFloat64Var"
// 	var i float64
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Float64Var,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  float64(1),
// 			Var:      &i,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if i != 1 {
// 		t.Errorf("%s: received %f expected 1", fname, i)
// 	}
// }

// func TestConfFloat64VarTypeError(t *testing.T) {
// 	const fname = "TestConfFloat64VarTypeError"
// 	var i string
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Float64Var,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  float64(2),
// 			Var:      &i,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestConfFloat64VarDefaultTypeError(t *testing.T) {
// 	const fname = "TestConfFloat64VarDefaultTypeError"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Float64Var,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  "wrongType",
// 			Var:      float64(2),
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// String

// func TestConfString(t *testing.T) {
// 	const fname = "TestConfString"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     String,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  "string",
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	i, err := config.ValueString("one")
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if i != "string" {
// 		t.Errorf("%s: received %q expected \"string\"", fname, i)
// 	}
// }

// func TestConfStringDefaultTypeError(t *testing.T) {
// 	const fname = "TestConfStringDefaultTypeError"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     String,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  1,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	_, err = config.ValueString("one")
// 	if !errors.Is(err, errStored) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestValueStringNotThere(t *testing.T) {
// 	const fname = "TestValueStringNotThere"
// 	_, err := global.ValueString("notThere")
// 	if !errors.Is(err, errNoKey) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestConfStringNoData(t *testing.T) {
// 	const fname = "TestConfStringNoData"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     String,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  "string",
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	config.options["one"].data = nil
// 	_, err = config.ValueString("one")
// 	if !errors.Is(err, errNoData) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestValueString(t *testing.T) {
// 	const fname = "TestValueString"
// 	i, err := ValueString("string")
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if i != "string" {
// 		t.Errorf("%s: received %s expected 1", fname, i)
// 	}
// }

// StringVar

// func TestConfStringVar(t *testing.T) {
// 	const fname = "TestConfStringVar"
// 	var str string
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     StringVar,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  "string",
// 			Var:      &str,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if str != "string" {
// 		t.Errorf("%s: received %q expected \"string\"", fname, str)
// 	}
// }

// func TestConfStringVarTypeError(t *testing.T) {
// 	const fname = "TestConfStringVarTypeError"
// 	var str int
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     StringVar,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  "string",
// 			Var:      &str,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestConfStringVarDefaultTypeError(t *testing.T) {
// 	const fname = "TestConfStringVarDefaultTypeError"
// 	config := Config{}
// 	str := "string"
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     StringVar,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  1,
// 			Var:      str,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// Bool

// func TestConfBool(t *testing.T) {
// 	const fname = "TestConfBool"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Bool,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  true,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	i, err := config.ValueBool("one")
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if i != true {
// 		t.Errorf("%s: received %t expected \"true\"", fname, i)
// 	}
// }

// func TestConfBoolDefaultTypeError(t *testing.T) {
// 	const fname = "TestConfBoolDefaultTypeError"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Bool,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  1,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	_, err = config.ValueBool("one")
// 	if !errors.Is(err, errStored) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestValueBoolNotThere(t *testing.T) {
// 	const fname = "TestValueBoolNotThere"
// 	_, err := global.ValueBool("notThere")
// 	if !errors.Is(err, errNoKey) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestConfBoolNoData(t *testing.T) {
// 	const fname = "TestConfBoolNoData"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Bool,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  true,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	config.options["one"].data = nil
// 	_, err = config.ValueBool("one")
// 	if !errors.Is(err, errNoData) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestValueBool(t *testing.T) {
// 	const fname = "TestValueBool"
// 	i, err := ValueBool("bool")
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if i != true {
// 		t.Errorf("%s: received %t expected \"true\"", fname, i)
// 	}
// }

// // BoolVar

// func TestConfBoolVar(t *testing.T) {
// 	const fname = "TestConfBoolVar"
// 	var b bool
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     BoolVar,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  true,
// 			Var:      &b,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if b != true {
// 		t.Errorf("%s: received %t expected \"true\"", fname, b)
// 	}
// }

// func TestConfBoolVarTypeError(t *testing.T) {
// 	const fname = "TestConfBoolVarTypeError"
// 	var b int
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     BoolVar,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  true,
// 			Var:      &b,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestConfBoolVarDefaultTypeError(t *testing.T) {
// 	const fname = "TestConfBoolVarDefaultTypeError"
// 	config := Config{}
// 	b := true
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     BoolVar,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  1,
// 			Var:      &b,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// Duration

// func TestConfDuration(t *testing.T) {
// 	const fname = "TestConfDuration"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Duration,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  time.Duration(0),
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	i, err := config.ValueDuration("one")
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if i != time.Duration(0) {
// 		t.Errorf("%s: received %q expected \"0s\"", fname, i)
// 	}
// }

// func TestConfDurationDefaultTypeError(t *testing.T) {
// 	const fname = "TestConfDurationDefaultTypeError"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Duration,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  1,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	_, err = config.ValueDuration("one")
// 	if !errors.Is(err, errStored) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestValueDurationNotThere(t *testing.T) {
// 	const fname = "TestValueDurationNotThere"
// 	_, err := global.ValueDuration("notThere")
// 	if !errors.Is(err, errNoKey) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestConfDurationNoData(t *testing.T) {
// 	const fname = "TestConfDurationNoData"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Duration,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  time.Duration(0),
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	config.options["one"].data = nil
// 	_, err = config.ValueDuration("one")
// 	if !errors.Is(err, errNoData) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestValueDuration(t *testing.T) {
// 	const fname = "TestValueDuration"
// 	i, err := ValueDuration("duration")
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if i != time.Duration(0) {
// 		t.Errorf("%s: received %q expected \"0s\"", fname, i)
// 	}
// }

// // DurationVar

// func TestConfDurationVar(t *testing.T) {
// 	const fname = "TestConfDurationVar"
// 	var d time.Duration
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     DurationVar,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  time.Duration(0),
// 			Var:      &d,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if d != time.Duration(0) {
// 		t.Errorf("%s: received %q expected \"0s\"", fname, d)
// 	}
// }

// func TestConfDurationVarTypeError(t *testing.T) {
// 	const fname = "TestConfDurationVarTypeError"
// 	var b int
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     DurationVar,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  time.Duration(0),
// 			Var:      &b,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// func TestConfDurationVarDefaultTypeError(t *testing.T) {
// 	const fname = "TestConfDurationVarDefaultTypeError"
// 	config := Config{}
// 	b := time.Duration(0)
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     DurationVar,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  1,
// 			Var:      &b,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// Value

// func TestConfValue(t *testing.T) {
// 	const fname = "TestConfInt"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Int,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  1,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	i, _, err := config.Value("one")
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if *i.(*int) != 1 {
// 		t.Errorf("%s: received %d expected 1", fname, i)
// 	}
// }

//func TestConfValueError(t *testing.T) {
//	const fname = "TestConfInt"
//	config := Config{}
//	cmd := config.Setup("Usage Heading", "Mode Heading")
//	opts := []Option{
//		{Name: "one",
//			Type:  Int,
//			Flag:  "i",
//			Usage: "do it like this",
//			//Default:  1,
//			Commands: cmd,
//		},
//	}
//	err := config.Options(opts...)
//	if !errors.Is(err, errConfig) {
//		t.Errorf("%s: error: %s", fname, err)
//	}
//	err = config.Parse()
//	if !errors.Is(err, errConfig) {
//		t.Errorf("%s: error: %s", fname, err)
//	}
//	_, _, err = config.Value("one")
//	if !errors.Is(err, errType) {
//		t.Errorf("%s: error: %s", fname, err)
//	}
//}

//func TestConfValueNotThere(t *testing.T) {
//	const fname = "TestConfValueNotThere"
//	_, _, err := global.Value("notThere")
//	if !errors.Is(err, errNoKey) {
//		t.Errorf("%s: error: %s", fname, err)
//	}
//}

//func TestConfValueNoData(t *testing.T) {
//	const fname = "TestConfDurationNoData"
//	config := Config{}
//	cmd := config.Setup("Usage Heading", "Mode Heading")
//	opts := []Option{
//		{Name: "one",
//			Type:     Int,
//			Flag:     "i",
//			Usage:    "do it like this",
//			Default:  1,
//			Commands: cmd,
//		},
//	}
//	err := config.Options(opts...)
//	if err != nil {
//		t.Errorf("%s: error: %s", fname, err)
//	}
//	err = config.Parse()
//	if err != nil {
//		t.Errorf("%s: error: %s", fname, err)
//	}
//	config.options["one"].data = nil
//	_, _, err = config.Value("one")
//	if !errors.Is(err, errNoData) {
//		t.Errorf("%s: error: %s", fname, err)
//	}
//}

//func TestValue(t *testing.T) {
//	const fname = "TestValue"
//	i, _, err := Value("int")
//	if err != nil {
//		t.Errorf("%s: error: %s", fname, err)
//	}
//	if *i.(*int) != 1 {
//		t.Errorf("%s: received %d expected 1", fname, i)
//	}
//}

// Var

// type testValue struct {
// 	str string
// }

// func (t testValue) String() string {
// 	return t.str
// }

// func (t *testValue) Set(str string) error {
// 	return nil
// }

// func TestConfVar(t *testing.T) {
// 	const fname = "TestConfVar"
// 	thing := testValue{str: "string"}
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Var,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  testValue{str: ""},
// 			Value:    &thing,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if thing.str != "string" {
// 		t.Errorf("%s: received %q expected \"string\"", fname, thing.str)
// 	}
// }

// func TestConfVarDefaultTypeError(t *testing.T) {
// 	const fname = "TestConfVarDefaultTypeError"
// 	_ = testValue{str: "string"}
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Var,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  testValue{str: "string"},
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// Nil

// func TestConfNil(t *testing.T) {
// 	const fname = "TestConfNil"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Nil,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  time.Duration(0),
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }

// default

// func TestConfDefault(t *testing.T) {
// 	const fname = "TestConfDefault"
// 	config := Config{}
// 	cmd := config.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Default,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  1,
// 			Commands: cmd,
// 		},
// 	}
// 	err := config.Options(opts...)
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	err = config.Parse()
// 	if !errors.Is(err, errConfig) {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// }
