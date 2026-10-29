package bot

import (
	"strings"
	"unicode"
)

// isInvisible reports whether r is a character that renders as nothing
// (or whitespace) and should never be significant at the edges of a message.
func isInvisible(r rune) bool {
	switch {
	case unicode.IsSpace(r): // includes NBSP, U+3000, etc.
		return true
	case unicode.IsControl(r): // Cc
		return true
	case unicode.Is(unicode.Cf, r): // zero-width space/joiners, soft hyphen, BOM, tag chars E0001+
		return true
	case r == 0x034F: // combining grapheme joiner
		return true
	case r >= 0xE0000 && r <= 0xE007F: // tag block, incl. unassigned U+E0000
		return true
	case r >= 0xFE00 && r <= 0xFE0F: // variation selectors
		return true
	}
	return false
}

// SanitizeText removes invisible characters from the start and end of a message.
// Inner characters are left alone so emoji sequences (ZWJ, variation selectors)
// keep working.
func SanitizeText(s string) string {
	return strings.TrimFunc(s, isInvisible)
}
