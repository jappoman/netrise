package mathx

import (
	"errors"
	"strconv"
)

const Scale int64 = 1000

const MaxUnits int64 = int64(^uint64(0)>>1) / Scale

var ErrOverflow = errors.New("fixed-point arithmetic overflow")

type Amount int64

func NewChecked(units int64) (Amount, error) {
	if units > MaxUnits || units < -MaxUnits {
		return 0, ErrOverflow
	}
	return Amount(units * Scale), nil
}
func New(units int64) Amount {
	amount, err := NewChecked(units)
	if err != nil {
		panic(err)
	}
	return amount
}
func (a Amount) Units() int64 { return int64(a) / Scale }
func (a Amount) Add(b Amount) (Amount, error) {
	c := int64(a) + int64(b)
	if (b > 0 && c < int64(a)) || (b < 0 && c > int64(a)) {
		return 0, ErrOverflow
	}
	return Amount(c), nil
}
func (a Amount) MulUnits(n int64) (Amount, error) {
	left := int64(a)
	const maxInt64 = int64(^uint64(0) >> 1)
	const minInt64 = -maxInt64 - 1
	if left == 0 || n == 0 {
		return 0, nil
	}
	if (left == minInt64 && n == -1) || (n == minInt64 && left == -1) {
		return 0, ErrOverflow
	}
	switch {
	case left > 0 && n > 0 && left > maxInt64/n:
		return 0, ErrOverflow
	case left > 0 && n < 0 && n < minInt64/left:
		return 0, ErrOverflow
	case left < 0 && n > 0 && left < minInt64/n:
		return 0, ErrOverflow
	case left < 0 && n < 0 && n < maxInt64/left:
		return 0, ErrOverflow
	}
	return Amount(left * n), nil
}
func (a Amount) String() string { return strconv.FormatInt(a.Units(), 10) }
