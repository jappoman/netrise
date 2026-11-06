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

func TestNewCheckedRejectsUnitOverflow(t *testing.T) {
	if _, err := NewChecked(MaxUnits + 1); err != ErrOverflow {
		t.Fatalf("positive unit overflow returned %v, want %v", err, ErrOverflow)
	}
	if _, err := NewChecked(-MaxUnits - 1); err != ErrOverflow {
		t.Fatalf("negative unit overflow returned %v, want %v", err, ErrOverflow)
	}
	if got, err := NewChecked(MaxUnits); err != nil || got.Units() != MaxUnits {
		t.Fatalf("maximum representable units failed: amount=%d err=%v", got, err)
	}
}

func TestNewPanicsInsteadOfSilentlyOverflowing(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("New silently accepted an overflowing unit value")
		}
	}()
	_ = New(MaxUnits + 1)
}
