package finance

import (
	"errors"
	"math"
)

const (
	// MaxIterations determines the maximum number of iterations performed by the Newton-Raphson algorithm.
	// MaxIterations确定由Newton-Raphson算法执行的最大迭代次数。
	MaxIterations = 30
	// Precision determines how close to the solution the Newton-Raphson algorithm should arrive before stopping.
	// 精度确定了牛顿-拉夫森算法在停止之前应该到达解决方案的程度。
	Precision = 1E-6
)

func newton(guess float64, function func(float64) float64, derivative func(float64) float64, numIt int) (float64, error) {
	x := guess - function(guess)/derivative(guess)
	if math.Abs(x-guess) < Precision {
		return x, nil
	} else if numIt >= MaxIterations {
		return 0, errors.New("solution didn't converge")
	} else {
		return newton(x, function, derivative, numIt+1)
	}
}
