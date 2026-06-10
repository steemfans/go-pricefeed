package price

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type CryptoCompare struct{}

func (CryptoCompare) Name() string { return "Cryptocompare" }

func (CryptoCompare) Fetch(ctx context.Context, client *http.Client) (float64, error) {
	body, err := getJSON(ctx, client, "https://min-api.cryptocompare.com/data/price?fsym=STEEM&tsyms=USDT", nil)
	if err != nil {
		return 0, err
	}
	var payload struct {
		USDT float64 `json:"USDT"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, err
	}
	if payload.USDT == 0 {
		return 0, fmt.Errorf("missing USDT")
	}
	return payload.USDT, nil
}
