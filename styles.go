package prompt

import "github.com/charmbracelet/lipgloss"

// Styles holds all lipgloss styles used by bubble-prompt.
// You can replace individual fields to customise the look.
type Styles struct {
	// Prefix is the style applied to the prompt prefix string (e.g. ">>> ").
	Prefix lipgloss.Style

	// Placeholder is applied when the input is empty and a placeholder is set.
	Placeholder lipgloss.Style

	// CompletionBox is the outer container of the completion popup.
	// It typically sets borders, padding and background.
	CompletionBox lipgloss.Style

	// Item is the style for a non-selected completion row.
	Item lipgloss.Style

	// SelectedItem is the style for the currently highlighted completion row.
	SelectedItem lipgloss.Style

	// Description is the style applied to the right-hand description column.
	Description lipgloss.Style

	// Scrollbar is the style for the "n/total" scroll indicator shown when
	// the list is longer than MaxSuggestions.
	Scrollbar lipgloss.Style
}

// DefaultStyles returns the default built-in style set.
func DefaultStyles() Styles {
	accent := lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}
	subtle := lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#6B7280"}
	selectedBg := lipgloss.AdaptiveColor{Light: "#EDE9FE", Dark: "#2E1065"}
	fg := lipgloss.AdaptiveColor{Light: "#111827", Dark: "#F3F4F6"}

	return Styles{
		Prefix: lipgloss.NewStyle().
			Foreground(accent).
			Bold(true),

		Placeholder: lipgloss.NewStyle().
			Foreground(subtle),

		CompletionBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Padding(0, 1),

		Item: lipgloss.NewStyle().
			Foreground(fg),

		SelectedItem: lipgloss.NewStyle().
			Background(selectedBg).
			Foreground(accent).
			Bold(true),

		Description: lipgloss.NewStyle().
			Foreground(subtle),

		Scrollbar: lipgloss.NewStyle().
			Foreground(subtle),
	}
}
