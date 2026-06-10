package price

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Cloudflare struct{}

func (Cloudflare) Name() string { return "Cloudflare" }

func (Cloudflare) Fetch(ctx context.Context, client *http.Client) (float64, error) {
	body, err := getJSON(ctx, client, "https://ticker.justyy.com/query/?s=STEEM+USDT", nil)
	if err != nil {
		return 0, err
	}
	var payload struct {
		Result []string `json:"result"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, err
	}
	if len(payload.Result) == 0 {
		return 0, fmt.Errorf("empty result")
	}
	parts := strings.Fields(payload.Result[0])
	if len(parts) < 4 {
		return 0, fmt.Errorf("unexpected result format: %q", payload.Result[0])
	}
	return parseFloat(parts[3])
}
