package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadApplyDefaultsAndNormalizeRPC(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{
  "feed_steem_account": "witness",
  "feed_steem_active_key": "5Jxxx",
  "exchanges": ["coingecko"],
  "rpc_node": "https://node.example"
}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Interval != 60 {
		t.Fatalf("Interval = %d, want 60", cfg.Interval)
	}
	if len(cfg.RPCNodes) != 1 || cfg.RPCNodes[0] != "https://node.example" {
		t.Fatalf("RPCNodes = %#v", cfg.RPCNodes)
	}
}

func TestLoadEnvFallback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{
  "exchanges": ["coingecko"]
}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("feed_steem_account", "envwitness")
	t.Setenv("feed_steem_active_key", "5Jenv")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.FeedSteemAccount != "envwitness" {
		t.Fatalf("FeedSteemAccount = %q", cfg.FeedSteemAccount)
	}
}

func TestValidateRequiresCoinMarketCapKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{
  "feed_steem_account": "witness",
  "feed_steem_active_key": "5Jxxx",
  "exchanges": ["coinmarketcap"]
}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("expected error for missing coinmarketcap_api_key")
	}
}
