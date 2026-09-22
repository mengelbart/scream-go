// Copyright (c) 2020 Mathis Engelbart All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scream

/*
#include "ScreamTxC.h"
#include "RtpQueueC.h"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"runtime/cgo"
	"time"
)

var (
	// ErrDuplicateSSRC is returned when a stream is registered with an ssrc
	// that is already in use. SCReAM resolves a stream by its ssrc and would
	// never reach the second registration.
	ErrDuplicateSSRC = errors.New("ssrc already registered")

	// ErrTooManyStreams is returned when more streams are registered than
	// SCReAM can hold.
	ErrTooManyStreams = errors.New("too many streams")
)

// stream holds the per-stream resources that Close has to free.
type stream struct {
	ssrc uint32
	// handle is the callback context passed to C.
	handle   cgo.Handle
	rtpQueue *C.RtpQueueC
}

// Tx implements the sender side of SCReAM. A Tx is not safe for concurrent use.
// The RTPQueue methods are invoked as callbacks from within Tx calls, so a
// RTPQueue must be safe for use from whichever goroutine is calling the Tx at
// the time.
type Tx struct {
	screamTx *C.ScreamV2Tx
	streams  []stream
	created  time.Time
}

// NewTx creates a new Tx instance. Close must be called to free the resources
// held by the returned Tx.
func NewTx() *Tx {
	return &Tx{
		screamTx: C.ScreamTxInit(),
		created:  time.Now(),
	}
}

func (t *Tx) check() {
	if t.screamTx == nil {
		panic("scream: Tx used after Close")
	}
}

// RegisterNewStream registers a new stream with ssrc using rtpQueue.
// Priority is in the range ]0.0..1.0] where 1.0 denotes the highest priority.
// It is recommended that at least one stream has priority 1.0.
// Bitrates are specified in bps
//
// It returns ErrDuplicateSSRC if ssrc is already registered, and
// ErrTooManyStreams if the Tx is full.
func (t *Tx) RegisterNewStream(rtpQueue RTPQueue, ssrc uint32, priority, minBitrate, startBitrate, maxBitrate float64) error {
	t.check()
	for _, s := range t.streams {
		if s.ssrc == ssrc {
			return fmt.Errorf("%w: %v", ErrDuplicateSSRC, ssrc)
		}
	}
	if len(t.streams) >= int(C.ScreamTxMaxStreams()) {
		return fmt.Errorf("%w: limit is %v", ErrTooManyStreams, C.ScreamTxMaxStreams())
	}

	h := cgo.NewHandle(rtpQueue)
	rtpQueueC := C.RtpQueueCInit(C.uintptr_t(h))
	t.streams = append(t.streams, stream{ssrc: ssrc, handle: h, rtpQueue: rtpQueueC})
	C.ScreamTxRegisterNewStream(t.screamTx, rtpQueueC, C.uint32_t(ssrc), C.float(priority), C.float(minBitrate), C.float(startBitrate), C.float(maxBitrate))
	return nil
}

// Close closes t and frees all resources. Any other method called after Close
// panics. Close is idempotent.
func (t *Tx) Close() error {
	if t.screamTx == nil {
		return nil
	}
	C.ScreamTxFree(t.screamTx)
	t.screamTx = nil

	for _, s := range t.streams {
		C.RtpQueueCFree(s.rtpQueue)
		s.handle.Delete()
	}
	t.streams = nil
	return nil
}

// NewMediaFrame should be called for each new video frame.
// IsOkToTransmit should be called after newMediaFrame
func (t *Tx) NewMediaFrame(ts time.Time, ssrc uint32, bytesRTP int, isMarker bool) {
	t.check()
	C.ScreamTxNewMediaFrame(t.screamTx, C.uint32_t(toNTP32(ts)), C.uint32_t(ssrc), C.int(bytesRTP), C.bool(isMarker))
}

// IsOkToTransmit determines if an RTP packet can be transmitted and returns
// the SSRC of the stream that SCReAM selected. The returned SSRC is only valid
// if the returned float64 is 0.0.
// Returns:
//
// 0.0: RTP packet with the returned ssrc can be immediately transmitted.
// AddTransmitted must be called if packet is transmitted as a result of this.
//
// >0.0: Time [s] until this function should be called again. This can be used to
// start a timer.
// Note that a call to NewMediaFrame or IncomingFeedback should cause an immediate
// call to isOkToTransmit.
//
// -1.0: No RTP packet available to transmit or send window is not large enough
func (t *Tx) IsOkToTransmit(ts time.Time) (float64, uint32) {
	t.check()
	var ssrc C.uint32_t
	res := C.ScreamTxIsOkToTransmit(t.screamTx, C.uint32_t(toNTP32(ts)), &ssrc)
	return float64(res), uint32(ssrc)
}

// AddTransmitted adds a packet to list of transmitted packets. Should be called when an
// RTP packet was transmitted. AddTransmitted returns the time until IsOkToTransmit can be
// called again.
func (t *Tx) AddTransmitted(ts time.Time, ssrc uint32, size int, seqNr uint16, isMark bool) float64 {
	t.check()
	return float64(C.ScreamTxAddTransmitted(t.screamTx, C.uint32_t(toNTP32(ts)), C.uint32_t(ssrc), C.int(size), C.uint16_t(seqNr), C.bool(isMark)))
}

// IncomingStandardizedFeedback parses an incoming standardized feedback according to
// https://tools.ietf.org/wg/avtcore/draft-ietf-avtcore-cc-feedback-message/
// Current implementation implements -02 version and assumes that SR/RR or other
// non-CC feedback is stripped.
func (t *Tx) IncomingStandardizedFeedback(ts time.Time, buf []byte) {
	t.check()
	buffer := C.CBytes(buf)
	defer C.free(buffer)
	C.ScreamTxIncomingStdFeedbackBuf(t.screamTx, C.uint32_t(toNTP32(ts)), (*C.uchar)(buffer), C.int(len(buf)))
}

// GetTargetBitrate returns the target bitrate for the stream with ssrc.
// NOTE!, Because SCReAM operates on RTP packets, the target bitrate will
// also include the RTP overhead. This means that a subsequent call to set the
// media coder target bitrate must subtract an estimate of the RTP + framing
// overhead. This is not critical for Video bitrates but can be important
// when SCReAM is used to congestion control e.g low bitrate audio streams.
//
// Function returns -1 if a loss is detected, this signal can be used to
// request a new key frame from a video encoder.
func (t *Tx) GetTargetBitrate(ts time.Time, ssrc uint32) float64 {
	t.check()
	return float64(C.ScreamTxGetTargetBitrate(t.screamTx, C.uint32_t(toNTP32(ts)), C.uint32_t(ssrc)))
}

// GetStatistics returns some overall SCReAM statistics. The summary is
// timestamped with the elapsed time since t was created.
func (t *Tx) GetStatistics(ts time.Time) string {
	t.check()
	buffer := C.malloc(C.size_t(C.SCREAM_TX_STATISTICS_SIZE))
	defer C.free(buffer)
	C.ScreamTxGetStatistics(t.screamTx, C.float(ts.Sub(t.created).Seconds()), (*C.char)(buffer))
	return C.GoString((*C.char)(buffer))
}
