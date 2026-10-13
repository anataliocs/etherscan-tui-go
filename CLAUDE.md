# CLAUDE.md

Guidance for Claude Code when working in this repository. Adapted from `.junie/guidelines.md` and the CI pipeline in `.github/workflows/go.yml` — keep the three in sync.

## Project Overview

A Bubble Tea TUI for exploring Ethereum blocks and transactions via the Etherscan API. Go 1.26 (CI pins `1.26.9`).

```
cmd/ethereum-explorer/   # entry point
internal/config/         # configuration (.env, see .env.template)
internal/etherscan/      # Etherscan REST client
internal/model/          # domain types
internal/tui/            # root TUI model
  context/               # ProgramContext (shared global state)
  theme/                 # centralized styles and colors
  components/            # one package per component (block, transaction, header, footer, input, loader, errorview)
test/                    # E2E tests
```

## Commands

Go may not be on `PATH` by default — prefix commands with `export PATH=$PATH:/usr/local/go/bin` if `go` is not found.

| Task | Command |
|---|---|
| All checks | `make all` |
| Build | `go build -v ./...` |
| Lint | `make lint` (`golangci-lint run ./...`, config in `.golangci.yml`) |
| Vulnerability check | `make vulncheck` (`govulncheck ./...`) |
| Unit tests | `make test` (`go test ./... -v`) |
| E2E tests | `make test-e2e` (`go test ./test/... -v`) |
| Tidy modules | `go mod tidy && go mod verify` |

## Definition of Done

CI runs build → golangci-lint → govulncheck → unit tests → E2E tests on every push/PR to `main`. A task is NOT done until all of the following pass locally:

1. `go build -v ./...`
2. `make lint`
3. `make vulncheck`
4. `make test`
5. `make test-e2e`
6. `go mod tidy && go mod verify` (no diff in `go.mod`/`go.sum`)

## General Principles

- **Simplicity**: Write clear, simple code. Avoid over-engineering.
- **Explicit over Implicit**: Prefer explicit error handling and configuration.
- **Consistency**: Follow the existing style and patterns in the codebase.

## Planning

- When adding a new field to a component, verify the Etherscan REST API request actually returns the required value.
- Before modifying a file, map its upstream callers and downstream dependencies, then write a 2-sentence impact analysis explaining how the local change affects the broader architecture.

## Modern Go Idioms (Go 1.26+)

### Error Handling
- `errors.Is(err, target)` instead of `err == target`.
- `errors.AsType[T](err)` for type assertions on errors.
- Wrap with `fmt.Errorf("context: %w", err)` to preserve the chain.
- `errors.Join(err1, err2)` to combine errors.

### Types & Generics
- `any` instead of `interface{}`.
- `new(val)` for pointers to values (e.g. `new(30)`, `new(true)`).
- `reflect.TypeFor[T]()` instead of `reflect.TypeOf((*T)(nil)).Elem()`.

### Slices & Maps
- `slices.Contains`, `slices.Max`, `slices.Min`, `slices.IndexFunc`.
- `maps.Clone`, `maps.Copy`, `clear(m)`.
- `maps.Keys` / `maps.Values` iterators with `slices.Collect` or `slices.Sorted`.

### Loops & Control Flow
- `for i := range n` for simple counted loops.
- Built-in `max(a, b)` / `min(a, b)`.
- `cmp.Or(a, b, "default")` for the first non-zero value.
- `strings.SplitSeq` / `strings.FieldsSeq` when iterating over split results.

### Concurrency
- `wg.Go(fn)` with `sync.WaitGroup`.
- `atomic.Bool`, `atomic.Int64`, etc.
- `context.WithCancelCause` / `context.WithTimeoutCause` to record cancellation reasons.
- `context.AfterFunc(ctx, fn)` for cleanup on cancellation.

### Testing
- `t.Context()` when a test needs a context.
- `b.Loop()` for benchmark loops.

### JSON Struct Tags
- Use `omitzero` instead of `omitempty` for `time.Duration`, `time.Time`, structs, slices, and maps.

### Naming & Dependencies
- Standard Go naming (`CamelCase`, short local names).
- Non-public code lives under `internal/`.
- Keep dependencies minimal.

## TUI Architecture (inspired by gh-dash)

### Components
- **Package per component** under `internal/tui/components/`.
- **MVU**: every component implements its own `Model`, `Update(msg tea.Msg) (Model, tea.Cmd)`, and `View() string`.
- **ProgramContext** (`internal/tui/context`) shares terminal dimensions, config, and styles.
- Components expose `UpdateProgramContext(*context.ProgramContext)` to react to global changes (e.g. resize).
- The root model delegates `Update` and `View` to children.

### Bubble Tea
- All side effects (API calls, I/O) run as `tea.Cmd` and report back via custom message types.
- Prefer specific message types for inter-component communication.
- Define key bindings with `github.com/charmbracelet/bubbles/key`.
- Keep the UI responsive — never block in `Update`.
- Use functional options for component constructors.

### Lipgloss Styling
- Use `lipgloss.AdaptiveColor` for light/dark terminal support.
- Define styles centrally in `internal/tui/theme` and distribute via `ProgramContext`.
- Compute widths/heights from terminal size; no hardcoded dimensions.
- Compose layouts with `lipgloss.JoinVertical` / `lipgloss.JoinHorizontal`, not manual concatenation.
- Extend base styles with `.Inherit()` or layering.
- Pre-define styles as variables; don't allocate them inside `View()`.

## Systems Architecture

### Strong Invariants
- Make invalid states unrepresentable instead of coding defensively.
- Wrap complex domain concepts in strong types rather than passing loose primitives.
- Handle nil at system boundaries (e.g. API entry points), not deep in business logic.
- Before adding an `if` check or fallback for a bad state, consider whether refactoring input types or signatures can prevent it.

### Fail Fast
- On unexpected state, fail loudly with a descriptive error. No silent fallbacks or mock data that hide errors.

### DRY
- **Discover first**: search the repo (especially `internal/tui/components`, `internal/model`, `internal/tui/theme`) for existing logic before writing new code. If something solves 80% of the problem, extend it rather than duplicating it.
- **Rule of Three**: a second occurrence may stay inline or use a small local helper; a third occurrence must be extracted into a shared abstraction, and the earlier call sites refactored to use it. If a change requires editing the same logic in more than two places, stop and propose a shared abstraction.
- **Single source of truth**: no duplicated hardcoded strings, magic numbers, status codes, API routes, or validation logic — use constants, config, or a single validation function at the boundary.
- **Guardrail**: don't build deep or hyper-generic abstractions to save a couple of lines. Code that looks alike but serves different domains and changes for different reasons should stay separate.
