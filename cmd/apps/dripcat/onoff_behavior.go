package main

import (
	"time"
)

type OnOffBehavior struct {
	OnDuration  time.Duration
	OffDuration time.Duration
	BlockSize   int
	isOn        bool
	stateChange time.Time
}

func (b *OnOffBehavior) NextBlockSize() int {
	now := time.Now()
	if now.After(b.stateChange) {
		b.isOn = !b.isOn
		if b.isOn {
			b.stateChange = now.Add(b.OnDuration)
		} else {
			b.stateChange = now.Add(b.OffDuration)
		}
	}
	if b.isOn {
		return b.BlockSize
	}
	return 0
}

func (b *OnOffBehavior) Sleep() {
	time.Sleep(100 * time.Millisecond) // Fixed sleep time for simplicity
}
