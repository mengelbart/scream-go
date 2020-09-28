#ifndef SCREAM_RX_H
#define SCREAM_RX_H


#ifdef __cplusplus
extern "C" {
#endif

#include <stdbool.h>

    typedef void* ScreamRxC;
    ScreamRxC* ScreamRxInit(int ssrc);
    void ScreamRxFree(ScreamRxC*);

    void ScreamRxReceive(ScreamRxC*, unsigned int, void*, unsigned int, int, unsigned int, unsigned char);
    bool ScreamRxIsFeedback(ScreamRxC*, unsigned int);
    bool ScreamRxGetFeedback(ScreamRxC*, unsigned int, bool, unsigned char *, int);

#ifdef __cplusplus
}
#endif

#endif