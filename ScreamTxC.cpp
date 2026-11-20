// Copyright (c) Mathis Engelbart. All rights reserved.
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

ScreamTxWrapper* ScreamTxInit(const ScreamTxConfig* c) {
  ScreamTxWrapper* wrapper = new ScreamTxWrapper();

  wrapper->scream = new ScreamV2Tx(c->lossBeta, c->ecnCeBeta, c->queueDelayTargetMin,
                        c->cwnd, c->packetPacingHeadroom,
                        c->maxAdaptivePacingRateScale, c->bytesInFlightHeadRoom,
                        c->multiplicativeIncreaseScalefactor, c->isL4s,
                        c->maxWindowHeadroom, c->enableSbd,
                        c->enableClockDriftCompensation);

  return wrapper;
}

void ScreamTxFree(ScreamTxWrapper* s) {
  if (s->logFile != nullptr) {
    s->scream->setDetailedLogFp(nullptr);
    fclose(s->logFile);
    s->logFile = nullptr;
  }

  delete s->scream;
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

void ScreamTxRegisterNewStream(ScreamTxWrapper* s,
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

  s->scream->registerNewStream(rtpq, ssrc, priority, minBitrate, startBitrate,
                       maxBitrate, maxRtpQueueDelay, isAdaptiveTargetRateScale,
                       hysteresis, enableFrameSizeOverhead);
}

void ScreamTxNewMediaFrame(ScreamTxWrapper* s,
                           uint32_t time_ntp,
                           uint32_t ssrc,
                           int bytesRtp,
                           bool isMarker) {
  s->scream->newMediaFrame(time_ntp, ssrc, bytesRtp, isMarker);
}

float ScreamTxIsOkToTransmit(ScreamTxWrapper* s, uint32_t time_ntp, uint32_t* ssrc) {
  return s->scream->isOkToTransmit(time_ntp, *ssrc);
}

float ScreamTxAddTransmitted(ScreamTxWrapper* s,
                             uint32_t time_ntp,
                             uint32_t ssrc,
                             int size,
                             uint16_t seqNr,
                             bool isMark) {
  return s->scream->addTransmitted(time_ntp, ssrc, size, seqNr, isMark);
}

void ScreamTxIncomingStdFeedbackBuf(ScreamTxWrapper* s,
                                    uint32_t time_ntp,
                                    unsigned char* buf,
                                    int size) {
  char str[100];
  sprintf(str, "%1.4f", time_ntp / 65536.0f);
  s->scream->setTimeString(str);
  s->scream->incomingStandardizedFeedback(time_ntp, buf, size);
}

void ScreamTxIncomingStdFeedback(ScreamTxWrapper* s,
                                 uint32_t time_ntp,
                                 int streamId,
                                 uint32_t timestamp,
                                 uint16_t seqNr,
                                 uint8_t ceBits,
                                 bool isLast) {
  char str[100];
  sprintf(str, "%1.4f", time_ntp / 65536.0f);
  s->scream->setTimeString(str);
  s->scream->incomingStandardizedFeedback(time_ntp, streamId, timestamp, seqNr, ceBits,
                                  isLast);
}

float ScreamTxGetTargetBitrate(ScreamTxWrapper* s,
                               uint32_t time_ntp,
                               uint32_t ssrc) {
  return s->scream->getTargetBitrate(time_ntp, ssrc);
}

void ScreamTxEnablePacketPacing(ScreamTxWrapper* s, bool enable) {
  s->scream->enablePacketPacing(enable);
}

void ScreamTxEnableRelaxedPacing(ScreamTxWrapper* s, bool enable) {
  s->scream->enableRelaxedPacing(enable);
}

void ScreamTxSetEnableCyclicPacing(ScreamTxWrapper* s, bool enable) {
  s->scream->setEnableCyclicPacing(enable);
}

void ScreamTxSetCwndMin(ScreamTxWrapper* s, int value) {
  s->scream->setCwndMin(value);
}

void ScreamTxAutoTuneMinCwnd(ScreamTxWrapper* s, bool enable) {
  s->scream->autoTuneMinCwnd(enable);
}

void ScreamTxSetDetailedLog(ScreamTxWrapper* s, char* filename) {
  if (s->logFile != nullptr) {
    s->scream->setDetailedLogFp(nullptr);
    fclose(s->logFile);
    s->logFile = nullptr;
  }

  if (filename != nullptr) {
    s->logFile = fopen(filename, "w");
    if (s->logFile != nullptr) {
      s->scream->setDetailedLogFp(s->logFile);
    }
  }
}

void ScreamTxSetEnableRatePolicerProtection(ScreamTxWrapper* s, bool enable) {
  s->scream->isEnableRatePolicerProtection(enable);
}

void ScreamTxSetExpFeature(ScreamTxWrapper* s, int value) {
  s->scream->setExpFeature(value);
}

void ScreamTxSetPostCongestionDelayRtts(ScreamTxWrapper* s, int value) {
  s->scream->setPostCongestionDelayRtts(value);
}

void ScreamTxSetReorderTime(ScreamTxWrapper* s, float value) {
  s->scream->setReorderTime(value);
}

void ScreamTxSetEnableRateUpdate(ScreamTxWrapper* s, bool enable) {
  s->scream->setEnableRateUpdate(enable);
}

void ScreamTxSetLogTag(ScreamTxWrapper* s, char* logTag) {
  s->scream->setLogTag(logTag);
}

void ScreamTxSetMaxTotalBitrate(ScreamTxWrapper* s, float bitrate) {
  s->scream->setMaxTotalBitrate(bitrate);
}

bool ScreamTxSetMssListMinPacketsInFlight(ScreamTxWrapper* s,
                                          int* mssList,
                                          int nMssListItems,
                                          int minPacketsInFlight) {
  if (nMssListItems < 0 || nMssListItems > kMssListSize) {
    return false;
  }
  s->scream->setMssListMinPacketsInFlight(mssList, nMssListItems, minPacketsInFlight);
  return true;
}

void ScreamTxUpdateBitrateStream(ScreamTxWrapper* s,
                                 uint32_t ssrc,
                                 float minBitrate,
                                 float maxBitrate) {
  s->scream->updateBitrateStream(ssrc, minBitrate, maxBitrate);
}

void ScreamTxSetTargetPriority(ScreamTxWrapper* s, uint32_t ssrc, float priority) {
  s->scream->setTargetPriority(ssrc, priority);
}

void ScreamTxGetStatistics(ScreamTxWrapper* s, float time, char* result) {
  result[0] = '\0';
  s->scream->getStatistics(time, result);
}
