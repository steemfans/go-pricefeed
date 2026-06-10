package publisher

import (
	"context"
	"time"

	"github.com/steemit/steemgosdk"
	"github.com/steemfans/go-pricefeed/internal/config"
	"github.com/steemfans/go-pricefeed/internal/log"
)

// Publisher signs and broadcasts witness price feeds.
type Publisher struct {
	cfg       *config.Config
	account   string
	activeWIF string
	nodeIndex int
	broadcast Broadcaster
}

// NewPublisher creates a publisher using the first configured RPC node.
func NewPublisher(cfg *config.Config) *Publisher {
	node := cfg.RPCNodes[0]
	client := steemgosdk.GetClient(node)
	return &Publisher{
		cfg:       cfg,
		account:   cfg.FeedSteemAccount,
		activeWIF: cfg.FeedSteemActiveKey,
		nodeIndex: 0,
		broadcast: &SteemBroadcast{Inner: client.GetBroadcast()},
	}
}

// Publish broadcasts feed_publish with retry and RPC failover until success or ctx cancel.
func (p *Publisher) Publish(ctx context.Context, price float64) error {
	rate := BuildExchangeRate(price, p.cfg.PegMulti)
	log.Printf("Broadcasting feed_publish transaction: {\"base\":\"%s\",\"quote\":\"%s\"}", rate.Base, rate.Quote)

	for retries := 0; ; retries++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		_, err := p.broadcast.FeedPublish(p.account, rate, p.activeWIF)
		if err == nil {
			log.Println("Broadcast successful!")
			return nil
		}

		log.Printf("Error broadcasting feed_publish transaction: %v", err)

		if ShouldFailover(retries, p.cfg.FeedPublishFailRetry) {
			p.failover()
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(p.cfg.RetryInterval) * time.Second):
		}
	}
}

func (p *Publisher) failover() {
	if len(p.cfg.RPCNodes) <= 1 {
		return
	}
	p.nodeIndex = NextRPCIndex(p.nodeIndex, len(p.cfg.RPCNodes))
	node := p.cfg.RPCNodes[p.nodeIndex]
	client := steemgosdk.GetClient(node)
	p.broadcast = &SteemBroadcast{Inner: client.GetBroadcast()}
	log.Println("***********************************************")
	log.Printf("Failing over to: %s", node)
	log.Println("***********************************************")
}

// NodeIndex exposes current RPC index for tests.
func (p *Publisher) NodeIndex() int {
	return p.nodeIndex
}

// SetBroadcaster replaces the broadcaster implementation (tests).
func (p *Publisher) SetBroadcaster(b Broadcaster) {
	p.broadcast = b
}
