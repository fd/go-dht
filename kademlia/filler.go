package kademlia

import (
	"time"

	"golang.org/x/net/context"
)

func (dht *DHT) runFiller(ctx context.Context) {
	timer := time.NewTicker(15 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			dht.fillAll(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (dht *DHT) fillAll(ctx context.Context) {
	for i := 0; i < numBuckets; i++ {
		// var (
		// 	fill = dht.getBucket(i-1).GetSize() > 0 ||
		// 		dht.getBucket(i).GetSize() > 0 ||
		// 		dht.getBucket(i+1).GetSize() > 0
		// )
		// if !fill {
		// 	continue
		// }

		go dht.fillBucket(ctx, i)
	}
}

func (dht *DHT) fillBucket(ctx context.Context, bucket int) {
	logf("FILL begin bucket=%d", bucket)
	defer logf("FILL end bucket=%d", bucket)

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	dst := randKeyInBucket(dht.key, bucket)
	peers, err := dht.Seek(ctx, dst, 25)
	if err == context.Canceled || err == context.DeadlineExceeded {
		err = nil
	}
	if err != nil {
		// handle error
		// do not return
	}

	for _, peer := range peers {
		dht.Add(peer)
	}
}
