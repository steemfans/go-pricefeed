package price

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/steemfans/go-pricefeed/internal/config"
	"github.com/steemfans/go-pricefeed/internal/log"
)

// Fetcher loads STEEM price in USD from an external source.
type Fetcher interface {
	Name() string
	Fetch(ctx context.Context, client *http.Client) (float64, error)
}

// Aggregator collects prices from configured fetchers within a fixed wait window.
type Aggregator struct {
	cfg      *config.Config
	fetchers []Fetcher
	client   *http.Client
}

// NewAggregator builds fetchers from config.
func NewAggregator(cfg *config.Config) *Aggregator {
	return NewAggregatorWithFetchers(cfg, NewRegistry(cfg))
}

// NewAggregatorWithFetchers creates an aggregator with explicit fetchers (tests).
func NewAggregatorWithFetchers(cfg *config.Config, fetchers []Fetcher) *Aggregator {
	return &Aggregator{
		cfg:      cfg,
		fetchers: fetchers,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewRegistry returns fetchers for configured exchange names.
func NewRegistry(cfg *config.Config) []Fetcher {
	var fetchers []Fetcher
	for _, name := range cfg.Exchanges {
		switch name {
		case "cloudflare":
			fetchers = append(fetchers, Cloudflare{})
		case "slowapi":
			fetchers = append(fetchers, SlowAPI{})
		case "coingecko":
			fetchers = append(fetchers, CoinGecko{})
		case "cryptocompare":
			fetchers = append(fetchers, CryptoCompare{})
		case "coinmarketcap":
			fetchers = append(fetchers, CoinMarketCap{APIKey: cfg.CoinMarketCapAPIKey})
		default:
			log.Printf("unknown exchange %q, skipping", name)
		}
	}
	return fetchers
}

// Collect starts parallel fetches, waits feed_publish_interval, and returns valid prices.
func (a *Aggregator) Collect(ctx context.Context) ([]float64, error) {
	if len(a.fetchers) == 0 {
		return nil, fmt.Errorf("no prices found")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make(chan float64, len(a.fetchers))
	for _, fetcher := range a.fetchers {
		go func(f Fetcher) {
			price, err := fetchWithRetry(ctx, a.cfg, a.client, f)
			if err != nil {
				log.Printf("Error loading STEEM price from %s: %v", f.Name(), err)
				return
			}
			log.Printf("Loaded STEEM Price from %s: %v", f.Name(), price)
			select {
			case results <- price:
			case <-ctx.Done():
			}
		}(fetcher)
	}

	var prices []float64
	timer := time.NewTimer(time.Duration(a.cfg.FeedPublishInterval) * time.Second)
	defer timer.Stop()

collectLoop:
	for {
		select {
		case price := <-results:
			prices = append(prices, price)
		case <-timer.C:
			break collectLoop
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if len(prices) == 0 {
		return nil, fmt.Errorf("no prices found")
	}
	return prices, nil
}

// Average returns the mean of valid prices.
func Average(prices []float64) float64 {
	if len(prices) == 0 {
		return 0
	}
	sum := 0.0
	for _, p := range prices {
		sum += p
	}
	return sum / float64(len(prices))
}

func fetchWithRetry(ctx context.Context, cfg *config.Config, client *http.Client, fetcher Fetcher) (float64, error) {
	var lastErr error
	for attempt := 0; attempt <= cfg.PriceFeedMaxRetry; attempt++ {
		price, err := fetcher.Fetch(ctx, client)
		if err == nil {
			return price, nil
		}
		lastErr = err
		if attempt == cfg.PriceFeedMaxRetry {
			break
		}
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(time.Duration(cfg.RetryInterval) * time.Second):
		}
	}
	return 0, lastErr
}

func getJSON(ctx context.Context, client *http.Client, url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return body, nil
}
