/*
 * Copyright (c) Mathis Engelbart. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

#ifndef SCREAM_TX_H
#define SCREAM_TX_H

#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>

#include "RtpQueueC.h"

#ifdef __cplusplus
#include "ScreamTx.h"
extern "C" {
#else
typedef struct ScreamV2Tx ScreamV2Tx;
#endif

/*
 * The ScreamV2Tx constructor arguments. Fill it with ScreamTxDefaultConfig to
 * get the upstream defaults, then override what is needed.
 */
typedef struct ScreamTxConfig {
  float lossBeta;
  float ecnCeBeta;
  float queueDelayTargetMin;
  int cwnd;
  float packetPacingHeadroom;
  float maxAdaptivePacingRateScale;
  float bytesInFlightHeadRoom;
  float multiplicativeIncreaseScalefactor;
  bool isL4s;
  float maxWindowHeadroom;
  bool enableSbd;
  bool enableClockDriftCompensation;
} ScreamTxConfig;

void ScreamTxDefaultConfig(ScreamTxConfig*);

// 

/*
 * ScreamTxWrapper wraps a ScreamV2Tx and its log file pointer, so that the C 
 * API can manage both.
 */
typedef struct ScreamTxWrapper {
  ScreamV2Tx* scream;
  FILE* logFile;
} ScreamTxWrapper;

ScreamTxWrapper* ScreamTxInit(const ScreamTxConfig*);
void ScreamTxFree(ScreamTxWrapper*);

/*
 * Max number of streams one ScreamV2Tx accepts. registerNewStream writes into a
 * fixed size array without bounds checking, so the caller must not exceed this.
 */
int ScreamTxMaxStreams();

/* Default maxRtpQueueDelay of a registered stream, in seconds. */
float ScreamTxDefaultMaxRtpQueueDelay();

/*
 * Max number of entries setMssListMinPacketsInFlight accepts. It memcpy's into
 * a fixed size array without bounds checking.
 */
int ScreamTxMaxMssListSize();

void ScreamTxRegisterNewStream(ScreamTxWrapper*,
                               RtpQueueC*,
                               uint32_t ssrc,
                               float priority,
                               float minBitrate,
                               float startBitrate,
                               float maxBitrate,
                               float maxRtpQueueDelay,
                               bool isAdaptiveTargetRateScale,
                               float hysteresis,
                               bool enableFrameSizeOverhead);
void ScreamTxNewMediaFrame(ScreamTxWrapper*, uint32_t, uint32_t, int, bool);
/* ssrc is an output, it is only written when the return value is 0.0 */
float ScreamTxIsOkToTransmit(ScreamTxWrapper*, uint32_t time_ntp, uint32_t* ssrc);
float ScreamTxAddTransmitted(ScreamTxWrapper*,
                             uint32_t,
                             uint32_t,
                             int,
                             uint16_t,
                             bool);
void ScreamTxIncomingStdFeedbackBuf(ScreamTxWrapper*,
                                    uint32_t,
                                    unsigned char*,
                                    int size);
void ScreamTxIncomingStdFeedback(ScreamTxWrapper*,
                                 uint32_t,
                                 int,
                                 uint32_t,
                                 uint16_t,
                                 uint8_t,
                                 bool);
float ScreamTxGetTargetBitrate(ScreamTxWrapper*, uint32_t, uint32_t);

/*
 * Settings applied after construction. Those that SCReAM reads continuously
 * can also be changed while running.
 */
void ScreamTxEnablePacketPacing(ScreamTxWrapper*, bool);
void ScreamTxEnableRelaxedPacing(ScreamTxWrapper*, bool);
void ScreamTxSetEnableCyclicPacing(ScreamTxWrapper*, bool);
void ScreamTxSetCwndMin(ScreamTxWrapper*, int);
void ScreamTxAutoTuneMinCwnd(ScreamTxWrapper*, bool);
void ScreamTxSetDetailedLog(ScreamTxWrapper*, char*);
void ScreamTxSetEnableRatePolicerProtection(ScreamTxWrapper*, bool);
void ScreamTxSetExpFeature(ScreamTxWrapper*, int);
void ScreamTxSetPostCongestionDelayRtts(ScreamTxWrapper*, int);
void ScreamTxSetReorderTime(ScreamTxWrapper*, float);
void ScreamTxSetEnableRateUpdate(ScreamTxWrapper*, bool);
/* SCReAM stores the pointer without copying, it must outlive the ScreamTxWrapper */
void ScreamTxSetLogTag(ScreamTxWrapper*, char*);
void ScreamTxSetMaxTotalBitrate(ScreamTxWrapper*, float);
/* Returns false if nMssListItems exceeds ScreamTxMaxMssListSize */
bool ScreamTxSetMssListMinPacketsInFlight(ScreamTxWrapper*,
                                          int* mssList,
                                          int nMssListItems,
                                          int minPacketsInFlight);

void ScreamTxUpdateBitrateStream(ScreamTxWrapper*,
                                 uint32_t ssrc,
                                 float minBitrate,
                                 float maxBitrate);
void ScreamTxSetTargetPriority(ScreamTxWrapper*, uint32_t ssrc, float priority);

/*
 * Size of the buffer ScreamTxGetStatistics writes into.
 * ScreamTx::Statistics::getSummary (see ScreamTx.cpp) sprintf's a single fixed
 * summary line of roughly 200 characters plus the log tag, and does not bounds
 * check the buffer.
 */
#define SCREAM_TX_STATISTICS_SIZE 1024

void ScreamTxGetStatistics(ScreamTxWrapper*, float time, char* result);

#ifdef __cplusplus
}
#endif

#endif
