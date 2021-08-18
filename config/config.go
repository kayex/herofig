package config

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
)

type Config map[string]string

func (c Config) Hash() Hash {
	lines := make([]string, 0, len(c))
	for k, v := range c {
		lines = append(lines, fmt.Sprintf("%s=%s\n", k, v))
	}
	sort.Strings(lines)

	h := sha1.New()
	for _, l := range lines {
		h.Write([]byte(l))
	}

	return h.Sum(nil)
}

type Hash []byte

func (h Hash) Mnemonic() string {
	// We can fit 2 byte per part.
	parts := make([]string, 0, len(h)/2)

	for i := 0; i < 4; i += 2 {
		parts = append(parts, string(mnemonize(binary.BigEndian.Uint16(h[i:i+2]))))
	}

	return strings.Join(parts, "-")
}

func mnemonize(x uint16) []byte {
	vowels := []byte("aiou")
	consonants := []byte("bdfghjklmnprstvz")

	cons3 := x & 0x0f
	x >>= 4
	vow2 := x & 0x03
	x >>= 2
	cons2 := x & 0x0f
	x >>= 4
	vow1 := x & 0x03
	x >>= 2
	cons1 := x & 0x0f

	m := make([]byte, 5)
	m[0] = consonants[cons1]
	m[1] = vowels[vow1]
	m[2] = consonants[cons2]
	m[3] = vowels[vow2]
	m[4] = consonants[cons3]

	return m
}
