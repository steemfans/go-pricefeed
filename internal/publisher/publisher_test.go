package publisher

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/steemit/steemgosdk/broadcast"
	"github.com/steemfans/go-pricefeed/internal/config"
)

type mockBroadcaster struct {
	calls     atomic.Int32
	failUntil int32
	lastRate  broadcast.ExchangeRate
}

func (m *mockBroadcaster) FeedPublish(publisher string, rate broadcast.ExchangeRate, privKeyWif string) ([]byte, error) {
	m.calls.Add(1)
	m.lastRate = rate
	if m.calls.Load() <= m.failUntil {
		return nil, errors.New("broadcast failed")
	}
	return []byte("ok"), nil
}

func TestBuildExchangeRate(t *testing.T) {
	rate := BuildExchangeRate(0.5, 1)
	if rate.Base != "0.500 SBD" || rate.Quote != "1.000 STEEM" {
		t.Fatalf("rate = %#v", rate)
	}
}

func TestShouldFailover(t *testing.T) {
	cases := []struct {
		retries int
		want    bool
	}{
		{0, true},
		{1, false},
		{5, true},
	}
	for _, tc := range cases {
		if got := ShouldFailover(tc.retries, 5); got != tc.want {
			t.Fatalf("ShouldFailover(%d) = %v, want %v", tc.retries, got, tc.want)
		}
	}
}

func TestNextRPCIndex(t *testing.T) {
	if got := NextRPCIndex(0, 3); got != 1 {
		t.Fatalf("NextRPCIndex() = %d, want 1", got)
	}
	if got := NextRPCIndex(2, 3); got != 0 {
		t.Fatalf("NextRPCIndex() = %d, want 0", got)
	}
}

func TestPublishRetriesUntilSuccess(t *testing.T) {
	cfg := testConfig()
	cfg.FeedPublishFailRetry = 100
	pub := NewPublisher(cfg)
	mock := &mockBroadcaster{failUntil: 2}
	pub.SetBroadcaster(mock)

	if err := pub.Publish(context.Background(), 0.123); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if mock.calls.Load() != 3 {
		t.Fatalf("calls = %d, want 3", mock.calls.Load())
	}
	if mock.lastRate.Base != "0.123 SBD" {
		t.Fatalf("lastRate = %#v", mock.lastRate)
	}
}

func TestPublishFailoverRotatesRPCNode(t *testing.T) {
	cfg := testConfig()
	cfg.RPCNodes = []string{"https://node-a", "https://node-b"}
	pub := NewPublisher(cfg)
	if pub.NodeIndex() != 0 {
		t.Fatalf("NodeIndex = %d, want 0", pub.NodeIndex())
	}
	pub.failover()
	if pub.NodeIndex() != 1 {
		t.Fatalf("NodeIndex = %d, want 1", pub.NodeIndex())
	}
}

func testConfig() *config.Config {
	return &config.Config{
		RPCNodes:             []string{"https://api.steemit.com"},
		FeedSteemAccount:     "witness",
		FeedSteemActiveKey:   "5Jxxx",
		FeedPublishFailRetry: 5,
		RetryInterval:        0,
		PegMulti:             1,
	}
}
