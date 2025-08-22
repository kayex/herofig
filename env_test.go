package main_test

import (
	"bytes"
	"maps"
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
		{" KEY=value", "KEY", "value"},
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

func TestParse(t *testing.T) {
	cases := []struct {
		e    string
		want Config
	}{
		{
			`KEY1=value

			KEY2=value`,
			Config{
				"KEY1": "value",
				"KEY2": "value",
			},
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			cfg, err := Parse(bytes.NewBufferString(c.e))
			if err != nil {
				t.Fatalf("Parse(%q): %v", c.e, err)
			}

			if !maps.Equal(cfg, c.want) {
				t.Errorf("Parse(%q) = %v; want %v", c.e, cfg, c.want)
			}
		})
	}
}
