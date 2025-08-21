package hash

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
)

type Hash []byte

func Map(m map[string]string) Hash {
	lines := make([]string, 0, len(m))
	for k, v := range m {
		lines = append(lines, fmt.Sprintf("%s=%s\n", k, v))
	}
	sort.Strings(lines)

	h := sha1.New()
	for _, l := range lines {
		h.Write([]byte(l))
	}

	return h.Sum(nil)
}

func (h Hash) Mnemonic(length uint) string {
	maxLength := uint(len(h) / 2)
	if length > maxLength {
		panic(fmt.Errorf("length of %d exceeds maximum possible length %d for hash of length %d", length, maxLength, len(h)*8))
	}

	parts := make([]string, 0, length)
	for i := uint(0); i < length*2; i += 2 {
		parts = append(parts, string(proquint(binary.BigEndian.Uint16(h[i:i+2]))))
	}

	return strings.Join(parts, "-")
}

// proquint returns a deterministic, pronounceable quintuplet of alternating unambiguous consonants and vowels
// based on the value of x.
// https://arxiv.org/html/0901.4016
func proquint(x uint16) []byte {
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
