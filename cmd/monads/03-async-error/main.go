package main

import (
	"context"
	"fmt"
	"time"
)

type Result[T any] struct {
	Value T
	Err   error
}

type Error string

func (e Error) Error() string { return string(e) }

type Monad[T any] struct {
	value <-chan Result[T]
}

func Bind[T any, U any](ctx context.Context, m Monad[T], transform func(value T) Monad[U]) Monad[U] {
	return Monad[U]{
		value: func() <-chan Result[U] {
			c := make(chan Result[U])
			go func() {
				defer close(c)
				for {
					select {
					case r, ok := <-m.value:
						if !ok {
							return
						}
						if r.Err != nil {
							c <- Result[U]{Err: r.Err}
							return
						}
						for u := range transform(r.Value).value {
							c <- u
						}
					case <-ctx.Done():
						c <- Result[U]{Err: Error("context done")}
						return
					}
				}
			}()
			return c
		}(),
	}
}

func Resolve[T any](value T) Monad[T] {
	c := make(chan Result[T], 1)
	c <- Result[T]{Value: value}
	close(c)
	return Monad[T]{value: c}
}

func Reject[T any](err error) Monad[T] {
	c := make(chan Result[T], 1)
	c <- Result[T]{Err: err}
	close(c)
	return Monad[T]{value: c}
}

func TestErrorHandling() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	m := Reject[int](Error("expected error"))
	m = Bind(ctx, m, func(value int) Monad[int] {
		fmt.Println("This will not be printed")
		return Resolve(value + 1)
	})

	v := <-m.value
	if v.Err == nil || v.Err.Error() != "expected error" {
		fmt.Println("Expected error, got ", v.Err)
	}
	if v.Value != 0 {
		fmt.Println("Expected 0, got ", v.Value)
	}
	fmt.Println("Done, error: ", v.Err)
}

func TestSuccess() {
	ctx := context.Background()

	m := Resolve(1)
	m = Bind(ctx, m, func(value int) Monad[int] {
		return Resolve(value + 1)
	})
	m = Bind(ctx, m, func(value int) Monad[int] {
		return Resolve(value + 1)
	})
	m = Bind(ctx, m, func(value int) Monad[int] {
		// wait 2 seconds before finishing the computation, but watch ctx for cancellation.
		select {
		case <-ctx.Done():
			return Reject[int](ctx.Err())
		default:
			return Resolve(value + 1)
		}
	})

	v := <-m.value
	if v.Err != nil {
		fmt.Println("Expected no error, got ", v.Err)
	}
	if v.Value != 4 {
		fmt.Println("Expected 4, got ", v.Value)
	}
	fmt.Println("Done, error: ", v.Err, "value:", v.Value)
}

func main() {
	TestErrorHandling()
	TestSuccess()
}
