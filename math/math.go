package math

func MinFloat32(a float32, b float32) float32 {
	if a < b {
		return a
	} else {
		return b
	}
}

func MaxFloat32(a float32, b float32) float32 {
	if a > b {
		return a
	} else {
		return b
	}
}

func MaxUint8(a uint8, b uint8) uint8 {
	if a > b {
		return a
	} else {
		return b
	}
}
