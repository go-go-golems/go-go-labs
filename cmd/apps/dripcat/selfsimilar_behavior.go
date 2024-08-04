package main

import (
	"math"
	"math/rand"
	"time"
)

type SelfSimilarBehavior struct {
	HurstParameter float64
	Mean           float64
	Variance       float64
}

func (b *SelfSimilarBehavior) NextBlockSize() int {
	fGn := b.generateFractionalGaussianNoise()
	return int(math.Max(1, b.Mean+math.Sqrt(b.Variance)*fGn))
}

func (b *SelfSimilarBehavior) Sleep() {
	time.Sleep(100 * time.Millisecond) // Fixed sleep time for simplicity
}

func (b *SelfSimilarBehavior) generateFractionalGaussianNoise() float64 {
	x1 := rand.NormFloat64()
	x2 := rand.NormFloat64()
	r := math.Pow(0.5, 2-2*b.HurstParameter)
	return x1*r + x2*math.Sqrt(1-r*r)
}
