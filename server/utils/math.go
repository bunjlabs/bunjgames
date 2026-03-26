package utils

func Abs[T int | int8 | int16 | int32 | int64](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

func Max[T int | int8 | int16 | int32 | int64](arr []T) (int, T) {
	maxIndex, maxValue := -1, T(0)
	for index, value := range arr {
		if maxIndex < 0 || value > maxValue {
			maxIndex, maxValue = index, value
		}
	}
	return maxIndex, maxValue
}
