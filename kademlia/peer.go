package kademlia

import (
	"time"

	"golang.org/x/net/context"
)

type Peer interface {
	KademliaKey() Key
	Ping(ctx context.Context) error
	Lookup(ctx context.Context, key Key, n int) ([]Peer, error)
}

func ping(ctx context.Context, peer Peer) (latency time.Duration, err error) {
	start := time.Now()

	ctx, cancel := context.WithTimeout(ctx, 55*time.Second)
	defer cancel()

	err = peer.Ping(ctx)

	return time.Since(start), err
}
