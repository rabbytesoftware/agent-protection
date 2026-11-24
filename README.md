# Agent Protection

Deterministic, machine-enforced coding-standards gate for AI coding agents.
One universal manifest per language (Go, with Rust/TypeScript growing),
applied identically to every repo that uses that language -- no per-repo
or per-project-archetype variants.

## Why

AI agents reliably drift from coding standards given only as prose
(READMEs, CLAUDE.md). Agent Protection converts the mechanically-checkable
parts of a coding manifest into hard CI/pre-commit gates instead, and
explicitly excludes from its manifest anything that cannot be checked with
certainty -- no judgment-call rules, no LLM review step, no suppression
mechanism an agent could use to bypass a rule.

## Usage

One `ap` process == one language == one component folder, matching how
this is wired in CI: a separate job runs `ap check .` inside a Go module's
root, inside a Cargo.toml directory, inside a package.json directory --
never one process scanning a whole multi-language monorepo.

```sh
ap check .
```

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

Go: 32 analyzer and architecture rules plus the golangci-lint proxies,
covering the full Rabbyte Go manifest wherever a rule is mechanically
checkable: one parameter per line (R01), at most two indentation levels
(R02), guard clauses with no else (R03), named results and no naked
returns (R04), the comment policy (R07), one type per file (R08),
stateless types as functions (R11, R12), interfaces over exported structs
(R13), package layering and internal/ boundaries (R16, R19, R47-R51),
naming (R21-R23), no panic (R27), 1:1 test files (R42, R43), file and
directory size caps (R44, R46), and paragraph layout: a blank line after every closed block, checks attached to what they check (R55). Agent Protection passes its own gate.
Rust: unwrap/expect/panic-risk detection through clippy, excluding inline
`#[cfg(test)]` modules.
TypeScript: complexity, magic-number, no-else, no-debug-output,
no-param-reassign, modern-syntax and one-class-per-file proxies via ESLint
core rules.

No rule has ever been weakened or carve-out'd to pass on a specific repo;
a repo failing a rule means the repo needs fixing, not the manifest.
