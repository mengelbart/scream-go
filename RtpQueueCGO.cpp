#include "RtpQueueCGO.h"

RtpQueueCGO::RtpQueueCGO(int id) {
}

void RtpQueueCGO::clear() {
    goClear(this->id);
}

int RtpQueueCGO::sizeOfNextRtp() {
    return goSiz;
}

int RtpQueueCGO::seqNrOfNextRtp() {
    return 0;
}

int RtpQueueCGO::bytesInQueue() {
    return 0;
}

int RtpQueueCGO::sizeOfQueue() {
    return 0;
}

float RtpQueueCGO::getDelay(float currTs) {
    return 0.0;
}

int RtpQueueCGO::getSizeOfLastFrame() {
    return 0;
}

