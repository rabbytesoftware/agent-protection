package rules

type Rule struct {
	ID          string `yaml:"id"`
	Level       string `yaml:"level"`
	Text        string `yaml:"text"`
	Enforcement string `yaml:"enforcement"`
	Rationale   string `yaml:"rationale"`
	FixStrategy string `yaml:"fix_strategy"`
}

// Registry is the hand-authored, hardcoded v1 registry. In production this
// loads from an embedded rules.yaml; inlined here as Go structs for the
// pilot build to avoid pulling in a YAML dependency for a one-shot test.
var Registry = map[string]Rule{
	"R32": {
		ID: "R32", Level: "L0", Enforcement: "analyzer",
		Text:      "No boolean flag parameters that silently branch a function's behavior",
		Rationale: "A boolean parameter means the function secretly has two behaviors selected by the caller; this is a fork in behavior disguised as data, and the call site can't tell what true/false means without opening the function.",
		FixStrategy: "Split into two explicitly named functions, or replace the bool with a typed enum describing the real choice.",
	},
	"R33": {
		ID: "R33", Level: "L0", Enforcement: "analyzer",
		Text:      "Command-Query Separation: a function either mutates state or returns a value, never both",
		Rationale: "Mixing \"tell me something\" and \"change something\" in one call makes call sites unpredictable and hard to reason about.",
		FixStrategy: "Split into a Compute/Get query function (no mutation) and a separate Set/Apply/Update command function (no return value beyond error).",
	},
	"R34": {
		ID: "R34", Level: "L0", Enforcement: "analyzer",
		Text:      "No hidden side effects: a function's name must reflect everything observable it does",
		Rationale: "A reader trusts a Get/Find/Compute-prefixed name to be cheap and side-effect-free; violating that is a common source of bugs (accidental writes, accidental caching).",
		FixStrategy: "Rename to reflect the real behavior (Fetch/Load/Sync/Save) or remove the side effect from the function entirely.",
	},
	"R16": {
		ID: "R16", Level: "L1", Enforcement: "arch",
		Text:      "Sibling internal packages never import each other",
		Rationale: "Two packages at the same layer importing each other is undeclared coupling that makes the dependency graph impossible to reason about and blocks independent evolution of either package.",
		FixStrategy: "Extract the shared contract into a third, lower package that both siblings depend on, or merge the two packages if they are genuinely one responsibility.",
	},
	"R42": {
		ID: "R42", Level: "L3", Enforcement: "script",
		Text:      "Every non-test source file has exactly one corresponding _test.go file",
		Rationale: "A source file with no test file has unverified behavior by construction; this is the floor, not a strong guarantee of quality.",
		FixStrategy: "Add a _test.go file covering at least the exported surface of the source file.",
	},
	"R46": {
		ID: "R46", Level: "L0", Enforcement: "script",
		Text:      "A directory (excluding project root) holds at most 8 non-test source files",
		Rationale: "A directory with too many files is a package doing too much work; this is the directory-level analog of a per-file line cap.",
		FixStrategy: "Split into subdirectories by responsibility, never just to get under the count.",
	},
	"PROXY-COMPLEXITY": {
		ID: "PROXY-COMPLEXITY", Level: "L0", Enforcement: "lint",
		Text:      "Function exceeds complexity/length/nesting budget (proxy for SRP / does-one-thing)",
		Rationale: "A function under complexity, length and nesting budgets is structurally incapable of doing many unrelated things.",
		FixStrategy: "Extract named helper functions per branch/step; replace nested conditionals with early returns.",
	},
	"PROXY-MAGIC-NUMBER": {
		ID: "PROXY-MAGIC-NUMBER", Level: "L0", Enforcement: "lint",
		Text:      "Unnamed numeric/string literal used outside a const declaration",
		Rationale: "Unnamed literals hide intent and make the meaning of a threshold invisible at the call site.",
		FixStrategy: "Hoist the literal into a named const with a descriptive identifier.",
	},
	"PROXY-DUPLICATION": {
		ID: "PROXY-DUPLICATION", Level: "L0", Enforcement: "lint",
		Text:      "Code block duplicated 3+ times above a token threshold",
		Rationale: "Two occurrences can be coincidental; three is a pattern, and the point past which drift between copies becomes a real risk.",
		FixStrategy: "Extract the duplicated block into a shared helper, parameterizing the differences.",
	},
	"R03": {
		ID: "R03", Level: "L1", Enforcement: "analyzer",
		Text:      "Guard clauses first, no else, success return last",
		Rationale: "An else branch means the happy path and the exceptional path are nested at the same level instead of the exceptional path exiting early -- harder to scan, one more indentation level to track.",
		FixStrategy: "Invert the condition and return/continue early; keep the function body flat.",
	},
	"R04": {
		ID: "R04", Level: "L1", Enforcement: "analyzer",
		Text:      "Named returns instead of var in the body; never naked return",
		Rationale: "A naked return on a named-result function hides what's actually being returned at the call site that matters most -- the function's exit.",
		FixStrategy: "Write the explicit `return value, err` even though the names are declared in the signature.",
	},
	"R08": {
		ID: "R08", Level: "L1", Enforcement: "analyzer",
		Text:      "One type per file, named after the type",
		Rationale: "A file with multiple types has no single, name-matched home, making it harder for both agents and humans to find or size any one type's code.",
		FixStrategy: "Split into one file per type, named after that type.",
	},
	"R21": {
		ID: "R21", Level: "L1", Enforcement: "analyzer",
		Text:      "Package names are short, lowercase, singular nouns, no underscores",
		Rationale: "A package name is read at every call site (pkg.Func); inconsistent casing or underscores add friction everywhere that name is used.",
		FixStrategy: "Rename the package to a short, lowercase, no-underscore noun.",
	},
	"R22": {
		ID: "R22", Level: "L1", Enforcement: "analyzer",
		Text:      "No util/common/helpers/base/misc packages",
		Rationale: "A generically-named package becomes an unbounded dumping ground with no single responsibility, and its name tells a reader nothing about what's inside.",
		FixStrategy: "Name the package after what it actually does, and split it if it's doing more than one thing.",
	},
	"R23": {
		ID: "R23", Level: "L1", Enforcement: "analyzer",
		Text:      "No naming stutter (vault.Vault is the one allowed exception)",
		Rationale: "pkg.PkgThing repeats information the caller already has from the package qualifier -- it's noise at every call site.",
		FixStrategy: "Drop the repeated package name from the type name.",
	},
	"R25": {
		ID: "R25", Level: "L1", Enforcement: "analyzer",
		Text:      "context.Context is always the first parameter and never stored in a struct",
		Rationale: "A context not positioned first breaks the one universal Go convention every reader expects; storing it in a struct hides its lifetime and invites using a stale or wrong context.",
		FixStrategy: "Move ctx to the first parameter position; pass it through call chains explicitly instead of storing it.",
	},
	"R44": {
		ID: "R44", Level: "L3", Enforcement: "script",
		Text:      "Source files stay under 500-600 lines",
		Rationale: "A file this long is a sign the abstraction is wrong, not that it needs a bigger limit.",
		FixStrategy: "Split by responsibility, not by line count.",
	},
	"R41": {
		ID: "R41", Level: "L3", Enforcement: "script",
		Text:      "No real infrastructure in unit tests by default; use in-memory fakes",
		Rationale: "A unit test that opens a real socket or DB connection is slow, flaky, and stops being a unit test.",
		FixStrategy: "Inject an in-memory fake through the same interface the real implementation satisfies, or move the test to an integration-tagged file.",
	},
	"R24": {
		ID: "R24", Level: "L1", Enforcement: "lint",
		Text:      "No init() and no mutable package-level state",
		Rationale: "init() runs implicitly with no caller-visible trigger, and package-level mutable state is shared, hidden, and hard to reason about under concurrency or tests.",
		FixStrategy: "Move init logic into an explicit constructor; pass state through a struct field instead of a package var.",
	},
	"R26": {
		ID: "R26", Level: "L1", Enforcement: "lint",
		Text:      "Never ignore an error; wrap with %w and compare with errors.Is/As",
		Rationale: "A silently discarded error is a silently discarded failure -- the single most common way real bugs get masked in Go.",
		FixStrategy: "Check the error; wrap it with context via %w if propagating.",
	},
	"R28": {
		ID: "R28", Level: "L1", Enforcement: "lint",
		Text:      "Keep interfaces small (1-3 methods), defined in the consuming package",
		Rationale: "A large interface forces every implementor and every fake to satisfy methods most callers never use, and couples far more callers to far more surface than necessary.",
		FixStrategy: "Split the interface along how it's actually consumed; define each smaller interface where it's used.",
	},
	"TS-PROXY-COMPLEXITY": {
		ID: "TS-PROXY-COMPLEXITY", Level: "L1", Enforcement: "lint",
		Text:      "Function exceeds complexity/length/nesting/param-count budget (TypeScript)",
		Rationale: "A function under these budgets is structurally incapable of doing many unrelated things -- same proxy principle as the Go manifest, applied via ESLint core rules.",
		FixStrategy: "Extract named helper functions per branch/step; replace nested conditionals with early returns.",
	},
	"TS-PROXY-MAGIC-NUMBER": {
		ID: "TS-PROXY-MAGIC-NUMBER", Level: "L1", Enforcement: "lint",
		Text:      "Unnamed numeric literal used outside a const declaration (TypeScript)",
		Rationale: "Unnamed literals hide intent and make the meaning of a threshold invisible at the call site.",
		FixStrategy: "Hoist the literal into a named const with a descriptive identifier.",
	},
	"TS-NO-ELSE": {
		ID: "TS-NO-ELSE", Level: "L0", Enforcement: "lint",
		Text:      "Guard clauses first, no else, success return last (TypeScript)",
		Rationale: "An else branch nests the happy path and exceptional path at the same level instead of exiting early.",
		FixStrategy: "Invert the condition and return early; keep the function body flat.",
	},
	"TS-NO-DEBUG-OUTPUT": {
		ID: "TS-NO-DEBUG-OUTPUT", Level: "L0", Enforcement: "lint",
		Text:      "No console.log/debug output left in source",
		Rationale: "Debug output left behind is dead signal in production logs and usually a sign a real logger call was never added.",
		FixStrategy: "Remove it, or replace with the project's real logging mechanism.",
	},
	"TS-NO-PARAM-REASSIGN": {
		ID: "TS-NO-PARAM-REASSIGN", Level: "L0", Enforcement: "lint",
		Text:      "A function must not reassign or mutate its own parameters",
		Rationale: "Mutating a parameter hides a side effect from the caller, who passed a value expecting it to remain theirs.",
		FixStrategy: "Copy into a local variable before modifying, or return a new value instead of mutating in place.",
	},
	"TS-MODERN-SYNTAX": {
		ID: "TS-MODERN-SYNTAX", Level: "L1", Enforcement: "lint",
		Text:      "Use modern, safe syntax: no var, prefer const, strict equality",
		Rationale: "var's function-scoping and == 's type coercion are both long-documented sources of subtle bugs that let/const and === eliminate entirely.",
		FixStrategy: "Replace var with let/const, and == / != with === / !==.",
	},
	"TS-ONE-TYPE-PER-FILE": {
		ID: "TS-ONE-TYPE-PER-FILE", Level: "L1", Enforcement: "lint",
		Text:      "One class per file, named after it",
		Rationale: "A file with multiple classes has no single, name-matched home, making it harder to find or size any one class's code.",
		FixStrategy: "Split into one file per class, named after that class.",
	},
	"PROXY-GLOBAL-STATE": {
		ID: "PROXY-GLOBAL-STATE", Level: "L1", Enforcement: "lint",
		Text:      "Mutable package-level variable",
		Rationale: "Package-level mutable state is implicit, shared, and hard to reason about under concurrency or testing.",
		FixStrategy: "Pass the state explicitly through a constructor/struct field instead.",
	},
}
