package main

import (
	"math/rand"
	"time"
)

type BurstyBehavior struct {
	NormalBlockSize    int
	BurstBlockSize     int
	NormalInterval     time.Duration
	BurstInterval      time.Duration
	BurstProbability   float64
	BurstStateDuration time.Duration
	inBurst            bool
	burstStateEnd      time.Time
}

func (b *BurstyBehavior) NextBlockSize() int {
	if b.inBurst {
		return b.BurstBlockSize
	}
	return b.NormalBlockSize
}

func (b *BurstyBehavior) Sleep() {
	now := time.Now()
	if b.inBurst && now.After(b.burstStateEnd) {
		b.inBurst = false
	} else if !b.inBurst && rand.Float64() < b.BurstProbability {
		b.inBurst = true
		b.burstStateEnd = now.Add(b.BurstStateDuration)
	}

	if b.inBurst {
		time.Sleep(b.BurstInterval)
	} else {
		time.Sleep(b.NormalInterval)
	}
}
