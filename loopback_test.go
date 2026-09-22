// Copyright (c) 2020 Mathis Engelbart All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scream

import (
	"strings"
	"testing"
	"time"
)

// TestLoopback drives a Tx and an Rx against each other over an in-process
// link, exercising the full call sequence: NewMediaFrame, IsOkToTransmit,
// AddTransmitted, Receive, CreateStandardizedFeedback and
// IncomingStandardizedFeedback.
func TestLoopback(t *testing.T) {
	const (
		ssrc        = 1234
		frames      = 400
		framePeriod = 10 * time.Millisecond
		frameSize   = 1000
	)

	tx := NewTx(WithLogTag("loopback"))
	defer tx.Close()
	rx := NewRx(4321)
	defer rx.Close()

	q := NewQueue[testPacket]()
	if err := tx.RegisterNewStream(q, ssrc, 1.0, 1e5, 5e5, 2e6); err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	seq := uint16(0)
	transmitted, feedbacks := 0, 0

	for i := 0; i < frames; i++ {
		now := start.Add(time.Duration(i) * framePeriod)

		q.Enqueue(testPacket{ts: now, seq: seq, size: frameSize})
		tx.NewMediaFrame(now, ssrc, frameSize, true)
		seq++

		// Drain whatever SCReAM allows this tick.
		for {
			res, sendSSRC := tx.IsOkToTransmit(now)
			if res != 0.0 {
				break
			}
			p, ok := q.Dequeue()
			if !ok {
				break
			}
			tx.AddTransmitted(now, sendSSRC, p.Size(), p.SequenceNumber(), true)
			rx.Receive(now, sendSSRC, p.Size(), p.SequenceNumber(), 0, true)
			transmitted++
		}

		if rx.IsFeedback(now) {
			if fb, ok := rx.CreateStandardizedFeedback(now, true); ok {
				// Feed it back one frame period later, so the RTT is non-zero.
				tx.IncomingStandardizedFeedback(now.Add(framePeriod), fb)
				feedbacks++
			}
		}
	}

	if transmitted == 0 {
		t.Fatal("no packets were transmitted")
	}
	if feedbacks == 0 {
		t.Fatal("no feedback was generated")
	}

	end := start.Add(frames * framePeriod)
	bitrate := tx.GetTargetBitrate(end, ssrc)
	if bitrate <= 0 {
		t.Fatalf("target bitrate = %v, want a positive rate", bitrate)
	}

	stats := tx.GetStatistics(end)
	if !strings.Contains(stats, "RTT = ") {
		t.Fatalf("statistics = %q, want an RTT field", stats)
	}
	// The feedback loop has to have produced a measurable RTT, otherwise
	// nothing actually round tripped.
	if strings.Contains(stats, "RTT = 0.000s") {
		t.Errorf("statistics = %q, want a non-zero RTT", stats)
	}

	t.Logf("transmitted %v packets, %v feedbacks, target bitrate %.0f bps",
		transmitted, feedbacks, bitrate)
	t.Logf("%s", stats)
}

// Several streams sharing one Tx must all be served, and IsOkToTransmit must
// only ever name a registered stream.
func TestLoopbackMultiStream(t *testing.T) {
	tx := NewTx()
	defer tx.Close()

	ssrcs := []uint32{11, 22, 33}
	queues := map[uint32]*Queue[testPacket]{}
	served := map[uint32]int{}

	for _, ssrc := range ssrcs {
		q := NewQueue[testPacket]()
		queues[ssrc] = q
		if err := tx.RegisterNewStream(q, ssrc, 1.0, 1e5, 5e5, 1e6); err != nil {
			t.Fatal(err)
		}
	}

	start := time.Now()
	for i := 0; i < 300; i++ {
		now := start.Add(time.Duration(i) * 10 * time.Millisecond)
		for _, ssrc := range ssrcs {
			queues[ssrc].Enqueue(testPacket{ts: now, seq: uint16(i)})
			tx.NewMediaFrame(now, ssrc, 1000, true)
		}
		for {
			res, ssrc := tx.IsOkToTransmit(now)
			if res != 0.0 {
				break
			}
			q, ok := queues[ssrc]
			if !ok {
				t.Fatalf("IsOkToTransmit named unregistered ssrc %v", ssrc)
			}
			p, ok := q.Dequeue()
			if !ok {
				break
			}
			tx.AddTransmitted(now, ssrc, p.Size(), p.SequenceNumber(), true)
			served[ssrc]++
		}
	}

	for _, ssrc := range ssrcs {
		if served[ssrc] == 0 {
			t.Errorf("stream %v was never served", ssrc)
		}
	}
	t.Logf("served: %v", served)
}
