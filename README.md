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

## Status

Go: 16 rules, validated against 3 independent real codebases across
several false-positive-hunting rounds.
Rust: unwrap/expect/panic-risk detection, validated against 2 codebases
(inline `#[cfg(test)]` module exclusion fixed).
TypeScript: complexity/magic-number/no-else/no-debug-output/no-param-reassign/
modern-syntax/one-class-per-file proxies via ESLint core rules, validated
against 2 codebases. No custom structural checks yet (an attempted
boolean-flag-param check was cut after two failed implementations).

No rule has ever been weakened or carve-out'd to pass on a specific repo;
a repo failing a rule means the repo needs fixing, not the manifest.
