package main

import (
	"fmt"
	"github.com/tiktoken-go/tokenizer"
	"log"
	"regexp"
	"strings"
)

// SplitSentence splits a given text at the nearest sentence or paragraph boundary below a character size
func SplitSentence(text string, charSize int, separators []string) (string, string) {
	if len(text) <= charSize {
		return text, ""
	}

	// Combine all separators into a regular expression
	var sepString string
	for i, sep := range separators {
		sepString += regexp.QuoteMeta(sep)
		if i < len(separators)-1 {
			sepString += "|"
		}
	}

	// Split the text based on the combined separators
	re := regexp.MustCompile(sepString)
	sentences := re.Split(text, -1)

	// Find the nearest boundary below the character size
	for i := len(sentences) - 1; i >= 0; i-- {
		partial := strings.Join(sentences[:i+1], " ")
		if len(partial)+1 <= charSize { // +1 accounts for the trailing separator
			return partial + ".", strings.Join(sentences[i+1:], " ")
		}
	}
	return "", text
}

func SplitSentenceByToken(text string, tokenCount int, separators []string, codecName tokenizer.Model) (string, string) {
	codec, err := tokenizer.ForModel(codecName)
	if err != nil {
		log.Fatalf("Error getting codec: %v", err)
	}

	var currentTokenCount int
	var headTokens, tailTokens []uint

	// Create a regex pattern to capture the separators
	var sepString string
	for i, sep := range separators {
		sepString += regexp.QuoteMeta(sep)
		if i < len(separators)-1 {
			sepString += "|"
		}
	}
	sepPattern := "(" + sepString + ")"
	re := regexp.MustCompile(sepPattern)

	// Use FindAllStringIndex to find start and end indexes of all separators
	indexes := re.FindAllStringIndex(text, -1)

	// Initialize an index for the last token added to head
	var lastIndex int

	// Iterate over the indexes and tokenize sentences along with their separators
	for _, idx := range indexes {
		// Substring with separator
		sentence := text[lastIndex:idx[1]]
		tokenIds, tokens, err := codec.Encode(sentence)
		_ = tokens
		if err != nil {
			log.Fatalf("Error encoding sentence: %v", err)
		}

		// Check token count
		if currentTokenCount+len(tokenIds) > tokenCount {
			break
		}

		// Update variables
		currentTokenCount += len(tokenIds)
		headTokens = append(headTokens, tokenIds...)
		lastIndex += len(sentence)
	}

	// catch last segment
	sentence := text[lastIndex:]
	tokenIds, _, err := codec.Encode(sentence)
	if err != nil {
		log.Fatalf("Error encoding sentence: %v", err)
	}

	// Check token count
	if currentTokenCount+len(tokenIds) <= tokenCount {
		headTokens = append(headTokens, tokenIds...)
		lastIndex += len(sentence)
	}

	// Remaining tokens for tail
	sentence = text[lastIndex:]
	tailTokens, _, err = codec.Encode(sentence)

	// Decode tokens to form 'head' and 'tail'
	head, err := codec.Decode(headTokens)
	if err != nil {
		log.Fatalf("Error decoding head tokens: %v", err)
	}

	tail, err := codec.Decode(tailTokens)
	if err != nil {
		log.Fatalf("Error decoding tail tokens: %v", err)
	}

	return head, tail
}

func main() {
	separators := []string{".", "!", "?", "\n\n"}
	head, tail := SplitSentenceByToken("This is a test. Another test is here! And here is a third one?", 10, separators, tokenizer.GPT4)
	fmt.Println("Head:", head)
	fmt.Println("Tail:", tail)

	head, tail = SplitSentence("This is a test. Another test is here! And here is a third one?", 25, separators)
	fmt.Println("Head:", head)
	fmt.Println("Tail:", tail)

	codec, err := tokenizer.ForModel(tokenizer.GPT4)
	if err != nil {
		log.Fatalf("Error getting codec: %v", err)
	}

	dfa := computeDFA(separators, codec)
	fmt.Println("DFA:", dfa)

	text := "This is a sentence. Another sentence! Yet another one?"
	tokenIds, _, err := codec.Encode(text)
	if err != nil {
		log.Fatalf("Error encoding text: %v", err)
	}

	headIds, tailIds := splitTokenIdsByDFA(dfa, tokenIds, 10, codec)

	headString, _ := codec.Decode(headIds)
	tailString, _ := codec.Decode(tailIds)
	fmt.Println("HeadIds:", headIds)
	fmt.Println("Head", headString)
	fmt.Println("TailIds:", tailIds)
	fmt.Println("Tail", tailString)
}
