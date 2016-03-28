package utils

func MinF32(a float32, b float32) float32 {
	if a < b {
		return a
	} else {
		return b
	}
}

func MaxF32(a float32, b float32) float32 {
	if a > b {
		return a
	} else {
		return b
	}
}

func MinU8(a uint8, b uint8) uint8 {
	if a < b {
		return a
	} else {
		return b
	}
}

func MaxU8(a uint8, b uint8) uint8 {
	if a > b {
		return a
	} else {
		return b
	}
}

func MinInt(a int, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func MaxInt(a int, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
