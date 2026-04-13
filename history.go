package prompt

// history manages the command history and navigation state.
type history struct {
	entries []string
	pos     int    // -1 = not browsing; otherwise index into entries
	tmp     string // the live input saved when the user starts browsing
}

func newHistory(initial []string) history {
	entries := make([]string, len(initial))
	copy(entries, initial)
	return history{pos: -1}
}

// Add appends s to the history. Empty strings and consecutive duplicates are
// ignored. The browsing position is reset.
func (h *history) Add(s string) {
	if s == "" {
		return
	}
	if len(h.entries) > 0 && h.entries[len(h.entries)-1] == s {
		return
	}
	h.entries = append(h.entries, s)
	h.pos = -1
}

// Older moves toward older entries (↑).  current is the live input to save
// when browsing begins.  Returns (entry, true) on success.
func (h *history) Older(current string) (string, bool) {
	if len(h.entries) == 0 {
		return "", false
	}
	if h.pos == -1 {
		h.tmp = current
		h.pos = len(h.entries) - 1
	} else if h.pos > 0 {
		h.pos--
	}
	return h.entries[h.pos], true
}

// Newer moves toward newer entries (↓).  When the end is reached it restores
// the live input that was saved when browsing began.
func (h *history) Newer() (string, bool) {
	if h.pos == -1 {
		return "", false
	}
	if h.pos < len(h.entries)-1 {
		h.pos++
		return h.entries[h.pos], true
	}
	h.pos = -1
	return h.tmp, true
}

// Reset resets the browsing position without modifying entries.
func (h *history) Reset() {
	h.pos = -1
	h.tmp = ""
}
