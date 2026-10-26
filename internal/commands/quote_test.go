package commands

import "testing"

func TestQuotesURLForChannel(t *testing.T) {
	tests := []struct {
		channel string
		want    string
	}{
		{"pectha", quotesPageURL + "?channel=pectha"},
		{" Pizza_TM ", quotesPageURL + "?channel=pizza_tm"},
		{"", quotesPageURL},
	}

	for _, test := range tests {
		if got := quotesURLForChannel(test.channel); got != test.want {
			t.Errorf("quotesURLForChannel(%q) = %q, want %q", test.channel, got, test.want)
		}
	}
}
