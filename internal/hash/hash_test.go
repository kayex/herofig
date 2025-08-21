package hash_test

import (
	"bytes"
	"testing"

	. "github.com/kayex/herofig/internal/hash"
)

func Test_HashMap(t *testing.T) {
	cases := []struct {
		m    map[string]string
		hash []byte
	}{
		{
			map[string]string{"KEY": "value"},
			[]byte{105, 250, 214, 32, 35, 38, 216, 52, 114, 192, 226, 120, 253, 183, 126, 37, 171, 15, 216, 247},
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			h := Map(c.m)
			if !bytes.Equal(h, c.hash) {
				t.Errorf("Hash() = %v, want %v", h, c.hash)
			}
		})
	}
}
