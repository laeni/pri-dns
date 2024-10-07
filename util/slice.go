package util

import (
	"github.com/coredns/coredns/plugin/pkg/rand"
	"time"
)

var (
	rn = rand.New(time.Now().UnixNano())
)

// SliceRandom 返回经过乱序后的 p
func SliceRandom[T any](p []T) []T {
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

// SliceDeduplication 对切片元素进行去重，但是需要确保切片是可比较的
func SliceDeduplication[T comparable](ts []T) []T {
	l := len(ts)
	hisMap := make(map[T]struct{}, l)
	for _, h := range ts {
		hisMap[h] = struct{}{}
	}
	i := 0
	for s := range hisMap {
		ts[i] = s
		i++
	}
	return ts[:i]
}

// SliceEqual 比较切片元素的值是否相等，忽略切片元素的顺序
func SliceEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for _, v := range a {
		found := false
		for _, v2 := range b {
			if v == v2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
