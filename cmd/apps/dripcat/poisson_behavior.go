package main

import (
	"math"
	"math/rand"
	"time"
)

type PoissonBehavior struct {
	Rate float64
}

func (b *PoissonBehavior) NextBlockSize() int {
	return 1 // Poisson process typically models event occurrences, so we'll keep block size at 1
}

func (b *PoissonBehavior) Sleep() {
	interval := -math.Log(1.0-rand.Float64()) / b.Rate
	time.Sleep(time.Duration(interval * float64(time.Second)))
}
