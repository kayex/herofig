package hash_test

import (
	"bytes"
	"testing"

	. "github.com/kayex/herofig/internal/hash"
)

func Test_New(t *testing.T) {
	cases := []struct {
		s    []string
		hash []byte
	}{
		{
			[]string{"KEY=value"},
			[]byte{164, 64, 154, 240, 97, 235, 134, 185, 61, 69, 79, 191, 125, 194, 149, 0, 9, 39, 31, 89},
		},
	}

	for _, c := range cases {
		t.Run(c.s[0], func(t *testing.T) {
			h := New(c.s)
			if !bytes.Equal(h, c.hash) {
				t.Errorf("Hash() = %v, want %v", h, c.hash)
			}
		})
	}
}
