package commands

import "testing"

func TestURLRegex(t *testing.T) {
	matches := []string{
		"google.com",
		"http://twitch.tv",
		"https://www.youtube.com/",
		"https://pizzason.stumburs.id.lv/bertstats",
		"pizzason.stumburs.id.lv/bertstats",
		"https://pizzason.stumburs.id.lv/bertstats",
		"https://pizzason.stumburs.id.lv/panel",
		"pizzason.stumburs.id.lv/panel",
		"https://www.twitch.tv/pizza_tm/clip/DifferentFurtiveCaterpillarArsonNoSexy-QfPFQwCIKsZ_MmVT?filter=clips&range=all",
		"www.twitch.tv/pizza_tm/clip/DifferentFurtiveCaterpillarArsonNoSexy-QfPFQwCIKsZ_MmVT?filter=clips&range=all",
		"https://twitch.tv/pizza_tm/clip/DifferentFurtiveCaterpillarArsonNoSexy-QfPFQwCIKsZ_MmVT?filter=clips&range=all",
		"twitch.tv/pizza_tm/clip/DifferentFurtiveCaterpillarArsonNoSexy-QfPFQwCIKsZ_MmVT?filter=clips&range=all",
	}

	nonMatches := []string{}

	for _, s := range matches {
		if !urlRegex.MatchString(s) {
			t.Errorf("expected %q to match urlRegex", s)
		}
	}

	for _, s := range nonMatches {
		if urlRegex.MatchString(s) {
			t.Errorf("expected %q NOT to match urlRegex", s)
		}
	}
}
