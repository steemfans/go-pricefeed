package publisher

import (
	"fmt"

	"github.com/steemit/steemgosdk/broadcast"
)

// BuildExchangeRate formats witness feed_publish exchange_rate fields.
func BuildExchangeRate(price, pegMulti float64) broadcast.ExchangeRate {
	return broadcast.ExchangeRate{
		Base:  fmt.Sprintf("%.3f SBD", price),
		Quote: fmt.Sprintf("%.3f STEEM", 1.0/pegMulti),
	}
}

// NextRPCIndex returns the next RPC node index after failover.
func NextRPCIndex(current, total int) int {
	if total <= 1 {
		return current
	}
	return (current + 1) % total
}

// ShouldFailover reports whether RPC failover should run for the given retry count.
func ShouldFailover(retries, failRetry int) bool {
	return retries%failRetry == 0
}
