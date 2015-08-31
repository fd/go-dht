package kademlia

import (
	"sync"
	"time"

	"golang.org/x/net/context"
)

func (dht *DHT) Seek(ctx context.Context, dst Key, n int) ([]Peer, error) {
	if n <= 0 {
		n = 25
	}

	s := &seek{
		dht:       dht,
		dst:       dst,
		n:         n,
		smallSema: newSema(5),
		largeSema: newSema(5),
		queue:     make(chan Peer),
		results:   make([]*seekEntry, n),
	}

	return s.Seek(ctx)
}

type seek struct {
	dht *DHT
	dst Key
	n   int

	smallSema semaphore
	largeSema semaphore
	queue     chan Peer

	wg      sync.WaitGroup
	mtx     sync.Mutex
	results []*seekEntry
}

type seekEntry struct {
	distance Key
	peer     Peer
}

func (s *seek) AddEntry(newEntry *seekEntry) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	for idx, entry := range s.results {

		if entry != nil {
			if less(entry.distance, newEntry.distance) {
				continue
			}

			if entry.distance == newEntry.distance {
				// ignore; already in result
				// logf("ignored %x", newEntry.peer.KademliaKey())
				return
			}
		}

		// insert new entry
		copy(s.results[idx+1:], s.results[idx:])
		s.results[idx] = newEntry
		// logf("added %x", newEntry.peer.KademliaKey())

		s.wg.Add(1)
		s.queue <- newEntry.peer
		return
	}
}

func (s *seek) HandleError(err error) {
	logf("DHT/error: %s", err)
}

func (s *seek) Seek(ctx context.Context) ([]Peer, error) {
	// logf("SEEK begin %x", s.dst)
	// defer logf("SEEK end %x", s.dst)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	s.wg.Add(1)
	go s.localSeek()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case peer := <-s.queue:
				go s.remoteSeek(ctx, peer)
			}
		}
	}()

	s.wg.Wait()

	peers := make([]Peer, len(s.results))
	for i, entry := range s.results {
		if entry == nil {
			peers = peers[:i]
			break
		}
		peers[i] = entry.peer
	}

	return peers, ctx.Err()
}

func (s *seek) localSeek() {
	// logf("SEEK local %x (%x)", s.dht.key, s.dst)
	defer s.wg.Done()

	peers := s.dht.LookupN(s.dst, s.n)
	// logf("SEEK local %x (%d)", s.dht.key, len(peers))
	for _, peer := range peers {
		// logf("Attempt add %x", peer.KademliaKey())
		s.AddEntry(&seekEntry{
			distance: keyDistance(s.dst, peer.KademliaKey()),
			peer:     peer,
		})
	}
}

func (s *seek) remoteSeek(ctx context.Context, peer Peer) {
	// logf("SEEK remote %x(%x)", peer.KademliaKey(), s.dst)

	if peer.KademliaKey() == s.dht.key {
		s.localSeek()
		return
	}

	defer s.wg.Done()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	// acquire spot in fast-lane
	if !s.smallSema.acquire(ctx) {
		return
	}
	go func() {
		defer s.smallSema.release()
		smallCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()
		<-smallCtx.Done()
	}()

	// acquire spot in slow-lane
	if !s.largeSema.acquire(ctx) {
		return
	}
	defer s.largeSema.release()

	peers, err := peer.Lookup(ctx, s.dst, s.n)
	if err != nil {
		s.HandleError(err)
		return
	}

	for _, peer := range peers {
		s.AddEntry(&seekEntry{
			distance: keyDistance(s.dst, peer.KademliaKey()),
			peer:     peer,
		})
	}
}
