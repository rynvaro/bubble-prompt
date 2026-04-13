package prompt

import "unicode"

// textInput is an internal component that manages the editable text buffer.
// All cursor positions are in rune units (not bytes), so multi-byte Unicode
// characters are handled correctly.
type textInput struct {
	value  []rune
	cursor int // rune index; invariant: 0 <= cursor <= len(value)
}

func newTextInput() textInput {
	return textInput{}
}

// Value returns the current text as a string.
func (t *textInput) Value() string {
	return string(t.value)
}

// SetValue replaces the input text and moves the cursor to the end.
func (t *textInput) SetValue(s string) {
	t.value = []rune(s)
	t.cursor = len(t.value)
}

// Reset clears the input and moves the cursor to position 0.
func (t *textInput) Reset() {
	t.value = t.value[:0]
	t.cursor = 0
}

// TextBeforeCursor returns text to the left of the cursor.
func (t *textInput) TextBeforeCursor() string {
	return string(t.value[:t.cursor])
}

// TextAfterCursor returns text to the right of the cursor.
func (t *textInput) TextAfterCursor() string {
	return string(t.value[t.cursor:])
}

// Document builds a Document snapshot for use by the Completer.
func (t *textInput) Document() Document {
	return Document{
		Text:           string(t.value),
		CursorPosition: t.cursor,
	}
}

// --- Editing -----------------------------------------------------------------

// InsertRune inserts r at the cursor position and advances the cursor.
func (t *textInput) InsertRune(r rune) {
	t.value = append(t.value[:t.cursor:t.cursor], append([]rune{r}, t.value[t.cursor:]...)...)
	t.cursor++
}

// Backspace deletes the rune immediately to the left of the cursor.
func (t *textInput) Backspace() {
	if t.cursor == 0 {
		return
	}
	t.value = append(t.value[:t.cursor-1], t.value[t.cursor:]...)
	t.cursor--
}

// Delete deletes the rune immediately to the right of the cursor.
func (t *textInput) Delete() {
	if t.cursor >= len(t.value) {
		return
	}
	t.value = append(t.value[:t.cursor], t.value[t.cursor+1:]...)
}

// DeleteWordBackward deletes from the cursor leftward to the start of the
// previous word (Ctrl+W / Ctrl+Backspace behaviour).
func (t *textInput) DeleteWordBackward() {
	if t.cursor == 0 {
		return
	}
	end := t.cursor
	// step over trailing whitespace
	for t.cursor > 0 && unicode.IsSpace(t.value[t.cursor-1]) {
		t.cursor--
	}
	// step over word characters
	for t.cursor > 0 && !unicode.IsSpace(t.value[t.cursor-1]) {
		t.cursor--
	}
	t.value = append(t.value[:t.cursor], t.value[end:]...)
}

// DeleteToLineStart deletes from the cursor to the beginning of the line
// (Ctrl+U behaviour).
func (t *textInput) DeleteToLineStart() {
	t.value = t.value[t.cursor:]
	t.cursor = 0
}

// DeleteToLineEnd deletes from the cursor to the end of the line (Ctrl+K).
func (t *textInput) DeleteToLineEnd() {
	t.value = t.value[:t.cursor]
}

// --- Cursor movement ---------------------------------------------------------

// MoveLeft moves the cursor one rune to the left.
func (t *textInput) MoveLeft() {
	if t.cursor > 0 {
		t.cursor--
	}
}

// MoveRight moves the cursor one rune to the right.
func (t *textInput) MoveRight() {
	if t.cursor < len(t.value) {
		t.cursor++
	}
}

// MoveWordLeft moves the cursor to the start of the previous word.
func (t *textInput) MoveWordLeft() {
	for t.cursor > 0 && unicode.IsSpace(t.value[t.cursor-1]) {
		t.cursor--
	}
	for t.cursor > 0 && !unicode.IsSpace(t.value[t.cursor-1]) {
		t.cursor--
	}
}

// MoveWordRight moves the cursor past the end of the next word.
func (t *textInput) MoveWordRight() {
	for t.cursor < len(t.value) && unicode.IsSpace(t.value[t.cursor]) {
		t.cursor++
	}
	for t.cursor < len(t.value) && !unicode.IsSpace(t.value[t.cursor]) {
		t.cursor++
	}
}

// MoveToStart moves the cursor to position 0 (Home / Ctrl+A).
func (t *textInput) MoveToStart() {
	t.cursor = 0
}

// MoveToEnd moves the cursor past the last rune (End / Ctrl+E).
func (t *textInput) MoveToEnd() {
	t.cursor = len(t.value)
}
