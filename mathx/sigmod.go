package mathx

import "math"

// 平滑 & 平移的 Sigmoid 函数
func SmoothShiftedSigmoid(x, alpha, x0 float64) float64 {
	return 1 / (1 + math.Exp(-alpha*(x-x0)))
}
