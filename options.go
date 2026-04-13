package prompt

// Option is a functional option for configuring a Prompt.
type Option func(*Model)

// WithPrefix sets a static prompt prefix string (default: ">>> ").
func WithPrefix(s string) Option {
	return func(m *Model) {
		m.prefix = s
		m.dynamicPrefix = nil
	}
}

// WithDynamicPrefix sets a function that is called on every render to supply
// the prefix string (e.g. to show the current time or directory).
func WithDynamicPrefix(fn func() string) Option {
	return func(m *Model) {
		m.dynamicPrefix = fn
	}
}

// WithPlaceholder sets the grey hint text shown when the input is empty.
func WithPlaceholder(s string) Option {
	return func(m *Model) {
		m.placeholder = s
	}
}

// WithHistory pre-populates the command history.
func WithHistory(entries []string) Option {
	return func(m *Model) {
		m.history = newHistory(entries)
	}
}

// WithMaxSuggestions sets the maximum number of completion items visible at
// once (default: 8).
func WithMaxSuggestions(n int) Option {
	return func(m *Model) {
		m.completion.maxItems = n
	}
}

// WithStyles replaces the default styles with a custom Styles value.
func WithStyles(s Styles) Option {
	return func(m *Model) {
		m.styles = s
		m.completion.styles = s
	}
}
