// Copyright (c) Mathis Engelbart. All rights reserved.
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
	"encoding/binary"
	"errors"
	"fmt"
	"runtime/cgo"
	"time"
	"unsafe"
)

var (
	// ErrDuplicateSSRC is returned when a stream is registered with an ssrc
	// that is already in use. SCReAM resolves a stream by its ssrc and would
	// never reach the second registration.
	ErrDuplicateSSRC = errors.New("ssrc already registered")

	// ErrTooManyStreams is returned when more streams are registered than
	// SCReAM can hold.
	ErrTooManyStreams = errors.New("too many streams")

	// ErrMssListTooLong is returned when the MSS list is longer than SCReAM
	// can hold.
	ErrMssListTooLong = errors.New("mss list too long")

	// ErrMalformedFeedback is returned when a congestion control feedback
	// packet is not laid out the way SCReAM's parser assumes.
	ErrMalformedFeedback = errors.New("malformed congestion control feedback")
)

// Layout of an RFC 8888 congestion control feedback packet.
const (
	// feedbackHeaderLen is the RTCP header plus the sender SSRC.
	feedbackHeaderLen = 8
	// feedbackBlockHeaderLen is a report block's SSRC, begin_seq and
	// num_reports.
	feedbackBlockHeaderLen = 8
	// feedbackTimestampLen is the report timestamp trailing the packet.
	feedbackTimestampLen = 4
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
	// logTag is kept alive because SCReAM stores the pointer, see WithLogTag.
	logTag *C.char
}

// NewTx creates a new Tx instance. Close must be called to free the resources
// held by the returned Tx.
func NewTx(opts ...TxOption) *Tx {
	cfg := defaultTxConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	t := &Tx{
		screamTx: C.ScreamTxInit(&cfg.c),
		created:  time.Now(),
	}
	C.ScreamTxEnablePacketPacing(t.screamTx, C.bool(cfg.packetPacing))
	for _, apply := range cfg.post {
		apply(t)
	}
	return t
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
func (t *Tx) RegisterNewStream(rtpQueue RTPQueue, ssrc uint32, priority, minBitrate, startBitrate, maxBitrate float64, opts ...StreamOption) error {
	t.check()
	for _, s := range t.streams {
		if s.ssrc == ssrc {
			return fmt.Errorf("%w: %v", ErrDuplicateSSRC, ssrc)
		}
	}
	if len(t.streams) >= int(C.ScreamTxMaxStreams()) {
		return fmt.Errorf("%w: limit is %v", ErrTooManyStreams, C.ScreamTxMaxStreams())
	}

	cfg := defaultStreamConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	h := cgo.NewHandle(rtpQueue)
	rtpQueueC := C.RtpQueueCInit(C.uintptr_t(h))
	t.streams = append(t.streams, stream{ssrc: ssrc, handle: h, rtpQueue: rtpQueueC})
	C.ScreamTxRegisterNewStream(
		t.screamTx,
		rtpQueueC,
		C.uint32_t(ssrc),
		C.float(priority),
		C.float(minBitrate),
		C.float(startBitrate),
		C.float(maxBitrate),
		cfg.maxRTPQueueDelay,
		cfg.adaptiveTargetRateScale,
		cfg.hysteresis,
		cfg.frameSizeOverhead,
	)
	return nil
}

// UpdateBitrateStream updates the min and max bitrate of an already registered
// stream. Bitrates are specified in bps.
func (t *Tx) UpdateBitrateStream(ssrc uint32, minBitrate, maxBitrate float64) {
	t.check()
	C.ScreamTxUpdateBitrateStream(t.screamTx, C.uint32_t(ssrc), C.float(minBitrate), C.float(maxBitrate))
}

// SetTargetPriority sets the priority of an already registered stream.
// Priority is in the range ]0.0..1.0] where 1.0 denotes the highest priority.
func (t *Tx) SetTargetPriority(ssrc uint32, priority float64) {
	t.check()
	C.ScreamTxSetTargetPriority(t.screamTx, C.uint32_t(ssrc), C.float(priority))
}

// SetMaxTotalBitrate sets the maximum total bitrate in bps across all streams.
func (t *Tx) SetMaxTotalBitrate(bitrate float64) {
	t.check()
	C.ScreamTxSetMaxTotalBitrate(t.screamTx, C.float(bitrate))
}

// SetMssListMinPacketsInFlight sets the list of MSS values SCReAM picks a
// recommended MSS from, and the minimum number of packets in flight.
// It returns an error if mssList is longer than SCReAM accepts.
func (t *Tx) SetMssListMinPacketsInFlight(mssList []int, minPacketsInFlight int) error {
	t.check()
	if max := int(C.ScreamTxMaxMssListSize()); len(mssList) > max {
		return fmt.Errorf("%w: limit is %v", ErrMssListTooLong, max)
	}
	if len(mssList) == 0 {
		return nil
	}

	list := make([]C.int, len(mssList))
	for i, mss := range mssList {
		list[i] = C.int(mss)
	}
	if !bool(C.ScreamTxSetMssListMinPacketsInFlight(t.screamTx, &list[0], C.int(len(list)), C.int(minPacketsInFlight))) {
		return ErrMssListTooLong
	}
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

	if t.logTag != nil {
		C.free(unsafe.Pointer(t.logTag))
		t.logTag = nil
	}
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
//
// The buffer is validated and report blocks SCReAM cannot parse are stripped
// before it reaches C, see sanitizeStandardizedFeedback. A buffer that fails
// validation is dropped and ErrMalformedFeedback is returned.
func (t *Tx) IncomingStandardizedFeedback(ts time.Time, buf []byte) error {
	t.check()
	sanitized, err := sanitizeStandardizedFeedback(buf)
	if err != nil {
		return err
	}
	if sanitized == nil {
		return nil
	}
	buffer := C.CBytes(sanitized)
	defer C.free(buffer)
	C.ScreamTxIncomingStdFeedbackBuf(t.screamTx, C.uint32_t(toNTP32(ts)), (*C.uchar)(buffer), C.int(len(sanitized)))
	return nil
}

// sanitizeStandardizedFeedback validates an RFC 8888 congestion control
// feedback packet and returns a buffer that SCReAM's parser can walk, or nil if
// nothing is left to report.
//
// SCReAM reads the packet without any bounds checking: it derives the metric
// block count as num_reports-1 in a uint16, so an empty report block underflows
// it to 65535 and reads 128 KiB past the buffer, and it ends the block loop on
// ptr == size-4, so a block that does not land exactly there loops off the end.
// Both are reachable from a conforming remote, so the wrapper has to reject
// what SCReAM would misparse and drop empty blocks, which RFC 8888 §3.1 allows
// a receiver to send for a source it saw no packets from.
func sanitizeStandardizedFeedback(buf []byte) ([]byte, error) {
	// header + sender ssrc + one report block + report timestamp
	if len(buf) < feedbackHeaderLen+feedbackBlockHeaderLen+feedbackTimestampLen || len(buf)%4 != 0 {
		return nil, fmt.Errorf("%w: length %d", ErrMalformedFeedback, len(buf))
	}
	// SCReAM locates the report timestamp at length*4, which must be the last
	// word of the packet.
	if int(binary.BigEndian.Uint16(buf[2:4]))*4 != len(buf)-feedbackTimestampLen {
		return nil, fmt.Errorf("%w: length field %d does not match %d octets",
			ErrMalformedFeedback, binary.BigEndian.Uint16(buf[2:4]), len(buf))
	}

	end := len(buf) - feedbackTimestampLen
	kept := make([]byte, feedbackHeaderLen, len(buf))
	copy(kept, buf[:feedbackHeaderLen])
	dropped := false
	for ptr := feedbackHeaderLen; ptr != end; {
		if ptr+feedbackBlockHeaderLen > end {
			return nil, fmt.Errorf("%w: truncated report block at octet %d", ErrMalformedFeedback, ptr)
		}
		numReports := int(binary.BigEndian.Uint16(buf[ptr+6 : ptr+8]))
		blockLen := feedbackBlockHeaderLen + 2*numReports
		if numReports%2 == 1 {
			// odd number of metric blocks is zero padded to a whole word
			blockLen += 2
		}
		if ptr+blockLen > end {
			return nil, fmt.Errorf("%w: report block at octet %d claims %d reports",
				ErrMalformedFeedback, ptr, numReports)
		}
		if numReports == 0 {
			dropped = true
		} else {
			kept = append(kept, buf[ptr:ptr+blockLen]...)
		}
		ptr += blockLen
	}
	if !dropped {
		return buf, nil
	}
	if len(kept) == feedbackHeaderLen {
		return nil, nil
	}
	kept = append(kept, buf[end:]...)
	binary.BigEndian.PutUint16(kept[2:4], uint16(len(kept)/4-1))
	return kept, nil
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
