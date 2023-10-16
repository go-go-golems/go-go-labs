package main

import (
	"context"
	"fmt"
	"time"
)

type Monad[T any] struct {
	value <-chan T
}

func Bind[T any, U any](ctx context.Context, m Monad[T], transform func(value T) Monad[U]) Monad[U] {
	return Monad[U]{
		value: func() <-chan U {
			c := make(chan U)
			go func() {
				defer close(c)
				// we only expect a single value from a monad's value
				for {
					select {
					case v, ok := <-m.value:
						if !ok {
							return
						}
						for u := range transform(v).value {
							c <- u
						}
					case <-ctx.Done():
						fmt.Println("context done")
						return
					}
				}
			}()
			return c
		}(),
	}
}

func Resolve[T any](value T) Monad[T] {
	c := make(chan T, 1)
	c <- value
	close(c)
	return Monad[T]{value: c}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel() // cancel when we are finished consumings ints or when main returns

	m1 := Resolve(1)
	m2 := Bind(ctx, m1, func(value int) Monad[int] {
		return Resolve(value + 1)
	})
	m3 := Bind(ctx, m2, func(value int) Monad[int] {
		return Resolve(value + 1)
	})
	m4After2Seconds := Bind(ctx, m3, func(value int) Monad[int] {
		// wait 2 seconds before finishing the computation, but watch ctx for cancellation.
		select {
		case <-time.After(2 * time.Second):
		case <-ctx.Done():
			fmt.Println("context done in sleep")
			return Resolve(69)
		}

		return Resolve(value + 1)
	})

	v := <-m4After2Seconds.value
	println(v)
}
