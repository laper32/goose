package slice

func RemoveDuplicate[T comparable](in []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range in {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
