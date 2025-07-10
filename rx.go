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

// Rx implements the receiver side of SCReAM
type Rx struct {
	screamRx *C.ScreamRx
}

// NewRx creates a new Rx instance. One Rx is created for each source SSRC
func NewRx(ssrc uint32) *Rx {
	return &Rx{
		screamRx: C.ScreamRxInit(C.uint(ssrc)),
	}
}

// Receive needs to be called each time an RTP packet is received
func (r *Rx) Receive(ntpTime uint32, ssrc uint32, size int, seqNr uint16, ceBits uint8, isMark bool) {
	// currently sets timestamp to 0 since it is not used in ScreamRx.cpp
	C.ScreamRxReceive(
		r.screamRx,
		C.uint32_t(ntpTime),
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
func (r *Rx) IsFeedback(ntpTime uint32) bool {
	return bool(C.ScreamRxIsFeedback(r.screamRx, C.uint32_t(ntpTime)))
}

// CreateStandardizedFeedback creates a feedback packet according to
// https://tools.ietf.org/wg/avtcore/draft-ietf-avtcore-cc-feedback-message/
// Current implementation implements -02 version
// It is up to the wrapper application to prepend this RTCP with SR or RR when needed
func (r *Rx) CreateStandardizedFeedback(ntpTime uint32, isMark bool) ([]byte, bool) {

	buf := make([]byte, 2048)
	cbuf := C.CBytes(buf)
	defer C.free(cbuf)
	var size C.int

	ret := C.ScreamRxGetFeedback(r.screamRx, C.uint32_t(ntpTime), C.bool(isMark), (*C.uchar)(cbuf), &size)

	result := C.GoBytes(cbuf, size)
	return result, bool(ret)
}
