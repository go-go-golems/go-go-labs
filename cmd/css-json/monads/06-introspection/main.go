package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log"
	"os"
	"sync"
	"time"
)

// State contains the data that can be introspected from a monad
// NOTE(manuel, 2023-08-04) I don't think this is necessarily the right place to store all of this.
// This represents a result, not the transformer itself, but we might want to get some information
// about the transformation steps, in fact we probably want mostly to get information about the
// current state of transformation, while the state of the output is mostly gathering the input / output
// metadata. That's because so often these things are not even created yet, so how could we introspect them?
//
// So the main first step is to create an actual structure that represents the actual transformers,
// instead of just using a lambda, and then modeling around that. The transformers are created once
// at the start of the app, but the Monad is created multiple times.
type State[T any] struct {
	UUID uuid.UUID
	Name string
	// NOTE(manuel, 2023-08-04) I'm not sure why I decided to make multiple inputs...
	Input []interface{}
	// TODO(manuel, 2023-07-28) Add input/output types for monads
	// theoretically this could be a generic type too...
	Output []T
	Meta   map[string]interface{}
	mu     sync.Mutex
}

func (s *State[T]) AddInput(value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Input = append(s.Input, value)
}

func (s *State[T]) AddOutput(value T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Output = append(s.Output, value)
}

type Monad[T any] struct {
	value  <-chan T
	logger *log.Logger
	state  *State[T]
}

func NewMonad[T any](name string, meta map[string]interface{}, logger *log.Logger) Monad[T] {
	return Monad[T]{
		logger: logger,
		state: &State[T]{
			UUID:   uuid.New(),
			Name:   name,
			Input:  make([]interface{}, 0),
			Output: make([]T, 0),
			Meta:   meta,
		},
	}
}

func (m Monad[T]) GetState() *State[T] {
	return m.state
}

func Bind[T any, U any](
	ctx context.Context,
	m Monad[T],
	// NOTE(manuel, 2023-08-04)
	// the transform is the main component that we are building and as such it should probably be its own struct.
	// to model the monad, we should make a:
	// - emitter that sends out a few values (urls, for example)
	// - a curl step that takes the urls and returns the body
	// - a retry step that takes a step and if it fails retries it a couple of times
	// - a step that generates multiple outputs from one input
	transform func(value T) Monad[U],
) Monad[U] {
	c := make(chan U)
	var wg sync.WaitGroup

	ret := NewMonad[U](fmt.Sprintf("%s-%s", m.state.Name, "Bind"), m.state.Meta, m.logger)
	go func() {
		for v := range m.value {
			// the vs' here are so to say the inputs of m2
			//fmt.Printf("Adding input %v to monad %s\n", v, ret.state.Name)
			ret.state.AddInput(v)
			// in order to fully rerun something from the inputs, we need actually input + transform,
			// put to restore the result monad from persistence, we just need the outputs.
			// So the connection from input to output through transform is the graph that is compiled.

			wg.Add(1)
			go func(v T) {
				defer wg.Done()
				m2 := transform(v)
				ch := m2.value

			L:
				for {
					select {
					case u, ok := <-ch:
						// u here is an "output" of m2 (ie, it's what the next bound monad will read as input)
						if ok {
							//fmt.Printf("Adding output %v to monad %s\n", u, ret.state.Name)
							ret.state.AddOutput(u)
							//fmt.Printf("ret state %v\n", ret.state)
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

	ret.value = c
	return ret
}

func Resolve[S any, T any](m Monad[S], values ...T) Monad[T] {
	ret := NewMonad[T](
		fmt.Sprintf("%s-%s", m.state.Name, "Resolve"),
		m.state.Meta,
		m.logger,
	)

	c := make(chan T, len(values))
	for _, v := range values {
		ret.state.AddOutput(v)
		c <- v
	}
	close(c)

	ret.value = c
	return ret
}

func printState[T any](m *Monad[T]) {
	state := m.GetState()
	//fmt.Printf("state: %v\n", state)
	fmt.Printf("UUID: %s\n", state.UUID)
	fmt.Printf("Name: %s\n", state.Name)
	if state.Input != nil && len(state.Input) > 0 {
		fmt.Printf("Input: %v\n", state.Input)
	}
	if state.Output != nil && len(state.Output) > 0 {
		fmt.Printf("Output: %v\n", state.Output)
	}
	if state.Meta != nil {
		fmt.Printf("Meta: %v\n", state.Meta)
	}
}

func TestThreeByThree() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	logger := log.New(os.Stdout, "Async Monad: ", log.Ltime)

	m1 := NewMonad[int]("m1", nil, logger)
	println("m1")
	printState(&m1)
	m1 = Resolve(m1, 1, 2, 3)
	println("m1 after Resolve")
	printState(&m1)
	m2 := Bind(ctx, m1, func(value int) Monad[string] {
		return Resolve(m1, fmt.Sprintf("%d-%d", value, 1), fmt.Sprintf("%d-%d", value, 2), fmt.Sprintf("%d-%d", value, 3))
	})

	println("m1 after Bind")
	printState(&m1)
	println("m2 after Bind")
	printState(&m2)

	for v := range m2.value {
		println(v)
	}

	println("m1 after range")
	printState(&m1)
	println("m2 after range")
	printState(&m2)

}

func TestThreeByOne() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	logger := log.New(os.Stdout, "Async Monad: ", log.Ltime)

	m1 := NewMonad[int]("m1", nil, logger)
	printState(&m1)
	m1 = Resolve(m1, 1, 2, 3)
	printState(&m1)
	m2 := Bind(ctx, m1, func(value int) Monad[string] {
		return Resolve(m1, fmt.Sprintf("%d-%d", value, 1))
	})
	printState(&m2)

	for v := range m2.value {
		println(v)
	}
	printState(&m2)
}

func main() {
	TestThreeByThree()
	println()
	TestThreeByOne()
}
