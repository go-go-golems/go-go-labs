package main

import "time"

type Monad[T any] struct {
	value <-chan T
}

func Bind[T any, U any](m Monad[T], transform func(value T) Monad[U]) Monad[U] {
	return Monad[U]{
		value: func() <-chan U {
			c := make(chan U)
			go func() {
				defer close(c)
				// we only expect a single value from a monad's value
				for v := range m.value {
					for u := range transform(v).value {
						c <- u
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

	m1 := Resolve(1)
	m2 := Bind(m1, func(value int) Monad[int] {
		return Resolve(value + 1)
	})
	m3 := Bind(m2, func(value int) Monad[int] {
		return Resolve(value + 1)
	})
	m4After2Seconds := Bind(m3, func(value int) Monad[int] {
		// wait 2 seconds before finishing the computation
		time.Sleep(2 * time.Second)
		return Resolve(value + 1)
	})

	v := <-m4After2Seconds.value
	println(v)
}
