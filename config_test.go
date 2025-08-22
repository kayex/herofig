package main_test

import (
	"reflect"
	"testing"

	. "github.com/kayex/herofig"
)

func TestConfig_Ordered(t *testing.T) {
	cases := []struct {
		cfg  Config
		want []Var
	}{
		{
			Config{
				"A": "value",
				"B": "value",
				"C": "value",
			},
			[]Var{
				{"A", "value"},
				{"B", "value"},
				{"C", "value"},
			},
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			got := c.cfg.Ordered()
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("Ordered() = %v; want %v", got, c.want)
			}
		})
	}
}
