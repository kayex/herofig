package main_test

import (
	"testing"

	. "github.com/kayex/herofig"
)

func TestParseVar(t *testing.T) {
	cases := []struct {
		v     string
		key   string
		value string
	}{
		{"KEY=value", "KEY", "value"},
		{"KEY=value=value", "KEY", "value=value"},
	}

	for _, c := range cases {
		t.Run(c.v, func(t *testing.T) {
			key, value, err := ParseVar(c.v)
			if err != nil {
				t.Fatalf("ParseVar(%s): %v", c.v, err)
			}

			if key != c.key || value != c.value {
				t.Errorf("ParseVar(%s) = %s, %s; got %s, %s", c.v, c.key, c.value, key, value)
			}
		})
	}
}

func TestParseVar_Errors(t *testing.T) {
	cases := []struct {
		name string
		v    string
	}{
		{"no key", "=value"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			key, value, err := ParseVar(c.v)
			if err == nil {
				t.Errorf("ParseVar(%s) = %q, %q, %v; want error", c.v, key, value, err)
			}
		})
	}
}
