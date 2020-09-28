#include "ScreamTx.h"
#include "ScreamTxC.h"
#include "RtpQueueCGO.h"

RtpQueueIfaceC* RtpQueueIfaceInit(int id) {
    RtpQueueIface* ret = new RtpQueueCGO(id);
    return (void**) ret;
}

ScreamTxC* ScreamTxInit() {
    ScreamTx* ret = new ScreamTx();
    return (void**) ret;
}

void ScreamTxFree(ScreamTxC* s) {
    ScreamTx* stx = (ScreamTx*) s;
    delete stx;
}

void ScreamTxRegisterNewStream(ScreamTxC* s,
        RtpQueueIfaceC* rtpQueue,
        unsigned int ssrc,
        float priority,
        float minBitrate,
        float startBitrate,
        float maxBitrate) {

    ScreamTx* stx = (ScreamTx*) s;
    RtpQueueIface* rtpq = (RtpQueueIface*) rtpQueue;

    stx->registerNewStream(rtpq,
            ssrc,
            priority,
            minBitrate,
            startBitrate,
            maxBitrate);
}

void ScreamTxNewMediaFrame(ScreamTxC* s, unsigned int time_ntp, unsigned int ssrc, int bytesRtp) {
    ScreamTx* stx = (ScreamTx*) s;
    stx->newMediaFrame(time_ntp, ssrc, bytesRtp);
}

float ScreamTxIsOkToTransmit(ScreamTxC* s, unsigned int time_ntp, unsigned int ssrc) {
    ScreamTx* stx = (ScreamTx*) s;
    return stx->isOkToTransmit(time_ntp, ssrc);
}

float ScreamTxAddTransmitted(ScreamTxC* s, unsigned int time_ntp, unsigned int ssrc, int size, unsigned int seqNr, bool isMark) {
    ScreamTx* stx = (ScreamTx*) s;
    return stx->addTransmitted(time_ntp, ssrc, size, seqNr, isMark);
}

void ScreamTxIncomingStdFeedback(ScreamTxC* s,
        unsigned int time_ntp,
        int streamId,
        unsigned int timestamp,
        unsigned int seqNr,
        unsigned char ceBits,
        bool isLast) {
    ScreamTx* stx = (ScreamTx*) s;
    stx->incomingStandardizedFeedback(time_ntp, streamId, timestamp, seqNr, ceBits, isLast);
}

float ScreamTxGetTargetBitrate(ScreamTxC* s, unsigned int ssrc) {
    ScreamTx* stx = (ScreamTx*) s;
    return stx->getTargetBitrate(ssrc);
}