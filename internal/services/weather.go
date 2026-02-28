package services

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func GetTemperature(location string) (string, error) {
	endpoint := fmt.Sprintf("https://wttr.in/%s?format=%%t", url.QueryEscape(location))

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "curl/7.68.0")

	client := &http.Client{
		Transport: &http.Transport{
			TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		},
	}

	resp, err := client.Do(req)
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
