package prompt

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model is the bubbletea Model for bubble-prompt.  It implements tea.Model and
// can be embedded in a larger application model.
type Model struct {
	// ---- configuration (set via options) ----
	prefix        string
	dynamicPrefix func() string
	placeholder   string
	styles        Styles

	// ---- internal components ----
	textInput  textInput
	completion completionList
	history    history

	// ---- user-provided functions ----
	completer Completer

	// ---- runtime state ----
	width int

	// submitted / quitting flags are read by Prompt.Run() after prog.Run().
	lastInput string
	submitted bool
	quitting  bool
}

func newModel(completer Completer) Model {
	styles := DefaultStyles()
	return Model{
		prefix:     ">>> ",
		styles:     styles,
		textInput:  newTextInput(),
		completion: newCompletionList(8, styles),
		history:    newHistory(nil),
		completer:  completer,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()

	switch k {
	// ---- quit ---------------------------------------------------------------
	case "ctrl+c", "ctrl+d":
		m.quitting = true
		return m, tea.Quit

	// ---- submit / accept ----------------------------------------------------
	case "enter":
		if m.completion.IsVisible() {
			m.acceptCompletion()
		} else {
			return m.submit()
		}
		return m, nil

	// ---- completion ---------------------------------------------------------
	case "tab":
		if m.completion.IsVisible() {
			m.completion.Next()
		} else {
			m.refreshCompletion()
		}
		return m, nil

	case "shift+tab":
		if m.completion.IsVisible() {
			m.completion.Prev()
		}
		return m, nil

	case "esc":
		m.completion.Close()
		return m, nil

	// ---- navigation (context-sensitive: completion OR history) --------------
	case "up":
		if m.completion.IsVisible() {
			m.completion.Prev()
		} else {
			if entry, ok := m.history.Older(m.textInput.Value()); ok {
				m.textInput.SetValue(entry)
			}
		}
		return m, nil

	case "down":
		if m.completion.IsVisible() {
			m.completion.Next()
		} else {
			if entry, ok := m.history.Newer(); ok {
				m.textInput.SetValue(entry)
			}
		}
		return m, nil

	// ---- cursor movement ----------------------------------------------------
	case "left":
		m.textInput.MoveLeft()
		return m, nil

	case "right":
		m.textInput.MoveRight()
		return m, nil

	case "ctrl+left", "alt+b":
		m.textInput.MoveWordLeft()
		return m, nil

	case "ctrl+right", "alt+f":
		m.textInput.MoveWordRight()
		return m, nil

	case "home", "ctrl+a":
		m.textInput.MoveToStart()
		return m, nil

	case "end", "ctrl+e":
		m.textInput.MoveToEnd()
		return m, nil

	// ---- editing ------------------------------------------------------------
	case "backspace":
		m.textInput.Backspace()
		m.refreshCompletion()
		return m, nil

	case "delete":
		m.textInput.Delete()
		m.refreshCompletion()
		return m, nil

	case "ctrl+w":
		m.textInput.DeleteWordBackward()
		m.refreshCompletion()
		return m, nil

	case "ctrl+u":
		m.textInput.DeleteToLineStart()
		m.refreshCompletion()
		return m, nil

	case "ctrl+k":
		m.textInput.DeleteToLineEnd()
		m.refreshCompletion()
		return m, nil

	// ---- rune input ---------------------------------------------------------
	default:
		// tea.KeySpace is space (0x20); msg.Runes is nil for it, so insert
		// the space rune explicitly. tea.KeyRunes covers all other printable chars.
		if msg.Type == tea.KeySpace {
			m.textInput.InsertRune(' ')
			m.refreshCompletion()
		} else if msg.Type == tea.KeyRunes {
			for _, r := range msg.Runes {
				m.textInput.InsertRune(r)
			}
			m.refreshCompletion()
		}
	}
	return m, nil
}

// submit commits the current input and requests a program quit so that
// Prompt.Run() can call the executor and then restart the TUI for the next
// command.
func (m Model) submit() (tea.Model, tea.Cmd) {
	m.lastInput = m.textInput.Value()
	m.history.Add(m.lastInput)
	m.history.Reset()
	m.textInput.Reset()
	m.completion.Close()
	m.submitted = true
	return m, tea.Quit
}

// acceptCompletion replaces the current word with the selected suggestion.
func (m *Model) acceptCompletion() {
	sel := m.completion.Selected()
	if sel == nil {
		return
	}

	doc := m.textInput.Document()
	word := doc.CurrentWord()
	text := []rune(doc.Text)
	cursor := doc.CursorPosition
	wordLen := len([]rune(word))

	// Build the new rune slice: text before word + completion text + text after cursor.
	newRunes := make([]rune, 0, len(text)-wordLen+len([]rune(sel.Text)))
	newRunes = append(newRunes, text[:cursor-wordLen]...)
	newRunes = append(newRunes, []rune(sel.Text)...)
	newRunes = append(newRunes, text[cursor:]...)

	m.textInput.value = newRunes
	m.textInput.cursor = cursor - wordLen + len([]rune(sel.Text))
	m.completion.Close()
}

// refreshCompletion re-runs the completer and updates the popup.
func (m *Model) refreshCompletion() {
	if m.completer == nil {
		m.completion.SetItems(nil)
		return
	}
	suggestions := m.completer(m.textInput.Document())
	m.completion.SetItems(suggestions)
}

func (m Model) currentPrefix() string {
	if m.dynamicPrefix != nil {
		return m.dynamicPrefix()
	}
	return m.prefix
}

// View implements tea.Model.
func (m Model) View() string {
	prefix := m.currentPrefix()
	prefixRendered := m.styles.Prefix.Render(prefix)

	var inputDisplay string
	if m.textInput.Value() == "" && m.placeholder != "" {
		inputDisplay = m.styles.Placeholder.Render(m.placeholder)
	} else {
		before := m.textInput.TextBeforeCursor()
		afterRunes := []rune(m.textInput.TextAfterCursor())

		var cursorChar, rest string
		if len(afterRunes) > 0 {
			cursorChar = lipgloss.NewStyle().Reverse(true).Render(string(afterRunes[0]))
			rest = string(afterRunes[1:])
		} else {
			// Cursor is at end of line: show a reversed space.
			cursorChar = lipgloss.NewStyle().Reverse(true).Render(" ")
		}
		inputDisplay = before + cursorChar + rest
	}

	promptLine := prefixRendered + inputDisplay

	if m.completion.IsVisible() {
		indent := lipgloss.Width(prefixRendered)
		popup := lipgloss.NewStyle().MarginLeft(indent).Render(m.completion.View())
		return promptLine + "\n" + popup
	}
	return promptLine
}
