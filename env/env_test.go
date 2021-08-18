package env

import (
	"testing"
)

func TestParsePair(t *testing.T) {
	cases := []struct {
		pair  string
		key   string
		value string
		err   bool
	}{
		{"KEY=value", "KEY", "value", false},
		{"=value", "", "", true},
		{"VALUE_WITH_EQUALS_SIGN=base64:QmVvbm9kZQ==", "VALUE_WITH_EQUALS_SIGN", "base64:QmVvbm9kZQ==", false},
	}

	for _, c := range cases {
		key, value, err := ParsePair(c.pair)

		if err == nil {
			if c.err {
				t.Errorf("expected ParsePair(%s) to return error, got nil.", c.pair)
				continue
			}
		} else {
			if !c.err {
				t.Errorf("ParsePair(%s): %v", c.pair, err)
				continue
			}
		}

		if key != c.key || value != c.value {
			t.Errorf("expected ParsePair(%s) to return (%s, %s), got (%s, %s)", c.pair, c.key, c.value, key, value)
		}
	}
}
