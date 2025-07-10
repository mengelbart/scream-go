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
	"runtime"
	"runtime/cgo"
	"time"
	"unsafe"
)

// Tx implements the sender side of SCReAM
type Tx struct {
	screamTx *C.ScreamV2Tx
	pinner   *runtime.Pinner
	queues   []*cgo.Handle
}

// NewTx creates a new Tx instance.
func NewTx() *Tx {
	return &Tx{
		screamTx: C.ScreamTxInit(),
		pinner:   &runtime.Pinner{},
	}
}

// RegisterNewStream registers a new stream with ssrc using rtpQueue.
// Priority is in the range ]0.0..1.0] where 1.0 denotes the highest priority.
// It is recommended that at least one stream has priority 1.0.
// Bitrates are specified in bps
func (t *Tx) RegisterNewStream(rtpQueue RTPQueue, ssrc uint32, priority, minBitrate, startBitrate, maxBitrate float64) {
	h := cgo.NewHandle(rtpQueue)
	t.queues = append(t.queues, &h)
	t.pinner.Pin(&h)
	rtpQueueC := C.RtpQueueCInit(unsafe.Pointer(&h))
	C.ScreamTxRegisterNewStream(t.screamTx, rtpQueueC, C.uint(ssrc), C.float(priority), C.float(minBitrate), C.float(startBitrate), C.float(maxBitrate))
}

// Close closes t and frees all resources.
func (t *Tx) Close() error {
	t.pinner.Unpin()
	for _, q := range t.queues {
		q.Delete()
	}
	return nil
}

// NewMediaFrame should be called for each new video frame.
// IsOkToTransmit should be called after newMediaFrame
func (t *Tx) NewMediaFrame(ts time.Time, ssrc uint32, bytesRTP int, isMarker bool) {
	C.ScreamTxNewMediaFrame(t.screamTx, C.uint32_t(toNTP32(ts)), C.uint32_t(ssrc), C.int(bytesRTP), C.bool(isMarker))
}

// IsOkToTransmit determines if an RTP packet with ssrc can be transmitted.
// Returns:
//
// 0.0: RTP packet with ssrc can be immediately transmitted. AddTransmitted must be
// called if packet is transmitted as a result of this.
//
// >0.0: Time [s] until this function should be called again. This can be used to
// start a timer.
// Note that a call to NewMediaFrame or IncomingFeedback should cause an immediate
// call to isOkToTransmit.
//
// -1.0: No RTP packet available to transmit or send window is not large enough
func (t *Tx) IsOkToTransmit(ts time.Time, ssrc uint32) float64 {
	return float64(C.ScreamTxIsOkToTransmit(t.screamTx, C.uint(toNTP32(ts)), C.uint(ssrc)))
}

// AddTransmitted adds a packet to list of transmitted packets. Should be called when an
// RTP packet was transmitted. AddTransmitted returns the time until IsOkToTransmit can be
// called again.
func (t *Tx) AddTransmitted(ts time.Time, ssrc uint32, size int, seqNr uint16, isMark bool) float64 {
	return float64(C.ScreamTxAddTransmitted(t.screamTx, C.uint32_t(toNTP32(ts)), C.uint32_t(ssrc), C.int(size), C.uint16_t(seqNr), C.bool(isMark)))
}

// IncomingStandardizedFeedback parses an incoming standardized feedback according to
// https://tools.ietf.org/wg/avtcore/draft-ietf-avtcore-cc-feedback-message/
// Current implementation implements -02 version and assumes that SR/RR or other
// non-CC feedback is stripped.
func (t *Tx) IncomingStandardizedFeedback(ts time.Time, buf []byte) {
	buffer := C.CBytes(buf)
	defer C.free(buffer)
	C.ScreamTxIncomingStdFeedbackBuf(t.screamTx, C.uint(toNTP32(ts)), (*C.uchar)(buffer), C.int(len(buf)))
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
	return float64(C.ScreamTxGetTargetBitrate(t.screamTx, C.uint32_t(toNTP32(ts)), C.uint32_t(ssrc)))
}

// GetStatistics returns some overall SCReAM statistics.
func (t *Tx) GetStatistics(ntpTime time.Time) string {
	buffer := C.malloc(C.size_t(1024))
	defer C.free(buffer)
	tsf := float64(toNTP(ntpTime)) / 65536.0
	C.ScreamTxGetStatistics(t.screamTx, C.float(tsf), (*C.char)(buffer))
	return C.GoString((*C.char)(buffer))
}
