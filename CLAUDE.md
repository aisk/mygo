# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**ego** is an experimental Go preprocessor/transpiler that introduces the `?` operator for more concise error handling. It transforms `.ego` source files into standard `.go` files.

Example transformation:
```go
// .ego input
result := someFunc()?

// .go output
result, err := someFunc()
if err != nil {
    return err
}
```

## Development Commands

```bash
# Build
go build -o ego

# Run all tests
go test -v ./...

# Run a specific package's tests
go test -v ./transpiler

# Run a specific test case
go test -v ./transpiler -run TestTranspiler/simple.ego
```

## Usage

```bash
# From stdin
cat input.ego | ./ego > output.go

# Specific files or directories
ego file.ego
ego ./folder          # non-recursive
ego ./...             # recursive
```

## Architecture

The project is a fork/extension of Go's standard library packages (`go/ast`, `go/parser`, etc.). Each sub-package mirrors its stdlib counterpart with ego-specific modifications.

### Data Flow

```
.ego source → parser → AST (with TryExpr nodes) → transpiler → modified AST → printer/format → .go output
```

### Key Extension Points

1. **`ast/ast.go:399-403`** — `TryExpr` struct: new AST node for `expr?` syntax
   ```go
   TryExpr struct {
       X        Expr      // the expression before ?
       Question token.Pos // position of "?"
   }
   ```

2. **`parser/parser.go`** — Extended Go parser that recognizes `?` and wraps the preceding expression in a `TryExpr` node.

3. **`transpiler/transpiler.go`** — Core transformation logic using `astutil.Apply` (two-pass: pre/post visit). Handles three `TryExpr` contexts:
   - `AssignStmt` RHS: `result := f()?` → appends `err` to LHS, inserts `if err != nil` after
   - `ExprStmt`: `f()?` → replaces with `if err := f(); err != nil`
   - `IfStmt` condition: `if f()? > 0` → restructures with nested if for error check first

4. **`containers/stack.go`** — Generic stack used to track the enclosing `FuncType` during AST traversal, needed to generate correct zero-value returns.

### Return Value Generation

`genResults` in `transpiler/transpiler.go` generates the zero-value return expressions for each return type when propagating an error. Supported types: numeric types, `bool`, `string`, `error`, pointer types (`*T`), and qualified types (`pkg.T`). Unhandled types cause a transpile error.

## Testing

Transpiler tests use a file-based fixture pattern in `transpiler/testdata/`:
- Input: `<name>.ego`
- Expected output: `<name>_expected.go`

To add a new test case, create a `<name>.ego` and `<name>_expected.go` pair in that directory — the test runner discovers them automatically.

## Constraints

- `?` is only valid inside functions whose **last** return type is `error`
- `?` applies to the immediately preceding expression (a call that returns `(..., error)`)
- Nested `?` in complex expressions (beyond `BinaryExpr`, `UnaryExpr`, `ParenExpr`, `CallExpr`) may not be handled
