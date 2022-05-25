/*
 * Copyright (c) 2020 Mathis Engelbart All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

#ifndef RTP_QUEUE_CGO
#define RTP_QUEUE_CGO

#include "RtpQueue.h"

#ifdef __cplusplus
extern "C" {
#endif


    void goClear(int);
    int goSizeOfNextRtp(int);
    int goSeqNrOfNextRtp(int);
    int goBytesInQueue(int);
    int goSizeOfQueue(int);
    float goGetDelay(int, float);
    int goGetSizeOfLastFrame(int);

#ifdef __cplusplus
}
#endif

class RtpQueueCGO : public RtpQueueIface {
public:
    RtpQueueCGO(int);

    void clear();
    int sizeOfNextRtp();
    int seqNrOfNextRtp();
    int bytesInQueue(); // Number of bytes in queue
    int sizeOfQueue();  // Number of items in queue
    float getDelay(float currTs);
    int getSizeOfLastFrame();

private:
    int id;
};

#endif
