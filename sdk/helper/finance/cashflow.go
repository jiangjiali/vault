package finance

import (
	"errors"
	"math"
	"time"
)

// NetPresentValue returns the Net Present Value of a cash flow series given a discount rate
// NetPresentValue在给定折现率的情况下返回现金流量系列的净现值
// Excel equivalent: NPV
// Excel等效项：NPV
func NetPresentValue(rate float64, values []float64) float64 {
	npv := 0.0
	nper := len(values)
	for i := 1; i <= nper; i++ {
		npv += values[i-1] / math.Pow(1+rate, float64(i))
	}
	return npv
}

// InternalRateOfReturn returns the internal rate of return of a cash flow series.
// InternalRateOfReturn返回现金流量系列的内部收益率。
// Guess is a guess for the rate, used as a starting point for the iterative algorithm.
// Guess是对速率的猜测，用作迭代算法的起点。
// Excel equivalent: IRR
// Excel等效项：IRR
func InternalRateOfReturn(values []float64, guess float64) (float64, error) {
	min, max := minMaxSlice(values)
	if min*max >= 0 {
		return 0, errors.New("the cash flow must contain at least one positive value and one negative value")
	}

	function := func(rate float64) float64 {
		return NetPresentValue(rate, values)
	}
	derivative := func(rate float64) float64 {
		return dNetPresentValue(rate, values)
	}
	return newton(guess, function, derivative, 0)
}

func dNetPresentValue(rate float64, values []float64) float64 {
	dnpv := 0.0
	nper := len(values)
	for i := 1; i <= nper; i++ {
		dnpv -= values[i-1] * float64(i) / math.Pow(1+rate, float64(i+1))
	}
	return dnpv
}

// ModifiedInternalRateOfReturn returns the internal rate of return of a cash flow series, considering both financial and reinvestment rates
// ModifiedInternalRateOfReturn返回考虑财务和再投资率的现金流量系列的内部收益率
// financeRate is the rate on the money used in the cash flow.
// financeRate是现金流量中所用资金的比率。
// reinvestRate is the rate received when reinvested
// reinvestRate是再投资时收到的汇率
// Excel equivalent: MIRR
// Excel等效项：MIRR
func ModifiedInternalRateOfReturn(values []float64, financeRate float64, reinvestRate float64) (float64, error) {
	min, max := minMaxSlice(values)
	if min*max >= 0 {
		return 0, errors.New("the cash flow must contain at least one positive value and one negative value")
	}
	positiveFlows := make([]float64, 0)
	negativeFlows := make([]float64, 0)
	for _, value := range values {
		if value >= 0 {
			positiveFlows = append(positiveFlows, value)
			negativeFlows = append(negativeFlows, 0)
		} else {
			positiveFlows = append(positiveFlows, 0)
			negativeFlows = append(negativeFlows, value)
		}
	}
	nper := len(values)
	return math.Pow(-NetPresentValue(reinvestRate, positiveFlows)*math.Pow(1+reinvestRate, float64(nper))/NetPresentValue(financeRate, negativeFlows)/(1+financeRate), 1/float64(nper-1)) - 1, nil
}

// ScheduledNetPresentValue returns the Net Present Value of a scheduled cash flow series given a discount rate
// ScheduledNetPresentValue返回给定折扣率的计划现金流量系列的净现值
// Excel equivalent: XNPV
// Excel等效项：XNPV
func ScheduledNetPresentValue(rate float64, values []float64, dates []time.Time) (float64, error) {
	if len(values) != len(dates) {
		return 0, errors.New("values and dates must have the same length")
	}

	xnpv := 0.0
	nper := len(values)
	for i := 1; i <= nper; i++ {
		exp := dates[i-1].Sub(dates[0]).Hours() / 24.0 / 365.0
		xnpv += values[i-1] / math.Pow(1+rate, exp)
	}
	return xnpv, nil
}

// ScheduledInternalRateOfReturn returns the internal rate of return of a scheduled cash flow series.
// ScheduledInternalRateOfReturn返回计划的现金流量系列的内部收益率。
// Guess is a guess for the rate, used as a starting point for the iterative algorithm.
// Guess是对速率的猜测，用作迭代算法的起点。
// Excel equivalent: XIRR
// Excel等效项：XIRR
func ScheduledInternalRateOfReturn(values []float64, dates []time.Time, guess float64) (float64, error) {
	min, max := minMaxSlice(values)
	if min*max >= 0 {
		return 0, errors.New("the cash flow must contain at least one positive value and one negative value")
	}
	if len(values) != len(dates) {
		return 0, errors.New("values and dates must have the same length")
	}

	function := func(rate float64) float64 {
		r, _ := ScheduledNetPresentValue(rate, values, dates)
		return r
	}
	derivative := func(rate float64) float64 {
		r, _ := dScheduledNetPresentValue(rate, values, dates)
		return r
	}
	return newton(guess, function, derivative, 0)
}

func dScheduledNetPresentValue(rate float64, values []float64, dates []time.Time) (float64, error) {
	if len(values) != len(dates) {
		return 0, errors.New("values and dates must have the same length")
	}

	dxnpv := 0.0
	nper := len(values)
	for i := 1; i <= nper; i++ {
		exp := dates[i-1].Sub(dates[0]).Hours() / 24.0 / 365.0
		dxnpv -= values[i-1] * exp / math.Pow(1+rate, exp+1)
	}
	return dxnpv, nil
}

func minMaxSlice(values []float64) (float64, float64) {
	min := math.MaxFloat64
	max := -min
	for _, value := range values {
		if value > max {
			max = value
		}
		if value < min {
			min = value
		}
	}
	return min, max
}
