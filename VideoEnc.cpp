// Copyright (c) 2015, Ericsson AB. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without modification,
// are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice, this
// list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice, this
// list of conditions and the following disclaimer in the documentation and/or other
// materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
// IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT,
// INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
// NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
// PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY,
// WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY
// OF SUCH DAMAGE.

#include "VideoEnc.h"
#include "RtpQueue.h"
#include <string.h>
#include <iostream>
#include <stdio.h>
#include <stdlib.h>
#include <algorithm>

using namespace std;

static const int kMaxRtpSize = 1200;
static const int kRtpOverHead = 12;

VideoEnc::VideoEnc(RtpQueue* rtpQueue_, float frameRate_, char *fname, int ixOffset_) {
    rtpQueue = rtpQueue_;
    frameRate = frameRate_;
    ix = ixOffset_;
    nFrames = 0;
    seqNr = 0;
    nominalBitrate = 0.0;
    FILE *fp = fopen(fname,"r");
    char s[100];
    float sum = 0.0;
    while (fgets(s,99,fp)) {
        if (nFrames < MAX_FRAMES - 1) {
            float x = atof(s);
            frameSize[nFrames] = x;
            nFrames++;
            sum += x;
        }
    }
    float t = nFrames / frameRate;
    nominalBitrate = sum * 8 / t;
    fclose(fp);
}

void VideoEnc::setTargetBitrate(float targetBitrate_) {
    targetBitrate = targetBitrate_;
}

int VideoEnc::encode(float time) {
	int rtpBytes = 0;
	char rtpPacket[2000];
	int bytes = (int)(frameSize[ix] / nominalBitrate * targetBitrate);
	nominalBitrate = 0.95*nominalBitrate + 0.05*frameSize[ix] * frameRate * 8;
    ix++; if (ix == nFrames) ix = 0;
    while (bytes > 0) {
        int rtpSize = std::min(kMaxRtpSize, bytes);
        bytes -= rtpSize;
        rtpSize += kRtpOverHead;
        rtpBytes += rtpSize;
        rtpQueue->push(rtpPacket, rtpSize, seqNr, time);
        seqNr++;
    }
    rtpQueue->setSizeOfLastFrame(rtpBytes);
    return rtpBytes;
}

