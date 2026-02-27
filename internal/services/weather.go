package services

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func GetTemperature(location string) (string, error) {
	endpoint := fmt.Sprintf("https://wttr.in/%s?format=%%t", url.QueryEscape(location))

	resp, err := http.Get(endpoint)
	if err != nil {
		log.Println("[Weather] Request failed:", err)
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}
