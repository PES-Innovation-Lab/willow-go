package wgps

import (
	"context"
	"strconv"
)

type GuaranteedQueue struct {
	Guarantees    uint64
	Queue         []byte
	OutGoingBytes []byte
}

/** Add some bytes to the queue. */
func (q *GuaranteedQueue) Push(bytes []byte) {
	q.Queue = append(q.Queue, bytes...)
	q.UseGuarantees()
}

/** Add guarantees received from the server. */
func (q *GuaranteedQueue) AddGuarantees(bytes uint64) {
	q.Guarantees += bytes
	q.UseGuarantees()
}

/** Received a plea from the server to shrink the buffer to a certain size.
 *
 * This implementation always absolves them.
 */
func (q *GuaranteedQueue) Plead(targetSize uint64) uint64 {
	var AbsolveAmount = q.Guarantees - targetSize
	q.Guarantees -= AbsolveAmount
	return AbsolveAmount
}

func (q *GuaranteedQueue) UseGuarantees() {
	for len(q.Queue) > 0 {
		var peekHead = string(q.Queue[0])

		if len(peekHead) == 0 || len(peekHead) > len(strconv.FormatUint(q.Guarantees, 10)) {
			return
		}

		head := q.Queue[0]
		q.Queue = q.Queue[1:]
		q.OutGoingBytes = append(q.OutGoingBytes, head)
		q.Guarantees -= uint64(len(string(head)))
	}
}

func fetchOutgoingBytes(ctx context.Context, outgoingBytes [][]byte) <-chan []byte {
	ch := make(chan []byte)

	go func() {
		defer close(ch)
		for _, bytes := range outgoingBytes {
			select {
			case <-ctx.Done(): // Allows for graceful shutdown
				return
			case ch <- bytes:
				// Simulate some asynchronous operation
			}
		}
	}()

	return ch
}
