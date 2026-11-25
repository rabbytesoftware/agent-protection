# Agent Protection

Deterministic, machine-enforced coding-standards gate for AI coding agents.
One universal manifest per language (Go, with Rust/TypeScript growing),
applied identically to every repo that uses that language; no per-repo
or per-project-archetype variants.

## Why

AI agents reliably drift from coding standards given only as prose
(READMEs, CLAUDE.md). Agent Protection converts the mechanically-checkable
parts of a coding manifest into hard CI/pre-commit gates instead, and
explicitly excludes from its manifest anything that cannot be checked with
certainty; no judgment-call rules, no LLM review step, no suppression
mechanism an agent could use to bypass a rule.

## Usage

One `ap` process == one language == one component folder, matching how
this is wired in CI: a separate job runs `ap check .` inside a Go module's
root, inside a Cargo.toml directory, inside a package.json directory --
never one process scanning a whole multi-language monorepo.

```sh
ap check .
```

Each violated rule is reported once with what it requires, why (WHY), how
to fix it (FIX), and which changes silence it without fixing it (NOT A
FIX), such as moving implementation into a public sibling package, or
splitting an interface into pieces no consumer asks for. Agents follow the
report text, so the report also says what does not count as a fix.

A failing check ends with the architecture guide, which is also printed by:

```sh
ap guide
```

It describes the one shape AP enforces: no comments in any language, and
in Go the module model (a package with internal/ children exposes only its
API plus a container.go that wires those children; implementation always
lives under internal/; siblings form one layer).

The three pure-layout Go rules can be repaired automatically instead of by
hand: one parameter per line (R01), one argument per line (R52) and
paragraph layout (R55).

```sh
ap fix .
```

`ap fix` rewrites, in place, the same source and test files `ap check`
reads, and only the ones those three rules flag. It moves line breaks and
nothing else besides running gofmt: the syntax tree keeps its shape, the
result is gofmt-clean, and a second run changes nothing. It prints every rewritten
file and how many there were, and exits non-zero on a parse or write
error. It supports Go only; a Cargo.toml or package.json component is
refused. Every other rule still needs a human, so run `ap check .` after it.

See `.ap.yaml.example` for the one config file a consuming repo owns.

In GitHub Actions, use the composite action. On pull requests it keeps one
comment per component up to date: the violation count, a collapsed section
with the per-rule breakdown and the findings, and a link to the run, whose
`ap-report-*` artifact holds the full `ap-report.json`.

```yaml
permissions:
  contents: read
  pull-requests: write
steps:
  - uses: actions/checkout@v4
  - uses: rabbytesoftware/agent-protection@v0.1.0
    with:
      path: services/api
```

`ap markdown ap-report.json` prints that same comment body locally.

## Status

Go: analyzer and architecture rules plus the golangci-lint proxies,
covering the Rabbyte Go manifest wherever a rule is mechanically checkable.
Code: one parameter, argument and literal element per line (R01, R52), at
most two indentation levels (R02), guard clauses with no else (R03), named
results on every function and no naked returns (R04), no comments of any
kind in any file type, no prose smuggled into directives, strings or struct
tags (R07), one type per file and no local or anonymous struct types (R08),
methods live in their type's file and no numbered files (R66), stateless
types as functions unless they only implement an interface (R11, R12),
errors never dropped, overwritten or left unread (R26), no init or package
state (R24), no panic, log.Fatal or os.Exit outside main (R27), interfaces
and interface-embedding structs capped at 3 methods (R28), no bool flag
parameters even behind a named type (R32), command-query separation
including mutation through field method calls (R33), no hidden side
effects (R34), no content-free or numbered names, exported or not (R56),
gofmt (R62).
Architecture: public packages, module directories and command directories
expose only their API; container.go and main.go only delegate and return
interfaces or types the module declares (R14, R48, R64); shared vocabulary
lives in a declaration-only internal package and is re-exported with
aliases; no reaching into another module's subpackages (R15); siblings form
one layer, judged at the lowest common ancestor so cousin imports count,
and only the parent wires them (R16, R47); exported implementations stay
unexported (R13); vocabulary packages hold declarations only (R57);
interfaces are used and not only embedded (R60); no look-alike, escape
hatch or duplicated package names (R22, R58, R59); production never imports
examples or mocks (R61, R63); generated files are exempt only from style
rules and only with real generator provenance; internal/ boundaries (R19,
R50, R51), naming (R21-R23), 1:1 test files (R42, R43), and file and
directory size caps, the root included (R44, R46).
Agent Protection passes its own gate.
Every language, every file type (shell, SQL, YAML, Makefile, Dockerfile,
JS/TS, Rust, Python, go.mod): no comments (R07), machine directives only.
Rust: unwrap/expect/panic-risk detection through
clippy, excluding inline `#[cfg(test)]` modules.
TypeScript and JavaScript: complexity, magic-number, no-else, no-debug-output,
no-param-reassign, modern-syntax and one-class-per-file proxies via ESLint
core rules.

No rule has ever been weakened or carve-out'd to pass on a specific repo;
a repo failing a rule means the repo needs fixing, not the manifest.
