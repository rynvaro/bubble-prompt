package prompt

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// completionList is the internal component that stores suggestions and renders
// the popup.
type completionList struct {
	items    []Suggestion
	selected int // index of the highlighted item
	offset   int // scroll offset (first visible item)
	visible  bool
	maxItems int
	styles   Styles
}

func newCompletionList(maxItems int, styles Styles) completionList {
	if maxItems <= 0 {
		maxItems = 8
	}
	return completionList{maxItems: maxItems, styles: styles}
}

// SetItems replaces the suggestion list and resets the selection.
func (c *completionList) SetItems(items []Suggestion) {
	c.items = items
	c.selected = 0
	c.offset = 0
	c.visible = len(items) > 0
}

// IsVisible reports whether the popup should be shown.
func (c *completionList) IsVisible() bool {
	return c.visible && len(c.items) > 0
}

// Close hides the popup.
func (c *completionList) Close() {
	c.visible = false
}

// Next advances the selection by one (wraps around).
func (c *completionList) Next() {
	if !c.IsVisible() {
		return
	}
	c.selected = (c.selected + 1) % len(c.items)
	c.clampOffset()
}

// Prev moves the selection back one (wraps around).
func (c *completionList) Prev() {
	if !c.IsVisible() {
		return
	}
	c.selected = (c.selected - 1 + len(c.items)) % len(c.items)
	c.clampOffset()
}

func (c *completionList) clampOffset() {
	if c.selected < c.offset {
		c.offset = c.selected
	} else if c.selected >= c.offset+c.maxItems {
		c.offset = c.selected - c.maxItems + 1
	}
}

// Selected returns the currently highlighted suggestion, or nil.
func (c *completionList) Selected() *Suggestion {
	if !c.IsVisible() {
		return nil
	}
	return &c.items[c.selected]
}

// View renders the completion popup as a string.
func (c *completionList) View() string {
	if !c.IsVisible() {
		return ""
	}

	end := c.offset + c.maxItems
	if end > len(c.items) {
		end = len(c.items)
	}
	window := c.items[c.offset:end]

	// Measure column widths for alignment.
	maxText := 0
	maxDesc := 0
	for _, s := range window {
		display := displayText(s)
		if w := lipgloss.Width(display); w > maxText {
			maxText = w
		}
		if w := lipgloss.Width(s.Description); w > maxDesc {
			maxDesc = w
		}
	}
	if maxDesc > 32 {
		maxDesc = 32
	}

	rows := make([]string, 0, len(window)+1)
	for i, s := range window {
		absIdx := c.offset + i
		display := displayText(s)

		// Pad/truncate text column.
		textPad := maxText - lipgloss.Width(display)
		textCol := display + strings.Repeat(" ", textPad)

		// Build row content.
		var content string
		if maxDesc > 0 {
			desc := truncateRunes(s.Description, maxDesc)
			descPad := maxDesc - lipgloss.Width(desc)
			descCol := c.styles.Description.Render(desc + strings.Repeat(" ", descPad))
			content = textCol + "  " + descCol
		} else {
			content = textCol
		}

		if absIdx == c.selected {
			rows = append(rows, c.styles.SelectedItem.Render(content))
		} else {
			rows = append(rows, c.styles.Item.Render(content))
		}
	}

	// Scroll indicator when the list is taller than the window.
	if len(c.items) > c.maxItems {
		indicator := fmt.Sprintf("%d/%d", c.selected+1, len(c.items))
		rows = append(rows, c.styles.Scrollbar.
			Width(maxText+2+maxDesc).
			Align(lipgloss.Right).
			Render(indicator))
	}

	body := strings.Join(rows, "\n")
	return c.styles.CompletionBox.Render(body)
}

// --- helpers -----------------------------------------------------------------

func displayText(s Suggestion) string {
	if s.Display != "" {
		return s.Display
	}
	return s.Text
}

func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}
