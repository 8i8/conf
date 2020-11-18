package conf

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

func init() {
	test = true
}

var (
	m1     = Setup("Usage Heading", "Mode Heading")
	m2     = Mode("one", "one's heading")
	m3     = Mode("two", "two's heading")
	config Config
	cm1    = config.Setup("Usage Heading", "Mode Heading")
	cm2    = config.Command("cone", "cone's heading")
	cm3    = config.Command("ctwo", "ctwo's heading")
	ptInt  int
)

var opts = []Option{
	{Name: "int",
		Type:    Int,
		Flag:     "a",
		Usage:    "like this",
		Default: 1,
		Commands:   m1 | m2 | m3,
	},
	{Name: "intVar",
		Type:    IntVar,
		Flag:     "b",
		Usage:    "do it like this",
		Default: 2345,
		Var:     &ptInt,
		Commands:   m1,
	},
	{Name: "intNoDefault",
		Type:  Int,
		Flag:   "c",
		Usage:  "like that",
		Commands: m1,
	},
}

func TestToManyModes(t *testing.T) {
	const fname = "TestToManyModes"
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

func TestDoubleNameError(t *testing.T) {
	const fname = "TestDoubleNameError"
	config := Config{}
	m := config.Setup("", "")
	var opts = []Option{
		{Name: "errors",
			Type:    Int,
			Flag:     "d",
			Usage:    "like this",
			Default: 1,
			Commands:   m,
		},
		{Name: "errors",
			Type:    Int,
			Flag:     "d",
			Usage:    "like this",
			Default: 1,
			Commands:   m,
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
			Type:    Int,
			Flag:     "d",
			Usage:    "like this",
			Default: 1,
			Commands:   m,
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
			Type:    Int,
			Flag:     "a",
			Usage:    "like this",
			Default: 1,
			Commands:   m,
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
			Type:    Int,
			Flag:     "a",
			Usage:    "like this",
			Default: 1,
			Commands:   m,
		},
		{Name: "int",
			Type:    Int,
			Flag:     "b",
			Usage:    "like this",
			Default: 1,
			Commands:   m,
		},
	}
	err := config.Options(similarNames...)
	if !errors.Is(err, errConfig) {
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
	var similarNames = []Option{
		{Name: "int",
			Type:    Int,
			Flag:     "a",
			Usage:    "like this",
			Default: 1,
			Commands:   m1,
		},
		{Name: "int",
			Type:    Int,
			Flag:     "b",
			Usage:    "like this",
			Default: 1,
			Commands:   m2,
		},
	}
	err := config.Options(similarNames...)
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
			Type:    Int,
			Flag:     "a",
			Usage:    "like this",
			Default: 1,
			Commands:   m,
			Check: func(v interface{}) (interface{}, error) {
				i := *v.(*int)
				i++
				return &i, nil
			},
		},
		{Name: "int2",
			Type:    Int,
			Flag:     "b",
			Usage:    "like this",
			Default: 1,
			Commands:   m,
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
		t.Errorf("%s: recieved %d expected 2", fname, i)
	}

	_, err = config.ValueInt("int2")
	if !errors.Is(err, errCheck) {
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
			Type:    Int,
			Flag:     "a",
			Usage:    "like this",
			Default: 1,
			Commands:   m,
		},
		{Name: "similarKeys",
			Type:    Int,
			Flag:     "a",
			Usage:    "like this",
			Default: 1,
			Commands:   m,
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
			Type:    Int,
			Usage:    "like this",
			Default: 1,
			Commands:   m,
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
	m := 2
	var opts = []Option{
		{Name: "int",
			Type:    Int,
			Flag:     "a",
			Usage:    "like this",
			Default: 1,
			Commands:   m,
		},
	}
	err := config.Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
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
			Type:    Int,
			Flag:     "a",
			Usage:    "like this",
			Default: 1,
			Commands:   m1,
		},
		{Name: "similarKeys",
			Type:    Int,
			Flag:     "a",
			Usage:    "like this",
			Default: 1,
			Commands:   m2,
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
	config.Options(opts...)
	config.Parse()
	t.Run("ArgList", testArgList)
	t.Run("c.ArgList", testArgListConfig)
	t.Run("Modes", testModes)
	t.Run("c.Modes", testModesConfig)
	t.Run("flagIs", testFlagIs)
	t.Run("Int", testInt)
	t.Run("c.Int", testIntConfig)
	t.Run("IntVar", testIntVar)
}

// func testMultiErrors(t *testing.T) {
// }

func testArgList(t *testing.T) {
	const fname = "TestArgList"
	l := ArgList()
	if l == "" {
		t.Errorf("%s: recieved an empty string", fname)
	}
}

func testArgListConfig(t *testing.T) {
	const fname = "TestArgListConf"
	l := config.ArgString()
	if l == "" {
		t.Errorf("%s: recieved an empty string", fname)
	}
}

func testFlagIs(t *testing.T) {
	const fname = "TestFlagIs"
	v := c.cmdIs(0)
	if v {
		t.Errorf("%s: recieved true expected false", fname)
	}
	v = c.cmdIs(m1)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = c.cmdIs(m2)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = c.cmdIs(m3)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = c.cmdIs(m1 | m3)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = c.cmdIs(m1 | m2 | m3)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = c.cmdIs(c.index)
	if v {
		t.Errorf("%s: recieved true expected false", fname)
	}

	v = config.cmdIs(0)
	if v {
		t.Errorf("%s: recieved true expected false", fname)
	}
	v = config.cmdIs(cm1)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = config.cmdIs(cm2)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = config.cmdIs(cm3)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = config.cmdIs(cm1 | cm3)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = config.cmdIs(cm1 | cm2 | cm3)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = config.cmdIs(config.index)
	if v {
		t.Errorf("%s: recieved true expected false", fname)
	}
}

func testModes(t *testing.T) {
	const fname = "TestModes"
	m := GetMode()
	if m != "default" {
		t.Errorf("%s: expected \"default\" recieved %q", fname, m)
	}
	if !c.cmdIs(m1) {
		t.Errorf("%s: recived false expected true", fname)
	}
}

func testModesConfig(t *testing.T) {
	const fname = "TestModesConfig"
	m := c.GetCmd()
	if m != "default" {
		t.Errorf("%s: expected default recieved %s", fname, m)
	}
	if !config.cmdIs(cm1) {
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
		t.Errorf("%s: recieved %d expected 1", fname, i)
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

func testIntConfig(t *testing.T) {
	const fname = "testConfigInt"
	c := Config{}
	mode := c.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:    Int,
			Flag:     "i",
			Usage:    "do it like this",
			Default: 2,
			Commands:   mode,
		},
	}
	c.Options(opts...)
	c.Parse()
	i, err := c.ValueInt("one")
	if err != nil {
		t.Errorf("%s: error: %s", fname, err)
	}
	if i != 2 {
		t.Errorf("%s: recieved %d expected 2", fname, i)
	}
}

func testIntVar(t *testing.T) {
	const fname = "testIntVar"
	if ptInt != 2345 {
		t.Errorf("%s: recieved %d expected 2345", fname, ptInt)
	}
}
