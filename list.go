package main

import (
	"fmt"
	"slices"
	"strings"
)

type List []string

func (l List) String() string {
	return strings.Join(l, ",")
}

func (l List) Equal(other List) bool {
	if len(l) != len(other) {
		return false
	}
	for i, v := range l {
		if v != other[i] {
			return false
		}
	}
	return true
}

func (l List) Diff(other List) string {
	if l.Equal(other) {
		return ""
	}

	addition := []string{}
	deletion := []string{}
	// addition
	for _, v := range other {
		found := slices.Contains(l, v)
		if !found {
			addition = append(addition, v)
		}
	}
	// deletion
	for _, v := range l {
		found := slices.Contains(other, v)
		if !found {
			deletion = append(deletion, v)
		}
	}

	s := ""
	if len(addition) > 0 {
		s += fmt.Sprintf("新增: %s\n", strings.Join(addition, ","))
	}
	if len(deletion) > 0 {
		s += fmt.Sprintf("刪除: %s\n", strings.Join(deletion, ","))
	}

	return s
}
