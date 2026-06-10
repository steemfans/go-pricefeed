package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Config holds pricefeed runtime settings compatible with the Node.js pricefeed config.json.
type Config struct {
	RPCNodes             []string `json:"rpc_nodes"`
	RPCNode              string   `json:"rpc_node"`
	FeedSteemAccount     string   `json:"feed_steem_account"`
	FeedSteemActiveKey   string   `json:"feed_steem_active_key"`
	CoinMarketCapAPIKey  string   `json:"coinmarketcap_api_key"`
	Exchanges            []string `json:"exchanges"`
	Interval             int      `json:"interval"`
	FeedPublishInterval  int      `json:"feed_publish_interval"`
	FeedPublishFailRetry int      `json:"feed_publish_fail_retry"`
	PriceFeedMaxRetry    int      `json:"price_feed_max_retry"`
	RetryInterval        int      `json:"retry_interval"`
	PegMulti             float64  `json:"peg_multi"`
}

// Load reads config from path and applies env fallbacks, defaults, and normalization.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	applyEnvFallbacks(&cfg)
	cfg.ApplyDefaults()
	cfg.NormalizeRPC()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func applyEnvFallbacks(cfg *Config) {
	if cfg.FeedSteemAccount == "" {
		cfg.FeedSteemAccount = os.Getenv("feed_steem_account")
	}
	if cfg.FeedSteemActiveKey == "" {
		cfg.FeedSteemActiveKey = os.Getenv("feed_steem_active_key")
	}
	if cfg.CoinMarketCapAPIKey == "" {
		cfg.CoinMarketCapAPIKey = os.Getenv("coinmarketcap_api_key")
	}
}

// ApplyDefaults fills zero-value numeric fields with production-safe defaults.
func (c *Config) ApplyDefaults() {
	if c.Interval == 0 {
		c.Interval = 60
	}
	if c.FeedPublishInterval == 0 {
		c.FeedPublishInterval = 30
	}
	if c.FeedPublishFailRetry == 0 {
		c.FeedPublishFailRetry = 5
	}
	if c.PriceFeedMaxRetry == 0 {
		c.PriceFeedMaxRetry = 5
	}
	if c.RetryInterval == 0 {
		c.RetryInterval = 10
	}
	if c.PegMulti == 0 {
		c.PegMulti = 1
	}
}

// NormalizeRPC ensures rpc_nodes is populated from legacy rpc_node or default node list.
func (c *Config) NormalizeRPC() {
	if len(c.RPCNodes) == 0 && c.RPCNode != "" {
		c.RPCNodes = []string{c.RPCNode}
	}
	if len(c.RPCNodes) == 0 {
		c.RPCNodes = []string{"https://api.steemit.com"}
	}
}

// Validate checks required settings before startup.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.FeedSteemAccount) == "" {
		return errors.New("feed_steem_account not set in config.json or environment")
	}
	if strings.TrimSpace(c.FeedSteemActiveKey) == "" {
		return errors.New("feed_steem_active_key not set in config.json or environment")
	}
	if len(c.Exchanges) == 0 {
		return errors.New("no exchanges are specified")
	}
	if c.Interval <= 0 {
		return errors.New("interval must be greater than 0")
	}
	if c.FeedPublishInterval <= 0 {
		return errors.New("feed_publish_interval must be greater than 0")
	}
	if c.FeedPublishFailRetry <= 0 {
		return errors.New("feed_publish_fail_retry must be greater than 0")
	}
	if c.PegMulti <= 0 {
		return errors.New("peg_multi must be greater than 0")
	}
	if containsExchange(c.Exchanges, "coinmarketcap") && strings.TrimSpace(c.CoinMarketCapAPIKey) == "" {
		return errors.New("coinmarketcap_api_key not set in config.json or environment")
	}
	return nil
}

func containsExchange(exchanges []string, name string) bool {
	for _, ex := range exchanges {
		if ex == name {
			return true
		}
	}
	return false
}
