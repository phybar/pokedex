package main

import "testing"

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
            input:    "hack the box",
            expected: []string{"hack", "the", "box"},
        },
        {
            input: "gobbles the    turkey  nom nom noooms",
            expected: []string{"gobbles", "the", "turkey", "nom", "nom", "noooms"}, 
        },
    }

    for _, c := range cases {
        actual := cleanInput(c.input)
        
        if len(actual) != len(c.expected) {
            t.Errorf("got %d words, want %d words", len(actual), len(c.expected))
            continue
        }

        for i := range actual {
            word := actual[i]
            expectedWord := c.expected[i]
            if word != expectedWord {
                t.Errorf("got %q at position %d, want %q", word, i, expectedWord)
            }
        }
    }
}