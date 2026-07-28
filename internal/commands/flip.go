package commands

import (
	"strings"

	"pizza-son/internal/bot"
)

var flipMap = map[rune]string{
	'a': "ɐ", 'b': "q", 'c': "ɔ", 'd': "p", 'e': "ǝ", 'f': "ɟ",
	'g': "ɓ", 'h': "ɥ", 'i': "ᴉ", 'j': "ɾ", 'k': "ʞ", 'l': "ʅ",
	'm': "ɯ", 'n': "u", 'o': "o", 'p': "d", 'q': "b", 'r': "ɹ",
	's': "s", 't': "ʇ", 'u': "n", 'v': "ʌ", 'w': "ʍ", 'x': "x",
	'y': "ʎ", 'z': "z",
	'A': "∀", 'B': "𐐒", 'C': "Ↄ", 'D': "◖", 'E': "Ǝ", 'F': "Ⅎ",
	'G': "⅁", 'H': "H", 'I': "I", 'J': "ſ", 'K': "⋊", 'L': "⅂",
	'M': "W", 'N': "N", 'O': "O", 'P': "Ԁ", 'Q': "Ό", 'R': "ᴚ",
	'S': "S", 'T': "⊥", 'U': "∩", 'V': "Λ", 'W': "M", 'X': "X",
	'Y': "⅄", 'Z': "Z",
	'0': "0", '1': "⇂", '2': "↊", '3': "↋", '4': "ᔭ", '5': "ߓ",
	'6': "9", '7': "ㄥ", '8': "8", '9': "6",
	'!': "¡", '?': "¿", '.': "˙", ',': "'", '\'': ",", '"': ",,",
	'`': ",", '(': ")", ')': "(", '[': "]", ']': "[", '{': "}",
	'}': "{", '<': ">", '>': "<", '&': "⅋", '_': "‾",
	':': ":", ';': "؛",
}

func flipText(s string) string {
	runes := []rune(s)
	var b strings.Builder
	b.Grow(len(s))
	for i := len(runes) - 1; i >= 0; i-- {
		if flipped, ok := flipMap[runes[i]]; ok {
			b.WriteString(flipped)
		} else {
			b.WriteRune(runes[i])
		}
	}
	return b.String()
}

func init() {
	Register(bot.Command{
		Name:        "flip",
		Description: "Flips text vertically (upside down).",
		Usage:       "!flip <text>",
		Category:    bot.CategoryFun,
		Examples: []bot.CommandExample{
			{Input: "!flip hello", Output: "(╯°□°）╯ oʅʅǝɥ"},
		},
		Handler: func(ctx bot.CommandContext) {
			if len(ctx.Args) == 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "(╯°□°）╯  Provide some text to flip!")
				return
			}
			flipped := flipText(strings.Join(ctx.Args, " "))
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "(╯°□°）╯ "+flipped)
		},
	})
}
