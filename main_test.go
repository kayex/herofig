package main

import (
	"fmt"
	"slices"
	"testing"
)

func TestSubstringSearch(t *testing.T) {
	cases := []struct {
		haystack string
		needle   string
		want     []int
	}{
		{"SOME_KEY", "SOME", []int{0}},
		{"SOME_KEY", "some", []int{0}},
		{"some_key", "SOME", []int{0}},
		{"SOME_KEY", "KEY", []int{5}},
		{"A_A_KEY", "A", []int{0, 2}},
		{"SOME_KEY", "NOT_HERE", nil},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s in %s", c.needle, c.haystack), func(t *testing.T) {
			i := substringSearch(c.haystack, c.needle)
			if !slices.Equal(i, c.want) {
				t.Errorf("substringSearch(%s, %s) = %v; want %v", c.haystack, c.needle, i, c.want)
			}
		})
	}
}
