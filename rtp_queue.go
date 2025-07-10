package scream

import "C"
import (
	"runtime/cgo"
	"sync"
	"time"
	"unsafe"
)

// RTPQueue implements a simple RTP packet queue. One RTPQueue should be used
// per SSRC stream.
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

	// GetDelay returns the delay of the last item in the queue.
	// ts is given in seconds.
	GetDelay(ts float64) float64

	// GetSizeOfLastFrame returns the size of the latest pushed item.
	GetSizeOfLastFrame() int

	// Clear empties the queue.
	Clear() int
}

//export goClear
func goClear(context unsafe.Pointer) C.int {
	h := *(*cgo.Handle)(context)
	queue := h.Value().(RTPQueue)
	return C.int(queue.Clear())
}

//export goSizeOfNextRtp
func goSizeOfNextRtp(context unsafe.Pointer) C.int {
	h := *(*cgo.Handle)(context)
	queue := h.Value().(RTPQueue)
	return C.int(queue.SizeOfNextRTP())
}

//export goSeqNrOfNextRtp
func goSeqNrOfNextRtp(context unsafe.Pointer) C.int {
	h := *(*cgo.Handle)(context)
	queue := h.Value().(RTPQueue)
	return C.int(queue.SeqNrOfNextRTP())
}

//export goSeqNrOfLastRtp
func goSeqNrOfLastRtp(context unsafe.Pointer) C.int {
	h := *(*cgo.Handle)(context)
	queue := h.Value().(RTPQueue)
	return C.int(queue.SeqNrOfLastRTP())
}

//export goBytesInQueue
func goBytesInQueue(context unsafe.Pointer) C.int {
	h := *(*cgo.Handle)(context)
	queue := h.Value().(RTPQueue)
	return C.int(queue.BytesInQueue())
}

//export goSizeOfQueue
func goSizeOfQueue(context unsafe.Pointer) C.int {
	h := *(*cgo.Handle)(context)
	queue := h.Value().(RTPQueue)
	return C.int(queue.SizeOfQueue())
}

//export goGetDelay
func goGetDelay(context unsafe.Pointer, currTs C.float) C.float {
	h := *(*cgo.Handle)(context)
	queue, ok := h.Value().(RTPQueue)
	if !ok {
		panic("got invalid pointer")
	}
	x := queue.GetDelay(float64(currTs))
	d := C.float(x)
	return d
}

//export goGetSizeOfLastFrame
func goGetSizeOfLastFrame(context unsafe.Pointer) C.int {
	h := *(*cgo.Handle)(context)
	queue := h.Value().(RTPQueue)
	return C.int(queue.GetSizeOfLastFrame())
}

var _ RTPQueue = (*Queue[Packet])(nil)

type Packet interface {
	Size() int
	SequenceNumber() uint16
	Timestamp() time.Time
}

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
	t := q.data[0].Timestamp()
	tsf := float64(toNTP(t)) / 65536.0
	d := ts - tsf
	return max(0, d)
}

// GetSizeOfLastFrame implements RTPQueue.
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
