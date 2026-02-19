package services

import "strings"

func FillPlaceholders(prompt string) string {
	// TODO: Replace with proper placeholders
	replacer := strings.NewReplacer(
		"{{game_name}}", "trackmania",
		"{{channel_name}}", "pizza_tm",
		"{{stream_title}}", "Creamstream",
	)

	result := replacer.Replace(prompt)

	return result
}
