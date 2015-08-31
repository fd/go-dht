package kademlia

import (
	"sort"
	"sync"
	"time"
)

const (
	maxBucketSize             = 128
	latencyEMAPeriods float64 = 5
	latencyEMAAlpha   float64 = 2 / (1 + latencyEMAPeriods)
)

type bucket struct {
	mtx         sync.RWMutex
	modified    bool
	peers       []*peerInfo
	lookupTable *lookupTable
}

type lookupTable struct {
	fastest []*peerInfo // sorted by latency (fastest first)
	oldest  []*peerInfo // sorted by first seen (oldest first)
	peers   []*peerInfo // sorted by key
}

func (b *bucket) Touch(peer Peer, latency time.Duration) {
	// add to bucket if not full
	// replace with slowest if faster than most

	now := time.Now()

	b.mtx.Lock()
	defer b.mtx.Unlock()

	if b.peers == nil {
		b.peers = make([]*peerInfo, 0, maxBucketSize)
	}

	var (
		key              = peer.KademliaKey()
		info             *peerInfo
		slowestPeerIndex int
		maxLatency       time.Duration
		avgLatency       time.Duration
	)

	for idx, i := range b.peers {
		avgLatency += i.latency

		if i.latency > maxLatency {
			maxLatency = i.latency
			slowestPeerIndex = idx
		}

		if i.key == key {
			info = i
		}
	}
	if n := len(b.peers); n > 0 {
		avgLatency /= time.Duration(n)
	}

	// Update existing record
	if info != nil {
		if latency > 0 {
			info.latency = time.Duration((float64(latency) * latencyEMAAlpha) + ((1 - latencyEMAAlpha) * float64(latency)))
		}
		info.lastSeen = now
		b.modified = true
		return
	}

	// Add entry; bucket is not full
	if len(b.peers) < maxBucketSize {
		logf("BUCKET add (more) %x", key)
		if latency <= 0 {
			latency = 1 * time.Minute
		}
		info = &peerInfo{
			key:       key,
			peer:      peer,
			lastSeen:  now,
			firstSeen: now,
			latency:   latency,
		}
		b.peers = append(b.peers, info)
		sort.Sort(lexographicPeers(b.peers))
		b.modified = true
		return
	}

	// Add entry; faster than most
	if latency < avgLatency {
		logf("BUCKET add (faster) %x", key)
		if latency <= 0 {
			latency = 1 * time.Minute
		}
		info = &peerInfo{
			key:       key,
			peer:      peer,
			lastSeen:  now,
			firstSeen: now,
			latency:   latency,
		}
		// replace slowest peer with new peer
		b.peers[slowestPeerIndex] = info
		sort.Sort(lexographicPeers(b.peers))
		b.modified = true
		return
	}
}

func (b *bucket) Remove(peer Peer) {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	var key = peer.KademliaKey()

	for idx, info := range b.peers {
		if info.key == key {
			copy(b.peers[idx:], b.peers[idx+1:])
			b.peers = b.peers[:len(b.peers)-1]
			b.modified = true
			return
		}
	}
}

func (b *bucket) GetSize() int {
	if b == nil {
		return 0
	}

	b.mtx.RLock()
	s := len(b.peers)
	b.mtx.RUnlock()

	return s
}

func (b *bucket) GetLookupTable() *lookupTable {
	if b == nil {
		return nil
	}

	b.mtx.RLock()
	table, modified := b.lookupTable, b.modified
	b.mtx.RUnlock()

	if !modified {
		return table
	}

	b.mtx.Lock()
	if b.modified {
		table = &lookupTable{
			fastest: make([]*peerInfo, len(b.peers)),
			oldest:  make([]*peerInfo, len(b.peers)),
			peers:   make([]*peerInfo, len(b.peers)),
		}

		copy(table.fastest, b.peers)
		copy(table.oldest, b.peers)
		copy(table.peers, b.peers)

		sort.Sort(fastestPeers(table.fastest))
		sort.Sort(oldestPeers(table.oldest))

		b.lookupTable = table
		b.modified = false
	} else {
		table = b.lookupTable
	}
	b.mtx.Unlock()

	if table == nil {
		table = &lookupTable{}
	}

	return table
}
