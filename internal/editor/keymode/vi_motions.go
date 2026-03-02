package keymode

import (
	"unicode"
)

// viWordBoundaryRight finds the end of the next word (Vi 'w' motion)
// Vi words: sequences of word chars or sequences of non-word non-space chars
func viWordBoundaryRight(line string, col int) int {
	runes := []rune(line)
	if col >= len(runes) {
		return len(runes)
	}
	i := col
	if i < len(runes) && isWordChar(runes[i]) {
		// Skip word chars
		for i < len(runes) && isWordChar(runes[i]) {
			i++
		}
	} else if i < len(runes) && !unicode.IsSpace(runes[i]) {
		// Skip punctuation
		for i < len(runes) && !unicode.IsSpace(runes[i]) && !isWordChar(runes[i]) {
			i++
		}
	}
	// Skip whitespace
	for i < len(runes) && unicode.IsSpace(runes[i]) {
		i++
	}
	return i
}

// viWordBoundaryLeft finds the start of the previous word (Vi 'b' motion)
func viWordBoundaryLeft(line string, col int) int {
	runes := []rune(line)
	if col <= 0 {
		return 0
	}
	i := col - 1
	// Skip whitespace
	for i > 0 && unicode.IsSpace(runes[i]) {
		i--
	}
	if i >= 0 && isWordChar(runes[i]) {
		for i > 0 && isWordChar(runes[i-1]) {
			i--
		}
	} else if i >= 0 && !unicode.IsSpace(runes[i]) {
		for i > 0 && !unicode.IsSpace(runes[i-1]) && !isWordChar(runes[i-1]) {
			i--
		}
	}
	return i
}

// viWordEndRight finds the end of the current/next word (Vi 'e' motion)
func viWordEndRight(line string, col int) int {
	runes := []rune(line)
	if col >= len(runes)-1 {
		return len(runes) - 1
	}
	i := col + 1
	// Skip whitespace
	for i < len(runes) && unicode.IsSpace(runes[i]) {
		i++
	}
	if i < len(runes) && isWordChar(runes[i]) {
		for i < len(runes)-1 && isWordChar(runes[i+1]) {
			i++
		}
	} else {
		for i < len(runes)-1 && !unicode.IsSpace(runes[i+1]) && !isWordChar(runes[i+1]) {
			i++
		}
	}
	return i
}

// firstNonBlank returns the column of the first non-whitespace character
func firstNonBlank(line string) int {
	for i, ch := range line {
		if !unicode.IsSpace(ch) {
			return i
		}
	}
	return 0
}

// findCharForward finds the next occurrence of ch after col in line
func findCharForward(line string, col int, ch rune) int {
	runes := []rune(line)
	for i := col + 1; i < len(runes); i++ {
		if runes[i] == ch {
			return i
		}
	}
	return -1
}

// findCharBefore finds the column just before the next occurrence of ch (for 't')
func findCharBefore(line string, col int, ch rune) int {
	pos := findCharForward(line, col, ch)
	if pos > 0 {
		return pos - 1
	}
	return -1
}

func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}
