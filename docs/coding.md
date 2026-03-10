---
kind: coding
description: Use this file only when the task explicitly points to repository-specific implementation rules.
---

# Coding Conventions

Use this file only when the task explicitly points to repository-specific implementation rules.

## Runtime Boundaries

- Use `log/slog` for structured logs.
- Do not commit `fmt.Print*` debugging.
- Do not `panic` outside process-boundary code such as `main` or `cmd`.
- Thread inherited `context.Context` through call paths. Do not call `context.Background()` or `context.TODO()` in application code.
- Inject a clock or time source instead of calling `time.Now()` directly.
- Build HTTP requests with context and use an explicit `http.Client`; avoid package-level `http.Get` / `http.Post` helpers.

## Errors and APIs

- Return errors and let the caller decide whether to log them. Do not log and return the same error.
- Use `%w` only when callers should inspect the wrapped error with `errors.Is` or `errors.As`; otherwise wrap with `%v` at the boundary.
- Put `context.Context` first in function signatures and never store it in a struct.
- Prefer synchronous functions. Add concurrency at the caller boundary unless the API is inherently asynchronous.
- Define interfaces in the consuming package and return concrete types from constructors.

## Imports

- Avoid package aliases. Use them only when an import path conflict makes the default name ambiguous.
- When an alias is required, form it from the last two path segments concatenated: `github.com/shuymn/pd/internal/metadata` → `internalmetadata`.
- Do not rename the package itself to sidestep a conflict (e.g. declaring `package internalmetadata`). Prefer an alias over a distorted package name.

## Variable Shadowing

- Fix shadow lint errors by pre-declaring the new variable with `var` and switching `:=` to `=`. Do not rename `err` (e.g., `kindErr`).

```go
// Bad – renames err to avoid shadow
k, kindErr := metadata.ParseKind(*cmd.Kind)

// Bad – re-declares err with :=
k, err := metadata.ParseKind(*cmd.Kind)

// Good – pre-declare k, reuse err
var k metadata.Kind
k, err = metadata.ParseKind(*cmd.Kind)
```

## Comments and Receivers

- Every exported identifier (type, func, var, const) must have a doc comment starting with the identifier name.
- Receiver names must be an abbreviation of the type name (typically its first letter or two). Never use generic names such as `cmd`, `self`, or `this`.
- Use a value receiver only when all methods leave the type immutable and the struct is small. Add a `// NOTE:` comment on the type explaining the rationale and the conditions for switching to a pointer receiver: a mutating method is added, the struct exceeds ~4 fields, or the type appears in hot copy paths.

## Struct Discipline

- Types named `Config`, `Options`, `Params`, `Query`, or `Event` are treated as configuration or data boundaries.
- In non-test code, initialize those structs explicitly so `exhaustruct` can catch drift.
- See [.golangci.yaml](../.golangci.yaml) for the exact enforced bans and exceptions.
