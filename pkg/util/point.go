package util

type PointAble interface {
	// string |
	// 	int | int64 | int32 | uint | uint64 |
	// 	float32 | float64 |
	// 	time.Time | time.Duration
}

func Point[T PointAble](v T) *T {
	return &v
}
