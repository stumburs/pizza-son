package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type geoResult struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
	Country   string  `json:"country"`
}

type geoResponse struct {
	Results []geoResult `json:"results"`
}

func GetGeoForCity(city string) (*geoResult, error) {
	u := "https://geocoding-api.open-meteo.com/v1/search?" + url.Values{
		"name":  {city},
		"count": {"1"},
	}.Encode()

	resp, err := http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("geocode request failed: %w", err)
	}
	defer resp.Body.Close()

	var gr geoResponse
	if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return nil, fmt.Errorf("geocode decode failed: %w", err)
	}
	if len(gr.Results) == 0 {
		return nil, fmt.Errorf("no location found for %q", city)
	}
	return &gr.Results[0], nil
}
