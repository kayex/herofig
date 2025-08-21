package main_test

import (
	"testing"

	. "github.com/kayex/herofig"
)

func TestParsePair(t *testing.T) {
	cases := []struct {
		pair  string
		key   string
		value string
	}{
		{"KEY=value", "KEY", "value"},
		{"KEY=value=value", "KEY", "value=value"},
	}

	for _, c := range cases {
		t.Run(c.pair, func(t *testing.T) {
			key, value, err := ParsePair(c.pair)
			if err != nil {
				t.Fatalf("ParsePair(%s): %v", c.pair, err)
			}

			if key != c.key || value != c.value {
				t.Errorf("ParsePair(%s) = %s, %s; got %s, %s", c.pair, c.key, c.value, key, value)
			}
		})
	}
}

func TestParsePair_Errors(t *testing.T) {
	cases := []struct {
		name string
		pair string
	}{
		{"no key", "=value"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			key, value, err := ParsePair(c.pair)
			if err == nil {
				t.Errorf("ParsePair(%s) = %q, %q, %v; want error", c.pair, key, value, err)
			}
		})
	}
}
