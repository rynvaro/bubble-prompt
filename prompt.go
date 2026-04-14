// Package prompt provides an interactive command-line prompt with tab
// completion, powered by the Bubble Tea TUI framework.
//
// Basic usage:
//
//	p := prompt.New(myExecutor, myCompleter,
//	    prompt.WithPrefix("$ "),
//	)
//	if err := p.Run(); err != nil {
//	    log.Fatal(err)
//	}
package prompt

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// ErrExit is a sentinel error that an Executor can return to signal that the
// prompt loop should exit cleanly.  No error message is printed.
//
//	example:
//	    func executor(input string) error {
//	        if input == "exit" { return prompt.ErrExit }
//	        ...
//	    }
var ErrExit = errors.New("exit")

// Suggestion is a single completion candidate.
type Suggestion struct {
	// Text is the string that gets inserted when this suggestion is accepted.
	Text string

	// Display is the label shown in the popup.  Defaults to Text when empty.
	Display string

	// Description is shown in the right-hand column of the popup.
	Description string

	// Category is an optional grouping label (used in future versions).
	Category string
}

// Completer is called on every keystroke and should return the suggestions
// that match the current Document.
type Completer func(d Document) []Suggestion

// Executor is called after the user presses Enter.  Any output the executor
// wants to display should be written to os.Stdout (or os.Stderr) directly;
// the TUI is paused between submissions so output appears cleanly in the
// terminal scroll-back.
type Executor func(input string) error

// Prompt is the top-level object.  Create one with New and call Run to start
// the interactive REPL loop.
type Prompt struct {
	model    Model
	executor Executor
}

// New creates a new Prompt.
func New(executor Executor, completer Completer, opts ...Option) *Prompt {
	m := newModel(completer)
	for _, opt := range opts {
		opt(&m)
	}
	return &Prompt{
		model:    m,
		executor: executor,
	}
}

// Run starts the interactive REPL loop.  It blocks until the user quits with
// Ctrl+C or Ctrl+D.
//
// The loop works as follows:
//  1. Start a Bubble Tea program to render the prompt and handle input.
//  2. When the user presses Enter, the program quits and returns the input.
//  3. Run calls the Executor with that input (output goes to the terminal).
//  4. Repeat from step 1 with preserved history.
func (p *Prompt) Run() error {
	for {
		// Take a copy of the model for this iteration so the run is isolated.
		m := p.model

		prog := tea.NewProgram(m, tea.WithInput(os.Stdin), tea.WithOutput(os.Stderr))
		result, err := prog.Run()
		if err != nil {
			return err
		}

		finalModel, ok := result.(Model)
		if !ok {
			return nil
		}

		// User pressed Ctrl+C / Ctrl+D without submitting.
		if finalModel.quitting && !finalModel.submitted {
			return nil
		}

		// Preserve history for the next iteration.
		p.model.history = finalModel.history

		if finalModel.submitted {
			input := finalModel.lastInput

			// Empty input: print a newline so the terminal scrolls down
			// naturally (like a real shell), then show a new prompt.
			if input == "" {
				fmt.Fprintln(os.Stderr)
				continue
			}

			if p.executor != nil {
				if execErr := p.executor(input); execErr != nil {
					if errors.Is(execErr, ErrExit) {
						return nil
					}
					fmt.Fprintln(os.Stderr, "error:", execErr)
				}
			}
		}
	}
}

// Model returns the underlying bubbletea Model so that the prompt can be
// embedded in a larger application:
//
//	type AppModel struct {
//	    prompt prompt.Model
//	    // ...
//	}
func (p *Prompt) Model() Model {
	return p.model
}
