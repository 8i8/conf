package conf

import "testing"

// func TestConfig(t *testing.T) {
// 	c := Config{}
// 	mode := c.Setup("Usage Heading", "Mode Heading")
// 	opts := []Option{
// 		{Name: "one",
// 			Type:    Int,
// 			Key:     "i",
// 			Help:    "like this",
// 			Default: 2,
// 			Modes:   mode,
// 		},
// 	}
// 	c.Options(opts...)
// }

func TestFlagIs(t *testing.T) {
	const fname = "TestFlagIs"
	v := c.flagIs(0)
	if v {
		t.Errorf("%s: recieved true expected false", fname)
	}
	m1 := c.Mode("one", "help1")
	v = c.flagIs(m1)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	m2 := c.Mode("two", "help2")
	v = c.flagIs(m2)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	m3 := c.Mode("three", "help3")
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
	v = c.flagIs(m3 << 1)
	if v {
		t.Errorf("%s: recieved true expected false", fname)
	}
	v = c.flagIs(64)
	if v {
		t.Errorf("%s: recieved true expected false", fname)
	}
}
