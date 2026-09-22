// Copyright (c) Mathis Engelbart. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scream

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestTxRegisterNewStream(t *testing.T) {
	tx := NewTx()
	defer tx.Close()

	if err := tx.RegisterNewStream(NewQueue[testPacket](), 1234, 1.0, 1e5, 5e5, 1e6); err != nil {
		t.Fatalf("first registration: %v", err)
	}
}

func TestTxRegisterDuplicateSSRC(t *testing.T) {
	tx := NewTx()
	defer tx.Close()

	if err := tx.RegisterNewStream(NewQueue[testPacket](), 1234, 1.0, 1e5, 5e5, 1e6); err != nil {
		t.Fatal(err)
	}
	err := tx.RegisterNewStream(NewQueue[testPacket](), 1234, 1.0, 1e5, 5e5, 1e6)
	if !errors.Is(err, ErrDuplicateSSRC) {
		t.Fatalf("got %v, want ErrDuplicateSSRC", err)
	}
}

// SCReAM registers streams into a fixed size array without bounds checking, so
// the wrapper has to refuse the overflowing registration.
func TestTxRegisterTooManyStreams(t *testing.T) {
	tx := NewTx()
	defer tx.Close()

	registered := 0
	for ssrc := uint32(1); ssrc < 100; ssrc++ {
		err := tx.RegisterNewStream(NewQueue[testPacket](), ssrc, 1.0, 1e5, 5e5, 1e6)
		if err != nil {
			if !errors.Is(err, ErrTooManyStreams) {
				t.Fatalf("got %v, want ErrTooManyStreams", err)
			}
			break
		}
		registered++
	}
	if registered == 0 || registered >= 99 {
		t.Fatalf("registered %v streams, expected a limit to be enforced", registered)
	}
	t.Logf("stream limit is %v", registered)
}

func TestTxCloseIsIdempotent(t *testing.T) {
	tx := NewTx()
	if err := tx.RegisterNewStream(NewQueue[testPacket](), 1, 1.0, 1e5, 5e5, 1e6); err != nil {
		t.Fatal(err)
	}
	if err := tx.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := tx.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestTxUseAfterClosePanics(t *testing.T) {
	// Calling into a freed ScreamV2Tx would be a segfault, so the wrapper
	// turns it into a Go panic instead.
	for name, call := range map[string]func(*Tx){
		"NewMediaFrame":     func(tx *Tx) { tx.NewMediaFrame(time.Now(), 1, 1000, true) },
		"IsOkToTransmit":    func(tx *Tx) { tx.IsOkToTransmit(time.Now()) },
		"AddTransmitted":    func(tx *Tx) { tx.AddTransmitted(time.Now(), 1, 1000, 1, true) },
		"GetTargetBitrate":  func(tx *Tx) { tx.GetTargetBitrate(time.Now(), 1) },
		"GetStatistics":     func(tx *Tx) { tx.GetStatistics(time.Now()) },
		"RegisterNewStream": func(tx *Tx) { tx.RegisterNewStream(NewQueue[testPacket](), 1, 1.0, 1e5, 5e5, 1e6) },
	} {
		t.Run(name, func(t *testing.T) {
			tx := NewTx()
			tx.Close()
			defer func() {
				if recover() == nil {
					t.Error("expected a panic after Close")
				}
			}()
			call(tx)
		})
	}
}

// IsOkToTransmit reports which stream SCReAM picked. The value is only
// meaningful when the returned time is 0.0.
func TestTxIsOkToTransmitReportsSSRC(t *testing.T) {
	tx := NewTx()
	defer tx.Close()

	const ssrc = 4321
	q := NewQueue[testPacket]()
	if err := tx.RegisterNewStream(q, ssrc, 1.0, 1e5, 5e5, 1e6); err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	opportunities := 0
	for i := 0; i < 50; i++ {
		now := start.Add(time.Duration(i) * 10 * time.Millisecond)
		q.Enqueue(testPacket{ts: now, seq: uint16(i)})
		tx.NewMediaFrame(now, ssrc, 1000, true)

		res, got := tx.IsOkToTransmit(now)
		if res != 0.0 {
			continue
		}
		opportunities++
		if got != ssrc {
			t.Fatalf("IsOkToTransmit returned ssrc %v, want %v", got, ssrc)
		}
		if p, ok := q.Dequeue(); ok {
			tx.AddTransmitted(now, got, p.Size(), p.SequenceNumber(), true)
		}
	}
	if opportunities == 0 {
		t.Fatal("IsOkToTransmit never allowed a transmission, the ssrc was never checked")
	}
}

func TestTxOptionsReachSCReAM(t *testing.T) {
	// The log tag is the one setting that is observable through the public
	// API, it prefixes the statistics summary.
	tx := NewTx(
		WithLogTag("tag-under-test"),
		WithL4S(true),
		WithPacketPacing(true),
		WithCwndMinLow(5000),
		WithInitialCwnd(20000),
		WithQueueDelayTargetMin(0.15),
		WithClockDriftCompensation(true),
		WithRelaxedPacing(true),
		WithReorderTime(0.05),
		WithMaxTotalBitrate(2e6),
	)
	defer tx.Close()

	if got := tx.GetStatistics(tx.created); !strings.HasPrefix(got, "tag-under-test ") {
		t.Errorf("statistics = %q, want the log tag as a prefix", got)
	}
}

// SCReAM stores the log tag pointer without copying it, so the Go side has to
// keep the C string alive for the lifetime of the Tx.
func TestTxLogTagOutlivesGC(t *testing.T) {
	tx := NewTx(WithLogTag("tag-alive"))
	defer tx.Close()

	for i := 0; i < 20; i++ {
		runtime.GC()
	}
	if got := tx.GetStatistics(tx.created); !strings.HasPrefix(got, "tag-alive ") {
		t.Errorf("statistics = %q, log tag did not survive", got)
	}
}

// The statistics timestamp is the elapsed time since the Tx was created, not
// an absolute NTP value.
func TestTxStatisticsTimestampIsRelative(t *testing.T) {
	tx := NewTx()
	defer tx.Close()

	for _, tc := range []struct {
		elapsed time.Duration
		want    string
	}{
		{0, " summary   0.0 "},
		{12500 * time.Millisecond, " summary  12.5 "},
	} {
		got := tx.GetStatistics(tx.created.Add(tc.elapsed))
		if !strings.HasPrefix(got, tc.want) {
			t.Errorf("statistics after %v = %q, want prefix %q", tc.elapsed, got, tc.want)
		}
	}
}

// The wrapper has always run with packet pacing disabled, unlike SCReAM whose
// own default is enabled. Adding the option must not have changed that.
//
// This checks the configuration the wrapper builds rather than the resulting
// behaviour, because whether pacing delays a transmission depends on the send
// window and is not observable through the public API.
func TestTxPacketPacingDefaultsOff(t *testing.T) {
	if defaultTxConfig().packetPacing {
		t.Error("packet pacing defaults to enabled, the wrapper has always disabled it")
	}

	for _, want := range []bool{true, false} {
		cfg := defaultTxConfig()
		WithPacketPacing(want)(cfg)
		if cfg.packetPacing != want {
			t.Errorf("WithPacketPacing(%v) set %v", want, cfg.packetPacing)
		}
	}
}

// The constructor options have to land in the struct handed to ScreamTxInit,
// which is the only place they can still take effect.
func TestTxConstructorOptions(t *testing.T) {
	cfg := defaultTxConfig()
	if cfg.c.isL4s {
		t.Error("L4S defaults to enabled, want disabled")
	}

	WithL4S(true)(cfg)
	WithInitialCwnd(20000)(cfg)
	WithQueueDelayTargetMin(0.15)(cfg)
	WithClockDriftCompensation(true)(cfg)
	WithSharedBottleneckDetection(true)(cfg)

	if !cfg.c.isL4s {
		t.Error("WithL4S(true) did not set isL4s")
	}
	if got, want := int(cfg.c.cwnd), 20000; got != want {
		t.Errorf("cwnd = %v, want %v", got, want)
	}
	if got, want := float64(cfg.c.queueDelayTargetMin), 0.15; math.Abs(got-want) > 1e-6 {
		t.Errorf("queueDelayTargetMin = %v, want %v", got, want)
	}
	if !cfg.c.enableClockDriftCompensation {
		t.Error("WithClockDriftCompensation(true) did not set the field")
	}
	if !cfg.c.enableSbd {
		t.Error("WithSharedBottleneckDetection(true) did not set the field")
	}
}

// The stream options have to reach registerNewStream, which takes them as
// arguments rather than through a setter.
func TestStreamConfigOptions(t *testing.T) {
	cfg := defaultStreamConfig()
	if !cfg.adaptiveTargetRateScale {
		t.Error("adaptive target rate scale defaults to off, want on")
	}
	if !cfg.frameSizeOverhead {
		t.Error("frame size overhead defaults to off, want on")
	}
	if float64(cfg.maxRTPQueueDelay) <= 0 {
		t.Errorf("default maxRTPQueueDelay = %v, want a positive value", cfg.maxRTPQueueDelay)
	}

	WithHysteresis(0.1)(cfg)
	WithMaxRTPQueueDelay(0.3)(cfg)
	WithAdaptiveTargetRateScale(false)(cfg)
	WithFrameSizeOverhead(false)(cfg)

	if got, want := float64(cfg.hysteresis), 0.1; math.Abs(got-want) > 1e-6 {
		t.Errorf("hysteresis = %v, want %v", got, want)
	}
	if got, want := float64(cfg.maxRTPQueueDelay), 0.3; math.Abs(got-want) > 1e-6 {
		t.Errorf("maxRTPQueueDelay = %v, want %v", got, want)
	}
	if cfg.adaptiveTargetRateScale {
		t.Error("WithAdaptiveTargetRateScale(false) did not clear the field")
	}
	if cfg.frameSizeOverhead {
		t.Error("WithFrameSizeOverhead(false) did not clear the field")
	}
}

// A negative ackDiff tells SCReAM to derive it, which is what the wrapper
// defaults to.
func TestRxConfigOptions(t *testing.T) {
	cfg := defaultRxConfig()
	if cfg.ackDiff >= 0 {
		t.Errorf("default ackDiff = %v, want a negative value", cfg.ackDiff)
	}
	if cfg.reportedRTPPackets <= 0 {
		t.Errorf("default reportedRTPPackets = %v, want a positive value", cfg.reportedRTPPackets)
	}

	WithAckDiff(4)(cfg)
	WithReportedRTPPackets(16)(cfg)
	if got, want := int(cfg.ackDiff), 4; got != want {
		t.Errorf("ackDiff = %v, want %v", got, want)
	}
	if got, want := int(cfg.reportedRTPPackets), 16; got != want {
		t.Errorf("reportedRTPPackets = %v, want %v", got, want)
	}
}

func TestTxStreamOptions(t *testing.T) {
	tx := NewTx()
	defer tx.Close()

	err := tx.RegisterNewStream(NewQueue[testPacket](), 1234, 1.0, 1e5, 5e5, 1e6,
		WithHysteresis(0.1),
		WithMaxRTPQueueDelay(0.3),
		WithAdaptiveTargetRateScale(false),
		WithFrameSizeOverhead(false),
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTxRuntimeSetters(t *testing.T) {
	tx := NewTx()
	defer tx.Close()

	const ssrc = 1234
	if err := tx.RegisterNewStream(NewQueue[testPacket](), ssrc, 1.0, 1e5, 5e5, 1e6); err != nil {
		t.Fatal(err)
	}

	tx.UpdateBitrateStream(ssrc, 2e5, 2e6)
	tx.SetTargetPriority(ssrc, 0.5)
	tx.SetMaxTotalBitrate(3e6)

	// The new minimum takes effect, the target bitrate cannot fall below it.
	if got := tx.GetTargetBitrate(time.Now(), ssrc); got < 2e5 {
		t.Errorf("target bitrate = %v, want at least the new minimum 200000", got)
	}
}

// SCReAM memcpy's the MSS list into a fixed size array without bounds
// checking, so an oversized list has to be refused.
func TestTxSetMssList(t *testing.T) {
	tx := NewTx()
	defer tx.Close()

	if err := tx.SetMssListMinPacketsInFlight(nil, 4); err != nil {
		t.Errorf("empty list: %v", err)
	}
	if err := tx.SetMssListMinPacketsInFlight([]int{1200, 1400}, 4); err != nil {
		t.Errorf("short list: %v", err)
	}
	if err := tx.SetMssListMinPacketsInFlight(make([]int, 1000), 4); !errors.Is(err, ErrMssListTooLong) {
		t.Errorf("got %v, want ErrMssListTooLong", err)
	}
}

// The RTPQueue callbacks cross from C++ back into Go. Forcing collections in
// between catches a handle that was not kept alive.
func TestTxCallbacksUnderGC(t *testing.T) {
	tx := NewTx()
	defer tx.Close()

	const ssrc = 1234
	q := NewQueue[testPacket]()
	if err := tx.RegisterNewStream(q, ssrc, 1.0, 1e5, 5e5, 1e6); err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	for i := 0; i < 100; i++ {
		now := start.Add(time.Duration(i) * 10 * time.Millisecond)
		q.Enqueue(testPacket{ts: now, seq: uint16(i)})
		tx.NewMediaFrame(now, ssrc, 1000, true)
		runtime.GC()
		if res, got := tx.IsOkToTransmit(now); res == 0.0 {
			if p, ok := q.Dequeue(); ok {
				tx.AddTransmitted(now, got, p.Size(), p.SequenceNumber(), true)
			}
		}
		runtime.GC()
		tx.GetTargetBitrate(now, ssrc)
	}
}

// ccfb builds an RFC 8888 feedback packet with one report block per numReports
// entry. A zero entry produces an empty report block.
func ccfb(ssrc uint32, numReports ...int) []byte {
	buf := make([]byte, 8)
	buf[0] = 0x80 | 11
	buf[1] = 205
	binary.BigEndian.PutUint32(buf[4:], 1)
	for _, n := range numReports {
		block := make([]byte, 8+2*n)
		binary.BigEndian.PutUint32(block, ssrc)
		binary.BigEndian.PutUint16(block[4:], 100)
		binary.BigEndian.PutUint16(block[6:], uint16(n))
		for i := 0; i < n; i++ {
			binary.BigEndian.PutUint16(block[8+2*i:], 0x8000)
		}
		if n%2 == 1 {
			block = append(block, 0, 0)
		}
		buf = append(buf, block...)
	}
	buf = append(buf, 0, 0, 0, 0)
	binary.BigEndian.PutUint16(buf[2:], uint16(len(buf)/4-1))
	return buf
}

// SCReAM derives its metric block count as num_reports-1 in a uint16, so an
// empty report block, which RFC 8888 allows, reads far past the buffer.
func TestTxIncomingFeedbackEmptyReportBlock(t *testing.T) {
	tx := NewTx()
	defer tx.Close()

	const ssrc = 1234
	if err := tx.RegisterNewStream(NewQueue[testPacket](), ssrc, 1.0, 1e5, 5e5, 1e6); err != nil {
		t.Fatal(err)
	}
	if err := tx.IncomingStandardizedFeedback(time.Now(), ccfb(ssrc, 0)); err != nil {
		t.Fatalf("empty report block: %v", err)
	}
	if err := tx.IncomingStandardizedFeedback(time.Now(), ccfb(ssrc, 0, 2, 0)); err != nil {
		t.Fatalf("mixed report blocks: %v", err)
	}
}

func TestSanitizeStandardizedFeedback(t *testing.T) {
	const ssrc = 1234
	for _, tc := range []struct {
		name string
		buf  []byte
		want []byte
	}{
		{"single block", ccfb(ssrc, 2), ccfb(ssrc, 2)},
		{"odd block padded", ccfb(ssrc, 3), ccfb(ssrc, 3)},
		{"empty block dropped", ccfb(ssrc, 0, 2), ccfb(ssrc, 2)},
		{"trailing empty block dropped", ccfb(ssrc, 3, 0), ccfb(ssrc, 3)},
		{"only empty blocks", ccfb(ssrc, 0, 0), nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := sanitizeStandardizedFeedback(tc.buf)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, tc.want) {
				t.Fatalf("got % x, want % x", got, tc.want)
			}
		})
	}
}

func TestSanitizeStandardizedFeedbackMalformed(t *testing.T) {
	const ssrc = 1234
	truncatedBlock := ccfb(ssrc, 2)
	binary.BigEndian.PutUint16(truncatedBlock[14:], 3)
	leftoverOctets := ccfb(ssrc, 1)
	binary.BigEndian.PutUint16(leftoverOctets[14:], 0)
	badLength := ccfb(ssrc, 2)
	binary.BigEndian.PutUint16(badLength[2:], 9)

	for _, tc := range []struct {
		name string
		buf  []byte
	}{
		{"empty", nil},
		{"header only", ccfb(ssrc)},
		{"not word aligned", append(ccfb(ssrc, 2), 0)},
		{"length field too large", badLength},
		{"block longer than packet", truncatedBlock},
		{"leftover octets after last block", leftoverOctets},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := sanitizeStandardizedFeedback(tc.buf); !errors.Is(err, ErrMalformedFeedback) {
				t.Fatalf("got %v, want ErrMalformedFeedback", err)
			}
		})
	}
}
