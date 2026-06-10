package price

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type SlowAPI struct{}

func (SlowAPI) Name() string { return "SlowAPI" }

func (SlowAPI) Fetch(ctx context.Context, client *http.Client) (float64, error) {
	body, err := getJSON(ctx, client, "https://slowapi.com/api/yf/", nil)
	if err != nil {
		return 0, err
	}
	var payload struct {
		Data map[string]struct {
			RegularMarketPrice float64 `json:"regularMarketPrice"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, err
	}
	entry, ok := payload.Data["STEEM-USD"]
	if !ok {
		return 0, fmt.Errorf("STEEM-USD not found")
	}
	return entry.RegularMarketPrice, nil
}
