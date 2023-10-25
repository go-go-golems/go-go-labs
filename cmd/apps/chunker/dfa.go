package main

import (
	"github.com/rs/zerolog/log"
	"github.com/tiktoken-go/tokenizer" // Adjust the import path based on your setup
)

type State int
type TokenId = uint
type Action string

const (
	SentenceSeparator Action = "SentenceSeparator"
	None              Action = "None"
)

// DFA Structure
type DFA struct {
	States       map[State]bool              // set of states; true if terminal
	Transitions  map[State]map[TokenId]State // state transitions
	FinalActions map[State]Action            // action to take at final states
}

// Function to compute the DFA based on encoded separators
func computeDFA(separators []string, codec tokenizer.Codec) DFA {
	dfa := DFA{
		States:       make(map[State]bool),
		Transitions:  make(map[State]map[TokenId]State),
		FinalActions: make(map[State]Action),
	}

	var nextState State = 1
	for _, separator := range separators {
		tokenIds, tokens, err := codec.Encode(separator)
		if err != nil {
			log.Fatal().Msgf("Error encoding separator: %v", err)
		}
		log.Debug().Strs("tokens", tokens).Uints("tokenIds", tokenIds).Str("separator", separator).Msg("Encoded separator")

		currentState := State(0) // start state is always 0
		for _, token := range tokenIds {
			if _, ok := dfa.Transitions[currentState]; !ok {
				dfa.Transitions[currentState] = make(map[TokenId]State)
			}

			if next, ok := dfa.Transitions[currentState][token]; ok {
				currentState = next
			} else {
				dfa.Transitions[currentState][token] = nextState
				currentState = nextState
				nextState++
			}
		}

		dfa.States[currentState] = true // mark as terminal state
		dfa.FinalActions[currentState] = SentenceSeparator
	}

	return dfa
}

func splitTokenIdsByDFA(dfa DFA, tokenIds []TokenId, tokenLimit int, codec tokenizer.Codec) ([]TokenId, []TokenId) {
	currentState := State(0)
	headIds := []TokenId{}
	tailIds := []TokenId{}
	currentTokenCount := 0

	for _, tokenId := range tokenIds {
		nextState, ok := dfa.Transitions[currentState][tokenId]

		tailString, err := codec.Decode(tailIds)
		if err != nil {
			log.Fatal().Msgf("Error decoding tail tokens: %v", err)
		}
		tokenString, err := codec.Decode([]TokenId{tokenId})

		log.Debug().Uint("tokenId", tokenId).
			Str("token", tokenString).
			Uint("currentState", uint(currentState)).
			Uint("nextState", uint(nextState)).
			Bool("ok", ok).
			Str("tail", tailString).
			Msg("Transition")

		// If transition is undefined, enqueue token into tailIds
		if !ok {
			nextState = State(0) // Reset to initial state
		}

		// Update the current state
		currentState = nextState
		currentTokenCount += 1

		// If token limit is exceeded, break and return
		if currentTokenCount > tokenLimit {
			return headIds, tokenIds[len(headIds):]
		}

		// If we're at a separator and token count is ok, append tailIds to headIds and reset tailIds
		if dfa.States[currentState] {
			headIds = append(headIds, tailIds...)
			headIds = append(headIds, tokenId)
			tailIds = []TokenId{}
			headString, err := codec.Decode(headIds)
			if err != nil {
				log.Fatal().Msgf("Error decoding head tokens: %v", err)
			}
			log.Debug().Str("head", headString).Msg("Head")
		} else {
			tailIds = append(tailIds, tokenId)
		}
	}

	// if head + tail < tokenCount, concatenate.
	if len(headIds)+len(tailIds) <= tokenLimit {
		headIds = append(headIds, tailIds...)
		tailIds = []TokenId{}
	}

	return headIds, tailIds
}
