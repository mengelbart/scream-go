#include "ScreamRx.h"
#include "ScreamRxC.h"

ScreamRxC* ScreamRxInit(int ssrc) {
    ScreamRx* ret = new ScreamRx(ssrc);
    return (void**) ret;
}

void ScreamRxFree(ScreamRxC* s) {
    ScreamRx* srx = (ScreamRx*) s;
    delete srx;
}

void ScreamRxReceive(ScreamRxC* s,
        unsigned int time_ntp,
        void* rtpPacket,
        unsigned int ssrc,
        int size,
        unsigned int seqNr,
        unsigned char ceBits
    ){

    ScreamRx* srx = (ScreamRx*) s;
    srx->receive(time_ntp, rtpPacket, ssrc, size, seqNr, ceBits);
}

bool ScreamRxIsFeedback(ScreamRxC* s, unsigned int time_ntp) {
    ScreamRx* srx = (ScreamRx*) s;
    return srx->isFeedback(time_ntp);
}

bool ScreamRxGetFeedback(ScreamRxC* s, unsigned int time_ntp, bool isMark, unsigned char *buf, int size) {
    ScreamRx* srx = (ScreamRx*) s;
    return srx->createStandardizedFeedback(time_ntp, isMark, buf, size);
}
