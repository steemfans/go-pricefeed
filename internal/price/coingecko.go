package price

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type CoinGecko struct{}

func (CoinGecko) Name() string { return "Coingecko" }

func (CoinGecko) Fetch(ctx context.Context, client *http.Client) (float64, error) {
	body, err := getJSON(ctx, client, "https://api.coingecko.com/api/v3/simple/price?ids=steem&vs_currencies=usd", nil)
	if err != nil {
		return 0, err
	}
	var payload struct {
		Steem struct {
			USD float64 `json:"usd"`
		} `json:"steem"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, err
	}
	if payload.Steem.USD == 0 {
		return 0, fmt.Errorf("missing steem.usd")
	}
	return payload.Steem.USD, nil
}
