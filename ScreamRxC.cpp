// Copyright (c) 2020 Mathis Engelbart All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include <cstdlib>

#include "ScreamRx.h"
#include "ScreamRxC.h"

ScreamRxC* ScreamRxInit(unsigned int ssrc) {
    ScreamRx* ret = new ScreamRx(ssrc);
    return (void**) ret;
}

void ScreamRxFree(ScreamRxC* s) {
    ScreamRx* srx = (ScreamRx*) s;
    delete srx;
}

void ScreamRxReceive(ScreamRxC* s,
        uint32_t time_ntp,
        void* rtpPacket,
        uint32_t ssrc,
        int size,
        uint16_t seqNr,
        uint8_t ceBits,
        bool isMark,
        uint32_t timeStamp
    ){

    ScreamRx* srx = (ScreamRx*) s;
    srx->receive(time_ntp, rtpPacket, ssrc, size, seqNr, ceBits, isMark, timeStamp);
}

bool ScreamRxIsFeedback(ScreamRxC* s, uint32_t time_ntp) {
    ScreamRx* srx = (ScreamRx*) s;
    return srx->isFeedback(time_ntp);
}

Feedback* ScreamRxGetFeedback(ScreamRxC* s, uint32_t time_ntp, bool isMark, unsigned char* buf) {
    ScreamRx* srx = (ScreamRx*) s;
    Feedback *fb = (Feedback*)malloc(sizeof(Feedback));
    fb->result = srx->createStandardizedFeedback(time_ntp, isMark, buf, fb->size);
    return fb;
}

bool ScreamRxGetFeedbackResult(Feedback* fb) {
    return fb->result;
}

int ScreamRxGetFeedbackSize(Feedback* fb) {
    return fb->size;
}
