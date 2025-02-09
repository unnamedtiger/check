# `check` - a static analysis tool

`check` is a tool for [static analysis](https://en.wikipedia.org/wiki/Static_program_analysis) of code.
It is inspired by [Go vet](https://pkg.go.dev/cmd/vet) and similar tools for other programming languages, like [Clang-Tidy](https://clang.llvm.org/extra/clang-tidy/).
Built as a plugin architecture, it is easy to extend `check` to handle additional rules.
It is designed to handle many different programming languages, allowing for reuse of plugins over multi-language codebases as long as the individual programming languages aren't too different.
This allows you to slim down on the additional tools you're running.

## Design Decisions / Limitations

There are two major limitations that a `check` plugin has to contend with.
They are:

* A `check` plugin is passed a representation of the abstract syntax tree of code.
    It's not possible to build a plugin that needs more context, like an already run preprocessor or code generator, information about struct layouts, or similar.
    This means that several subgroups of static analysis tasks can't be implemented with `check`.
* A `check` plugin gets every code file individually in an unspecified order.
    While a plugin can store information gathered from one file, it won't be able to reliably evaluate a holistic view of the entire codebase until the end of the run.

## Architecture

`check` follows a three-part plugin architecture.

* The library `common` does the heavy lifting and provides types and functions for the other parts to use
* The plugins export a `common.Plugin` and report violations
* The main executable `wrapper` collects all plugins with a single method call into an executable, powered by the `common` library

Additionally, the `test` package facilitates tests of the entire system by running the plugins against real code and ensuring

* all violations in the code files part of the testsuite are justified and therefore known
* all justifications are found

See [Testing](#testing) below for more.

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

Here `unwanted-imports` is the tag to look for.
It is name of the plugin, optionally followed by the error code the plugin produced, e.g. `unwanted-imports/E001`.
A single justification can handle multiple tags separated by commas.
The text after the colon is your comment on why this violation is okay.
The justification comment may only be one line long.

## Testing

There are **unit tests** in the `common` library; they are handled like normal in Go.

All test files for the **system tests** go into `test/data`.
The test files are passed to all plugins and resulting violations are collected.
Every violation has to be justified, allowing for self-documenting test cases.
Justification messages have to be unique over all test cases.
Unjustified violations are reported as errors as well as superflous justifications.

All tests (unit tests in the `common` library and system tests in `test`) are run like this together:

```sh
go test ./common ./test
```

To include coverage, use this call:

```sh
go test -v -coverpkg=./... -coverprofile=cov.out ./common ./test && go tool cover -html=cov.out -o=cov.html
```
