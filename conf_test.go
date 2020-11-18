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
	m2     = Command("one", "one's heading")
	m3     = Command("two", "two's heading")
	config Config
	cm1    = config.Setup("Usage Heading", "Mode Heading")
	cm2    = config.Command("cone", "cone's heading")
	cm3    = config.Command("ctwo", "ctwo's heading")
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

// TestCmdIs test the fuction that tests if a cmd exists.
func TestCmdIs(t *testing.T) {
	const fname = "TestCmdIs"
	config := Config{}
	_ = config.Setup("", "")
	_ = config.Command("cmd", "")
	if !config.cmdNameIs("cmd") {
		t.Errorf("%s: recieved false expected true", fname)
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
	config.Options(opts...)
	config.Parse()
	t.Run("ArgList", testArgList)
	t.Run("c.ArgList", testArgListConfig)
	t.Run("Modes", testModes)
	t.Run("c.Modes", testModesConfig)
	t.Run("flagIs", testFlagIs)
	t.Run("Int", testInt)
	//t.Run("c.Int", testIntConfig)
	t.Run("IntVar", testIntVar)
}

func testArgList(t *testing.T) {
	const fname = "TestArgList"
	l := ArgString()
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
	v := c.cmdTokenIs(0)
	if v {
		t.Errorf("%s: recieved true expected false", fname)
	}
	v = c.cmdTokenIs(m1)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = c.cmdTokenIs(m2)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = c.cmdTokenIs(m3)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = c.cmdTokenIs(m1 | m3)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = c.cmdTokenIs(m1 | m2 | m3)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = c.cmdTokenIs(c.index)
	if v {
		t.Errorf("%s: recieved true expected false", fname)
	}

	v = config.cmdTokenIs(0)
	if v {
		t.Errorf("%s: recieved true expected false", fname)
	}
	v = config.cmdTokenIs(cm1)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = config.cmdTokenIs(cm2)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = config.cmdTokenIs(cm3)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = config.cmdTokenIs(cm1 | cm3)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = config.cmdTokenIs(cm1 | cm2 | cm3)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = config.cmdTokenIs(config.index)
	if v {
		t.Errorf("%s: recieved true expected false", fname)
	}
}

func testModes(t *testing.T) {
	const fname = "TestModes"
	m := GetCmd()
	if m != "default" {
		t.Errorf("%s: expected \"default\" recieved %q", fname, m)
	}
	if !c.cmdTokenIs(m1) {
		t.Errorf("%s: recived false expected true", fname)
	}
}

func testModesConfig(t *testing.T) {
	const fname = "TestModesConfig"
	m := c.GetCmd()
	if m != "default" {
		t.Errorf("%s: expected default recieved %s", fname, m)
	}
	if !config.cmdTokenIs(cm1) {
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

// func testIntConfig(t *testing.T) {
// 	const fname = "testConfigInt"
// 	c := Config{}
// 	mode := c.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:     Int,
// 			Flag:     "i",
// 			Usage:    "do it like this",
// 			Default:  2,
// 			Commands: mode,
// 		},
// 	}
// 	c.Options(opts...)
// 	c.Parse()
// 	i, err := c.ValueInt("one")
// 	if err != nil {
// 		t.Errorf("%s: error: %s", fname, err)
// 	}
// 	if i != 2 {
// 		t.Errorf("%s: recieved %d expected 2", fname, i)
// 	}
// }

func testIntVar(t *testing.T) {
	const fname = "testIntVar"
	if ptInt != 2345 {
		t.Errorf("%s: recieved %d expected 2345", fname, ptInt)
	}
}

// Int

func TestConfInt(t *testing.T) {
	const fname = "TestConfInt"
	config := Config{}
	mode := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     Int,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  1,
			Commands: mode,
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
	i, err := config.ValueInt("one")
	if err != nil {
		t.Errorf("%s: error: %s", fname, err)
	}
	if i != 1 {
		t.Errorf("%s: recieved %d expected 1", fname, i)
	}
}

func TestConfIntDefaultTypeError(t *testing.T) {
	const fname = "TestConfIntDefaultTypeError"
	config := Config{}
	mode := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     Int,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  "wrongType",
			Commands: mode,
		},
	}
	err := config.Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	err = config.Parse()
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	_, err = config.ValueInt("one")
	if !errors.Is(err, errType) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

// IntVar

func TestConfIntVar(t *testing.T) {
	const fname = "TestConfIntVar"
	var i int
	config := Config{}
	mode := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     IntVar,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  1,
			Var:      &i,
			Commands: mode,
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
	if i != 1 {
		t.Errorf("%s: recieved %d expected 1", fname, i)
	}
}

func TestConfIntVarTypeError(t *testing.T) {
	const fname = "TestConfIntVarTypeError"
	var i string
	config := Config{}
	mode := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     IntVar,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  2,
			Var:      &i,
			Commands: mode,
		},
	}
	err := config.Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	err = config.Parse()
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

func TestConfIntVarDefaultTypeError(t *testing.T) {
	const fname = "TestConfIntVarDefaultTypeError"
	config := Config{}
	mode := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     IntVar,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  "wrongType",
			Var:      2,
			Commands: mode,
		},
	}
	err := config.Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	err = config.Parse()
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

// Int64

func TestConfInt64(t *testing.T) {
	const fname = "TestConfInt64"
	config := Config{}
	mode := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     Int64,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  int64(1),
			Commands: mode,
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
	i, err := config.ValueInt64("one")
	if err != nil {
		t.Errorf("%s: error: %s", fname, err)
	}
	if i != 1 {
		t.Errorf("%s: recieved %d expected 1", fname, i)
	}
}

func TestConfInt64DefaultTypeError(t *testing.T) {
	const fname = "TestConfInt64DefaultTypeError"
	config := Config{}
	mode := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     Int64,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  2,
			Commands: mode,
		},
	}
	err := config.Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	err = config.Parse()
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	_, err = config.ValueInt64("one")
	if !errors.Is(err, errType) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

// Int64Var

func TestConfInt64Var(t *testing.T) {
	const fname = "TestConfInt64Var"
	var i int64
	config := Config{}
	mode := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     Int64Var,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  int64(1),
			Var:      &i,
			Commands: mode,
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
	if i != 1 {
		t.Errorf("%s: recieved %d expected 1", fname, i)
	}
}

func TestConfInt64VarTypeError(t *testing.T) {
	const fname = "TestConfInt64VarTypeError"
	var i string
	config := Config{}
	mode := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     Int64Var,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  int64(2),
			Var:      &i,
			Commands: mode,
		},
	}
	err := config.Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	err = config.Parse()
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

func TestConfInt64VarDefaultTypeError(t *testing.T) {
	const fname = "TestConfInt64VarDefaultTypeError"
	config := Config{}
	mode := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     Int64Var,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  "wrongType",
			Var:      int64(2),
			Commands: mode,
		},
	}
	err := config.Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	err = config.Parse()
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

// Uint

func TestConfUint(t *testing.T) {
	const fname = "TestConfUint"
	config := Config{}
	mode := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     Uint,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  uint(1),
			Commands: mode,
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
	i, err := config.ValueUint("one")
	if err != nil {
		t.Errorf("%s: error: %s", fname, err)
	}
	if i != 1 {
		t.Errorf("%s: recieved %d expected 1", fname, i)
	}
}

func TestConfUintDefaultTypeError(t *testing.T) {
	const fname = "TestConfUintDefaultTypeError"
	config := Config{}
	mode := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     Uint,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  "wrongType",
			Commands: mode,
		},
	}
	err := config.Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	err = config.Parse()
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	_, err = config.ValueUint("one")
	if !errors.Is(err, errType) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

// UintVar

func TestConfUintVar(t *testing.T) {
	const fname = "TestConfUintVar"
	var i uint
	config := Config{}
	mode := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     UintVar,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  uint(1),
			Var:      &i,
			Commands: mode,
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
	if i != 1 {
		t.Errorf("%s: recieved %d expected 1", fname, i)
	}
}

func TestConfUintVarTypeError(t *testing.T) {
	const fname = "TestConfUintVarTypeError"
	var i string
	config := Config{}
	mode := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     UintVar,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  uint(2),
			Var:      &i,
			Commands: mode,
		},
	}
	err := config.Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	err = config.Parse()
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

func TestConfUintVarDefaultTypeError(t *testing.T) {
	const fname = "TestConfUintVarDefaultTypeError"
	config := Config{}
	mode := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     UintVar,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  "wrongType",
			Var:      uint(2),
			Commands: mode,
		},
	}
	err := config.Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	err = config.Parse()
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

// Uint64

func TestConfUint64(t *testing.T) {
	const fname = "TestConfUint64"
	config := Config{}
	mode := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     Uint64,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  uint64(1),
			Commands: mode,
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
	i, err := config.ValueUint64("one")
	if err != nil {
		t.Errorf("%s: error: %s", fname, err)
	}
	if i != 1 {
		t.Errorf("%s: recieved %d expected 1", fname, i)
	}
}

func TestConfUint64DefaultTypeError(t *testing.T) {
	const fname = "TestConfUint64DefaultTypeError"
	config := Config{}
	mode := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     Uint64,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  "wrongType",
			Commands: mode,
		},
	}
	err := config.Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	err = config.Parse()
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	_, err = config.ValueUint64("one")
	if !errors.Is(err, errType) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

// Uint64Var

func TestConfUint64Var(t *testing.T) {
	const fname = "TestConfUint64Var"
	var i uint64
	config := Config{}
	mode := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     Uint64Var,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  uint64(1),
			Var:      &i,
			Commands: mode,
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
	if i != 1 {
		t.Errorf("%s: recieved %d expected 1", fname, i)
	}
}

func TestConfUint64VarTypeError(t *testing.T) {
	const fname = "TestConfUint64VarTypeError"
	var i string
	config := Config{}
	mode := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     Uint64Var,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  uint64(2),
			Var:      &i,
			Commands: mode,
		},
	}
	err := config.Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	err = config.Parse()
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

func TestConfUint64VarDefaultTypeError(t *testing.T) {
	const fname = "TestConfUint64VarDefaultTypeError"
	config := Config{}
	mode := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     Uint64Var,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  "wrongType",
			Var:      uint64(2),
			Commands: mode,
		},
	}
	err := config.Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	err = config.Parse()
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
}

// String

func TestConfString(t *testing.T) {
	const fname = "TestConfString"
	config := Config{}
	mode := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     String,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  "string",
			Commands: mode,
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
	i, err := config.ValueString("one")
	if err != nil {
		t.Errorf("%s: error: %s", fname, err)
	}
	if i != "string" {
		t.Errorf("%s: recieved %q expected \"string\"", fname, i)
	}
}

func TestConfStringDefaultTypeError(t *testing.T) {
	const fname = "TestConfStringDefaultTypeError"
	config := Config{}
	mode := config.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:     String,
			Flag:     "i",
			Usage:    "do it like this",
			Default:  1,
			Commands: mode,
		},
	}
	err := config.Options(opts...)
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	err = config.Parse()
	if !errors.Is(err, errConfig) {
		t.Errorf("%s: error: %s", fname, err)
	}
	_, err = config.ValueString("one")
	if !errors.Is(err, errType) {
		t.Errorf("%s: error: %s", fname, err)
	}
}
