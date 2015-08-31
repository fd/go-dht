package kademlia

import (
	"sort"
	"time"

	"golang.org/x/net/context"
)

func (dht *DHT) runPingScheduler(ctx context.Context) {
	timer := time.NewTimer(1 * time.Minute)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			go dht.runPinger(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (dht *DHT) runPinger(ctx context.Context) {
	// logf("run pinger")
	var peers = make([]*peerInfo, 0, 1024)

	for i := 0; i < numBuckets; i++ {
		table := dht.getBucket(i).GetLookupTable()
		if table == nil || len(table.peers) == 0 {
			continue
		}

		peers = append(peers, table.peers...)
	}

	sort.Sort(sort.Reverse(fastestPeers(peers)))

	var (
		ticker = tokenStream(len(peers), 1*time.Minute)
	)

	for len(peers) > 0 {
		select {
		case <-ctx.Done():
			return
		case <-ticker:
		}

		info := peers[0]
		peers = peers[1:]

		go dht.ping(ctx, info.peer)
	}
}

func (dht *DHT) ping(ctx context.Context, peer Peer) {
	// logf("ping: %x", peer.KademliaKey())
	latency, err := ping(ctx, peer)
	if err != nil {
		dht.Remove(peer)
	} else {
		dht.touch(peer, latency)
	}
}
