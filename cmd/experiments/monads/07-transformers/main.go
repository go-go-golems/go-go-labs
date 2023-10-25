package main

import (
	"context"
	"fmt"
)

// The goal in this attempt is to model the thing that actually works on a monad.
// The fact that a datastructure is a monad just means that a bind operation could be
// implemented on it (and that bind can be many things). However, we want to model
// the things that take an input and create a new monad with whatever they computed
// wrapped in it.

// For now, I will call these T -> Monad[U] functions "transformers", but I'm not sure
// and provide an interface for them, because I want to be able to introspect them.

// As an exercise, I'm going to build:
// - a producing transformer (takes the null monad and creates new outputs)
// - a failing transformer
// - a retry transformer (takes a failing transformer and retries it)

type Result[T any] struct {
	Value T
	Err   error
}

func (r *Result[T]) String() string {
	if r.Err != nil {
		return fmt.Sprintf("Error(%v)", r.Err)
	}
	return fmt.Sprintf("Value(%v)", r.Value)
}

type Error string

func (e Error) Error() string { return string(e) }

type Monad[T any] struct {
	value <-chan Result[T]
}

func (m Monad[T]) Return() []Result[T] {
	res := []Result[T]{}
	for r := range m.value {
		res = append(res, r)
	}
	return res
}

type Transformer[T any, U any] interface {
	Run(ctx context.Context, input T) Monad[U]
	// we probably need a Close method?
}

func Bind[T any, U any](
	ctx context.Context,
	m Monad[T],
	transformer Transformer[T, U],
) Monad[U] {
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
						for u := range transformer.Run(ctx, r.Value).value {
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

type IntProducer struct {
	Value int
}

func (p *IntProducer) Run(ctx context.Context, _ interface{}) Monad[int] {
	return Resolve[int](p.Value)
}

type FailingProducer struct {
}

func (p *FailingProducer) Run(ctx context.Context, v int) Monad[int] {
	fmt.Printf("FailingProducer: %v\n", v)
	if v > 0 {
		return Resolve[int](v + 10)
	} else {
		return Reject[int](Error("value too small"))
	}
}

type RetryProducer[U any] struct {
	Producer      Transformer[int, U]
	RetryCount    int
	MaxRetryCount int
}

func (p *RetryProducer[U]) Run(ctx context.Context, v int) Monad[U] {
	m := p.Producer.Run(ctx, v)

	return Monad[U]{
		value: func() <-chan Result[U] {
			c := make(chan Result[U])
			hasFailed := false

			go func() {
				defer close(c)
				for {
					select {
					case r, ok := <-m.value:
						fmt.Printf("RetryProducer: %s %v (%d/%d)\n", r.String(), ok, p.RetryCount, p.MaxRetryCount)
						if !ok {
							if !hasFailed {
								return
							}

							if p.RetryCount >= p.MaxRetryCount {
								c <- Result[U]{Err: Error("max retry count reached")}
								return
							}
							p.RetryCount++
							m = p.Producer.Run(ctx, v+p.RetryCount)
							continue
						}

						if r.Err != nil {
							hasFailed = true
						} else {
							fmt.Printf("RetryProducer: success %v\n", r)
							c <- r
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

type OutputTransformer[T any] struct {
}

func (t *OutputTransformer[T]) Run(_ context.Context, input T) Monad[T] {
	fmt.Printf("Output: %v\n", input)
	return ResolveNone[T]()
}

func Resolve[T any](value T) Monad[T] {
	c := make(chan Result[T], 1)
	c <- Result[T]{Value: value}
	close(c)
	return Monad[T]{value: c}
}

func ResolveNone[T any]() Monad[T] {
	c := make(chan Result[T], 1)
	close(c)
	return Monad[T]{value: c}
}

func Reject[T any](err error) Monad[T] {
	c := make(chan Result[T], 1)
	c <- Result[T]{Err: err}
	close(c)
	return Monad[T]{value: c}
}

func main() {
	i1 := &IntProducer{Value: -5}
	i2 := &IntProducer{Value: 10}

	f1 := &FailingProducer{}

	r1 := &RetryProducer[int]{
		Producer:      f1,
		RetryCount:    0,
		MaxRetryCount: 7,
	}
	r2 := &RetryProducer[int]{
		Producer:      f1,
		RetryCount:    0,
		MaxRetryCount: 2,
	}

	o := &OutputTransformer[int]{}

	ctx := context.Background()

	fmt.Println("start print one value")
	res := Bind[int, int](
		ctx,
		Bind[interface{}, int](
			ctx,
			Resolve[interface{}](nil),
			i1),
		o).Return()
	fmt.Printf("Result: %v\n", res)
	fmt.Println()

	u1 := Bind[interface{}, int](ctx, Resolve[interface{}](nil), i1)
	u2 := Bind[int, int](ctx, u1, r1)
	u3 := Bind[int, int](ctx, u2, o)

	res = u3.Return()
	fmt.Printf("Result: %v\n", res)
	fmt.Println()

	u4 := Bind[interface{}, int](ctx, Resolve[interface{}](nil), i2)
	u5 := Bind[int, int](ctx, u4, r2)
	u6 := Bind[int, int](ctx, u5, o)

	res = u6.Return()
	fmt.Printf("Result: %v\n", res)
}
