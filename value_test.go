// SPDX-License-Identifier: Apache-2.0

package product

import "testing"

// TestParse: each type's text reads back as the value that wrote it, and
// text that is not a value of the feature is refused with the feature's key.
func TestParse(t *testing.T) {
	size := Feature{Key: "a.size", Type: FeatureEnum, Values: []string{"s", "m", "l"}}
	for _, c := range []struct {
		f    Feature
		want Value
	}{
		{Feature{Key: "a.enabled", Type: FeatureBool}, Bool(false)},
		{Feature{Key: "a.max", Type: FeatureCount}, Count(30)},
		{Feature{Key: "a.window", Type: FeatureRange}, Range(-5, 30)},
		{size, Enum("m")},
		{Feature{Key: "a.region", Type: FeatureString}, String("eu west")},
	} {
		got, err := c.f.Parse(c.want.String())
		if err != nil || got != c.want {
			t.Errorf("Parse(%q) = %+v, %v; want %+v", c.want.String(), got, err, c.want)
		}
	}
	for _, c := range []struct {
		f    Feature
		text string
		want string
	}{
		{Feature{Key: "a.enabled", Type: FeatureBool}, "1", `feature "a.enabled": "1" is not true or false`},
		{Feature{Key: "a.max", Type: FeatureCount}, "-1", `feature "a.max": count -1 is negative`},
		{Feature{Key: "a.window", Type: FeatureRange}, "30..1", `feature "a.window": range 30..1 has its minimum above its maximum`},
		{Feature{Key: "a.window", Type: FeatureRange}, "30", `feature "a.window": "30" is not a range <min>..<max>`},
		{size, "xl", `feature "a.size": "xl" is not one of s, m, l`},
	} {
		if _, err := c.f.Parse(c.text); err == nil || err.Error() != c.want {
			t.Errorf("Parse(%q) = %v, want %q", c.text, err, c.want)
		}
	}
	if err := size.Check(Count(1)); err == nil || err.Error() != `"1" is of type count, not enum` {
		t.Errorf("Check of a count against an enum = %v", err)
	}
}

// TestNarrows: a value narrows an entitlement only when it allows no more.
func TestNarrows(t *testing.T) {
	size := Feature{Key: "a.size", Type: FeatureEnum, Values: []string{"s", "m", "l"}}
	for _, c := range []struct {
		name string
		f    Feature
		v, e Value
		want bool
	}{
		{"off within on", Feature{Type: FeatureBool}, Bool(false), Bool(true), true},
		{"off within off", Feature{Type: FeatureBool}, Bool(false), Bool(false), true},
		{"on beyond off", Feature{Type: FeatureBool}, Bool(true), Bool(false), false},
		{"lower count", Feature{Type: FeatureCount}, Count(5), Count(10), true},
		{"equal count", Feature{Type: FeatureCount}, Count(10), Count(10), true},
		{"higher count", Feature{Type: FeatureCount}, Count(11), Count(10), false},
		{"inner range", Feature{Type: FeatureRange}, Range(2, 8), Range(1, 10), true},
		{"range past the maximum", Feature{Type: FeatureRange}, Range(2, 11), Range(1, 10), false},
		{"range below the minimum", Feature{Type: FeatureRange}, Range(0, 8), Range(1, 10), false},
		{"earlier enum value", size, Enum("s"), Enum("m"), true},
		{"later enum value", size, Enum("l"), Enum("m"), false},
		{"unknown enum value", size, Enum("xs"), Enum("m"), false},
		{"same string", Feature{Type: FeatureString}, String("eu"), String("eu"), true},
		{"other string", Feature{Type: FeatureString}, String("us"), String("eu"), false},
	} {
		if got := c.f.Narrows(c.v, c.e); got != c.want {
			t.Errorf("%s: Narrows(%s, %s) = %t, want %t", c.name, c.v, c.e, got, c.want)
		}
	}
}
