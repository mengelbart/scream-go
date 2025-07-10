// Copyright (c) 2020 Mathis Engelbart All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "ScreamRxC.h"

#include "include/ScreamRx.h"

ScreamRx* ScreamRxInit(uint32_t ssrc) {
  return new ScreamRx(ssrc);
}

void ScreamRxFree(ScreamRx* s) {
  delete s;
}

void ScreamRxReceive(ScreamRx* s,
                     uint32_t time_ntp,
                     void* rtpPacket,
                     uint32_t ssrc,
                     int size,
                     uint16_t seqNr,
                     uint8_t ceBits,
                     bool isMark,
                     uint32_t timeStamp) {
  s->receive(time_ntp, rtpPacket, ssrc, size, seqNr, ceBits, isMark, timeStamp);
}

bool ScreamRxIsFeedback(ScreamRx* s, uint32_t time_ntp) {
  return s->isFeedback(time_ntp);
}

bool ScreamRxGetFeedback(ScreamRx* s,
                         uint32_t time_ntp,
                         bool isMark,
                         unsigned char* buf,
                         int* size) {
  return s->createStandardizedFeedback(time_ntp, isMark, buf, *size);
}
