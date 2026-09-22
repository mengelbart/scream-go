// Copyright (c) 2020 Mathis Engelbart All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scream

/*
#include "ScreamRxC.h"

#include <stdlib.h>
#include <stdint.h>
*/
import "C"
import "time"

// Rx implements the receiver side of SCReAM.
// An Rx is not safe for concurrent use.
type Rx struct {
	screamRx *C.ScreamRx
}

// NewRx creates a new Rx instance. One Rx is created for each source SSRC.
// Close must be called to free the resources held by the returned Rx.
func NewRx(ssrc uint32) *Rx {
	return &Rx{
		screamRx: C.ScreamRxInit(C.uint32_t(ssrc)),
	}
}

func (r *Rx) check() {
	if r.screamRx == nil {
		panic("scream: Rx used after Close")
	}
}

// Close closes r and frees all resources. Any other method called after Close
// panics. Close is idempotent.
func (r *Rx) Close() error {
	if r.screamRx == nil {
		return nil
	}
	C.ScreamRxFree(r.screamRx)
	r.screamRx = nil
	return nil
}

// Receive needs to be called each time an RTP packet is received
func (r *Rx) Receive(ts time.Time, ssrc uint32, size int, seqNr uint16, ceBits uint8, isMark bool) {
	r.check()
	// The RTP packet and its timestamp are passed as nil/0, ScreamRx only
	// stores them and never reads them back.
	C.ScreamRxReceive(
		r.screamRx,
		C.uint32_t(toNTP32(ts)),
		nil,
		C.uint32_t(ssrc),
		C.int(size),
		C.uint16_t(seqNr),
		C.uint8_t(ceBits),
		C.bool(isMark),
		C.uint32_t(0),
	)
}

// IsFeedback returns TRUE if an RTP packet has been received and there is pending feedback
func (r *Rx) IsFeedback(ts time.Time) bool {
	r.check()
	return bool(C.ScreamRxIsFeedback(r.screamRx, C.uint32_t(toNTP32(ts))))
}

// CreateStandardizedFeedback creates a feedback packet according to
// https://tools.ietf.org/wg/avtcore/draft-ietf-avtcore-cc-feedback-message/
// Current implementation implements -02 version
// It is up to the wrapper application to prepend this RTCP with SR or RR when needed
// It returns nil if there is no feedback to send.
func (r *Rx) CreateStandardizedFeedback(ts time.Time, isMark bool) ([]byte, bool) {
	r.check()

	cbuf := C.malloc(C.size_t(C.SCREAM_RX_MAX_FEEDBACK_SIZE))
	defer C.free(cbuf)
	var size C.int

	ret := C.ScreamRxGetFeedback(
		r.screamRx,
		C.uint32_t(toNTP32(ts)),
		C.bool(isMark),
		(*C.uchar)(cbuf),
		C.int(C.SCREAM_RX_MAX_FEEDBACK_SIZE),
		&size,
	)
	if !bool(ret) {
		return nil, false
	}
	return C.GoBytes(cbuf, size), true
}
