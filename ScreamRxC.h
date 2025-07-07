/*
 * Copyright (c) 2020 Mathis Engelbart All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

#ifndef SCREAM_RX_H
#define SCREAM_RX_H


#include <stdint.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

#include <stdbool.h>

    typedef struct {
        int size;
        bool result;
    } Feedback;

    typedef void* ScreamRxC;
    ScreamRxC* ScreamRxInit(unsigned int ssrc);
    void ScreamRxFree(ScreamRxC*);

    void ScreamRxReceive(ScreamRxC* s,
        uint32_t time_ntp,
        void* rtpPacket,
        uint32_t ssrc,
        int size,
        uint16_t seqNr,
        uint8_t ceBits,
        bool isMark,
        uint32_t timeStamp);
    bool ScreamRxIsFeedback(ScreamRxC*, unsigned int);
    Feedback* ScreamRxGetFeedback(ScreamRxC*, unsigned int, bool, unsigned char *buf);

    bool ScreamRxGetFeedbackResult(Feedback*);
    int ScreamRxGetFeedbackSize(Feedback*);
    unsigned char* ScreamRxGetFeedbackBuffer(Feedback*);

#ifdef __cplusplus
}
#endif

#endif
