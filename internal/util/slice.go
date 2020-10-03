package util

func Unique(xs []int64) []int64 {
	seen := make(map[int64]bool)
	unique := make([]int64, 0)
	for _, x := range xs {
		if _, value := seen[x]; !value {
			seen[x] = true
			unique = append(unique, x)
		}
	}
	return unique
}
