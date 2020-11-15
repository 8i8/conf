package conf

import "testing"

func TestFlagIs(t *testing.T) {
	const fname = "TestFlagIs"
	v := c.list.flagIs(0)
	if v {
		t.Errorf("%s: recieved true expected false", fname)
	}
	m1 := Mode("one", "help1")
	v = c.list.flagIs(m1)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	m2 := Mode("two", "help2")
	v = c.list.flagIs(m2)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	m3 := Mode("three", "help3")
	v = c.list.flagIs(m3)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = c.list.flagIs(m1 | m3)
	if !v {
		t.Errorf("%s: recieved false expected true", fname)
	}
	v = c.list.flagIs(m3 << 1)
	if v {
		t.Errorf("%s: recieved true expected false", fname)
	}
}
