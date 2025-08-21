package main

import (
	"fmt"
	"sort"

	"github.com/kayex/herofig/internal/hash"
)

type Config map[string]string

type Var struct {
	Key   string
	Value string
}

func (v Var) String() string {
	return fmt.Sprintf("%s=%s", v.Key, v.Value)
}

func (c Config) Hash() hash.Hash {
	ordered := c.Ordered()
	lines := make([]string, len(ordered))
	for i, v := range ordered {
		lines[i] = v.String()
	}
	return hash.New(lines)
}

func (c Config) Ordered() []Var {
	lines := make([]Var, 0, len(c))
	for k, v := range c {
		lines = append(lines, Var{k, v})
	}
	sort.Slice(lines, func(i, j int) bool {
		return lines[i].Key < lines[j].Key
	})
	return lines
}
