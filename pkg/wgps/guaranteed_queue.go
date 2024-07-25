package wgps

import (
	"context"
	"strconv"
)

type GuaranteedQueue struct {
	Guarantees    uint64
	Queue         chan []byte
	ReceivedBytes []byte
	OutGoingBytes []byte
}

/** Add some bytes to the queue. */
func (q *GuaranteedQueue) Push(bytes []byte) {
	q.Queue <- bytes
	q.ReceivedBytes = append(q.ReceivedBytes, bytes...)
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
		var peekHead = string((q.ReceivedBytes)[0])

		if len(peekHead) == 0 || len(peekHead) > len(strconv.FormatUint(q.Guarantees, 10)) {
			return
		}

		head := q.ReceivedBytes[0]
		q.ReceivedBytes = q.ReceivedBytes[1:]
		q.OutGoingBytes = append(q.OutGoingBytes, head)
		q.Guarantees -= uint64(len(string(head)))
	}
}

func FetchOutgoingBytes(ctx context.Context, outgoingBytes [][]byte) chan []byte {
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
