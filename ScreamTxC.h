/*
 * Copyright (c) 2020 Mathis Engelbart All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

#ifndef SCREAM_TX_H
#define SCREAM_TX_H

#include "RtpQueueC.h"

#ifdef __cplusplus
#include "include/ScreamTx.h"
extern "C" {
#else
typedef struct ScreamV2Tx ScreamV2Tx;
#endif

#include <stdbool.h>
#include <stdint.h>

ScreamV2Tx* ScreamTxInit();
void ScreamTxFree(ScreamV2Tx*);

/*
 * Max number of streams one ScreamV2Tx accepts. registerNewStream writes into a
 * fixed size array without bounds checking, so the caller must not exceed this.
 */
int ScreamTxMaxStreams();

void ScreamTxRegisterNewStream(ScreamV2Tx*,
                               RtpQueueC*,
                               uint32_t,
                               float,
                               float,
                               float,
                               float);
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
