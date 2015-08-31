package kademlia

import "testing"

func TestBucketForKey(t *testing.T) {
	const m = 255

	var tab = []struct {
		key    Key
		bucket int
	}{
		{Key{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 0},
		{Key{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, 1},
		{Key{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}, 2},
		{Key{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3}, 2},
		{Key{m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m}, 256},
	}

	var self Key

	for _, test := range tab {
		actual := bucketForKey(self, test.key)
		if actual != test.bucket {
			t.Errorf("expected bucket %d but was %d (key = %x)", test.bucket, actual, test.key)
		}
	}
}

func TestRandomKeyInBucket(t *testing.T) {
	for i := 0; i < 128; i++ {
		var self = randKey()
		var mustBeInBucket = 4

		for i := 0; i < 128; i++ {
			key := randKeyInBucket(self, mustBeInBucket)
			if bucket := bucketForKey(self, key); bucket != mustBeInBucket {
				t.Errorf("expected %x to be in bucket %d but was in %d", key, mustBeInBucket, bucket)
			}
		}
	}
}
