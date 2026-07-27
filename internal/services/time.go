package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type timeResponse struct {
	Datetime     string `json:"datetime"`
	Timezone     string `json:"timezone"`
	UTCOffset    string `json:"utc_offset"`
	Abbreviation string `json:"abbreviation"`
}

func GetCurrentTimeForTimezone(tz string) (*timeResponse, error) {
	u := "https://time.now/developer/api/timezone/" + url.PathEscape(tz)

	resp, err := http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("time request failed: %w", err)
	}
	defer resp.Body.Close()

	var tr timeResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, fmt.Errorf("time decode failed: %w", err)
	}
	return &tr, nil
}

func GetTimeForCity(city string) (*timeResponse, string, error) {
	loc, err := GetGeoForCity(city)
	if err != nil {
		return nil, "", err
	}

	t, err := GetCurrentTimeForTimezone(loc.Timezone)
	if err != nil {
		return nil, "", err
	}

	return t, loc.Timezone, nil
}

func DatetimeToHourMinute(datetime string) (string, error) {
	t, err := time.Parse(time.RFC3339, datetime)
	if err != nil {
		return "", fmt.Errorf("could not parse datetime %q: %w", datetime, err)
	}
	return t.Format("15:04"), nil
}
