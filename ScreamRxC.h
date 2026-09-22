/*
 * Copyright (c) 2020 Mathis Engelbart All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

#ifndef SCREAM_RX_H
#define SCREAM_RX_H

#include <stdbool.h>
#include <stdint.h>

#ifdef __cplusplus
#include "ScreamRx.h"
extern "C" {
#else
typedef struct ScreamRx ScreamRx;
#endif

#include <stdbool.h>

/*
 * Upper bound on the size of one standardized feedback packet.
 * ScreamRx::createStandardizedFeedback fills a buffer until kMaxRtcpSize (900
 * bytes, see ScreamRx.cpp) is reached and may append one more report block
 * after that, so it never writes more than roughly kMaxRtcpSize + 128 bytes.
 * It does not bounds check the buffer itself.
 */
#define SCREAM_RX_MAX_FEEDBACK_SIZE 2048

ScreamRx* ScreamRxInit(uint32_t ssrc);
void ScreamRxFree(ScreamRx*);

void ScreamRxReceive(ScreamRx* s,
                     uint32_t time_ntp,
                     void* rtpPacket,
                     uint32_t ssrc,
                     int size,
                     uint16_t seqNr,
                     uint8_t ceBits,
                     bool isMark,
                     uint32_t timeStamp);
bool ScreamRxIsFeedback(ScreamRx*, uint32_t);
/* Returns false without writing to buf if capacity is too small */
bool ScreamRxGetFeedback(ScreamRx*,
                         uint32_t,
                         bool,
                         unsigned char* buf,
                         int capacity,
                         int* size);

#ifdef __cplusplus
}
#endif

#endif
