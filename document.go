package prompt

import "unicode"

// Document provides a view into the current text input and cursor position.
// It is passed to the Completer on every keystroke.
type Document struct {
	Text           string
	CursorPosition int // rune index (number of runes to the left of the cursor)
}

func (d Document) runes() []rune {
	return []rune(d.Text)
}

// TextBeforeCursor returns the text to the left of the cursor.
func (d Document) TextBeforeCursor() string {
	r := d.runes()
	if d.CursorPosition > len(r) {
		return d.Text
	}
	return string(r[:d.CursorPosition])
}

// TextAfterCursor returns the text to the right of the cursor.
func (d Document) TextAfterCursor() string {
	r := d.runes()
	if d.CursorPosition >= len(r) {
		return ""
	}
	return string(r[d.CursorPosition:])
}

// CurrentWord returns the word immediately to the left of (or under) the
// cursor, split on whitespace. Useful for the completer to know what prefix
// to match against.
func (d Document) CurrentWord() string {
	before := []rune(d.TextBeforeCursor())
	start := len(before)
	for start > 0 && !unicode.IsSpace(before[start-1]) {
		start--
	}
	return string(before[start:])
}

// FilterHasPrefix returns suggestions whose Text starts with sub.
// When ignoreCase is true the comparison is case-insensitive.
func FilterHasPrefix(suggestions []Suggestion, sub string, ignoreCase bool) []Suggestion {
	if sub == "" {
		return suggestions
	}
	subR := []rune(sub)
	if ignoreCase {
		subR = toLowerRunes(subR)
	}
	out := make([]Suggestion, 0, len(suggestions))
	for _, s := range suggestions {
		textR := []rune(s.Text)
		if ignoreCase {
			textR = toLowerRunes(textR)
		}
		if hasRunePrefix(textR, subR) {
			out = append(out, s)
		}
	}
	return out
}

// FilterContains returns suggestions whose Text contains sub.
// When ignoreCase is true the comparison is case-insensitive.
func FilterContains(suggestions []Suggestion, sub string, ignoreCase bool) []Suggestion {
	if sub == "" {
		return suggestions
	}
	subR := []rune(sub)
	if ignoreCase {
		subR = toLowerRunes(subR)
	}
	out := make([]Suggestion, 0, len(suggestions))
	for _, s := range suggestions {
		textR := []rune(s.Text)
		if ignoreCase {
			textR = toLowerRunes(textR)
		}
		if containsRunes(textR, subR) {
			out = append(out, s)
		}
	}
	return out
}

// --- helpers -----------------------------------------------------------------

func toLowerRunes(r []rune) []rune {
	out := make([]rune, len(r))
	for i, v := range r {
		out[i] = unicode.ToLower(v)
	}
	return out
}

func hasRunePrefix(s, prefix []rune) bool {
	if len(prefix) > len(s) {
		return false
	}
	for i, r := range prefix {
		if s[i] != r {
			return false
		}
	}
	return true
}

func containsRunes(s, sub []rune) bool {
	if len(sub) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		match := true
		for j, r := range sub {
			if s[i+j] != r {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
