package store

import (
	"errors"

	"github.com/shopspring/decimal"
)

var (
	forexBase  = "BTC"
	forexPairs = map[string]decimal.Decimal{
		"GBP": decimal.RequireFromString("40000.00"),
		"EUR": decimal.RequireFromString("45000.00"),
		"BTC": decimal.RequireFromString("1.00000000"),
		"ETH": decimal.RequireFromString("14.00000000"),
	}
)

func ForexRate(pairFrom, pairTo string) (ret decimal.Decimal, err error) {
	if pairFrom == pairTo {
		return decimal.RequireFromString("1"), nil
	}

	ret = decimal.RequireFromString("1")

	// If we're not converting from the base then we first need to convert the source amount
	// to the base currency
	if pairFrom != forexBase {
		baseRate, ok := forexPairs[pairFrom]
		if !ok {
			return decimal.Zero, errors.New("invalid currency pair")
		}

		ret = ret.Div(baseRate)
	}

	// Lookup the rate in the table
	rate, ok := forexPairs[pairTo]
	if !ok {
		return decimal.Zero, errors.New("invalid currency pair")
	}

	ret = ret.Mul(rate)

	return ret, nil
}
