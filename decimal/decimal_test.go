package decimal

import (
	"errors"
	"testing"

	"github.com/cockroachdb/apd"
	"github.com/stretchr/testify/require"
)

func TestParseCurrency(t *testing.T) {
	cases := []struct {
		a   string
		exp *apd.Decimal
		err error
	}{
		{
			a:   "ten",
			err: errors.New(`parse exponent: n: strconv.ParseInt: parsing "n": invalid syntax`),
		},
		{
			// No negative numbers
			a:   "-1",
			err: ErrNegative,
		},
		{
			a:   "-0",
			exp: apd.New(0, 0),
		},
		{
			a:   "0",
			exp: apd.New(0, 0),
		},
		{
			a:   "Inf",
			err: ErrNotFinite,
		},
		{
			a:   "NaN",
			err: ErrNotFinite,
		},
		{
			a:   "1.1234",
			err: ErrInvalidPrecision,
		},
		{
			a:   "1",
			exp: apd.New(1, 0),
		},
		{
			a:   "1.1",
			exp: apd.New(11, -1),
		},
		{
			a:   "0.00",
			exp: apd.New(0, -2),
		},
		{
			a:   "1.10",
			exp: apd.New(110, -2),
		},
		{
			a:   "123.45",
			exp: apd.New(12345, -2),
		},
	}

	for _, tc := range cases {
		t.Run(tc.a, func(t *testing.T) {
			exp, err := ParseCurrency(tc.a)
			if tc.err != nil {
				require.Error(t, err)
				require.Equal(t, tc.err.Error(), err.Error(), "%v != %v", tc.err, err)
				require.Nil(t, exp)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, tc.exp)
			require.NotNil(t, exp)
			require.Equal(t, 0, tc.exp.Cmp(exp))
		})
	}
}

func mustNewFromString(t *testing.T, a string) *apd.Decimal {
	x, cond, err := apd.NewFromString(a)
	require.NoError(t, err)
	require.False(t, cond.Any())
	return x
}

func TestValidateTransferAmount(t *testing.T) {
	cases := []struct {
		name   string
		amount *apd.Decimal
		err    error
	}{
		{
			name:   "nil amount",
			amount: nil,
			err:    ErrAmountNil,
		},
		{
			name:   "not finite",
			amount: mustNewFromString(t, "Inf"),
			err:    ErrNotFinite,
		},
		{
			name:   "not > 0",
			amount: apd.New(0, 0),
			err:    ErrAmountNotMoreThanZero,
		},
		{
			name:   "not > 0",
			amount: mustNewFromString(t, "-0"),
			err:    ErrAmountNotMoreThanZero,
		},
		{
			name:   "not > 0",
			amount: apd.New(-123, -2),
			err:    ErrAmountNotMoreThanZero,
		},
		{
			name:   "invalid precision",
			amount: apd.New(123, -3),
			err:    ErrInvalidPrecision,
		},

		// Valid cases
		{
			name:   "10",
			amount: apd.New(1, 1),
			err:    nil,
		},
		{
			name:   "1",
			amount: apd.New(1, 0),
			err:    nil,
		},
		{
			name:   "1.2",
			amount: apd.New(12, -1),
			err:    nil,
		},
		{
			name:   "1.23",
			amount: apd.New(123, -2),
			err:    nil,
		},
		{
			name:   "1.0",
			amount: apd.New(1, 0),
			err:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateTransferAmount(tc.amount)
			require.Equal(t, tc.err, err)
		})
	}
}
