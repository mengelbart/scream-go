// Copyright (c) 2020 Mathis Engelbart All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scream

import (
	"encoding/binary"
	"testing"
	"time"
)

func TestRxCloseIsIdempotent(t *testing.T) {
	rx := NewRx(1234)
	if err := rx.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := rx.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestRxUseAfterClosePanics(t *testing.T) {
	for name, call := range map[string]func(*Rx){
		"Receive":                    func(rx *Rx) { rx.Receive(time.Now(), 1, 1000, 1, 0, true) },
		"IsFeedback":                 func(rx *Rx) { rx.IsFeedback(time.Now()) },
		"CreateStandardizedFeedback": func(rx *Rx) { rx.CreateStandardizedFeedback(time.Now(), true) },
	} {
		t.Run(name, func(t *testing.T) {
			rx := NewRx(1234)
			rx.Close()
			defer func() {
				if recover() == nil {
					t.Error("expected a panic after Close")
				}
			}()
			call(rx)
		})
	}
}

func TestRxNoFeedbackBeforeReceive(t *testing.T) {
	rx := NewRx(1234)
	defer rx.Close()

	now := time.Now()
	if rx.IsFeedback(now) {
		t.Error("IsFeedback reported pending feedback before any packet arrived")
	}
	if fb, ok := rx.CreateStandardizedFeedback(now, true); ok || fb != nil {
		t.Errorf("CreateStandardizedFeedback = %v, %v, want nil, false", fb, ok)
	}
}

func TestRxCreateStandardizedFeedback(t *testing.T) {
	const rxSSRC = 0xDEADBEEF
	rx := NewRx(rxSSRC)
	defer rx.Close()

	now := time.Now()
	for i := 0; i < 8; i++ {
		rx.Receive(now, 1234, 1000, uint16(i), 0, true)
	}
	if !rx.IsFeedback(now) {
		t.Fatal("expected pending feedback after receiving packets")
	}

	fb, ok := rx.CreateStandardizedFeedback(now, true)
	if !ok {
		t.Fatal("CreateStandardizedFeedback returned not ok")
	}
	if len(fb) < 8 {
		t.Fatalf("feedback is %v bytes, too short to be an RTCP packet", len(fb))
	}

	// RFC 8888 CCFB: version 2, FMT 11, payload type 205, then the sender SSRC.
	if got, want := fb[0], byte(0x8B); got != want {
		t.Errorf("first byte = %#x, want %#x (V=2, FMT=11)", got, want)
	}
	if got, want := fb[1], byte(205); got != want {
		t.Errorf("payload type = %v, want %v", got, want)
	}
	if got := binary.BigEndian.Uint32(fb[4:8]); got != rxSSRC {
		t.Errorf("sender ssrc = %#x, want %#x", got, rxSSRC)
	}
}

// The feedback buffer is sized against SCReAM's internal limit, which it does
// not bounds check. Flood the receiver and make sure nothing overflows.
func TestRxFeedbackStaysWithinBuffer(t *testing.T) {
	rx := NewRx(1234)
	defer rx.Close()

	now := time.Now()
	for i := 0; i < 2000; i++ {
		rx.Receive(now, uint32(1000+i%4), 1200, uint16(i), 0, true)
	}

	fb, ok := rx.CreateStandardizedFeedback(now, true)
	if !ok {
		t.Fatal("expected feedback")
	}
	if len(fb) > 2048 {
		t.Fatalf("feedback is %v bytes, larger than the buffer it was written into", len(fb))
	}
	t.Logf("feedback for 2000 packets: %v bytes", len(fb))
}

func TestRxOptions(t *testing.T) {
	rx := NewRx(1234, WithAckDiff(4), WithReportedRTPPackets(16))
	defer rx.Close()

	now := time.Now()
	for i := 0; i < 8; i++ {
		rx.Receive(now, 4321, 1000, uint16(i), 0, true)
	}
	if !rx.IsFeedback(now) {
		t.Fatal("expected pending feedback")
	}
	if _, ok := rx.CreateStandardizedFeedback(now, true); !ok {
		t.Fatal("expected feedback")
	}
}
