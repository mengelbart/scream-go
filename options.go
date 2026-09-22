// Copyright (c) 2020 Mathis Engelbart All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scream

/*
#include "ScreamTxC.h"
#include "ScreamRxC.h"
#include <stdlib.h>
*/
import "C"
import "unsafe"

// A TxOption configures a Tx at construction. See NewTx.
type TxOption func(*txConfig)

// txConfig collects the ScreamV2Tx constructor arguments and the setters that
// have to run right after construction.
type txConfig struct {
	c C.ScreamTxConfig

	// packetPacing is tracked separately because the wrapper's default differs
	// from SCReAM's own, see WithPacketPacing.
	packetPacing bool

	// post runs against the constructed Tx, in the order the options were given.
	post []func(*Tx)
}

func defaultTxConfig() *txConfig {
	cfg := &txConfig{}
	C.ScreamTxDefaultConfig(&cfg.c)
	return cfg
}

// WithLossBeta sets the congestion window scale factor applied on loss.
func WithLossBeta(beta float64) TxOption {
	return func(c *txConfig) { c.c.lossBeta = C.float(beta) }
}

// WithECNCEBeta sets the congestion window scale factor applied on ECN-CE.
// It applies to classic ECN only.
func WithECNCEBeta(beta float64) TxOption {
	return func(c *txConfig) { c.c.ecnCeBeta = C.float(beta) }
}

// WithQueueDelayTargetMin sets the queue delay target in seconds.
func WithQueueDelayTargetMin(target float64) TxOption {
	return func(c *txConfig) { c.c.queueDelayTargetMin = C.float(target) }
}

// WithInitialCwnd sets the initial congestion window in bytes. It should be
// larger than twice the MSS, a useful value is initialBitrate/8*rtt.
func WithInitialCwnd(cwnd int) TxOption {
	return func(c *txConfig) { c.c.cwnd = C.int(cwnd) }
}

// WithPacketPacingHeadroom sets how much faster than the target bitrate video
// frames are transmitted.
func WithPacketPacingHeadroom(headroom float64) TxOption {
	return func(c *txConfig) { c.c.packetPacingHeadroom = C.float(headroom) }
}

// WithMaxAdaptivePacingRateScale sets the adaptive pacing rate scale. A value
// above 1.0 transmits large frames faster, which is mostly relevant with L4S.
func WithMaxAdaptivePacingRateScale(scale float64) TxOption {
	return func(c *txConfig) { c.c.maxAdaptivePacingRateScale = C.float(scale) }
}

// WithBytesInFlightHeadroom sets the bytes in flight headroom.
func WithBytesInFlightHeadroom(headroom float64) TxOption {
	return func(c *txConfig) { c.c.bytesInFlightHeadRoom = C.float(headroom) }
}

// WithMultiplicativeIncreaseScalefactor sets how much the congestion window
// may grow per RTT, 0.05 meaning at most 5%.
func WithMultiplicativeIncreaseScalefactor(factor float64) TxOption {
	return func(c *txConfig) { c.c.multiplicativeIncreaseScalefactor = C.float(factor) }
}

// WithL4S enables L4S mode.
func WithL4S(enabled bool) TxOption {
	return func(c *txConfig) { c.c.isL4s = C.bool(enabled) }
}

// WithMaxWindowHeadroom sets the maximum congestion window headroom.
func WithMaxWindowHeadroom(headroom float64) TxOption {
	return func(c *txConfig) { c.c.maxWindowHeadroom = C.float(headroom) }
}

// WithSharedBottleneckDetection enables shared bottleneck detection.
func WithSharedBottleneckDetection(enabled bool) TxOption {
	return func(c *txConfig) { c.c.enableSbd = C.bool(enabled) }
}

// WithClockDriftCompensation enables compensation for clock drift between
// sender and receiver.
func WithClockDriftCompensation(enabled bool) TxOption {
	return func(c *txConfig) { c.c.enableClockDriftCompensation = C.bool(enabled) }
}

// WithPacketPacing enables packet pacing. It is disabled by default, which
// differs from SCReAM's own default of enabled.
func WithPacketPacing(enabled bool) TxOption {
	return func(c *txConfig) { c.packetPacing = enabled }
}

// WithRelaxedPacing enables relaxed packet pacing.
func WithRelaxedPacing(enabled bool) TxOption {
	return func(c *txConfig) {
		c.post = append(c.post, func(t *Tx) {
			C.ScreamTxEnableRelaxedPacing(t.screamTx, C.bool(enabled))
		})
	}
}

// WithCyclicPacing enables cyclic packet pacing.
func WithCyclicPacing(enabled bool) TxOption {
	return func(c *txConfig) {
		c.post = append(c.post, func(t *Tx) {
			C.ScreamTxSetEnableCyclicPacing(t.screamTx, C.bool(enabled))
		})
	}
}

// WithCwndMinLow sets the lower bound on the minimum congestion window in
// bytes.
func WithCwndMinLow(cwnd int) TxOption {
	return func(c *txConfig) {
		c.post = append(c.post, func(t *Tx) {
			C.ScreamTxSetCwndMinLow(t.screamTx, C.int(cwnd))
		})
	}
}

// WithAutoTuneMinCwnd enables automatic tuning of the minimum congestion
// window.
func WithAutoTuneMinCwnd(enabled bool) TxOption {
	return func(c *txConfig) {
		c.post = append(c.post, func(t *Tx) {
			C.ScreamTxAutoTuneMinCwnd(t.screamTx, C.bool(enabled))
		})
	}
}

// WithAdaptiveWindowHeadroom enables adaptive congestion window headroom.
func WithAdaptiveWindowHeadroom(enabled bool) TxOption {
	return func(c *txConfig) {
		c.post = append(c.post, func(t *Tx) {
			C.ScreamTxSetEnableAdaptiveWindowHeadroom(t.screamTx, C.bool(enabled))
		})
	}
}

// WithRatePolicerProtection enables protection against rate policers.
func WithRatePolicerProtection(enabled bool) TxOption {
	return func(c *txConfig) {
		c.post = append(c.post, func(t *Tx) {
			C.ScreamTxSetEnableRatePolicerProtection(t.screamTx, C.bool(enabled))
		})
	}
}

// WithSchedulingJitterMargin sets the margin that absorbs scheduling jitter.
func WithSchedulingJitterMargin(margin float64) TxOption {
	return func(c *txConfig) {
		c.post = append(c.post, func(t *Tx) {
			C.ScreamTxSetSchedulingJitterMargin(t.screamTx, C.float(margin))
		})
	}
}

// WithPostCongestionDelayRtts sets how many RTTs the post congestion delay
// lasts.
func WithPostCongestionDelayRtts(rtts int) TxOption {
	return func(c *txConfig) {
		c.post = append(c.post, func(t *Tx) {
			C.ScreamTxSetPostCongestionDelayRtts(t.screamTx, C.int(rtts))
		})
	}
}

// WithReorderTime sets the packet reordering margin in seconds.
func WithReorderTime(seconds float64) TxOption {
	return func(c *txConfig) {
		c.post = append(c.post, func(t *Tx) {
			C.ScreamTxSetReorderTime(t.screamTx, C.float(seconds))
		})
	}
}

// WithRateUpdate enables updating of the target bitrate.
func WithRateUpdate(enabled bool) TxOption {
	return func(c *txConfig) {
		c.post = append(c.post, func(t *Tx) {
			C.ScreamTxSetEnableRateUpdate(t.screamTx, C.bool(enabled))
		})
	}
}

// WithLogTag sets the tag SCReAM prefixes its log lines with.
func WithLogTag(tag string) TxOption {
	return func(c *txConfig) {
		c.post = append(c.post, func(t *Tx) {
			// SCReAM keeps the pointer instead of copying, so the string has to
			// outlive the ScreamV2Tx. Close frees it.
			if t.logTag != nil {
				C.free(unsafe.Pointer(t.logTag))
			}
			t.logTag = C.CString(tag)
			C.ScreamTxSetLogTag(t.screamTx, t.logTag)
		})
	}
}

// WithMaxTotalBitrate sets the maximum total bitrate in bps across all
// streams. See also Tx.SetMaxTotalBitrate.
func WithMaxTotalBitrate(bitrate float64) TxOption {
	return func(c *txConfig) {
		c.post = append(c.post, func(t *Tx) {
			C.ScreamTxSetMaxTotalBitrate(t.screamTx, C.float(bitrate))
		})
	}
}

// A StreamOption configures a stream at registration.
// See Tx.RegisterNewStream.
type StreamOption func(*streamConfig)

type streamConfig struct {
	maxRTPQueueDelay        C.float
	adaptiveTargetRateScale C.bool
	hysteresis              C.float
	frameSizeOverhead       C.bool
}

func defaultStreamConfig() *streamConfig {
	return &streamConfig{
		maxRTPQueueDelay:        C.ScreamTxDefaultMaxRtpQueueDelay(),
		adaptiveTargetRateScale: C.bool(true),
		hysteresis:              C.float(0.0),
		frameSizeOverhead:       C.bool(true),
	}
}

// WithMaxRTPQueueDelay sets the RTP queue delay in seconds above which the
// stream's queue is flushed.
func WithMaxRTPQueueDelay(seconds float64) StreamOption {
	return func(c *streamConfig) { c.maxRTPQueueDelay = C.float(seconds) }
}

// WithAdaptiveTargetRateScale enables adaptive scaling of the target rate.
func WithAdaptiveTargetRateScale(enabled bool) StreamOption {
	return func(c *streamConfig) { c.adaptiveTargetRateScale = C.bool(enabled) }
}

// WithHysteresis sets how much the target rate must change before
// GetTargetBitrate reports a new value, as a fraction of the current rate.
func WithHysteresis(hysteresis float64) StreamOption {
	return func(c *streamConfig) { c.hysteresis = C.float(hysteresis) }
}

// WithFrameSizeOverhead enables compensation for varying frame sizes.
func WithFrameSizeOverhead(enabled bool) StreamOption {
	return func(c *streamConfig) { c.frameSizeOverhead = C.bool(enabled) }
}

// An RxOption configures an Rx at construction. See NewRx.
type RxOption func(*rxConfig)

type rxConfig struct {
	ackDiff            C.int
	reportedRTPPackets C.int
}

func defaultRxConfig() *rxConfig {
	return &rxConfig{
		// A negative ackDiff makes SCReAM derive it from reportedRTPPackets.
		ackDiff:            C.int(-1),
		reportedRTPPackets: C.ScreamRxDefaultReportedRtpPackets(),
	}
}

// WithAckDiff sets how many RTP packets are received between feedback
// messages. A value below 1 lets SCReAM derive it from the number of reported
// RTP packets.
func WithAckDiff(ackDiff int) RxOption {
	return func(c *rxConfig) { c.ackDiff = C.int(ackDiff) }
}

// WithReportedRTPPackets sets how many RTP packets one feedback message
// reports on.
func WithReportedRTPPackets(n int) RxOption {
	return func(c *rxConfig) { c.reportedRTPPackets = C.int(n) }
}
