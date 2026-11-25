package lint

import (
	"errors"
	"fmt"
	"go/build/constraint"
	"math/bits"
	"runtime"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	tagSeparator      = ","
	ignoreTag         = "ignore"
	toolsTag          = "tools"
	unixTag           = "unix"
	goVersionPrefix   = "go1."
	experimentPrefix  = "goexperiment."
	maxEnumeratedTags = 12
	noMask            = -1
)

func tagSets(
	root string,
	base []string,
) (sets [][]string, err error) {
	paths, err := source.Paths(root)
	if err != nil {
		return nil,
			fmt.Errorf(
				"collect build tags: %w",
				err,
			)
	}

	expressions, err := constraints(paths)
	if err != nil {
		return nil,
			fmt.Errorf(
				"collect build tags: %w",
				err,
			)
	}

	sets = [][]string{base}
	for _, expression := range expressions {
		sets = withCoverage(
			sets,
			expression,
		)
	}

	return sets,
		nil
}

func constraints(
	paths []string,
) (expressions []constraint.Expr, err error) {
	failures := []error{}
	for _, path := range paths {
		expression, constraintErr := source.BuildConstraint(path)
		expressions = append(
			expressions,
			expression,
		)
		failures = append(
			failures,
			constraintErr,
		)
	}

	return expressions,
		errors.Join(failures...)
}

func withCoverage(
	sets [][]string,
	expression constraint.Expr,
) (covered [][]string) {
	isCovered := expression == nil || slices.ContainsFunc(
		sets,
		satisfies(expression),
	)
	if isCovered {
		return sets
	}

	added, found := smallestSatisfying(
		expression,
		sets[0],
		customTags(expression),
	)
	if !found {
		return sets
	}

	return append(
		sets,
		added,
	)
}

func smallestSatisfying(
	expression constraint.Expr,
	base []string,
	candidates []string,
) (chosen []string, found bool) {
	if len(candidates) > maxEnumeratedTags {
		return nil,
			false
	}

	best := noMask
	for mask := range 1 << len(candidates) {
		best = preferred(
			expression,
			base,
			candidates,
			mask,
			best,
		)
	}

	if best == noMask {
		return nil,
			false
	}

	return withMask(
			base,
			candidates,
			best,
		),
		true
}

func preferred(
	expression constraint.Expr,
	base []string,
	candidates []string,
	mask int,
	best int,
) (chosen int) {
	isSmaller := best == noMask ||
		bits.OnesCount(uint(mask)) < bits.OnesCount(uint(best))
	isPreferred := isSmaller && satisfies(expression)(withMask(
		base,
		candidates,
		mask,
	))
	if !isPreferred {
		return best
	}

	return mask
}

func withMask(
	base []string,
	candidates []string,
	mask int,
) (enabled []string) {
	enabled = slices.Clone(base)
	for index, tag := range candidates {
		enabled = append(
			enabled,
			maskedTag(
				tag,
				index,
				mask,
			)...,
		)
	}

	return enabled
}

func maskedTag(
	tag string,
	index int,
	mask int,
) (enabled []string) {
	if mask&(1<<index) == 0 {
		return nil
	}

	return []string{tag}
}

func satisfies(
	expression constraint.Expr,
) (isSatisfied func([]string) bool) {
	return func(
		enabled []string,
	) (satisfied bool) {
		return expression.Eval(enabledTag(enabled))
	}
}

func enabledTag(
	enabled []string,
) (isEnabled func(string) bool) {
	return func(
		tag string,
	) (set bool) {
		return slices.Contains(
			enabled,
			tag,
		) || isHostTag(tag)
	}
}

func isHostTag(
	tag string,
) (host bool) {
	isUnix := tag == unixTag && slices.Contains(
		unixSystems(),
		runtime.GOOS,
	)
	return isUnix || tag == runtime.GOOS || tag == runtime.GOARCH || slices.Contains(
		toolchainTags(),
		tag,
	) || strings.HasPrefix(
		tag,
		goVersionPrefix,
	)
}

func customTags(
	expression constraint.Expr,
) (tags []string) {
	tags = slices.DeleteFunc(
		expressionTags(expression),
		isReservedTag,
	)
	slices.Sort(tags)
	return slices.Compact(tags)
}

func expressionTags(
	expression constraint.Expr,
) (tags []string) {
	switch node := expression.(type) {
	case *constraint.TagExpr:
		return []string{node.Tag}
	case *constraint.NotExpr:
		return expressionTags(node.X)
	case *constraint.AndExpr:
		return slices.Concat(
			expressionTags(node.X),
			expressionTags(node.Y),
		)
	case *constraint.OrExpr:
		return slices.Concat(
			expressionTags(node.X),
			expressionTags(node.Y),
		)
	}

	return nil
}

func isReservedTag(
	tag string,
) (reserved bool) {
	return tag == ignoreTag || tag == toolsTag || isHostTag(tag) || strings.HasPrefix(
		tag,
		goVersionPrefix,
	) || strings.HasPrefix(
		tag,
		experimentPrefix,
	) || slices.Contains(
		platformTags(),
		tag,
	)
}

func envTags(
	value string,
) (tags []string) {
	return slices.DeleteFunc(
		strings.Split(
			value,
			tagSeparator,
		),
		isBlankTag,
	)
}

func isBlankTag(
	tag string,
) (blank bool) {
	return strings.TrimSpace(tag) == ""
}

func unixSystems() (systems []string) {
	return []string{
		"aix",
		"android",
		"darwin",
		"dragonfly",
		"freebsd",
		"hurd",
		"illumos",
		"ios",
		"linux",
		"netbsd",
		"openbsd",
		"solaris",
	}
}

func toolchainTags() (tags []string) {
	return []string{"gc", "cgo"}
}

func platformTags() (tags []string) {
	return slices.Concat(
		unixSystems(),
		[]string{
			"gccgo",
			"js",
			"nacl",
			"plan9",
			"wasip1",
			"windows",
			"zos",
			unixTag,
		},
		[]string{
			"386",
			"amd64",
			"amd64p32",
			"arm",
			"armbe",
			"arm64",
			"arm64be",
			"loong64",
			"mips",
			"mipsle",
			"mips64",
			"mips64le",
			"mips64p32",
			"mips64p32le",
			"ppc",
			"ppc64",
			"ppc64le",
			"riscv",
			"riscv64",
			"s390",
			"s390x",
			"sparc",
			"sparc64",
			"wasm",
		},
	)
}
