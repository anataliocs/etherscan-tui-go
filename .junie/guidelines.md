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
