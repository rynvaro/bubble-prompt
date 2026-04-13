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

	// tab completion undo state — saved when Tab is first pressed for a word.
	// tabActive is false whenever the popup is driven by normal typing;
	// it becomes true only after the user explicitly presses Tab, ensuring
	// tabSavedWord/tabSavedCursor are always valid when applyCompletion runs.
	tabActive      bool
	tabSavedWord   string
	tabSavedCursor int

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
		// Popup may be open with text already applied by Tab — just close it
		// and submit whatever is currently in the input.
		m.completion.Close()
		return m.submit()

	// ---- completion ---------------------------------------------------------
	case "tab":
		if !m.completion.IsVisible() {
			m.refreshCompletion()
		}
		if m.completion.IsVisible() {
			if m.tabActive {
				// Already cycling: advance to next item.
				m.completion.Next()
			} else {
				// First Tab for this word (popup may have been opened by
				// typing). Save original word/cursor so Esc can revert.
				m.tabActive = true
				m.tabSavedWord = m.textInput.Document().CurrentWord()
				m.tabSavedCursor = m.textInput.cursor
			}
			m.applyCompletion()
		}
		return m, nil

	case "shift+tab":
		if m.completion.IsVisible() {
			if m.tabActive {
				m.completion.Prev()
			} else {
				m.tabActive = true
				m.tabSavedWord = m.textInput.Document().CurrentWord()
				m.tabSavedCursor = m.textInput.cursor
			}
			m.applyCompletion()
		}
		return m, nil

	case "esc":
		if m.completion.IsVisible() {
			if m.tabActive {
				m.revertCompletion()
				m.tabActive = false
			}
			m.completion.Close()
		}
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

// applyCompletion inserts the currently selected suggestion into the input,
// replacing the word from tabSavedCursor-len(tabSavedWord) to the current
// cursor position.  The popup stays open so the user can keep cycling.
func (m *Model) applyCompletion() {
	sel := m.completion.Selected()
	if sel == nil {
		return
	}
	selRunes := []rune(sel.Text)
	wordStart := m.tabSavedCursor - len([]rune(m.tabSavedWord))
	cur := m.textInput.cursor
	newVal := make([]rune, 0, wordStart+len(selRunes)+len(m.textInput.value)-cur)
	newVal = append(newVal, m.textInput.value[:wordStart]...)
	newVal = append(newVal, selRunes...)
	newVal = append(newVal, m.textInput.value[cur:]...)
	m.textInput.value = newVal
	m.textInput.cursor = wordStart + len(selRunes)
}

// revertCompletion restores the word that was present before Tab was first
// pressed.  Called when the user presses Esc.
func (m *Model) revertCompletion() {
	wordStart := m.tabSavedCursor - len([]rune(m.tabSavedWord))
	cur := m.textInput.cursor
	savedRunes := []rune(m.tabSavedWord)
	newVal := make([]rune, 0, wordStart+len(savedRunes)+len(m.textInput.value)-cur)
	newVal = append(newVal, m.textInput.value[:wordStart]...)
	newVal = append(newVal, savedRunes...)
	newVal = append(newVal, m.textInput.value[cur:]...)
	m.textInput.value = newVal
	m.textInput.cursor = m.tabSavedCursor
}

// refreshCompletion re-runs the completer and updates the popup.
// It also resets tabActive so that the next Tab press saves a fresh baseline.
func (m *Model) refreshCompletion() {
	m.tabActive = false
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
