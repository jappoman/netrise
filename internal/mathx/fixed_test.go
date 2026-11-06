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

func TestMulUnitsChecksEverySignedOverflowDirection(t *testing.T) {
	const maxInt64 = int64(^uint64(0) >> 1)
	const minInt64 = -maxInt64 - 1
	cases := []struct {
		amount Amount
		units  int64
	}{
		{Amount(1 << 62), -3},
		{Amount(-(1 << 62)), 3},
		{Amount(minInt64), -1},
		{Amount(-1), minInt64},
		{Amount(minInt64), 2},
		{Amount(-2), minInt64},
	}
	for _, tc := range cases {
		if _, err := tc.amount.MulUnits(tc.units); err != ErrOverflow {
			t.Fatalf("MulUnits(%d, %d) returned %v, want %v", tc.amount, tc.units, err, ErrOverflow)
		}
	}
	if got, err := Amount(-7).MulUnits(-3); err != nil || got != 21 {
		t.Fatalf("valid signed multiplication failed: got=%d err=%v", got, err)
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

func TestAddSaturatedClampsBothDirections(t *testing.T) {
	max := Amount(^uint64(0) >> 1)
	min := Amount(-max - 1)
	if got := AddSaturated(max, 1); got != max {
		t.Fatalf("positive overflow was not saturated: %d", got)
	}
	if got := AddSaturated(min, -1); got != min {
		t.Fatalf("negative overflow was not saturated: %d", got)
	}
}
