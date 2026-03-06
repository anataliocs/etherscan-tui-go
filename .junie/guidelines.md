# Go Best Practices & Modern Idioms (Go 1.26+)

This project follows modern Go idioms and best practices as of Go 1.26. All contributions should adhere to these guidelines.

## General Principles

- **Simplicity**: Write clear, simple code. Avoid over-engineering.
- **Explicit over Implicit**: Explicit error handling and configuration are preferred.
- **Consistency**: Follow the existing style and patterns within the codebase.

## Modern Go Idioms (Go 1.26+)

### Error Handling
- Use `errors.Is(err, target)` instead of `err == target`.
- Use `errors.AsType[T](err)` for type assertion on errors.
- Wrap errors with `fmt.Errorf("context: %w", err)` to preserve the error chain.
- Use `errors.Join(err1, err2)` to combine multiple errors.

### Types & Generics
- Use `any` instead of `interface{}`.
- Use `new(val)` for pointers to values (e.g., `new(30)`, `new(true)`).
- Use `reflect.TypeFor[T]()` instead of `reflect.TypeOf((*T)(nil)).Elem()`.

### Slices & Maps
- Use `slices.Contains(items, x)` for membership checks.
- Use `slices.Max(items)` and `slices.Min(items)` for finding extremes.
- Use `slices.IndexFunc(items, predicate)` to find an index.
- Use `maps.Clone(m)` and `maps.Copy(dst, src)` for map operations.
- Use `clear(m)` to empty a map.
- Use `maps.Keys(m)` and `maps.Values(m)` iterators.
- Use `slices.Collect(iter)` to build a slice from an iterator.
- Use `slices.Sorted(iter)` for collecting and sorting in one step.

### Loops & Control Flow
- Use `for i := range n` for simple loops from 0 to n-1.
- Use `max(a, b)` and `min(a, b)` built-in functions.
- Use `cmp.Or(a, b, "default")` to select the first non-zero value.
- Use `strings.SplitSeq` and `strings.FieldsSeq` when iterating over split results.

### Concurrency
- Use `wg.Go(fn)` with `sync.WaitGroup` for spawning goroutines.
- Use `atomic.Bool`, `atomic.Int64`, etc., for atomic operations.
- Use `context.WithCancelCause` or `context.WithTimeoutCause` to provide reasons for cancellation.
- Use `context.AfterFunc(ctx, fn)` for cleanup on context cancellation.

### Testing
- Use `t.Context()` when a test needs a context.
- Use `b.Loop()` for the main loop in benchmark functions.

## Project-Specific Guidelines

### JSON Struct Tags
- Use `omitzero` instead of `omitempty` for types like `time.Duration`, `time.Time`, structs, slices, and maps.

### Naming
- Follow standard Go naming conventions (`CamelCase`, short names for local variables).
- Internal packages (`internal/`) should be used for code that is not intended for public use.

### Dependencies
- Keep dependencies minimal.
- Use `go mod tidy` to maintain the `go.mod` and `go.sum` files.

## TUI Architecture & Component Design (Inspired by gh-dash)

### Modular Component Structure
- **Package per Component**: Each reusable TUI element should reside in its own package under `internal/tui/components/`.
- **MVU (Model-View-Update)**: Every component MUST implement its own `Model`, `Update(msg tea.Msg) (Model, tea.Cmd)`, and `View() string`.
- **Centralized Context**: Use a `ProgramContext` struct (e.g., in `internal/tui/context`) to share global state like terminal dimensions, configuration, and shared styles.
- **Context Synchronization**: Components should have an `UpdateProgramContext(*context.ProgramContext)` method to react to global state changes (e.g., window resizing).
- **Delegation**: The main application model should delegate `Update` and `View` calls to its child components.

### Bubble Tea Idiomatic Usage
- **Command Dispatching**: Use `tea.Cmd` for all side effects (API calls, file I/O). Wrap them in custom message types.
- **Message Passing**: Prefer specific message types for internal communication between components.
- **Key Bindings**: Use `github.com/charmbracelet/bubbles/key` to define and manage keyboard shortcuts consistently.
- **Non-Blocking Operations**: Ensure long-running tasks are started as `tea.Cmd` to keep the UI responsive.
- **Functional Options**: Use the "Functional Options" pattern for component constructors to provide clean and extensible APIs.

### Lipgloss & Styling Best Practices
- **Adaptive Colors**: Always use `lipgloss.AdaptiveColor` to support both light and dark terminal themes.
- **Centralized Styles**: Define base styles and theme colors in a central location (e.g., `internal/tui/theme`) and distribute them via the `ProgramContext`.
- **Dynamic Layouts**: Calculate component widths and heights dynamically based on the current terminal size. Avoid hardcoded dimensions.
- **Composition**: Use `lipgloss.JoinVertical` and `lipgloss.JoinHorizontal` for complex layouts instead of manual string concatenation.
- **Style Inheritance**: Build complex styles by extending base styles using `.Inherit()` or by layering properties.
- **Performance**: Pre-define styles as variables where possible to avoid re-allocating them in the `View()` loop.
