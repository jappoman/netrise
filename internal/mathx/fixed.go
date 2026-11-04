package mathx

import (
	"errors"
	"strconv"
)

const Scale int64 = 1000

var ErrOverflow = errors.New("fixed-point arithmetic overflow")

type Amount int64

func New(units int64) Amount  { return Amount(units * Scale) }
func (a Amount) Units() int64 { return int64(a) / Scale }
func (a Amount) Add(b Amount) (Amount, error) {
	c := int64(a) + int64(b)
	if (b > 0 && c < int64(a)) || (b < 0 && c > int64(a)) {
		return 0, ErrOverflow
	}
	return Amount(c), nil
}
func (a Amount) MulUnits(n int64) (Amount, error) {
	if n != 0 && (int64(a) > 0 && n > (1<<63-1)/int64(a) || int64(a) < 0 && n < (-1<<63)/int64(a)) {
		return 0, ErrOverflow
	}
	return Amount(int64(a) * n), nil
}
func (a Amount) String() string { return strconv.FormatInt(a.Units(), 10) }
