package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type Monad[T any] struct {
	value  <-chan T
	logger *log.Logger
}

func NewMonad[T any](logger *log.Logger) Monad[T] {
	return Monad[T]{logger: logger}
}

func Bind[T any, U any](ctx context.Context, m Monad[T], transform func(value T) Monad[U]) Monad[U] {
	c := make(chan U)
	var wg sync.WaitGroup

	go func() {
		for v := range m.value {
			wg.Add(1)
			go func(v T) {
				defer wg.Done()
				ch := transform(v).value

			L:
				for {
					select {
					case u, ok := <-ch:
						if ok {
							select {
							case c <- u: // Send to channel if not closed and context not done.
							case <-ctx.Done():
								m.logger.Println("context done while sending to channel")
								break L
							}
						} else {
							break L
						}
					case <-ctx.Done():
						m.logger.Println("context done while reading from channel")
						break L
					}
				}

			}(v)
		}
		wg.Wait()
		close(c)
	}()

	return Monad[U]{value: c, logger: m.logger}
}

func Resolve[S any, T any](m Monad[S], values ...T) Monad[T] {
	c := make(chan T, len(values))
	for _, v := range values {
		c <- v
	}
	close(c)
	return Monad[T]{value: c, logger: m.logger}
}

func TestThreeByThree() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	logger := log.New(os.Stdout, "Async Monad: ", log.Ltime)

	m1 := NewMonad[int](logger)
	m1 = Resolve(m1, 1, 2, 3)
	m2 := Bind(ctx, m1, func(value int) Monad[string] {
		return Resolve(m1, fmt.Sprintf("%d-%d", value, 1), fmt.Sprintf("%d-%d", value, 2), fmt.Sprintf("%d-%d", value, 3))
	})

	for v := range m2.value {
		println(v)
	}
}

func TestThreeByOne() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	logger := log.New(os.Stdout, "Async Monad: ", log.Ltime)

	m1 := NewMonad[int](logger)
	m1 = Resolve(m1, 1, 2, 3)
	m2 := Bind(ctx, m1, func(value int) Monad[string] {
		return Resolve(m1, fmt.Sprintf("%d-%d", value, 1))
	})

	for v := range m2.value {
		println(v)
	}
}

func main() {
	TestThreeByThree()
	println()
	TestThreeByOne()
}
