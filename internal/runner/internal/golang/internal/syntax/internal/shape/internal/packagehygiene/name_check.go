package packagehygiene

type nameCheck struct {
	rule     string
	format   string
	violates func(string) bool
}

func checks() (all []nameCheck) {
	return []nameCheck{
		{
			rule:     formatRule,
			format:   "%s %q: rename it to a short lowercase noun without underscores",
			violates: isMalformed,
		},
		{
			rule:     formatRule,
			format:   "%s %q is plural: name it after the one thing it holds, as a singular noun",
			violates: isPlural,
		},
		{
			rule:     genericRule,
			format:   "%s %q is a generic dumping ground: name it after what it does",
			violates: isGeneric,
		},
		{
			rule:     genericRule,
			format:   "%s %q is an escape-hatch name: name it after what it does, not after how provisional or patched it is",
			violates: isEscapeHatch,
		},
		{
			rule:     genericRule,
			format:   "%s %q is a numbered overflow: merge it back into the package it split from, or name it after what it holds",
			violates: isNumbered,
		},
	}
}
