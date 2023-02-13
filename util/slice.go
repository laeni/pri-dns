package util

import (
	"github.com/coredns/coredns/plugin/pkg/rand"
	"time"
)

var (
	rn = rand.New(time.Now().UnixNano())
)

// SliceRemoveItem 删除切片中指定位置的元素，相比于常用的 `s = append(s[:i], s[i+1:]...)` 具有更高的效率
func SliceRemoveItem[T interface{}](s []T, i int) []T {
	l := len(s) - 1
	for i := i; i < l; i++ {
		s[i] = s[i+1]
	}
	return s[:l]
}

// RandomList 返回经过乱序后的 p
func RandomList[T interface{}](p []T) []T {
	switch len(p) {
	case 1:
		return p
	case 2:
		if rn.Int()%2 == 0 {
			return []T{p[1], p[0]} // swap
		}
		return p
	}

	perms := rn.Perm(len(p))
	rnd := make([]T, len(p))

	for i, p1 := range perms {
		rnd[i] = p[p1]
	}
	return rnd
}
