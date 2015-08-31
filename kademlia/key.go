package kademlia

import (
	"crypto/rand"
	"io"
)

const (
	keyLen = 32
)

type Key [keyLen]byte

func less(a, b Key) bool {
	for i, x := range a {
		y := b[i]
		if x < y {
			return true
		}
		if x > y {
			return false
		}
	}
	return false
}

func keyDistance(a, b Key) Key {
	var c Key
	for i, x := range a {
		c[i] = x ^ b[i]
	}
	return c
}

func bucketForKey(self, key Key) int {
	r := 256

	for i, x := range self {
		y := key[i]
		d := x ^ y

		if d == 0 {
			r -= 8
			continue
		}

		switch {
		case 0x80&d > 0:
			return r
		case 0x40&d > 0:
			return r - 1
		case 0x20&d > 0:
			return r - 2
		case 0x10&d > 0:
			return r - 3
		case 0x08&d > 0:
			return r - 4
		case 0x04&d > 0:
			return r - 5
		case 0x02&d > 0:
			return r - 6
		case 0x01&d > 0:
			return r - 7
		}
	}

	return r
}

func randKey() Key {
	var (
		rnd Key
	)

	_, err := io.ReadFull(rand.Reader, rnd[:])
	if err != nil {
		panic("crypto/rand failed")
	}

	return rnd
}

func randKeyInBucket(base Key, bucket int) Key {
	if bucket == 0 {
		return base
	}

	var (
		rnd = randKey()
		key Key
	)

	bucket--
	split := keyLen - (bucket / 8) - 1
	copy(key[:], base[:split])
	copy(key[split+1:], rnd[split+1:])

	shift := uint(bucket % 8)
	maskBase := byte(0xFF) << (shift + 1)
	maskRnd := ^(byte(0xFF) << (shift))

	key[split] = (maskBase & base[split]) |
		(maskRnd & rnd[split]) |
		((^base[split]) & (byte(1) << shift))

	return key
}
