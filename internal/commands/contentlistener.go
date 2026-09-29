package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"regexp"
	"strings"
)

var contentRegex = regexp.MustCompile(`[c\[({（［｛][^a-z]*o[^a-z]*n[^a-z]*t[^a-z]*e[^a-z]*n[^a-z]*t`)

var contentHomoglyphs = map[rune]rune{
	'0': 'o', '3': 'e', '7': 't', 'q': 'o', 'ʇ': 't', 'ǝ': 'e', '\u00BA': 'o', '\u00F8': 'o', '\u0254': 'o', '\u03B5': 'e',
	'\u03BF': 'o', '\u03C3': 'o', '\u03C4': 't', '\u03F2': 'c', '\u03F5': 'e', '\u0435': 'e', '\u043D': 'n',
	'\u043E': 'o', '\u043F': 'n', '\u0441': 'c', '\u0442': 't', '\u3007': 'o', '\uFF10': 'o', '\uFF13': 'e',
	'\uFF17': 't',

	'\u00C7': 'c', '\u00C8': 'e', '\u00C9': 'e', '\u00CA': 'e', '\u00CB': 'e', '\u00D1': 'n', '\u00D2': 'o',
	'\u00D3': 'o', '\u00D4': 'o', '\u00D5': 'o', '\u00D6': 'o', '\u00E7': 'c', '\u00E8': 'e', '\u00E9': 'e',
	'\u00EA': 'e', '\u00EB': 'e', '\u00F1': 'n', '\u00F2': 'o', '\u00F3': 'o', '\u00F4': 'o', '\u00F5': 'o',
	'\u00F6': 'o', '\u0106': 'c', '\u0107': 'c', '\u0108': 'c', '\u0109': 'c', '\u010A': 'c', '\u010B': 'c',
	'\u010C': 'c', '\u010D': 'c', '\u0112': 'e', '\u0113': 'e', '\u0114': 'e', '\u0115': 'e', '\u0116': 'e',
	'\u0117': 'e', '\u0118': 'e', '\u0119': 'e', '\u011A': 'e', '\u011B': 'e', '\u0143': 'n', '\u0144': 'n',
	'\u0145': 'n', '\u0146': 'n', '\u0147': 'n', '\u0148': 'n', '\u014C': 'o', '\u014D': 'o', '\u014E': 'o',
	'\u014F': 'o', '\u0150': 'o', '\u0151': 'o', '\u0162': 't', '\u0163': 't', '\u0164': 't', '\u0165': 't',
	'\u01A0': 'o', '\u01A1': 'o', '\u01D1': 'o', '\u01D2': 'o', '\u01EA': 'o', '\u01EB': 'o', '\u01F8': 'n',
	'\u01F9': 'n', '\u0204': 'e', '\u0205': 'e', '\u0206': 'e', '\u0207': 'e', '\u020C': 'o', '\u020D': 'o',
	'\u020E': 'o', '\u020F': 'o', '\u021A': 't', '\u021B': 't', '\u0228': 'e', '\u0229': 'e', '\u022E': 'o',
	'\u022F': 'o', '\u1E18': 'e', '\u1E19': 'e', '\u1E1A': 'e', '\u1E1B': 'e', '\u1E44': 'n', '\u1E45': 'n',
	'\u1E46': 'n', '\u1E47': 'n', '\u1E48': 'n', '\u1E49': 'n', '\u1E4A': 'n', '\u1E4B': 'n', '\u1E6A': 't',
	'\u1E6B': 't', '\u1E6C': 't', '\u1E6D': 't', '\u1E6E': 't', '\u1E6F': 't', '\u1E70': 't', '\u1E71': 't',
	'\u1E97': 't', '\u1EB8': 'e', '\u1EB9': 'e', '\u1EBA': 'e', '\u1EBB': 'e', '\u1EBC': 'e', '\u1EBD': 'e',
	'\u1ECC': 'o', '\u1ECD': 'o', '\u1ECE': 'o', '\u1ECF': 'o',

	'\u0274': 'n', '\u1D04': 'c', '\u1D07': 'e', '\u1D0F': 'o', '\u1D1B': 't', '\u1D49': 'e', '\u1D4B': 'e',
	'\u1D52': 'o', '\u1D53': 'o', '\u1D57': 't', '\u1D9C': 'c', '\u1DB0': 'n', '\u1DB1': 'o', '\u207F': 'n',
	'\u2091': 'e', '\u2092': 'o', '\u2099': 'n', '\u209C': 't', '\u2102': 'c', '\u2115': 'n', '\u212D': 'c',
	'\u212F': 'e', '\u2130': 'e', '\u2134': 'o', '\u2147': 'e', '\u249E': 'c', '\u24A0': 'e', '\u24A9': 'n',
	'\u24AA': 'o', '\u24AF': 't', '\u24B8': 'c', '\u24BA': 'e', '\u24C3': 'n', '\u24C4': 'o', '\u24C9': 't',
	'\u24D2': 'c', '\u24D4': 'e', '\u24DD': 'n', '\u24DE': 'o', '\u24E3': 't', '\uFF23': 'c', '\uFF25': 'e',
	'\uFF2E': 'n', '\uFF2F': 'o', '\uFF34': 't', '\uFF43': 'c', '\uFF45': 'e', '\uFF4E': 'n', '\uFF4F': 'o',
	'\uFF54': 't',

	'\U0001D402': 'c', '\U0001D404': 'e', '\U0001D40D': 'n', '\U0001D40E': 'o', '\U0001D413': 't',
	'\U0001D41C': 'c', '\U0001D41E': 'e', '\U0001D427': 'n', '\U0001D428': 'o', '\U0001D42D': 't',
	'\U0001D436': 'c', '\U0001D438': 'e', '\U0001D441': 'n', '\U0001D442': 'o', '\U0001D447': 't',
	'\U0001D450': 'c', '\U0001D452': 'e', '\U0001D45B': 'n', '\U0001D45C': 'o', '\U0001D461': 't',
	'\U0001D46A': 'c', '\U0001D46C': 'e', '\U0001D475': 'n', '\U0001D476': 'o', '\U0001D47B': 't',
	'\U0001D484': 'c', '\U0001D486': 'e', '\U0001D48F': 'n', '\U0001D490': 'o', '\U0001D495': 't',
	'\U0001D49E': 'c', '\U0001D4A9': 'n', '\U0001D4AA': 'o', '\U0001D4AF': 't', '\U0001D4B8': 'c',
	'\U0001D4C3': 'n', '\U0001D4C9': 't', '\U0001D4D2': 'c', '\U0001D4D4': 'e', '\U0001D4DD': 'n',
	'\U0001D4DE': 'o', '\U0001D4E3': 't', '\U0001D4EC': 'c', '\U0001D4EE': 'e', '\U0001D4F7': 'n',
	'\U0001D4F8': 'o', '\U0001D4FD': 't', '\U0001D508': 'e', '\U0001D511': 'n', '\U0001D512': 'o',
	'\U0001D517': 't', '\U0001D520': 'c', '\U0001D522': 'e', '\U0001D52B': 'n', '\U0001D52C': 'o',
	'\U0001D531': 't', '\U0001D53C': 'e', '\U0001D546': 'o', '\U0001D54B': 't', '\U0001D554': 'c',
	'\U0001D556': 'e', '\U0001D55F': 'n', '\U0001D560': 'o', '\U0001D565': 't', '\U0001D56E': 'c',
	'\U0001D570': 'e', '\U0001D579': 'n', '\U0001D57A': 'o', '\U0001D57F': 't', '\U0001D588': 'c',
	'\U0001D58A': 'e', '\U0001D593': 'n', '\U0001D594': 'o', '\U0001D599': 't', '\U0001D5A2': 'c',
	'\U0001D5A4': 'e', '\U0001D5AD': 'n', '\U0001D5AE': 'o', '\U0001D5B3': 't', '\U0001D5BC': 'c',
	'\U0001D5BE': 'e', '\U0001D5C7': 'n', '\U0001D5C8': 'o', '\U0001D5CD': 't', '\U0001D5D6': 'c',
	'\U0001D5D8': 'e', '\U0001D5E1': 'n', '\U0001D5E2': 'o', '\U0001D5E7': 't', '\U0001D5F0': 'c',
	'\U0001D5F2': 'e', '\U0001D5FB': 'n', '\U0001D5FC': 'o', '\U0001D601': 't', '\U0001D60A': 'c',
	'\U0001D60C': 'e', '\U0001D615': 'n', '\U0001D616': 'o', '\U0001D61B': 't', '\U0001D624': 'c',
	'\U0001D626': 'e', '\U0001D62F': 'n', '\U0001D630': 'o', '\U0001D635': 't', '\U0001D63E': 'c',
	'\U0001D640': 'e', '\U0001D649': 'n', '\U0001D64A': 'o', '\U0001D64F': 't', '\U0001D658': 'c',
	'\U0001D65A': 'e', '\U0001D663': 'n', '\U0001D664': 'o', '\U0001D669': 't', '\U0001D672': 'c',
	'\U0001D674': 'e', '\U0001D67D': 'n', '\U0001D67E': 'o', '\U0001D683': 't', '\U0001D68C': 'c',
	'\U0001D68E': 'e', '\U0001D697': 'n', '\U0001D698': 'o', '\U0001D69D': 't',
}

func normalizeContent(msg string) string {
	var b strings.Builder
	b.Grow(len(msg))
	for _, r := range strings.ToLower(msg) {
		if folded, ok := contentHomoglyphs[r]; ok {
			b.WriteRune(folded)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

var contentFlips = map[rune]rune{
	'ɔ': 'c',
	'u': 'n',
	'ʇ': 't',
	'ǝ': 'e',
	'⊥': 't',
}

func normalizeFlipped(msg string) string {
	runes := []rune(strings.ToLower(msg))
	var b strings.Builder
	b.Grow(len(msg))
	for i := len(runes) - 1; i >= 0; i-- {
		if flipped, ok := contentFlips[runes[i]]; ok {
			b.WriteRune(flipped)
			continue
		}
		b.WriteRune(runes[i])
	}
	return normalizeContent(b.String())
}

func containsContent(msg string) bool {
	return contentRegex.MatchString(normalizeContent(msg)) ||
		contentRegex.MatchString(normalizeFlipped(msg))
}

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "content",
		Description: "Times out people who say 'content'",
		Handler: func(ctx bot.CommandContext) bool {
			if containsContent(ctx.Message.Text) {
				services.TwitchServiceInstance.Timeout(ctx.Message.Channel, ctx.Message.User.ID, 69, "don't use the C word rar")
				ctx.Client.Say(ctx.Message.Channel, "don't use the C word rar")
				return true
			}
			return false
		},
	})
}
