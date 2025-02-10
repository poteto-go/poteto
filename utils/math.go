package utils

func IsSliceEqual[T comparable](as, bs []T) bool {
	if len(as) != len(bs) {
		return false
	}

	for i := 0; i < len(as); i++ {
		if as[i] != bs[i] {
			return false
		}
	}

	return true
}
