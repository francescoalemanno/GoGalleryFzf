package fzf

import (
	"sort"
	"unicode/utf8"

	"github.com/junegunn/fzf/src/algo"
	"github.com/junegunn/fzf/src/util"
)

// MatchResult represents a matched item with fzf score
type MatchResult struct {
	Text  string
	Score int
	Pos   []int // Positions of matched characters
}

// ByScore sorts matches by score descending
type ByScore []MatchResult

func (a ByScore) Len() int           { return len(a) }
func (a ByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByScore) Less(i, j int) bool {
	if a[i].Score != a[j].Score {
		return a[i].Score > a[j].Score // Higher score first
	}
	return a[i].Text < a[j].Text // Alphabetical tie-break
}

// FuzzyMatch performs fzf fuzzy matching on text with the given pattern
// Returns: (matched, score, positions)
func FuzzyMatch(pattern, text string, caseSensitive bool) (bool, int, []int) {
	if pattern == "" {
		return true, 0, nil
	}

	// Convert to runes for proper unicode handling
	patternRunes := []rune(pattern)
	
	// Create util.Chars from text using ToChars (returns value, need pointer)
	input := util.ToChars([]byte(text))

	// Use FuzzyMatchV2 - the optimal algorithm
	// FuzzyMatchV2(caseSensitive, normalize, forward, input, pattern, withPos, slab)
	result, positions := algo.FuzzyMatchV2(
		caseSensitive, // caseSensitive
		true,          // normalize (handle unicode normalization)
		true,          // forward matching
		&input,        // pointer to Chars
		patternRunes,
		true, // withPos - return matched positions
		nil,  // slab - memory allocator (nil = use default)
	)

	if result.Score <= 0 || result.Start == -1 {
		return false, 0, nil
	}

	// Convert positions from rune indices to byte indices
	var bytePositions []int
	if positions != nil {
		bytePositions = make([]int, len(*positions))
		runeIdx := 0
		byteIdx := 0
		for i := 0; i < len(text); {
			r, size := utf8.DecodeRuneInString(text[i:])
			if r == utf8.RuneError && size == 1 {
				// Invalid UTF-8, treat as single byte
				byteIdx = i
			} else {
				byteIdx = i
			}
			
			// Check if this rune index is in positions
			for _, pos := range *positions {
				if pos == runeIdx {
					bytePositions = append(bytePositions, byteIdx)
					break
				}
			}
			
			runeIdx++
			i += size
		}
	}

	return true, result.Score, bytePositions
}

// FuzzySearch searches for pattern in items and returns ranked results
// If limit > 0, returns at most limit results
func FuzzySearch(pattern string, items []string, limit int) []MatchResult {
	if pattern == "" {
		results := make([]MatchResult, len(items))
		for i, item := range items {
			results[i] = MatchResult{Text: item, Score: 0, Pos: nil}
		}
		return results
	}

	var results []MatchResult
	for _, item := range items {
		matched, score, pos := FuzzyMatch(pattern, item, false) // case insensitive
		if matched {
			results = append(results, MatchResult{
				Text:  item,
				Score: score,
				Pos:   pos,
			})
		}
	}

	// Sort by score descending
	sort.Sort(ByScore(results))

	if limit > 0 && len(results) > limit {
		return results[:limit]
	}
	return results
}

// FilterStrings filters strings using fzf matching, returns matched strings (no scores)
func FilterStrings(pattern string, items []string) []string {
	results := FuzzySearch(pattern, items, 0)
	filtered := make([]string, len(results))
	for i, r := range results {
		filtered[i] = r.Text
	}
	return filtered
}

// Score calculates fzf score for a match (0 if no match)
func Score(pattern, text string) int {
	_, score, _ := FuzzyMatch(pattern, text, false)
	return score
}

// HighlightPositions returns the byte positions of matched characters
func HighlightPositions(pattern, text string) []int {
	_, _, pos := FuzzyMatch(pattern, text, false)
	return pos
}

// ExactMatch performs exact substring matching with fzf scoring
func ExactMatch(pattern, text string, caseSensitive bool) (bool, int, []int) {
	if pattern == "" {
		return true, 0, nil
	}

	patternRunes := []rune(pattern)
	input := util.ToChars([]byte(text))

	// Use ExactMatchNaive for substring matching
	result, positions := algo.ExactMatchNaive(
		caseSensitive,
		true,  // normalize
		true,  // forward
		&input,
		patternRunes,
		true, // withPos
		nil,  // slab
	)

	if result.Score <= 0 || result.Start == -1 {
		return false, 0, nil
	}

	var bytePositions []int
	if positions != nil {
		for _, pos := range *positions {
			bytePositions = append(bytePositions, pos)
		}
	}

	return true, result.Score, bytePositions
}
