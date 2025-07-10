// Copyright (c) 2020 Mathis Engelbart All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "RtpQueueC.h"

RtpQueueC* RtpQueueCInit(void* ctx) {
  return new RtpQueueC(ctx);
}

void RtpQueueCFree(RtpQueueC* q) {
  delete q;
}

RtpQueueC::RtpQueueC(void* ctx) {
  this->ctx = ctx;
}

int RtpQueueC::clear() {
  return goClear(this->ctx);
}

int RtpQueueC::sizeOfNextRtp() {
  return goSizeOfNextRtp(this->ctx);
}

int RtpQueueC::seqNrOfNextRtp() {
  return goSeqNrOfNextRtp(this->ctx);
}

int RtpQueueC::seqNrOfLastRtp() {
  return goSeqNrOfLastRtp(this->ctx);
}

int RtpQueueC::bytesInQueue() {
  return goBytesInQueue(this->ctx);
}

int RtpQueueC::sizeOfQueue() {
  return goSizeOfQueue(this->ctx);
}

float RtpQueueC::getDelay(float currTs) {
  auto d = goGetDelay(this->ctx, currTs);
  return d;
}

int RtpQueueC::getSizeOfLastFrame() {
  return goGetSizeOfLastFrame(this->ctx);
}
