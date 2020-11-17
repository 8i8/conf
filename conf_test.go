package conf

import (
	"testing"
)

func init() {
	test = true
}

var (
	m1    = Setup("Usage Heading", "Mode Heading")
	m2    = Mode("one", "one's heading")
	m3    = Mode("two", "two's heading")
	tc    Config
	cm1   = tc.Setup("Usage Heading", "Mode Heading")
	cm2   = tc.Mode("cone", "cone's heading")
	cm3   = tc.Mode("ctwo", "ctwo's heading")
	ptInt int
)

var opts = []Option{
	{Name: "int",
		Type:    Int,
		Key:     "a",
		Help:    "like this",
		Default: 1,
		Modes:   m1 | m2 | m3,
	},
	{Name: "intVar",
		Type:    IntVar,
		Key:     "b",
		Help:    "do it like this",
		Default: 2345,
		Var:     &ptInt,
		Modes:   m1,
	},
	{Name: "intNoDefault",
		Type:  Int,
		Key:   "c",
		Help:  "like that",
		Modes: m1,
	},
}

func testFlagIs(t *testing.T) {
	const fname = "TestFlagIs"
	v := c.flagIs(0)
	if v {
		t.Errorf("%s: recieved true expected false", fname)
	}
	v = c.flagIs(m1)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = c.flagIs(m2)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = c.flagIs(m3)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = c.flagIs(m1 | m3)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = c.flagIs(m1 | m2 | m3)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = c.flagIs(c.index)
	if v {
		t.Errorf("%s: recieved true expected false", fname)
	}

	v = tc.flagIs(0)
	if v {
		t.Errorf("%s: recieved true expected false", fname)
	}
	v = tc.flagIs(cm1)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = tc.flagIs(cm2)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = tc.flagIs(cm3)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = tc.flagIs(cm1 | cm3)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = tc.flagIs(cm1 | cm2 | cm3)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = tc.flagIs(tc.index)
	if v {
		t.Errorf("%s: recieved true expected false", fname)
	}
}
func TestTypes(t *testing.T) {
	Options(opts...)
	Parse()
	tc.Options(opts...)
	tc.Parse()
	t.Run("Modes", testModes)
	t.Run("c.Modes", testModesConfig)
	t.Run("flagIs", testFlagIs)
	t.Run("Int", testInt)
	t.Run("c.Int", testConfigInt)
	t.Run("IntVar", testIntVar)
}

func testModes(t *testing.T) {
	const fname = "TestModes"
	m := GetMode()
	if m != "default" {
		t.Errorf("%s: expected \"default\" recieved %q", fname, m)
	}
	if !c.flagIs(m1) {
		t.Errorf("%s: recived false expected true", fname)
	}
}

func testModesConfig(t *testing.T) {
	const fname = "TestModesConfig"
	m := c.GetMode()
	if m != "default" {
		t.Errorf("%s: expected default recieved %s", fname, m)
	}
	if !tc.flagIs(cm1) {
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
	i, err = ValueInt("two")
	if err == nil {
		t.Errorf("%s: expected an error", fname)
	}
}

func testConfigInt(t *testing.T) {
	const fname = "testConfigInt"
	c := Config{}
	mode := c.Setup("Usage Heading", "Mode Heading")
	opts := []Option{
		{Name: "one",
			Type:    Int,
			Key:     "i",
			Help:    "do it like this",
			Default: 2,
			Modes:   mode,
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
