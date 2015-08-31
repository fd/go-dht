package kademlia

import (
	"bytes"
	"time"
)

type peerInfo struct {
	key  Key
	peer Peer

	lastSeen  time.Time
	firstSeen time.Time
	latency   time.Duration
}

type fastestPeers []*peerInfo

func (s fastestPeers) Len() int           { return len(s) }
func (s fastestPeers) Less(i, j int) bool { return s[i].latency < s[j].latency }
func (s fastestPeers) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type oldestPeers []*peerInfo

func (s oldestPeers) Len() int           { return len(s) }
func (s oldestPeers) Less(i, j int) bool { return s[i].firstSeen.Before(s[j].firstSeen) }
func (s oldestPeers) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type lexographicPeers []*peerInfo

func (s lexographicPeers) Len() int           { return len(s) }
func (s lexographicPeers) Less(i, j int) bool { return bytes.Compare(s[i].key[:], s[j].key[:]) < 0 }
func (s lexographicPeers) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
