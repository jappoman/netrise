package mathx

import "testing"

func TestAmount(t *testing.T) {
	a := New(25)
	if a.String() != "25" {
		t.Fatal(a)
	}
	if _, err := Amount(1 << 62).MulUnits(3); err == nil {
		t.Fatal("expected overflow")
	}
}
