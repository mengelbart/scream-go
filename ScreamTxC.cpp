// Copyright (c) 2020 Mathis Engelbart All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "ScreamTxC.h"

#include "include/ScreamTx.h"

ScreamV2Tx* ScreamTxInit(bool isL4s) {
  auto s = new ScreamV2Tx(
    0.7,
    0.7,
    0.06f,
    100 * 1000 / 8,
    1.5f,
    1.5f,
    2.0f,
    0.05f,
    isL4s,
    5.0f,
    false,
    false
  );
  s->enablePacketPacing(false);

  int mtu = 1200;
	s->setCwndMinLow((mtu+12)*2);
	s->enableRelaxedPacing(false);


  int mtuList[10];
  int nMtuListItems = 1;
  int minPktsInFlight = 0;
	s->setMssListMinPacketsInFlight(mtuList, nMtuListItems, minPktsInFlight);
  return s;
}

void ScreamTxFree(ScreamV2Tx* s) {
  delete s;
}

void ScreamTxRegisterNewStream(ScreamV2Tx* s,
                               RtpQueueC* rtpQueue,
                               uint32_t ssrc,
                               float priority,
                               float minBitrate,
                               float startBitrate,
                               float maxBitrate,
                               float maxRtpQueueDelay) {
  ScreamV2Tx* stx = (ScreamV2Tx*)s;
  RtpQueueIface* rtpq = (RtpQueueIface*)rtpQueue;

  stx->registerNewStream(rtpq, ssrc, priority, minBitrate, startBitrate,
                         maxBitrate, maxRtpQueueDelay);
}

void ScreamTxNewMediaFrame(ScreamV2Tx* s,
                           uint32_t time_ntp,
                           uint32_t ssrc,
                           int bytesRtp,
                           bool isMarker) {
  s->newMediaFrame(time_ntp, ssrc, bytesRtp, isMarker);
}

float ScreamTxIsOkToTransmit(ScreamV2Tx* s, uint32_t time_ntp, uint32_t ssrc) {
  return s->isOkToTransmit(time_ntp, ssrc);
}

float ScreamTxAddTransmitted(ScreamV2Tx* s,
                             uint32_t time_ntp,
                             uint32_t ssrc,
                             int size,
                             uint16_t seqNr,
                             bool isMark) {
  ScreamV2Tx* stx = (ScreamV2Tx*)s;
  return stx->addTransmitted(time_ntp, ssrc, size, seqNr, isMark);
}

void ScreamTxIncomingStdFeedbackBuf(ScreamV2Tx* s,
                                    uint32_t time_ntp,
                                    unsigned char* buf,
                                    int size) {
  s->incomingStandardizedFeedback(time_ntp, buf, size);
}

void ScreamTxIncomingStdFeedback(ScreamV2Tx* s,
                                 uint32_t time_ntp,
                                 int streamId,
                                 uint32_t timestamp,
                                 uint16_t seqNr,
                                 uint8_t ceBits,
                                 bool isLast) {
  s->incomingStandardizedFeedback(time_ntp, streamId, timestamp, seqNr, ceBits,
                                  isLast);
}

float ScreamTxGetTargetBitrate(ScreamV2Tx* s,
                               uint32_t time_ntp,
                               uint32_t ssrc) {
  return s->getTargetBitrate(time_ntp, ssrc);
}

void ScreamTxGetStatistics(ScreamV2Tx* s, float time_ntp, char* result) {
  s->getStatistics(time_ntp, result);
}
