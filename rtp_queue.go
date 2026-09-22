package scream

/*
#include <stdint.h>
*/
import "C"
import (
	"runtime/cgo"
	"sync"
	"time"
)

// RTPQueue implements a simple RTP packet queue. One RTPQueue should be used
// per SSRC stream.
//
// Throughout this interface "next" refers to the oldest item, i.e. the one that
// will be dequeued next, and "last" to the most recently enqueued item.
//
// The methods are called as callbacks from within Tx calls, on the goroutine
// that is calling the Tx. They must not call back into that Tx.
type RTPQueue interface {
	// SizeOfNextRTP returns the size of the next item in the queue.
	SizeOfNextRTP() int

	// SeqNrOfNextRTP returns the RTP sequence number of the next item in the queue
	SeqNrOfNextRTP() uint16

	// SeqNrOfLastRTP returns the RTP sequence number of the last item in the queue
	SeqNrOfLastRTP() uint16

	// BytesInQueue returns the total number of bytes in the queue, i.e. the
	// sum of the sizes of all items in the queue.
	BytesInQueue() int

	// SizeOfQueue returns the number of items in the queue.
	SizeOfQueue() int

	// GetDelay returns how long the next item has been in the queue.
	// ts is given in seconds in the NTP short format domain, i.e. it wraps
	// every 2^16 seconds.
	GetDelay(ts float64) float64

	// GetSizeOfLastFrame returns the total size of the most recently enqueued
	// frame, i.e. the sum of the sizes of its RTP packets.
	GetSizeOfLastFrame() int

	// Clear empties the queue and returns the number of items dropped.
	Clear() int
}

// queueFromContext resolves the callback context to the RTPQueue it was
// registered with. It returns nil if the context does not hold one, so that a
// callback returns a zero value instead of panicking across the C boundary.
func queueFromContext(context C.uintptr_t) RTPQueue {
	queue, _ := cgo.Handle(context).Value().(RTPQueue)
	return queue
}

//export goClear
func goClear(context C.uintptr_t) C.int {
	queue := queueFromContext(context)
	if queue == nil {
		return 0
	}
	return C.int(queue.Clear())
}

//export goSizeOfNextRtp
func goSizeOfNextRtp(context C.uintptr_t) C.int {
	queue := queueFromContext(context)
	if queue == nil {
		return 0
	}
	return C.int(queue.SizeOfNextRTP())
}

//export goSeqNrOfNextRtp
func goSeqNrOfNextRtp(context C.uintptr_t) C.int {
	queue := queueFromContext(context)
	if queue == nil {
		return 0
	}
	return C.int(queue.SeqNrOfNextRTP())
}

//export goSeqNrOfLastRtp
func goSeqNrOfLastRtp(context C.uintptr_t) C.int {
	queue := queueFromContext(context)
	if queue == nil {
		return 0
	}
	return C.int(queue.SeqNrOfLastRTP())
}

//export goBytesInQueue
func goBytesInQueue(context C.uintptr_t) C.int {
	queue := queueFromContext(context)
	if queue == nil {
		return 0
	}
	return C.int(queue.BytesInQueue())
}

//export goSizeOfQueue
func goSizeOfQueue(context C.uintptr_t) C.int {
	queue := queueFromContext(context)
	if queue == nil {
		return 0
	}
	return C.int(queue.SizeOfQueue())
}

//export goGetDelay
func goGetDelay(context C.uintptr_t, currTs C.float) C.float {
	queue := queueFromContext(context)
	if queue == nil {
		return 0
	}
	return C.float(queue.GetDelay(float64(currTs)))
}

//export goGetSizeOfLastFrame
func goGetSizeOfLastFrame(context C.uintptr_t) C.int {
	queue := queueFromContext(context)
	if queue == nil {
		return 0
	}
	return C.int(queue.GetSizeOfLastFrame())
}

var _ RTPQueue = (*Queue[Packet])(nil)

type Packet interface {
	Size() int
	SequenceNumber() uint16
	Timestamp() time.Time
}

// Queue is a RTPQueue backed by a slice. It is safe for concurrent use.
type Queue[T Packet] struct {
	lock sync.RWMutex
	data []T
}

func NewQueue[T Packet]() *Queue[T] {
	return &Queue[T]{
		lock: sync.RWMutex{},
		data: make([]T, 0),
	}
}

// BytesInQueue implements RTPQueue.
func (q *Queue[T]) BytesInQueue() int {
	q.lock.RLock()
	defer q.lock.RUnlock()
	res := 0
	for _, p := range q.data {
		res += p.Size()
	}
	return res
}

// Clear implements RTPQueue.
func (q *Queue[T]) Clear() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	size := len(q.data)
	q.data = make([]T, 0)
	return size
}

// GetDelay implements RTPQueue.
func (q *Queue[T]) GetDelay(ts float64) float64 {
	q.lock.RLock()
	defer q.lock.RUnlock()
	if len(q.data) == 0 {
		return 0
	}
	tsf := float64(toNTP32(q.data[0].Timestamp())) / 65536.0
	d := ts - tsf
	if d < -ntpShortHalfRange {
		// ts wrapped since the packet was enqueued
		d += ntpShortRange
	}
	return max(0, d)
}

// GetSizeOfLastFrame implements RTPQueue. Packet carries no marker bit, so
// Queue cannot detect frame boundaries and returns the size of the last packet
// instead of the size of the last frame. SCReAM does not currently use this
// value. Implement RTPQueue directly if the exact value is needed.
func (q *Queue[T]) GetSizeOfLastFrame() int {
	q.lock.RLock()
	defer q.lock.RUnlock()
	if len(q.data) == 0 {
		return 0
	}
	return q.data[len(q.data)-1].Size()
}

// SeqNrOfLastRTP implements RTPQueue.
func (q *Queue[T]) SeqNrOfLastRTP() uint16 {
	q.lock.RLock()
	defer q.lock.RUnlock()
	if len(q.data) == 0 {
		return 0
	}
	return q.data[len(q.data)-1].SequenceNumber()
}

// SeqNrOfNextRTP implements RTPQueue.
func (q *Queue[T]) SeqNrOfNextRTP() uint16 {
	q.lock.RLock()
	defer q.lock.RUnlock()
	if len(q.data) == 0 {
		return 0
	}
	return q.data[0].SequenceNumber()
}

// SizeOfNextRTP implements RTPQueue.
func (q *Queue[T]) SizeOfNextRTP() int {
	q.lock.RLock()
	defer q.lock.RUnlock()
	if len(q.data) == 0 {
		return 0
	}
	return q.data[0].Size()
}

// SizeOfQueue implements RTPQueue.
func (q *Queue[T]) SizeOfQueue() int {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return len(q.data)
}

func (q *Queue[T]) Enqueue(pkt T) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.data = append(q.data, pkt)
}

func (q *Queue[T]) Dequeue() (T, bool) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if len(q.data) == 0 {
		return *new(T), false
	}
	var next T
	next, q.data = q.data[0], q.data[1:]
	return next, true
}
