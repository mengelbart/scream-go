/*
 * Copyright (c) Mathis Engelbart. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

#ifndef SCREAM_TX_H
#define SCREAM_TX_H

#include <stdbool.h>
#include <stdint.h>

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

ScreamV2Tx* ScreamTxInit(const ScreamTxConfig*);
void ScreamTxFree(ScreamV2Tx*);

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

void ScreamTxRegisterNewStream(ScreamV2Tx*,
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
void ScreamTxNewMediaFrame(ScreamV2Tx*, uint32_t, uint32_t, int, bool);
/* ssrc is an output, it is only written when the return value is 0.0 */
float ScreamTxIsOkToTransmit(ScreamV2Tx*, uint32_t time_ntp, uint32_t* ssrc);
float ScreamTxAddTransmitted(ScreamV2Tx*,
                             uint32_t,
                             uint32_t,
                             int,
                             uint16_t,
                             bool);
void ScreamTxIncomingStdFeedbackBuf(ScreamV2Tx*,
                                    uint32_t,
                                    unsigned char*,
                                    int size);
void ScreamTxIncomingStdFeedback(ScreamV2Tx*,
                                 uint32_t,
                                 int,
                                 uint32_t,
                                 uint16_t,
                                 uint8_t,
                                 bool);
float ScreamTxGetTargetBitrate(ScreamV2Tx*, uint32_t, uint32_t);

/*
 * Settings applied after construction. Those that SCReAM reads continuously
 * can also be changed while running.
 */
void ScreamTxEnablePacketPacing(ScreamV2Tx*, bool);
void ScreamTxEnableRelaxedPacing(ScreamV2Tx*, bool);
void ScreamTxSetEnableCyclicPacing(ScreamV2Tx*, bool);
void ScreamTxSetCwndMinLow(ScreamV2Tx*, int);
void ScreamTxAutoTuneMinCwnd(ScreamV2Tx*, bool);
void ScreamTxSetEnableAdaptiveWindowHeadroom(ScreamV2Tx*, bool);
void ScreamTxSetEnableRatePolicerProtection(ScreamV2Tx*, bool);
void ScreamTxSetSchedulingJitterMargin(ScreamV2Tx*, float);
void ScreamTxSetPostCongestionDelayRtts(ScreamV2Tx*, int);
void ScreamTxSetReorderTime(ScreamV2Tx*, float);
void ScreamTxSetEnableRateUpdate(ScreamV2Tx*, bool);
/* SCReAM stores the pointer without copying, it must outlive the ScreamV2Tx */
void ScreamTxSetLogTag(ScreamV2Tx*, char*);
void ScreamTxSetMaxTotalBitrate(ScreamV2Tx*, float);
/* Returns false if nMssListItems exceeds ScreamTxMaxMssListSize */
bool ScreamTxSetMssListMinPacketsInFlight(ScreamV2Tx*,
                                          int* mssList,
                                          int nMssListItems,
                                          int minPacketsInFlight);

void ScreamTxUpdateBitrateStream(ScreamV2Tx*,
                                 uint32_t ssrc,
                                 float minBitrate,
                                 float maxBitrate);
void ScreamTxSetTargetPriority(ScreamV2Tx*, uint32_t ssrc, float priority);

/*
 * Size of the buffer ScreamTxGetStatistics writes into.
 * ScreamTx::Statistics::getSummary (see ScreamTx.cpp) sprintf's a single fixed
 * summary line of roughly 200 characters plus the log tag, and does not bounds
 * check the buffer.
 */
#define SCREAM_TX_STATISTICS_SIZE 1024

void ScreamTxGetStatistics(ScreamV2Tx*, float time, char* result);

#ifdef __cplusplus
}
#endif

#endif
