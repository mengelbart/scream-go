/*
 * Copyright (c) Mathis Engelbart. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

#ifndef RTP_QUEUE_C
#define RTP_QUEUE_C

#include <stdint.h>

#ifdef __cplusplus

#include "RtpQueue.h"

class RtpQueueC : public RtpQueueIface {
 public:
  RtpQueueC(uintptr_t);

  int clear();
  int sizeOfNextRtp();
  int seqNrOfNextRtp();
  int seqNrOfLastRtp();
  int bytesInQueue();  // Number of bytes in queue
  int sizeOfQueue();   // Number of items in queue
  float getDelay(float currTs);
  int getSizeOfLastFrame();

 private:
  /*
   * ctx is a runtime/cgo.Handle referring to the Go RTPQueue that backs this
   * queue. It is an integer, not a pointer, so it needs no pinning.
   */
  uintptr_t ctx;
};

extern "C" {
#else
typedef struct RtpQueueC RtpQueueC;
#endif

RtpQueueC* RtpQueueCInit(uintptr_t);
void RtpQueueCFree(RtpQueueC*);

int goClear(uintptr_t);
int goSizeOfNextRtp(uintptr_t);
int goSeqNrOfNextRtp(uintptr_t);
int goSeqNrOfLastRtp(uintptr_t);
int goBytesInQueue(uintptr_t);
int goSizeOfQueue(uintptr_t);
float goGetDelay(uintptr_t, float);
int goGetSizeOfLastFrame(uintptr_t);

#ifdef __cplusplus
}
#endif

#endif
