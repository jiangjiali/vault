package finance

import (
	"errors"
	"math"
)

// EffectiveRate returns the effective interest rate given the nominal rate and the number of compounding payments per year.
// EffectiveRate在给定名义利率和每年复利支付次数的情况下返回有效利率。
// Excel equivalent: EFFECT
// Excel等效项：EFFECT
func EffectiveRate(nominal float64, numPeriods int) (float64, error) {
	if numPeriods < 0 {
		return 0, errors.New("numPeriods must be strictly positive")
	}
	return math.Pow(1+nominal/float64(numPeriods), float64(numPeriods)) - 1, nil
}

// NominalRate returns the nominal interest rate given the effective rate and the number of compounding payments per year.
// NominalRate会根据实际利率和每年的复利支付次数返回名义利率。
// Excel equivalent: NOMINAL
// Excel等效项：NOMINAL
func NominalRate(effectiveRate float64, numPeriods int) (float64, error) {
	if numPeriods < 0 {
		return 0, errors.New("number of compounding payments per year must be strictly positive")
	}
	return float64(numPeriods) * (math.Pow(effectiveRate+1, 1/float64(numPeriods)) - 1), nil
}
