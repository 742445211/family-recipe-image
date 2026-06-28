package firdgemate

import "testing"

func TestIsSemanticLabel(t *testing.T) {
	cases := map[string]bool{
		"egg":       true,
		"tomato":    true,
		"17":        false,
		"9":         false,
		"undefined": false,
		"":          false,
	}
	for name, want := range cases {
		if got := IsSemanticLabel(name); got != want {
			t.Fatalf("%q: got %v want %v", name, got, want)
		}
	}
}
