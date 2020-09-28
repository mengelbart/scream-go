package scream

/*
#cgo CPPFLAGS: -Wno-overflow -Wno-write-strings
#include "ScreamTxC.h"
*/
import "C"
import "fmt"

//export goClear
func goClear(id C.int) {
	fmt.Println(id)
}

//export goSizeOfNextRtp
func goSizeOfNextRtp(id C.int) C.int {
	return 0
}

//export goSeqNrOfNextRtp
func goSeqNrOfNextRtp(id C.int) C.int {
	return 0
}

//export goBytesInQueue
func goBytesInQueue(id C.int) C.int {
	return 0
}

//export goSizeOfQueue
func goSizeOfQueue(id C.int) C.int {
	return 0
}

//export goGetDelay
func goGetDelay(id C.int, currTs C.float) C.float {
	return 0
}

//export goGetSizeOfLastFrame
func goGetSizeOfLastFrame(id C.int) C.int {
	return 0
}

type Tx struct {
	screamTx *C.ScreamTxC
}

type RTPQueue interface {
	Clear(int)
	SizeOfNextRTP()
	SeqNrOfNextRTP()
	BytesInQueue()
	SizeOfQueue()
	GetDelay(float64)
	SetSizeOfLastFrame()
}

func New() *Tx {
	return &Tx{
		screamTx: C.ScreamTxInit(),
	}
}

var rtpQueues = map[int]RTPQueue{}

func nextRTPQueueID() int {
	return 0
}

// TODO: Implement RtpQueueInterface handling
func (t *Tx) registerNewStream(rtpQueue RTPQueue, ssrc uint, priority, minBitrate, startBitrate, maxBitrate float64) {
	id := nextRTPQueueID()
	rtpQueues[id] = rtpQueue
	rtpQueueC := C.RtpQueueIfaceInit(C.int(id))
	C.ScreamTxRegisterNewStream(t.screamTx, rtpQueueC, C.uint(ssrc), C.float(priority), C.float(minBitrate), C.float(startBitrate), C.float(maxBitrate))
}

func (t *Tx) newMediaFrame(timeNTP, ssrc uint, bytesRTP int) {
	C.ScreamTxNewMediaFrame(t.screamTx, C.uint(timeNTP), C.uint(ssrc), C.int(bytesRTP))
}

func (t *Tx) isOkToTransmit(timeNTP uint, ssrc uint) float64 {
	return float64(C.ScreamTxIsOkToTransmit(t.screamTx, C.uint(timeNTP), C.uint(ssrc)))
}

func (t *Tx) addTransmitted(timeNTP uint, ssrc uint, size int, seqNr uint, isMark bool) float64 {
	return float64(C.ScreamTxAddTransmitted(t.screamTx, C.uint(timeNTP), C.uint(ssrc), C.int(size), C.uint(seqNr), C.bool(isMark)))
}

func (t *Tx) incomingStandardizedFeedback(timeNTP uint, streamID int, timestamp uint, seqNr uint, ceBits byte, isLast bool) {
	C.ScreamTxIncomingStdFeedback(t.screamTx, C.uint(timeNTP), C.int(streamID), C.uint(timestamp), C.uint(seqNr), C.uchar(ceBits), C.bool(isLast))
}

func (t *Tx) getTargetBitrate(ssrc uint) float64 {
	return float64(C.ScreamTxGetTargetBitrate(t.screamTx, C.uint(ssrc)))
}
