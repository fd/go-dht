package kademlia

import (
	"sync"
	"time"

	"golang.org/x/net/context"
)

const (
	numBuckets = (keyLen * 8) + 1
)

type DHT struct {
	key Key

	mtx     sync.RWMutex
	buckets [numBuckets]*bucket
}

func Start(ctx context.Context, self Key) *DHT {
	dht := &DHT{}
	dht.key = self

	go dht.runPingScheduler(ctx)
	go dht.runFiller(ctx)

	return dht
}

func (dht *DHT) getBucket(idx int) *bucket {
	if idx < 0 || numBuckets <= idx {
		return nil
	}

	dht.mtx.RLock()
	bkt := dht.buckets[idx]
	dht.mtx.RUnlock()

	if bkt == nil {
		dht.mtx.Lock()
		bkt = dht.buckets[idx]
		if bkt == nil {
			bkt = &bucket{}
			dht.buckets[idx] = bkt
		}
		dht.mtx.Unlock()
	}

	return bkt
}

func (dht *DHT) Add(peer Peer) error {
	latency, err := ping(context.Background(), peer)
	if err != nil {
		return err
	}

	dht.touch(peer, latency)
	return nil
}

func (dht *DHT) touch(peer Peer, latency time.Duration) {
	bucket := dht.getBucket(bucketForKey(dht.key, peer.KademliaKey()))
	if bucket != nil {
		bucket.Touch(peer, latency)
	}
}

func (dht *DHT) Remove(peer Peer) {
	bucket := dht.getBucket(bucketForKey(dht.key, peer.KademliaKey()))
	if bucket != nil {
		bucket.Remove(peer)
	}
}

func (dht *DHT) LookupN(key Key, n int) []Peer {
	var (
		bucketIndex  = bucketForKey(dht.key, key)
		bucketOffset = 0
		res          = make([]Peer, 0, n)
	)

	table := dht.getBucket(bucketIndex).GetLookupTable()
	if table == nil {
		return nil
	}

	// find exact match
	for _, info := range table.peers {
		if info.key == key {
			res = append(res, info.peer)
			break
		}
	}

	res = appendAtMost(res, table.fastest, n-len(res))

	for {
		bucketOffset++

		if bucketIndex+bucketOffset >= numBuckets && bucketIndex-bucketOffset < 0 {
			break
		}

		table1 := dht.getBucket(bucketIndex + bucketOffset).GetLookupTable()
		table2 := dht.getBucket(bucketIndex - bucketOffset).GetLookupTable()

		if table1 != nil {
			res = appendAtMost(res, table1.fastest, n-len(res))
			if len(res) >= n {
				break
			}
		}

		if table2 != nil {
			res = appendAtMost(res, table2.fastest, n-len(res))
			if len(res) >= n {
				break
			}
		}
	}

	return res
}

func appendAtMost(dst []Peer, src []*peerInfo, n int) []Peer {
	if n <= 0 {
		return dst
	}
	if len(src) > n {
		src = src[:n]
	}

	for _, info := range src {
		dst = append(dst, info.peer)
	}

	return dst
}
