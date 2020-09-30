// Copyright (c) 2020 Mathis Engelbart All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scream

/*
#cgo CPPFLAGS: -Wno-overflow -Wno-write-strings
#include "ScreamRxC.h"
#include <stdlib.h>
*/
import "C"
import "unsafe"

type Rx struct {
	screamRx *C.ScreamRxC
}

func NewRx() *Rx {
	return &Rx{
		screamRx: C.ScreamRxInit(0),
	}
}

// Receive needs to be called when new packets are received
// TODO: Determine better type for rtpPacket, convert to Cpointer and pass it to C
func (r *Rx) Receive(timeNTP uint, rtpPacket interface{}, ssrc int, size int, seqNr int, ceBits uint8) {
	C.ScreamRxReceive(r.screamRx, C.uint(timeNTP), nil, C.uint(ssrc), C.int(size), C.uint(seqNr), C.uchar(ceBits))
}

func (r *Rx) IsFeedback(timeNTP uint) bool {
	return bool(C.ScreamRxIsFeedback(r.screamRx, C.uint(timeNTP)))
}

// CreateStandardizedFeedback creates a feedback packet according to
// https://tools.ietf.org/wg/avtcore/draft-ietf-avtcore-cc-feedback-message/
func (r *Rx) CreateStandardizedFeedback(timeNTP uint, isMark bool) (bool, []byte) {
	ret := C.ScreamRxGetFeedback(r.screamRx, C.uint(timeNTP), C.bool(isMark))
	defer C.free(unsafe.Pointer(ret))
	size := C.ScreamRxGetFeedbackSize(ret)
	buf := C.ScreamRxGetFeedbackBuffer(ret)
	bs := C.GoBytes(unsafe.Pointer(buf), size)
	return bool(C.ScreamRxGetFeedbackResult(ret)), bs
}
