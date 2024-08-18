# `check` - a static analysis tool

`check` is a tool for [static analysis](https://en.wikipedia.org/wiki/Static_program_analysis) of code.
It is inspired by [Go vet](https://pkg.go.dev/cmd/vet) and similar tools for other programming languages, like [Clang-Tidy](https://clang.llvm.org/extra/clang-tidy/).
Built as a plugin architecture, it is easy to extend `check` to handle additional rules.
It is designed to handle many different programming languages, allowing for reuse of plugins over multi-language codebases as long as the individual programming languages aren't too different.


## Building

`check` depends on `go-tree-sitter` which is a Go binding for Tree Sitter.
It uses C code, so you'll have to enable CGo and have a C compiler available on your system.

Have an executable with a single name available that invokes `zig cc`, for example (on Windows) `zigcc.bat` next to `zig.exe`, with content `zig cc %*`.

```
set CGO_ENABLED=1
set CC="zigcc"
go build .
```
