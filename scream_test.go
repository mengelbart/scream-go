// Copyright (c) Mathis Engelbart. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scream

import (
	"math"
	"testing"
	"time"
)

// testPacket is the Packet implementation the tests queue up.
type testPacket struct {
	ts   time.Time
	seq  uint16
	size int
}

func (p testPacket) Size() int {
	if p.size == 0 {
		return 1000
	}
	return p.size
}

func (p testPacket) SequenceNumber() uint16 { return p.seq }
func (p testPacket) Timestamp() time.Time   { return p.ts }

// ntpEpochOffset is the number of seconds between the NTP and the Unix epoch.
const ntpEpochOffset = 2208988800

func TestToNTP(t *testing.T) {
	// 1 January 1970 is 2208988800 seconds into the NTP epoch, with a zero
	// fraction.
	if got, want := toNTP(time.Unix(0, 0)), uint64(ntpEpochOffset)<<32; got != want {
		t.Errorf("toNTP(unix 0) = %#x, want %#x", got, want)
	}

	// Half a second in gives the same seconds and a half fraction.
	half := toNTP(time.Unix(0, int64(time.Second/2)))
	if got, want := uint32(half>>32), uint32(ntpEpochOffset); got != want {
		t.Errorf("seconds = %v, want %v", got, want)
	}
	frac := float64(uint32(half)) / float64(1<<32)
	if math.Abs(frac-0.5) > 1e-6 {
		t.Errorf("fraction = %v, want ~0.5", frac)
	}
}

func TestToNTP32(t *testing.T) {
	// toNTP32 is the middle 32 bits of the NTP timestamp, so one unit is
	// 1/65536 s and the value wraps every 2^16 s.
	base := time.Unix(1000000, 0)
	for _, d := range []time.Duration{
		0,
		time.Millisecond,
		100 * time.Millisecond,
		time.Second,
		time.Minute,
	} {
		got := toNTP32(base.Add(d)) - toNTP32(base)
		want := uint32(d.Seconds() * 65536)
		if diff := int64(got) - int64(want); diff < -2 || diff > 2 {
			t.Errorf("toNTP32 delta for %v = %v, want ~%v", d, got, want)
		}
	}
}

func TestToNTP32Wraps(t *testing.T) {
	// The value must wrap rather than saturate, since SCReAM relies on
	// unsigned wraparound arithmetic.
	before := ntpShortTimeNear(t, ntpShortRange-1)
	after := before.Add(2 * time.Second)

	if toNTP32(after) > toNTP32(before) {
		t.Fatalf("expected a wrap, got %v then %v", toNTP32(before), toNTP32(after))
	}
	// Unsigned subtraction still gives the right delta across the wrap.
	if delta := toNTP32(after) - toNTP32(before); math.Abs(float64(delta)/65536.0-2) > 0.01 {
		t.Errorf("delta across wrap = %v units, want ~%v", delta, 2*65536)
	}
}

// ntpShortTimeNear returns a time whose NTP short format seconds are the given
// value, so that tests can exercise behaviour near the 2^16 second wrap.
func ntpShortTimeNear(t *testing.T, seconds int) time.Time {
	t.Helper()
	// NTP seconds of the Unix epoch, rounded down to a multiple of the wrap
	// period, plus the requested offset into the period.
	now := time.Now().Unix() + ntpEpochOffset
	period := int64(ntpShortRange)
	target := (now/period)*period + int64(seconds)
	return time.Unix(target-ntpEpochOffset, 0)
}
