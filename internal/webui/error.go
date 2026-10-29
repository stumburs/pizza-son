package webui

import "fmt"

func errorDocumentTitle(status int) string {
	switch status {
	case 404:
		return "404 · pizza-son"
	case 500:
		return "500 · pizza-son"
	default:
		return "Error · pizza-son"
	}
}

func errorHeading(status int) string {
	switch status {
	case 404:
		return "404 — page not found"
	case 500:
		return "500 — something broke"
	default:
		return fmt.Sprintf("%d — something went wrong", status)
	}
}

func errorMessage(status int) string {
	switch status {
	case 404:
		return "This page doesn't exist. Maybe a bert ate it."
	case 500:
		return "The server tripped over a slice. Try again in a moment."
	default:
		return "Something went wrong. Try again in a moment."
	}
}
