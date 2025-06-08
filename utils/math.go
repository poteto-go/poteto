package utils

import "strings"

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

func BuildString(strArrays ...string) string {
	// 文字列結合を最適化（StringBuilder pattern）
	var built string
	if len(strArrays) == 1 {
		built = strArrays[0]
	} else {
		// 事前に容量を計算してallocation回数を削減
		totalLen := 0
		for _, v := range strArrays {
			totalLen += len(v) + 2 // ", "の分
		}

		builder := strings.Builder{}
		builder.Grow(totalLen)
		for i, v := range strArrays {
			if i > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(v)
		}
		built = builder.String()
	}

	return built
}
