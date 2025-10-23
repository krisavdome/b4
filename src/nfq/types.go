package nfq

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/daniellavrushin/b4/sock"
	"github.com/florianl/go-nfqueue"
)

type Pool struct {
	workers []*Worker
}

type flowState struct {
	buf          []byte
	last         time.Time
	sniFound     bool
	sniProcessed bool
	sni          string
	packetCount  int
}

type Worker struct {
	cfg              atomic.Value
	qnum             uint16
	ctx              context.Context
	cancel           context.CancelFunc
	q                *nfqueue.Nfqueue
	wg               sync.WaitGroup
	mu               sync.Mutex
	flows            map[string]*flowState
	ttl              time.Duration
	limit            int
	matcher          atomic.Value
	sock             *sock.Sender
	packetsProcessed uint64
}
