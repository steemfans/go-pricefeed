package price

import (
	"context"
	"net/http"
	"testing"

	"github.com/steemfans/go-pricefeed/internal/config"
)

type stubFetcher struct {
	name  string
	price float64
}

func (s stubFetcher) Name() string { return s.name }

func (s stubFetcher) Fetch(ctx context.Context, client *http.Client) (float64, error) {
	return s.price, nil
}

func TestAggregatorCollectAveragesAvailablePrices(t *testing.T) {
	cfg := &config.Config{
		FeedPublishInterval: 1,
		PriceFeedMaxRetry:   0,
		RetryInterval:       0,
	}
	agg := NewAggregatorWithFetchers(cfg, []Fetcher{
		stubFetcher{name: "a", price: 1},
		stubFetcher{name: "b", price: 3},
	})

	prices, err := agg.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if len(prices) != 2 {
		t.Fatalf("len(prices) = %d, want 2", len(prices))
	}
	if got := Average(prices); got != 2 {
		t.Fatalf("Average() = %v, want 2", got)
	}
}

func TestAggregatorCollectNoPrices(t *testing.T) {
	cfg := &config.Config{
		FeedPublishInterval: 1,
		PriceFeedMaxRetry:   0,
	}
	agg := NewAggregatorWithFetchers(cfg, nil)
	if _, err := agg.Collect(context.Background()); err == nil {
		t.Fatal("expected error when no fetchers configured")
	}
}
