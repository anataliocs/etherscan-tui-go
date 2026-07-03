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

## Development Workflow

### Definition of Done
Before marking a task as completed and submitting your changes, you MUST:
1. **Run Linter**: Ensure the code passes linting by running `make lint` or `golangci-lint run ./...`.
2. **Run Vulnerability Check**: Ensure no known vulnerabilities exist by running `make vulncheck` or `govulncheck ./...`.
3. **Run Unit Tests**: All unit tests must pass. Run them with `make test` or `go test ./... -v`.
4. **Run E2E Tests**: Ensure end-to-end functionality is preserved by running `make test-e2e`.
5. **Go Mod Tidy**: Ensure `go.mod` and `go.sum` are up to date by running `go mod tidy && go mod verify`.

Failure to pass any of these checks means the task is NOT done.

## Project-Specific Guidelines

### JSON Struct Tags
- Use `omitzero` instead of `omitempty` for types like `time.Duration`, `time.Time`, structs, slices, and maps.

### Naming
- Follow standard Go naming conventions (`CamelCase`, short names for local variables).
- Internal packages (`internal/`) should be used for code that is not intended for public use.

### Dependencies
- Keep dependencies minimal.
- Use `go mod tidy && go mod verify` to maintain the `go.mod` and `go.sum` files.

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

## Systems Architecture

### Enforcing Strong Invariants
- PRINCIPLE: Make invalid states unrepresentable instead of defensive coding.
- Do not pass loose primitives (e.g., strings, dicts) for complex domain concepts. Wrap them in strongly typed objects or value types.
- If a variable can be nil, it must be explicitly handled at the boundaries of the system (e.g., API entry points), not deep within the business logic.
- Before adding an 'if' check or a fallback for a bad state, evaluate if the bad state can be prevented entirely by refactoring the input types or function signatures.

### Avoid Excessive Defensiveness
- PRINCIPLE: Fail Fast. If an unexpected state occurs, allow the application to crash or throw a loud, descriptive exception immediately. Do not build "silent fallbacks" or return mock data to paper over an error.

### Prevent Overly Local Reasoning
- Before writing code, locate existing implementations of similar logic across the entire repository. Use grep/search tools to check for duplicated patterns.
- PRINCIPLE: Reusable abstractions that can be shared across the codebase.
- Aim for max behavior per line of code without affecting readability.
- Task Sequence:
    1. Identify the file to modify.
    2. Map out all upstream callers and downstream dependencies of this file.
    3. Write a 2-sentence "Impact Analysis" explaining how your local change affects the broader system architecture.
- If a modification requires changing the same logic in more than two places, stop and propose a shared abstraction or utility function instead.

### THE DRY PRINCIPLE (DON'T REPEAT YOURSELF)

Treat code duplication as a critical bug. NEVER copy-paste existing logic or reinvent existing patterns.

#### 1. Mandatory Discovery Phase
Before writing new code, you MUST explicitly search the codebase.
- Use your grep/search tools to look for keywords related to the task.
- Check common shared directories (e.g., `/utils`, `/helpers`, `/shared`, `/components`, `/models`).
- If a function exists that solves 80% of your problem, do not write a new one. Refactor the existing function to accept a new parameter or handle the extra 20% variation safely.

#### 2. The Rule of Three (Refactoring Trigger)
- If you find yourself writing the exact same or highly similar logic for the second time, a local comment or small helper is acceptable.
- If you find the logic is needed in a third place, you are forbidden from writing it inline. You must stop, extract that logic into a single, reusable utility function or component, and refactor the previous two places to use your new shared implementation.

#### 3. Single Source of Truth
Every piece of knowledge, business logic, or data structure must have a single, unambiguous, authoritative representation within the system.
- Do not duplicate hardcoded strings, magic numbers, status codes, or API routes. Use enums, constants files, or configuration objects.
- Do not duplicate data validation logic. Use a single model schema or validation function at the boundaries.

#### 4. DRY vs. Over-Engineering Guardrail
While avoiding duplication, do not invent overly complex, deeply nested, or hyper-generalized abstractions (e.g., massive generic wrappers) just to save two lines of code. Prefer clean, readable, flat abstractions. If two pieces of code look identical but serve entirely different business domains and change for different reasons, treat them as separate.



