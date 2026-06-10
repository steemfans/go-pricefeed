package publisher

import "github.com/steemit/steemgosdk/broadcast"

// Broadcaster publishes signed Steem operations.
type Broadcaster interface {
	FeedPublish(publisher string, rate broadcast.ExchangeRate, privKeyWif string) ([]byte, error)
}

// SteemBroadcast adapts steemgosdk broadcast.Broadcast to Broadcaster.
type SteemBroadcast struct {
	Inner *broadcast.Broadcast
}

func (s *SteemBroadcast) FeedPublish(publisher string, rate broadcast.ExchangeRate, privKeyWif string) ([]byte, error) {
	return s.Inner.FeedPublish(publisher, rate, privKeyWif)
}
