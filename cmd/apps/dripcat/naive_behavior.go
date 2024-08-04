package main

import (
	"math/rand"
	"time"
)

type NaiveBehavior struct {
	BlockSize int
	SizeSpray int
	Interval  time.Duration
	TimeSpray time.Duration
}

func (b *NaiveBehavior) NextBlockSize() int {
	return b.BlockSize + rand.Intn(b.SizeSpray+1)
}

func (b *NaiveBehavior) Sleep() {
	if b.TimeSpray == 0 {
		time.Sleep(b.Interval)
	} else {
		time.Sleep(b.Interval + time.Duration(rand.Int63n(int64(b.TimeSpray))))
	}
}
