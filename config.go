package main

import "github.com/kayex/herofig/internal/hash"

type Config map[string]string

func (c Config) Hash() hash.Hash {
	return hash.Map(c)
}
