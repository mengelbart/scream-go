// Copyright (c) 2020 Mathis Engelbart All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scream

/*
#cgo CPPFLAGS: -Wno-overflow -Wno-write-strings
#include "ScreamRxC.h"
*/
import "C"

type Rx struct {
	screamRx *C.ScreamRxC
}

func NewRx() *Rx {
	return &Rx{
		screamRx: C.ScreamRxInit(0),
	}
}

// TODO: Determine better type for rtpPacket, convert to Cpointer and pass it to C
func (r *Rx) receive(timeNTP uint, rtpPacket interface{}, ssrc int, size int, seqNr int, ceBits uint8) {
	C.ScreamRxReceive(r.screamRx, C.uint(timeNTP), nil, C.uint(ssrc), C.int(size), C.uint(seqNr), C.uchar(ceBits))
}

func (r *Rx) isFeedback(timeNTP uint) bool {
	return bool(C.ScreamRxIsFeedback(r.screamRx, C.uint(timeNTP)))
}

// TODO: replace buffer by real memory pointer
func (r *Rx) createStandardizedFeedback(timeNTP uint, isMark bool, buf uint8, size int) bool {
	return bool(C.ScreamRxGetFeedback(r.screamRx, C.uint(timeNTP), C.bool(isMark), nil, C.int(size)))
}
