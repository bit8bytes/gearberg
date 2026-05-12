# CLAUDE.md - AI on a Leash for Go

## Project Architecture

- `cmd/` - Application entrypoints (main packages)
- `pkg/` - Public library code (importable by external projects)
- `internal/` - Private application code (not importable externally)
- Tests live next to the code they test: `foo.go` → `foo_test.go`

## Verification

Run the full verification suite:
```
make verify
```

Individual checks (all must pass):
```
make lint
```

## Go Code Standards

1. **Error handling**: Always wrap errors with context: `fmt.Errorf("funcName: %w", err)`
2. **Naming**: Follow Go conventions. MixedCaps, not underscores. Acronyms are all-caps (HTTP, URL, ID).
3. **Interfaces**: Accept interfaces, return structs. Define interfaces at the consumer, not the provider.
4. **Context**: First parameter when needed. Never store in structs.
5. **Concurrency**: Use channels for communication, mutexes for state. Always run tests with `-race`.
6. **Dependencies**: Use `internal/` for code that shouldn't be imported. Minimize third-party dependencies.
7. **Testing**: Table-driven tests. Property-based tests for pure functions. Test app for integration tests.

## Dependency Policy

- All dependencies pinned to exact versions in `go.sum`
- No deprecated packages
- `govulncheck` must pass clean

## AI-Specific Rules

1. **No tautological tests**: tests must encode expected outputs, not reimplement logic
2. **No hallucinated imports**: verify every dependency exists in the Go module ecosystem
3. **Human review required**: all code requires human review before merge
4. **Acceptance criteria first**: do not write code without Given/When/Then criteria
5. **Explain non-obvious decisions**: comment WHY, not WHAT
6. **Integration tests for multi-component features**: unit tests alone are not sufficient when components interact. Wire up real objects, not mocks, and prove data flows through the actual call chain
7. **No vaporware**: every package must be imported by non-test code. Every database table must have DML in non-test code. Code that compiles but isn't wired into the application is dead code