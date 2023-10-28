// Package kmp implements the Knuth-Morris-Pratt (KMP) algorithm specialized for matching lines of text.
// Instead of searching for a pattern of characters within a string, this implementation of KMP
// operates on sequences of lines. The algorithm preprocesses the pattern to identify prefixes
// that match suffixes at each point in the pattern (the "partial match" table, or "failure function").
// Using this information, the algorithm speeds up the search by skipping non-matching positions more efficiently.
// This avoids the less efficient backtracking characteristic of the naive search algorithm.
package kmp

// computeLPSArray calculates the Longest Prefix Suffix (LPS) array for a given pattern, represented as a slice of strings.
// Each string in the slice represents a line of text. The LPS array is a key component of the KMP algorithm,
// facilitating efficient pattern matching by allowing the algorithm to skip sections of the text that would
// create redundant comparisons.
//
// Parameters:
//   - pattern ([]string): A slice of strings, where each string is a line of text in the pattern to be searched for.
//
// Returns:
//   - ([]int): The LPS array, with each element representing the length of the longest proper prefix that is also a proper suffix.
//
// computeLPSArray calculates the Longest Prefix Suffix (LPS) array.
func computeLPSArray(pattern []string) []int {
	length := len(pattern)
	lps := make([]int, length)

	lengthOfPreviousLongestPrefixSuffix := 0
	i := 1

	for i < length {
		if pattern[i] == pattern[lengthOfPreviousLongestPrefixSuffix] {
			lengthOfPreviousLongestPrefixSuffix++
			lps[i] = lengthOfPreviousLongestPrefixSuffix
			i++
		} else {
			if lengthOfPreviousLongestPrefixSuffix != 0 {
				lengthOfPreviousLongestPrefixSuffix = lps[lengthOfPreviousLongestPrefixSuffix-1]
			} else {
				lps[i] = 0
				i++
			}
		}
	}

	return lps
}

// KMPSearch performs the KMP pattern searching algorithm on lines of text. It searches for the occurrences of a 'pattern'
// within the 'text', both represented as slices of strings. Each string in these slices is a line of text.
// The function prints the starting index of each match found in the 'text'. If no match is found, it produces no output.
//
// The search efficiency is improved by leveraging the LPS array, which allows the algorithm to skip non-matching characters
// without having to backtrack. This mechanism helps in avoiding unnecessary comparisons, speeding up the pattern-matching process.
//
// Parameters:
//
//   - text ([]string): A slice of strings, where each string is a line of text, representing the text to be searched.
//   - pattern ([]string): A slice of strings, where each string is a line of text, representing the pattern to be searched for.
//
// Returns:
//
//	(int): The index of the first match found in the 'text', or -1 if no match is found.
func KMPSearch(text []string, pattern []string) int {
	patternLength := len(pattern)
	textLength := len(text)

	if patternLength == 0 {
		return -1
	}

	if textLength == 0 {
		return -1
	}

	// Preprocess the pattern
	lps := computeLPSArray(pattern)

	i := 0 // index for text[]
	j := 0 // index for pattern[]

	for i < textLength {
		if pattern[j] == text[i] {
			j++
			i++
		}

		if j == patternLength {
			return i - j
		} else if i < textLength && pattern[j] != text[i] {
			if j != 0 {
				j = lps[j-1]
			} else {
				i = i + 1
			}
		}
	}

	return -1
}
