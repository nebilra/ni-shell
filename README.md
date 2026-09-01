# ni-shell — my own shell, built in Go

[![progress-banner](https://backend.codecrafters.io/progress/shell/5dcfa8e5-832c-4963-8a0f-f4f5db7d3f76)](https://app.codecrafters.io/users/nebilra?r=2qF)

I followed the [CodeCrafters "Build Your Own
Shell"](https://app.codecrafters.io/courses/shell/overview) challenge and wrote
a POSIX-ish shell from scratch in Go. This repo is the story of that build, from
the first `$` prompt to a shell with quoting, redirection, and tab completion.

## What the shell can do

A small REPL loop reads a line, parses it, and dispatches:

- **Builtins** — `exit`, `echo`, `pwd`, `cd` (with `~` = home dir), and `type`.
- **External programs** — anything on `PATH` is located and executed.
- **Parsing** — commands are tokenized character-by-character, respecting
  single quotes, double quotes, backslash escapes, and whitespace.
- **Redirection** — `>`, `1>`, `2>`, `2>1` (and append via `>>`) send stdout/stderr
  to files, for both builtins and external commands.
- **Tab completion** — builtins and executables on `PATH` complete on Tab; a second
  Tab lists the matches. Built on the
  [`github.com/chzyer/readline`](https://github.com/chzyer/readline) library.

The implementation lives in [`app/main.go`](app/main.go).

## Running it locally

Requires Go 1.26:

```sh
./ni-shell.sh
```

(The repo's CodeCrafters scripts are unchanged under `.codecrafters/`; `ni-shell.sh`
is the local runner I use instead of the starter's `your_program.sh`.)

## The journey

The git history is the real walkthrough — each commit message names the step:

1. **Prompt & REPL** — print `$`, loop reading lines, report unknown commands.
2. **Builtins** — `exit`, then `echo`, then `type`, then `pwd` and `cd`.
3. **PATH resolution** — several attempts to find executables by scanning `PATH`
   (skipping non-directories, checking the executable bit) before landing on a
   `exec.LookPath`-based approach.
4. **Executing programs** — run external commands found on `PATH`.
5. **Parsing & quoting** — a hand-rolled tokenizer grew support for single quotes,
   double quotes, backslash escaping, and finally became a single-pass parser.
6. **Redirection** — plumbing `>`, `1>`, `2>`, `2>1`, `>>` through a `Command`
   struct, for builtins and external commands alike.
7. **Interactive editing** — swapped the plain reader for `readline`, added Tab
   completion (builtins + `PATH`, shared-prefix completion, and a match listing),
   sorted suggestions, and deduped matches.

Along the way I learned (often by doing it wrong first) about REPL loops, shell
tokenization and quoting rules, `PATH` semantics, stream redirection, and how a
line editor like `readline` drives completion.
