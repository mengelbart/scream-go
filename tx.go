// Copyright (c) 2020 Mathis Engelbart All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scream

/*
#cgo CPPFLAGS: -Wno-overflow -Wno-write-strings
#include "ScreamTxC.h"
#include <stdlib.h>
*/
import "C"
import (
	"log"
	"sync"
	"unsafe"
)

type Tx struct {
	screamTx *C.ScreamTxC
}

func NewTx() *Tx {
	return &Tx{
		screamTx: C.ScreamTxInit(),
	}
}

func (t *Tx) RegisterNewStream(rtpQueue RTPQueue, ssrc uint, priority, minBitrate, startBitrate, maxBitrate float64) {
	id := nextRTPQueueID()
	rtpQueues[id] = rtpQueue
	rtpQueueC := C.RtpQueueIfaceInit(C.int(id))
	C.ScreamTxRegisterNewStream(t.screamTx, rtpQueueC, C.uint(ssrc), C.float(priority), C.float(minBitrate), C.float(startBitrate), C.float(maxBitrate))
}

func (t *Tx) NewMediaFrame(timeNTP, ssrc uint, bytesRTP int) {
	C.ScreamTxNewMediaFrame(t.screamTx, C.uint(timeNTP), C.uint(ssrc), C.int(bytesRTP))
}

func (t *Tx) IsOkToTransmit(timeNTP uint, ssrc uint) float64 {
	return float64(C.ScreamTxIsOkToTransmit(t.screamTx, C.uint(timeNTP), C.uint(ssrc)))
}

func (t *Tx) AddTransmitted(timeNTP uint, ssrc uint, size int, seqNr uint, isMark bool) float64 {
	return float64(C.ScreamTxAddTransmitted(t.screamTx, C.uint(timeNTP), C.uint(ssrc), C.int(size), C.uint(seqNr), C.bool(isMark)))
}

func (t *Tx) IncomingStandardizedFeedback(timeNTP uint, buf []byte) {
	c := make([]byte, len(buf))
	copy(c, buf)
	C.ScreamTxIncomingStdFeedback(t.screamTx, C.uint(timeNTP), unsafe.Pointer(&c[0]), C.int(len(c)))
}

func (t *Tx) IncomingFeedback(timeNTP uint, streamID int, timestamp uint, seqNr uint, ceBits byte, isLast bool) {
	C.ScreamTxIncomingFeedback(t.screamTx, C.uint(timeNTP), C.int(streamID), C.uint(timestamp), C.uint(seqNr), C.uchar(ceBits), C.bool(isLast))
}

func (t *Tx) GetTargetBitrate(ssrc uint) float64 {
	return float64(C.ScreamTxGetTargetBitrate(t.screamTx, C.uint(ssrc)))
}

func (t *Tx) GetStatistics(timeNTP uint) string {
	buf := C.ScreamTxGetStatistics(t.screamTx, C.uint(timeNTP))
	defer C.free(unsafe.Pointer(buf))
	return C.GoString(buf)
}

type RTPQueue interface {
	Clear()
	SizeOfNextRTP() int
	SeqNrOfNextRTP() int
	BytesInQueue() int
	SizeOfQueue() int
	GetDelay(float64) float64
	GetSizeOfLastFrame() int
}

var srcPipelinesLock sync.Mutex
var rtpQueues = map[int]RTPQueue{}
var nextPipelineID = 0

func nextRTPQueueID() int {
	defer func() {
		nextPipelineID++
	}()
	return nextPipelineID
}

//export goClear
func goClear(id C.int) {
	srcPipelinesLock.Lock()
	defer srcPipelinesLock.Unlock()
	log.Println("Clear queue")
	rtpQueues[int(id)].Clear()
}

//export goSizeOfNextRtp
func goSizeOfNextRtp(id C.int) C.int {
	srcPipelinesLock.Lock()
	defer srcPipelinesLock.Unlock()
	return C.int(rtpQueues[int(id)].SizeOfNextRTP())
}

//export goSeqNrOfNextRtp
func goSeqNrOfNextRtp(id C.int) C.int {
	srcPipelinesLock.Lock()
	defer srcPipelinesLock.Unlock()
	return C.int(rtpQueues[int(id)].SeqNrOfNextRTP())
}

//export goBytesInQueue
func goBytesInQueue(id C.int) C.int {
	srcPipelinesLock.Lock()
	defer srcPipelinesLock.Unlock()
	return C.int(rtpQueues[int(id)].BytesInQueue())
}

//export goSizeOfQueue
func goSizeOfQueue(id C.int) C.int {
	srcPipelinesLock.Lock()
	defer srcPipelinesLock.Unlock()
	return C.int(rtpQueues[int(id)].SizeOfQueue())
}

//export goGetDelay
func goGetDelay(id C.int, currTs C.float) C.float {
	srcPipelinesLock.Lock()
	defer srcPipelinesLock.Unlock()
	return C.float(rtpQueues[int(id)].GetDelay(float64(currTs)))
}

//export goGetSizeOfLastFrame
func goGetSizeOfLastFrame(id C.int) C.int {
	srcPipelinesLock.Lock()
	defer srcPipelinesLock.Unlock()
	return C.int(rtpQueues[int(id)].GetSizeOfLastFrame())
}
