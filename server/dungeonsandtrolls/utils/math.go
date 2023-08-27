package utils

import "golang.org/x/exp/constraints"

func Max[V constraints.Ordered](l, r V) V {
	if l < r {
		return r
	}
	return l
}
