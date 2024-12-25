package tests

import (
	"testing"

	"github.com/Shubham-Hazra/pokedexcli/utils"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "  hello  world  ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "Charmander     CHarizard  squirtle ",
			expected: []string{"charmander", "charizard", "squirtle"},
		},
	}
	for _, c := range cases {
		actual := utils.CleanInput(c.input)
		if len(actual) != len(c.expected) {
			t.Errorf("The len of actual: %v, does not match the len of expected: %v\n", len(actual), len(c.expected))
		}
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if word != expectedWord {
				t.Errorf("The actual word: %v, does not match the expected word: %v\n", word, expectedWord)
			}
		}
	}
}
