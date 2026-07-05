# mygo

![logo](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/tszc13irysyrnvg34lzp.png)

mygo is an experimental *toy* Go preprocessor/transpiler that adds a `?` operator for concise error handling.

For example, `mygo` transforms this:

```go
s := hello()?
```

into standard Go:

```go
s, err := hello()
if err != nil {
    return err
}
```

## Installation

```sh
$ go install github.com/aisk/mygo@latest
```

## Usage

Create `hello.mygo`:

```go
package main

import (
	"io"
	"os"
)

func hello() error {
	f := os.Open("hello.mygo")?
	defer f.Close()
	s := io.ReadAll(f)?
	println(string(s))
	return nil
}

func main() {
	hello()
}
```

`mygo` supports multiple ways to specify transpile targets:

```sh
$ cat hello.mygo | mygo > hello.go
$ mygo hello.mygo
$ mygo .
$ mygo ./...
$ mygo ...
```

## Why

mygo tries to add syntax sugar without adding runtime lock-in. The generated files are plain Go code:

- no runtime dependency
- no special library
- no custom Go compiler

## Constraints

To keep the generated Go readable, mygo intentionally avoids some rewrites:

- method chaining like `a()?.b()?` is not supported
- `?` in `for` initialization statements is not supported yet
- expression statements like `f()?` require `f` to return only `error`; use `_ = f()?` when discarding non-error return values

## TODO

- [ ] Implement a Go-compatible command-line tool that preprocesses `.mygo` files before running standard Go commands
