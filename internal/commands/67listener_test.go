package commands

import "testing"

func TestSixSevenRegex(t *testing.T) {
	matches := []string{
		"67",
		"6-7",
		"6 uh 7",
		"67cat",
		"67676767",
		"a 67 cat",
		"six 7",
		"6 seven",
		"six seven",
		"Six sEven",
		"sixty seven",
		"давай 67 раз",
		"б7",
		"sick seven",
		"sicks 7",
		"i wonder if 6 will get tested when i put 7 here",
		"sex seven",
		"VI VII",
		"VI seven",
		"six vii",
		"sicksty 7",
	}

	nonMatches := []string{
		"",
		"7",
		"68",
		"777",
		"six",
		"seven",
		"sixty",
		"six eight",
		"sixty eight",
		"69",
	}

	for _, s := range matches {
		if !sixSevenRegex.MatchString(s) {
			t.Errorf("expected %q to match sixSevenRegex", s)
		}
	}

	for _, s := range nonMatches {
		if sixSevenRegex.MatchString(s) {
			t.Errorf("expected %q NOT to match sixSevenRegex", s)
		}
	}
}
