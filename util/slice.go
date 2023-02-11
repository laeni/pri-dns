package util

// SliceRemoveItem 删除切片中指定位置的元素，相比于常用的 `s = append(s[:i], s[i+1:]...)` 具有更高的效率
func SliceRemoveItem[T interface{}](s []T, i int) []T {
	l := len(s) - 1
	for i := i; i < l; i++ {
		s[i] = s[i+1]
	}
	return s[:l]
}
