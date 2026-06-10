# Steem Witness Price Feed (Go)

Go rewrite of the [Node.js pricefeed](https://github.com/steemfans/pricefeed) tool for Steem witnesses. It fetches STEEM/USD prices from multiple APIs, averages successful sources, and publishes `feed_publish` transactions to the blockchain.

Docker image: `steemfans/go-pricefeed:latest`

## Features

- Compatible `config.json` schema with the Node.js version
- Five price sources: cloudflare, slowapi, coingecko, cryptocompare, coinmarketcap
- CoinMarketCap authentication via `X-CMC_PRO_API_KEY` header (not query string)
- RPC node failover on broadcast failures
- Fixed average calculation (uses count of successful prices only)
- Single-flight scheduler (skips overlapping cycles)
- Graceful shutdown on SIGINT/SIGTERM

## Private Key Requirements

Use your witness account **active private key** (`feed_steem_active_key`). This is **not** the witness **block signing key** used for block production. The chain requires active or owner authority for `feed_publish`.

## Installation

### Docker (recommended)

```bash
cp config.example.json config.json
# edit config.json with witness account and active key

docker run -itd \
  --name go-pricefeed \
  -v $(pwd)/config.json:/app/config.json \
  steemfans/go-pricefeed:latest

docker logs -f go-pricefeed
```

### Build from source

```bash
go build -o bin/pricefeed ./cmd/pricefeed
./bin/pricefeed -config config.json
```

## Configuration

See [config.example.json](config.example.json). Secrets can be set via environment variables when left empty in JSON:

| Environment variable | Config field |
|---------------------|--------------|
| `feed_steem_account` | `feed_steem_account` |
| `feed_steem_active_key` | `feed_steem_active_key` |
| `coinmarketcap_api_key` | `coinmarketcap_api_key` |

### Differences from Node.js pricefeed

| Item | Node.js | Go |
|------|---------|-----|
| Average price | Divides by total exchange slots | Divides by successful fetches only |
| CoinMarketCap auth | Query parameter | `X-CMC_PRO_API_KEY` header |
| Overlapping cycles | Allowed | Skipped (single-flight) |
| Shutdown | No signal handling | Graceful shutdown |

## Migration from Node.js pricefeed

1. Reuse your existing `config.json`
2. Replace Docker image `steemfans/pricefeed` with `steemfans/go-pricefeed`
3. No change required for RPC nodes, exchanges, or active key configuration

## License

MIT
