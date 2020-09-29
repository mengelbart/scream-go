/*
 * Copyright (c) 2015, Ericsson AB. All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without modification,
 * are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice, this
 * list of conditions and the following disclaimer.
 *
 * 2. Redistributions in binary form must reproduce the above copyright notice, this
 * list of conditions and the following disclaimer in the documentation and/or other
 * materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
 * ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
 * IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT,
 * INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
 * NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
 * PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY,
 * WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY
 * OF SUCH DAMAGE.
 */

#ifndef NET_QUEUE
#define NET_QUEUE


class NetQueueItem {
public:
    NetQueueItem();
    void* packet;
    unsigned int ssrc;
    int size;
    unsigned short seqNr;
    float tRelease;
    float tQueue;
    bool isCe;
    bool used;
};
const int NetQueueSize = 10000;
class NetQueue {
public:

    NetQueue(float delay, float rate=0.0f, float jitter=0.0f, bool isL4s = false);

    void insert(float time,
        void *rtpPacket,
        unsigned int ssrc,
        int size,
        unsigned short seqNr,
        bool isCe = false);
    bool extract(float time, 
        void *rtpPacket, 
        unsigned int &ssrc,
        int& size, 
        unsigned short &seqNr,
        bool &isCe);
    int sizeOfQueue();

    void updateRate(float time);

    NetQueueItem *items[NetQueueSize];
    int head; // Pointer to last inserted item
    int tail; // Pointer to the oldest item
    int nItems;
    float delay;
    float rate;
    float jitter;
    float nextTx;
    float lastQueueLow;
    bool isL4s;
    int bytesTx;
    float lastRateUpdateT;
    float pDrop;
    float prevRateFrac;
    float l4sTh;
	float tQueueAvg;
};

#endif