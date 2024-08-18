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

## Usage

Pass in the directory you want to analyze as a parameter.

There are multiple output format available:

* Use `-o json` to output JSON format
* Use `-o csv` to output CSV format
* By default the tool pretty-prints its results on the terminal

The `check` tool communicates status with exit codes:

* 2 means that an error happened during the run
* 1 means that there were violations found and at least one violation wasn't justified
* 0 means that no violations were found or all found violations were justified

## Justification

You can justify violations with a comment directly in code.
Put the justification comment directly above the offending line.

```c
// JUSTIFY(unwanted-imports): it's okay this time, I swear
```

Here `unwanted-imports` is the name of the plugin, you can have multiple plugins separated by commas.
The text after the colon is your comment on why this violation is okay.
The justification comment may only be one line long.
