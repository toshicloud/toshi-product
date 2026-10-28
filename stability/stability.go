// SPDX-License-Identifier: Apache-2.0

// Package stability is how mature a part of a Toshi build is: alpha, beta,
// stable or deprecated. Every product, store, object storage backend, cloud
// launcher and agent kind declares one, `toshi version` and discovery
// report it, and a deployment's products.min_stability turns off a product
// below it.
package stability

import (
	"errors"
	"fmt"
)

// Level is a part's stability level.
type Level string

// The levels, from least to most mature. Deprecated is apart: a deprecated
// part is going away, and Removal says in which version.
const (
	// Alpha may change or go away in any release, and is off in a
	// production build unless the deployment asks for it.
	Alpha Level = "alpha"
	// Beta is complete and tested; its configuration and API may still
	// change between minor releases, with notice in the release notes.
	Beta Level = "beta"
	// Stable changes only compatibly within a major version.
	Stable Level = "stable"
	// Deprecated still works as it did, and is removed in the version its
	// Stability names.
	Deprecated Level = "deprecated"
)

// Levels are the levels, in the order Parse accepts them.
var Levels = []Level{Alpha, Beta, Stable, Deprecated}

// rank orders the levels a minimum compares: a deprecated part ranks with
// a stable one until it is removed.
var rank = map[Level]int{Alpha: 1, Beta: 2, Stable: 3, Deprecated: 3}

// Parse returns the level s names.
func Parse(s string) (Level, error) {
	l := Level(s)
	if _, ok := rank[l]; !ok {
		return "", fmt.Errorf("stability %q: not alpha, beta, stable or deprecated", s)
	}
	return l, nil
}

// ParseMinimum returns the level s names as a minimum: alpha, beta or
// stable.
func ParseMinimum(s string) (Level, error) {
	l, err := Parse(s)
	if err == nil && l == Deprecated {
		err = errors.New(`stability "deprecated": a minimum is alpha, beta or stable`)
	}
	return l, err
}

// Stability is what a part declares.
type Stability struct {
	Level Level
	// Removal is the version a deprecated part is removed in; empty for
	// any other level.
	Removal string
}

// Of returns the stability of a part at level l, which is not Deprecated.
func Of(l Level) Stability { return Stability{Level: l} }

// DeprecatedUntil returns the stability of a deprecated part removed in
// version removal.
func DeprecatedUntil(removal string) Stability {
	return Stability{Level: Deprecated, Removal: removal}
}

// Check refuses a stability with no level or an unknown one, a deprecated
// one without its removal version, and a removal version on any other
// level.
func (s Stability) Check() error {
	if s.Level == "" {
		return errors.New("no stability")
	}
	if _, err := Parse(string(s.Level)); err != nil {
		return err
	}
	switch {
	case s.Level == Deprecated && s.Removal == "":
		return errors.New("stability deprecated: needs the version it is removed in")
	case s.Level != Deprecated && s.Removal != "":
		return fmt.Errorf("stability %s: only a deprecated part names a removal version", s.Level)
	}
	return nil
}

// AtLeast reports whether s meets the minimum level min.
func (s Stability) AtLeast(min Level) bool { return rank[s.Level] >= rank[min] }

// String returns the level, and for a deprecated part its removal version:
// "beta", "deprecated, removed in v0.9.0".
func (s Stability) String() string {
	if s.Level == Deprecated {
		return fmt.Sprintf("%s, removed in %s", s.Level, s.Removal)
	}
	return string(s.Level)
}
