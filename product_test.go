// SPDX-License-Identifier: Apache-2.0

package product

import "testing"

// Class is what a call does to the product's work: Stop wins over the
// permission, a .read permission reads and every other permission, its own
// and the operator's, starts work.
func TestProcedureClass(t *testing.T) {
	for _, c := range []struct {
		name       string
		permission string
		stop       bool
		want       Class
	}{
		{"a read permission", "runners.read", false, ClassRead},
		{"a write permission", "runners.write", false, ClassNewWork},
		{"a permission of its own", "debug.connect", false, ClassNewWork},
		{"the operator's permission", PermissionOperator, false, ClassNewWork},
		{"stop over a write permission", "runners.write", true, ClassStop},
		{"stop over a read permission", "runners.read", true, ClassStop},
		{"stop with the operator's permission", PermissionOperator, true, ClassStop},
	} {
		t.Run(c.name, func(t *testing.T) {
			p := Procedure{Name: "/toshi.runners.v1.RunnerService/Call", Permission: c.permission, Stop: c.stop}
			if got := p.Class(); got != c.want {
				t.Fatalf("Class of %q (stop %v) = %s, want %s", c.permission, c.stop, got, c.want)
			}
		})
	}
}
