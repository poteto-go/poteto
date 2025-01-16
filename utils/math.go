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

// EX:) []string{"hello", "world"} -> hello,world
func StrArrayToStr(targets []string) string {
	var valStr string
	for i, target := range targets {
		valStr += target
		if i != len(targets)-1 {
			valStr += ","
		}
	}
	return valStr
}
