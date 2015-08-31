package kademlia

import (
	"log"
	"os"
	"testing"
	"time"

	"golang.org/x/net/context"
)

func setup() {
	debug = true
	log.SetOutput(os.Stderr)
}

func TestIntegration(t *testing.T) {
	setup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		peers = make(map[Key]*testPeer)
		seed  Key
	)

	for i := 0; i < 100; i++ {
		key := randKey()

		if i == 0 {
			seed = key
		}

		peers[key] = &testPeer{
			Key:   key,
			DHT:   Start(ctx, key),
			peers: peers,
		}
	}

	for key, peer := range peers {
		if key == seed {
			continue
		}
		peers[seed].DHT.Add(peer)
		peer.DHT.Add(peers[seed])
	}

	time.Sleep(5 * time.Minute)
}

type testPeer struct {
	Key   Key
	DHT   *DHT
	peers map[Key]*testPeer
}

func (p *testPeer) KademliaKey() Key {
	return p.Key
}

func (p *testPeer) Ping(ctx context.Context) error {
	return nil
}

func (p *testPeer) Lookup(ctx context.Context, key Key, n int) ([]Peer, error) {
	return p.DHT.LookupN(key, n), nil
}
