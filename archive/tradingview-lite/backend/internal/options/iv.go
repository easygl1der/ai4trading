package options

import (
	"errors"
	"math"
)

const tradingDaysPerYear = 252.0

type ivInput struct {
	Kind          string
	Spot          float64
	Strike        float64
	TimeYears     float64
	RiskFreeRate  float64
	DividendYield float64
	Price         float64
}

func solveImpliedVol(input ivInput) (float64, error) {
	if input.Spot <= 0 || input.Strike <= 0 || input.TimeYears <= 0 || input.Price <= 0 {
		return 0, errors.New("invalid IV input: spot, strike, time, and price must be positive")
	}

	lowerBound := optionIntrinsicLowerBound(input)
	if input.Price < lowerBound-1e-6 {
		return 0, errors.New("invalid IV input: option price is below discounted intrinsic value")
	}

	lo := 1e-4
	hi := 5.0
	hiPrice := blackScholesPrice(input, hi)
	for hiPrice < input.Price && hi < 10.0 {
		hi *= 1.5
		hiPrice = blackScholesPrice(input, hi)
	}
	if hiPrice < input.Price {
		return 0, errors.New("unable to bracket implied volatility")
	}

	for i := 0; i < 100; i++ {
		mid := (lo + hi) / 2
		price := blackScholesPrice(input, mid)
		if math.Abs(price-input.Price) < 1e-6 {
			return mid, nil
		}
		if price > input.Price {
			hi = mid
		} else {
			lo = mid
		}
	}
	return (lo + hi) / 2, nil
}

func blackScholesPrice(input ivInput, sigma float64) float64 {
	sqrtT := math.Sqrt(input.TimeYears)
	if sigma <= 0 || sqrtT <= 0 {
		return optionIntrinsicLowerBound(input)
	}
	d1 := (math.Log(input.Spot/input.Strike) + (input.RiskFreeRate-input.DividendYield+0.5*sigma*sigma)*input.TimeYears) / (sigma * sqrtT)
	d2 := d1 - sigma*sqrtT
	discountedSpot := input.Spot * math.Exp(-input.DividendYield*input.TimeYears)
	discountedStrike := input.Strike * math.Exp(-input.RiskFreeRate*input.TimeYears)
	if input.Kind == "put" {
		return discountedStrike*normCDF(-d2) - discountedSpot*normCDF(-d1)
	}
	return discountedSpot*normCDF(d1) - discountedStrike*normCDF(d2)
}

func optionIntrinsicLowerBound(input ivInput) float64 {
	discountedSpot := input.Spot * math.Exp(-input.DividendYield*input.TimeYears)
	discountedStrike := input.Strike * math.Exp(-input.RiskFreeRate*input.TimeYears)
	if input.Kind == "put" {
		return math.Max(discountedStrike-discountedSpot, 0)
	}
	return math.Max(discountedSpot-discountedStrike, 0)
}

func normCDF(x float64) float64 {
	return 0.5 * (1 + math.Erf(x/math.Sqrt2))
}

func oneTradingDayMove(spot, annualizedIV float64) (float64, float64) {
	if spot <= 0 || annualizedIV <= 0 {
		return 0, 0
	}
	move := spot * annualizedIV / math.Sqrt(tradingDaysPerYear)
	return move, move / spot * 100
}
