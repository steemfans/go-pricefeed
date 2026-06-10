package price

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAverageUsesSuccessfulPricesOnly(t *testing.T) {
	got := Average([]float64{1, 2, 3})
	if got != 2.0 {
		t.Fatalf("Average() = %v, want 2", got)
	}
}

func TestParseCoinMarketCap(t *testing.T) {
	body := []byte(`{"data":{"STEEM":[{"quote":{"USD":{"price":0.123456}}}]}}`)
	price, err := parseCoinMarketCap(body)
	if err != nil {
		t.Fatalf("parseCoinMarketCap() error = %v", err)
	}
	if price != 0.123456 {
		t.Fatalf("price = %v", price)
	}
}

func TestCoinMarketCapUsesHeaderAuth(t *testing.T) {
	const apiKey = "test-api-key"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-CMC_PRO_API_KEY"); got != apiKey {
			t.Fatalf("header X-CMC_PRO_API_KEY = %q, want %q", got, apiKey)
		}
		if r.URL.Query().Get("CMC_PRO_API_KEY") != "" {
			t.Fatal("expected API key not in query string")
		}
		_, _ = w.Write([]byte(`{"data":{"STEEM":[{"quote":{"USD":{"price":0.5}}}]}}`))
	}))
	defer server.Close()

	body, err := getJSON(context.Background(), server.Client(), server.URL, map[string]string{
		"X-CMC_PRO_API_KEY": apiKey,
		"Accept":            "application/json",
	})
	if err != nil {
		t.Fatalf("getJSON() error = %v", err)
	}
	price, err := parseCoinMarketCap(body)
	if err != nil {
		t.Fatalf("parseCoinMarketCap() error = %v", err)
	}
	if price != 0.5 {
		t.Fatalf("price = %v", price)
	}
}

func TestCoinMarketCapRequestURLHasNoQueryKey(t *testing.T) {
	if CoinMarketCapUsesQueryKey(CoinMarketCapRequestURL()) {
		t.Fatal("expected CoinMarketCap URL without query API key")
	}
}

func TestParseCloudflarePayload(t *testing.T) {
	body := []byte(`{"result":["STEEM USDT 0.123 0.456 0.789"]}`)
	var payload struct {
		Result []string `json:"result"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	parts := strings.Fields(payload.Result[0])
	price, err := parseFloat(parts[3])
	if err != nil || price != 0.456 {
		t.Fatalf("price = %v err = %v", price, err)
	}
}

func TestParseCoingeckoPayload(t *testing.T) {
	body := []byte(`{"steem":{"usd":0.321}}`)
	var payload struct {
		Steem struct {
			USD float64 `json:"usd"`
		} `json:"steem"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Steem.USD != 0.321 {
		t.Fatalf("usd = %v", payload.Steem.USD)
	}
}
