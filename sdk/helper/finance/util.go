package finance

import "strconv"


// StrconvFloat 浮点转成百分数
func StrconvFloat(f float64) string {
	stringf := strconv.FormatFloat(f, 'f', 2, 64)
	return stringf + "%"
}
