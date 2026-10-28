// SPDX-License-Identifier: Apache-2.0

package product

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// MinStabilityKey is the key under [products] that holds the deployment's
// minimum stability, which no product may take as its ID.
const MinStabilityKey = "min_stability"

var idPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// Check refuses a manifest with no usable ID, version or stability (an ID
// that is products.min_stability's key is not usable), a procedure with
// no name or permission, a table or shared column that is not a plain SQL
// name, shared rows with no value, a feature that is not the product's own,
// declared twice or not typed, whose default is not one of its values, that
// is a narrowable string, bounded but not a limit or with a negative bound,
// or that falls back to a plan key, or names one twice, other than as a
// count, and one without its <id>.enabled switch.
func (m Manifest) Check() error {
	if !idPattern.MatchString(m.ID) {
		return fmt.Errorf("product %q: an ID is lower-case letters, digits and underscores, starting with a letter", m.ID)
	}
	if m.ID == MinStabilityKey {
		return fmt.Errorf("product %q: the ID is products.%s, the deployment's minimum stability", m.ID, MinStabilityKey)
	}
	var problems []error
	if m.Version == "" {
		problems = append(problems, errors.New("no version"))
	}
	if err := m.Stability.Check(); err != nil {
		problems = append(problems, err)
	}
	for _, p := range m.Procedures {
		if p.Name == "" || p.Permission == "" {
			problems = append(problems, fmt.Errorf("procedure %q: needs a name and a permission", p.Name))
		}
	}
	for _, t := range m.Tables {
		if !idPattern.MatchString(t) {
			problems = append(problems, fmt.Errorf("table %q: not a plain SQL name", t))
		}
	}
	for _, r := range m.Shared {
		if !idPattern.MatchString(r.Table) || !idPattern.MatchString(r.Column) {
			problems = append(problems, fmt.Errorf("shared rows %q.%q: not plain SQL names", r.Table, r.Column))
		}
		if len(r.Values) == 0 || slices.Contains(r.Values, "") {
			problems = append(problems, fmt.Errorf("shared rows %q.%q: no value", r.Table, r.Column))
		}
	}
	enabled := false
	keys, planKeys := map[string]bool{}, map[string]bool{}
	for _, f := range m.Features {
		name, ok := strings.CutPrefix(f.Key, m.ID+".")
		if !ok || name == "" {
			problems = append(problems, fmt.Errorf("feature %q: not named %s.<name>", f.Key, m.ID))
		}
		if keys[f.Key] {
			problems = append(problems, fmt.Errorf("feature %q: declared twice", f.Key))
		}
		keys[f.Key] = true
		switch {
		case f.Max < 0:
			problems = append(problems, fmt.Errorf("feature %q: a negative most, %d", f.Key, f.Max))
		case f.Max > 0 && !f.IsLimit():
			problems = append(problems, fmt.Errorf("feature %q: only a count or a range is bounded", f.Key))
		}
		if f.PlanKey != "" {
			if f.Type != FeatureCount {
				problems = append(problems, fmt.Errorf("feature %q: only a count falls back to a plan key", f.Key))
			}
			if !idPattern.MatchString(f.PlanKey) {
				problems = append(problems, fmt.Errorf("feature %q: plan key %q is not a plain key", f.Key, f.PlanKey))
			}
			if planKeys[f.PlanKey] {
				problems = append(problems, fmt.Errorf("feature %q: plan key %q: claimed by another feature too", f.Key, f.PlanKey))
			}
			planKeys[f.PlanKey] = true
		}
		switch f.Type {
		case FeatureEnum:
			if len(f.Values) == 0 {
				problems = append(problems, fmt.Errorf("feature %q: an enum needs its values", f.Key))
			}
		case FeatureBool, FeatureCount, FeatureRange, FeatureString:
			if len(f.Values) > 0 {
				problems = append(problems, fmt.Errorf("feature %q: values are for an enum only", f.Key))
			}
		default:
			problems = append(problems, fmt.Errorf("feature %q: type %q is not bool, count, range, enum or string", f.Key, f.Type))
			continue
		}
		if f.Default.Type == "" {
			problems = append(problems, fmt.Errorf("feature %q: needs a default", f.Key))
		} else if err := f.Check(f.Default); err != nil {
			problems = append(problems, fmt.Errorf("feature %q: default %w", f.Key, err))
		}
		if f.Type == FeatureString && f.Narrowable {
			problems = append(problems, fmt.Errorf("feature %q: a string cannot be narrowed", f.Key))
		}
		if f.Key == EnabledKey(m.ID) {
			enabled = f.Type == FeatureBool
		}
	}
	if !enabled {
		problems = append(problems, fmt.Errorf("feature %q: missing, or not a bool", EnabledKey(m.ID)))
	}
	if err := errors.Join(problems...); err != nil {
		return fmt.Errorf("product %q: %w", m.ID, err)
	}
	return nil
}
