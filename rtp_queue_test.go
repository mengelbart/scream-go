// Copyright (c) 2020 Mathis Engelbart All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scream

import (
	"math"
	"sync"
	"testing"
	"time"
)

func TestQueueEmpty(t *testing.T) {
	q := NewQueue[testPacket]()

	if got := q.SizeOfQueue(); got != 0 {
		t.Errorf("SizeOfQueue = %v, want 0", got)
	}
	if got := q.BytesInQueue(); got != 0 {
		t.Errorf("BytesInQueue = %v, want 0", got)
	}
	if got := q.SizeOfNextRTP(); got != 0 {
		t.Errorf("SizeOfNextRTP = %v, want 0", got)
	}
	if got := q.SeqNrOfNextRTP(); got != 0 {
		t.Errorf("SeqNrOfNextRTP = %v, want 0", got)
	}
	if got := q.SeqNrOfLastRTP(); got != 0 {
		t.Errorf("SeqNrOfLastRTP = %v, want 0", got)
	}
	if got := q.GetSizeOfLastFrame(); got != 0 {
		t.Errorf("GetSizeOfLastFrame = %v, want 0", got)
	}
	if got := q.GetDelay(1234); got != 0 {
		t.Errorf("GetDelay = %v, want 0", got)
	}
	if _, ok := q.Dequeue(); ok {
		t.Error("Dequeue on an empty queue returned ok")
	}
}

func TestQueueOrderAndAccounting(t *testing.T) {
	now := time.Now()
	q := NewQueue[testPacket]()
	q.Enqueue(testPacket{ts: now, seq: 10, size: 100})
	q.Enqueue(testPacket{ts: now, seq: 11, size: 200})
	q.Enqueue(testPacket{ts: now, seq: 12, size: 300})

	if got, want := q.SizeOfQueue(), 3; got != want {
		t.Errorf("SizeOfQueue = %v, want %v", got, want)
	}
	if got, want := q.BytesInQueue(), 600; got != want {
		t.Errorf("BytesInQueue = %v, want %v", got, want)
	}
	// "next" is the oldest packet, "last" the most recently enqueued one.
	if got, want := q.SeqNrOfNextRTP(), uint16(10); got != want {
		t.Errorf("SeqNrOfNextRTP = %v, want %v", got, want)
	}
	if got, want := q.SeqNrOfLastRTP(), uint16(12); got != want {
		t.Errorf("SeqNrOfLastRTP = %v, want %v", got, want)
	}
	if got, want := q.SizeOfNextRTP(), 100; got != want {
		t.Errorf("SizeOfNextRTP = %v, want %v", got, want)
	}

	p, ok := q.Dequeue()
	if !ok {
		t.Fatal("Dequeue returned not ok")
	}
	if got, want := p.SequenceNumber(), uint16(10); got != want {
		t.Errorf("dequeued seq = %v, want %v (FIFO)", got, want)
	}
	if got, want := q.BytesInQueue(), 500; got != want {
		t.Errorf("BytesInQueue after dequeue = %v, want %v", got, want)
	}
}

func TestQueueClear(t *testing.T) {
	q := NewQueue[testPacket]()
	for i := 0; i < 5; i++ {
		q.Enqueue(testPacket{ts: time.Now(), seq: uint16(i)})
	}

	if got, want := q.Clear(), 5; got != want {
		t.Errorf("Clear = %v, want %v dropped", got, want)
	}
	if got := q.SizeOfQueue(); got != 0 {
		t.Errorf("SizeOfQueue after Clear = %v, want 0", got)
	}
	if got := q.BytesInQueue(); got != 0 {
		t.Errorf("BytesInQueue after Clear = %v, want 0", got)
	}
}

// TestQueueGetDelay covers the unit mismatch that made GetDelay always return
// zero. currTs arrives from SCReAM as time_ntp/65536, i.e. seconds modulo
// 2^16, and the packet timestamp has to be converted into the same domain.
func TestQueueGetDelay(t *testing.T) {
	now := time.Now()

	for _, tc := range []struct {
		name string
		age  time.Duration
		want float64
	}{
		{"fresh", 0, 0},
		{"50ms", 50 * time.Millisecond, 0.05},
		{"1s", time.Second, 1},
		{"2.5s", 2500 * time.Millisecond, 2.5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			q := NewQueue[testPacket]()
			q.Enqueue(testPacket{ts: now.Add(-tc.age), seq: 1})

			got := q.GetDelay(float64(toNTP32(now)) / 65536.0)
			if math.Abs(got-tc.want) > 0.01 {
				t.Errorf("GetDelay = %v, want ~%v", got, tc.want)
			}
		})
	}
}

// A timestamp in the future must clamp to zero rather than wrap to ~2^16.
func TestQueueGetDelayFutureTimestamp(t *testing.T) {
	now := time.Now()
	q := NewQueue[testPacket]()
	q.Enqueue(testPacket{ts: now.Add(50 * time.Millisecond), seq: 1})

	if got := q.GetDelay(float64(toNTP32(now)) / 65536.0); got != 0 {
		t.Errorf("GetDelay = %v, want 0", got)
	}
}

// GetDelay must stay correct when currTs wraps past 2^16 seconds while a
// packet is still queued.
func TestQueueGetDelayAcrossWrap(t *testing.T) {
	// A packet enqueued one second before the wrap.
	enqueued := ntpShortTimeNear(t, ntpShortRange-1)
	q := NewQueue[testPacket]()
	q.Enqueue(testPacket{ts: enqueued, seq: 1})

	// Now read the delay one second after the wrap, so currTs is small while
	// the packet timestamp is near the top of the range.
	currTs := float64(toNTP32(enqueued.Add(2*time.Second))) / 65536.0
	got := q.GetDelay(currTs)
	if math.Abs(got-2) > 0.01 {
		t.Errorf("GetDelay across wrap = %v, want ~2", got)
	}
}

func TestQueueGetSizeOfLastFrame(t *testing.T) {
	now := time.Now()
	q := NewQueue[testPacket]()
	q.Enqueue(testPacket{ts: now, seq: 1, size: 100})
	q.Enqueue(testPacket{ts: now, seq: 2, size: 250})

	// Packet carries no marker bit, so Queue reports the last packet's size
	// rather than the last frame's. See the doc comment on the method.
	if got, want := q.GetSizeOfLastFrame(), 250; got != want {
		t.Errorf("GetSizeOfLastFrame = %v, want %v", got, want)
	}
}

// Queue documents itself as safe for concurrent use. Run under -race.
func TestQueueConcurrent(t *testing.T) {
	q := NewQueue[testPacket]()
	now := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(base int) {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				q.Enqueue(testPacket{ts: now, seq: uint16(base*200 + j)})
			}
		}(i)
	}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				q.Dequeue()
				q.BytesInQueue()
				q.SizeOfQueue()
				q.GetDelay(float64(toNTP32(now)) / 65536.0)
				q.SeqNrOfNextRTP()
				q.SeqNrOfLastRTP()
				q.GetSizeOfLastFrame()
			}
		}()
	}
	wg.Wait()
}
