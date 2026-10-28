// SPDX-License-Identifier: Apache-2.0

package product

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// Value is a feature's value, of the feature's type: Bool for a bool, Count
// for a count, Min and Max for a range, and Text for an enum or a string.
type Value struct {
	Type     FeatureType
	Bool     bool
	Count    int64
	Min, Max int64
	Text     string
}

// Bool returns the bool value b.
func Bool(b bool) Value { return Value{Type: FeatureBool, Bool: b} }

// Count returns the count n.
func Count(n int64) Value { return Value{Type: FeatureCount, Count: n} }

// Range returns the range from lo to hi, both included.
func Range(lo, hi int64) Value { return Value{Type: FeatureRange, Min: lo, Max: hi} }

// Enum returns the enum value s.
func Enum(s string) Value { return Value{Type: FeatureEnum, Text: s} }

// String returns the string value s.
func String(s string) Value { return Value{Type: FeatureString, Text: s} }

// String returns v as text, the form Parse reads: true or false, a count
// such as 30, a range such as 1..30, or an enum's or a string's text.
func (v Value) String() string {
	switch v.Type {
	case FeatureBool:
		return strconv.FormatBool(v.Bool)
	case FeatureCount:
		return strconv.FormatInt(v.Count, 10)
	case FeatureRange:
		return strconv.FormatInt(v.Min, 10) + ".." + strconv.FormatInt(v.Max, 10)
	default:
		return v.Text
	}
}

// Parse reads text, in the form Value.String writes, as a value of f.
func (f Feature) Parse(text string) (Value, error) {
	var v Value
	switch f.Type {
	case FeatureBool:
		if text != "true" && text != "false" {
			return Value{}, fmt.Errorf("feature %q: %q is not true or false", f.Key, text)
		}
		v = Bool(text == "true")
	case FeatureCount:
		n, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return Value{}, fmt.Errorf("feature %q: %q is not a count", f.Key, text)
		}
		v = Count(n)
	case FeatureRange:
		loText, hiText, ok := strings.Cut(text, "..")
		lo, errLo := strconv.ParseInt(loText, 10, 64)
		hi, errHi := strconv.ParseInt(hiText, 10, 64)
		if !ok || errLo != nil || errHi != nil {
			return Value{}, fmt.Errorf("feature %q: %q is not a range <min>..<max>", f.Key, text)
		}
		v = Range(lo, hi)
	case FeatureEnum:
		v = Enum(text)
	case FeatureString:
		v = String(text)
	default:
		return Value{}, fmt.Errorf("feature %q: type %q has no values", f.Key, f.Type)
	}
	if err := f.Check(v); err != nil {
		return Value{}, fmt.Errorf("feature %q: %w", f.Key, err)
	}
	return v, nil
}

// Check refuses v when it is not a value of f: of another type, a negative
// count, a range whose minimum is above its maximum, a count or a range end
// above f's Max, or an enum value f does not list. The error does not name
// f.
func (f Feature) Check(v Value) error {
	switch {
	case v.Type != f.Type:
		return fmt.Errorf("%q is of type %s, not %s", v.String(), v.Type, f.Type)
	case v.Type == FeatureCount && v.Count < 0:
		return fmt.Errorf("count %d is negative", v.Count)
	case v.Type == FeatureCount && f.Max > 0 && v.Count > f.Max:
		return fmt.Errorf("count %d is above the most, %d", v.Count, f.Max)
	case v.Type == FeatureRange && v.Min > v.Max:
		return fmt.Errorf("range %s has its minimum above its maximum", v)
	case v.Type == FeatureRange && f.Max > 0 && v.Max > f.Max:
		return fmt.Errorf("range %s ends above the most, %d", v, f.Max)
	case v.Type == FeatureEnum && !slices.Contains(f.Values, v.Text):
		return fmt.Errorf("%q is not one of %s", v.Text, strings.Join(f.Values, ", "))
	}
	return nil
}

// Narrows reports whether v is within the entitlement e, both values of f:
// false within any bool and true within true only, a count no higher, a
// range inside e's, an enum value no later in f's Values, and a string
// equal to e's.
func (f Feature) Narrows(v, e Value) bool {
	switch f.Type {
	case FeatureBool:
		return !v.Bool || e.Bool
	case FeatureCount:
		return v.Count <= e.Count
	case FeatureRange:
		return e.Min <= v.Min && v.Max <= e.Max
	case FeatureEnum:
		i := slices.Index(f.Values, v.Text)
		return i >= 0 && i <= slices.Index(f.Values, e.Text)
	default:
		return v.Text == e.Text
	}
}
