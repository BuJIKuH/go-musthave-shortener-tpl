package example

import "testing"

func TestReset(t *testing.T) {
	str := "hello"
	r := &ResetableStruct{
		i:     42,
		str:   "world",
		strP:  &str,
		s:     []int{1, 2, 3},
		m:     map[string]string{"a": "b"},
		child: &ResetableStruct{i: 100},
	}

	r.Reset()

	if r.i != 0 || r.str != "" || *r.strP != "" || len(r.s) != 0 || len(r.m) != 0 || r.child.i != 0 {
		t.Errorf("Reset() не сбросил поля корректно: %+v", r)
	}
}
