package price

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const coinMarketCapURL = "https://pro-api.coinmarketcap.com/v2/cryptocurrency/quotes/latest?symbol=STEEM"

type CoinMarketCap struct {
	APIKey string
}

func (CoinMarketCap) Name() string { return "CoinMarketCap" }

func (c CoinMarketCap) Fetch(ctx context.Context, client *http.Client) (float64, error) {
	headers := map[string]string{
		"X-CMC_PRO_API_KEY": c.APIKey,
		"Accept":            "application/json",
	}
	body, err := getJSON(ctx, client, coinMarketCapURL, headers)
	if err != nil {
		return 0, err
	}
	return parseCoinMarketCap(body)
}

func parseCoinMarketCap(body []byte) (float64, error) {
	var payload struct {
		Data map[string][]struct {
			Quote map[string]struct {
				Price float64 `json:"price"`
			} `json:"quote"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, err
	}
	entries, ok := payload.Data["STEEM"]
	if !ok || len(entries) == 0 {
		return 0, fmt.Errorf("STEEM data not found")
	}
	usd, ok := entries[0].Quote["USD"]
	if !ok || usd.Price == 0 {
		return 0, fmt.Errorf("USD price not found")
	}
	return usd.Price, nil
}

// CoinMarketCapRequestURL returns the request URL without exposing the API key in query params.
func CoinMarketCapRequestURL() string {
	return coinMarketCapURL
}

// CoinMarketCapUsesQueryKey reports whether the URL contains an API key query parameter.
func CoinMarketCapUsesQueryKey(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.Query().Get("CMC_PRO_API_KEY") != ""
}

func parseFloat(s string) (float64, error) {
	var value float64
	_, err := fmt.Sscanf(s, "%f", &value)
	if err != nil {
		return 0, err
	}
	return value, nil
}
