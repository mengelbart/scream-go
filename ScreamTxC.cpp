// Copyright (c) 2020 Mathis Engelbart All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "ScreamTxC.h"

#include <string.h>

#include "ScreamTx.h"

void ScreamTxDefaultConfig(ScreamTxConfig* c) {
  // Mirrors the ScreamV2Tx constructor defaults in ScreamTx.h.
  c->lossBeta = kLossBeta;
  c->ecnCeBeta = kEcnCeBeta;
  c->queueDelayTargetMin = kQueueDelayTargetMin;
  c->cwnd = 0;
  c->packetPacingHeadroom = kPacketPacingHeadRoom;
  c->maxAdaptivePacingRateScale = kMaxAdaptivePacingRateScale;
  c->bytesInFlightHeadRoom = kBytesInFlightHeadRoom;
  c->multiplicativeIncreaseScalefactor = kMultiplicativeIncreaseScalefactor;
  c->isL4s = false;
  c->maxWindowHeadroom = 3.0f;
  c->enableSbd = kEnableSbd;
  c->enableClockDriftCompensation = false;
}

ScreamV2Tx* ScreamTxInit(const ScreamTxConfig* c) {
  return new ScreamV2Tx(c->lossBeta, c->ecnCeBeta, c->queueDelayTargetMin,
                        c->cwnd, c->packetPacingHeadroom,
                        c->maxAdaptivePacingRateScale, c->bytesInFlightHeadRoom,
                        c->multiplicativeIncreaseScalefactor, c->isL4s,
                        c->maxWindowHeadroom, c->enableSbd,
                        c->enableClockDriftCompensation);
}

void ScreamTxFree(ScreamV2Tx* s) {
  delete s;
}

int ScreamTxMaxStreams() {
  return kMaxStreams;
}

float ScreamTxDefaultMaxRtpQueueDelay() {
  return kMaxRtpQueueDelay;
}

int ScreamTxMaxMssListSize() {
  return kMssListSize;
}

void ScreamTxRegisterNewStream(ScreamV2Tx* s,
                               RtpQueueC* rtpQueue,
                               uint32_t ssrc,
                               float priority,
                               float minBitrate,
                               float startBitrate,
                               float maxBitrate,
                               float maxRtpQueueDelay,
                               bool isAdaptiveTargetRateScale,
                               float hysteresis,
                               bool enableFrameSizeOverhead) {
  RtpQueueIface* rtpq = (RtpQueueIface*)rtpQueue;

  s->registerNewStream(rtpq, ssrc, priority, minBitrate, startBitrate,
                       maxBitrate, maxRtpQueueDelay, isAdaptiveTargetRateScale,
                       hysteresis, enableFrameSizeOverhead);
}

void ScreamTxNewMediaFrame(ScreamV2Tx* s,
                           uint32_t time_ntp,
                           uint32_t ssrc,
                           int bytesRtp,
                           bool isMarker) {
  s->newMediaFrame(time_ntp, ssrc, bytesRtp, isMarker);
}

float ScreamTxIsOkToTransmit(ScreamV2Tx* s, uint32_t time_ntp, uint32_t* ssrc) {
  return s->isOkToTransmit(time_ntp, *ssrc);
}

float ScreamTxAddTransmitted(ScreamV2Tx* s,
                             uint32_t time_ntp,
                             uint32_t ssrc,
                             int size,
                             uint16_t seqNr,
                             bool isMark) {
  return s->addTransmitted(time_ntp, ssrc, size, seqNr, isMark);
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

void ScreamTxEnablePacketPacing(ScreamV2Tx* s, bool enable) {
  s->enablePacketPacing(enable);
}

void ScreamTxEnableRelaxedPacing(ScreamV2Tx* s, bool enable) {
  s->enableRelaxedPacing(enable);
}

void ScreamTxSetEnableCyclicPacing(ScreamV2Tx* s, bool enable) {
  s->setEnableCyclicPacing(enable);
}

void ScreamTxSetCwndMinLow(ScreamV2Tx* s, int value) {
  s->setCwndMinLow(value);
}

void ScreamTxAutoTuneMinCwnd(ScreamV2Tx* s, bool enable) {
  s->autoTuneMinCwnd(enable);
}

void ScreamTxSetEnableAdaptiveWindowHeadroom(ScreamV2Tx* s, bool enable) {
  s->isEnableAdaptiveWindowHeadroom(enable);
}

void ScreamTxSetEnableRatePolicerProtection(ScreamV2Tx* s, bool enable) {
  s->isEnableRatePolicerProtection(enable);
}

void ScreamTxSetSchedulingJitterMargin(ScreamV2Tx* s, float value) {
  s->setSchedulingJitterMargin(value);
}

void ScreamTxSetPostCongestionDelayRtts(ScreamV2Tx* s, int value) {
  s->setPostCongestionDelayRtts(value);
}

void ScreamTxSetReorderTime(ScreamV2Tx* s, float value) {
  s->setReorderTime(value);
}

void ScreamTxSetEnableRateUpdate(ScreamV2Tx* s, bool enable) {
  s->setEnableRateUpdate(enable);
}

void ScreamTxSetLogTag(ScreamV2Tx* s, char* logTag) {
  s->setLogTag(logTag);
}

void ScreamTxSetMaxTotalBitrate(ScreamV2Tx* s, float bitrate) {
  s->setMaxTotalBitrate(bitrate);
}

bool ScreamTxSetMssListMinPacketsInFlight(ScreamV2Tx* s,
                                          int* mssList,
                                          int nMssListItems,
                                          int minPacketsInFlight) {
  if (nMssListItems < 0 || nMssListItems > kMssListSize) {
    return false;
  }
  s->setMssListMinPacketsInFlight(mssList, nMssListItems, minPacketsInFlight);
  return true;
}

void ScreamTxUpdateBitrateStream(ScreamV2Tx* s,
                                 uint32_t ssrc,
                                 float minBitrate,
                                 float maxBitrate) {
  s->updateBitrateStream(ssrc, minBitrate, maxBitrate);
}

void ScreamTxSetTargetPriority(ScreamV2Tx* s, uint32_t ssrc, float priority) {
  s->setTargetPriority(ssrc, priority);
}

void ScreamTxGetStatistics(ScreamV2Tx* s, float time, char* result) {
  result[0] = '\0';
  s->getStatistics(time, result);
}
