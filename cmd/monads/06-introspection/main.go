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

type State[T any] struct {
	UUID  uuid.UUID
	Name  string
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

			wg.Add(1)
			go func(v T) {
				defer wg.Done()
				m2 := transform(v)
				// NOTE(manuel, 2023-07-28) It'd be good to capture the submonads
				// here and percolate them upstream through GetState
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
