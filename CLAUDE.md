# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**mygo** is an experimental Go preprocessor/transpiler that introduces the `?` operator for more concise error handling. It transforms `.mygo` source files into standard `.go` files.

Example transformation:
```go
// .mygo input
result := someFunc()?

// .go output
result, err := someFunc()
if err != nil {
    return err
}
```

Custom error handling is supported by placing a block after `?`:
```go
// .mygo input
result := someFunc()? {
    return fmt.Errorf("someFunc failed: %w", err)
}

// .go output
result, err := someFunc()
if err != nil {
    return fmt.Errorf("someFunc failed: %w", err)
}
```

## Development Commands

```bash
# Build
go build -o mygo

# Run all tests
go test -v ./...

# Run a specific package's tests
go test -v ./transpiler

# Run a specific test case
go test -v ./transpiler -run TestTranspiler/simple.mygo
```

## Usage

```bash
# From stdin
cat input.mygo | ./mygo > output.go

# Specific files or directories
mygo file.mygo
mygo ./folder          # non-recursive
mygo ./...             # recursive

# Install globally
go install github.com/aisk/mygo@latest
```

## Architecture

The project is a fork/extension of Go's standard library packages (`go/ast`, `go/parser`, etc.). Each sub-package mirrors its stdlib counterpart with mygo-specific modifications.

### Data Flow

```
.mygo source → parser → AST (with TryExpr nodes) → transpiler → modified AST → printer/format → .go output
```

### Key Extension Points

1. **`ast/ast.go:399-404`** — `TryExpr` struct: new AST node for `expr?` syntax and optional custom error handler blocks
   ```go
   TryExpr struct {
       X        Expr       // the expression before ?
       Question token.Pos  // position of "?"
       Handler  *BlockStmt // optional block after ?
   }
   ```

2. **`parser/parser.go`** — Extended Go parser that recognizes `?` and wraps the preceding expression in a `TryExpr` node. If `?` is followed by `{ ... }` in a normal expression context, the block is stored as `TryExpr.Handler`; `if f()? { ... }` remains an if condition/body, not a handler.

3. **`transpiler/transpiler.go`** — Core transformation logic using `astutil.Apply` (two-pass: pre/post visit). Handles three `TryExpr` contexts:
   - `AssignStmt` RHS: `result := f()?` → appends `err` to LHS, inserts `if err != nil` after
   - `ExprStmt`: `f()?` → replaces with `if err := f(); err != nil`
   - `IfStmt` condition: `if f()? > 0` → restructures with nested if for error check first
   - Optional handlers: `result := f()? { return wrap(err) }` uses the provided block as the `if err != nil` body instead of generated zero-value returns

4. **`containers/stack.go`** — Generic stack used to track the enclosing `FuncType` during AST traversal, needed to generate correct zero-value returns.

### Self-Hosting (Bootstrap)

The transpiler is written in mygo itself. `transpiler/transpiler.mygo` is the source of truth, and `transpiler/transpiler.go` is **generated** from it:

```bash
go generate ./transpiler    # runs `go run .. transpiler.mygo` (see transpiler/generate.go)
```

Because `transpiler.go` is generated, edit `transpiler.mygo` and regenerate — direct edits to `transpiler.go` will be overwritten. The `.mygo` version uses the `?` operator (e.g. `expr := genEmptyValueExpr(field)?`) and custom handlers (e.g. `enclosingFunc := getEnclosingFuncType()? { ... }`) which expand to explicit `if err != nil` blocks in the `.go` version. Regenerating requires a working `mygo` binary, so changes that would break transpilation must be applied carefully (and may need the `.go` edited by hand first to bootstrap).

### Return Value Generation

`genResults` in `transpiler/transpiler.go` generates the zero-value return expressions for each return type when propagating an error. Supported types: numeric types, `bool`, `string`, `error`, pointer types (`*T`), and qualified types (`pkg.T`). Unhandled types cause a transpile error.

## Testing

Transpiler tests use a file-based fixture pattern in `transpiler/testdata/`:
- Input: `<name>.mygo`
- Expected output: `<name>_expected.go`

To add a new test case, create a `<name>.mygo` and `<name>_expected.go` pair in that directory — the test runner discovers them automatically.

## Constraints

- `?` is only valid inside functions whose **last** return type is `error`
- `?` applies to the immediately preceding expression (a call that returns `(..., error)`)
- `? { ... }` custom handlers can reference `err`; the block must return or otherwise handle control flow as appropriate
- Nested `?` in complex expressions (beyond `BinaryExpr`, `UnaryExpr`, `ParenExpr`, `CallExpr`) may not be handled
