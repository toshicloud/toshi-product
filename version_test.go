// SPDX-License-Identifier: Apache-2.0

package product_test

import (
	"regexp"
	"testing"

	"go.toshicloud.io/product"
)

func TestVersionIsMajorMinorPatch(t *testing.T) {
	if v := product.Version(); !regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`).MatchString(v) {
		t.Fatalf("VERSION = %q, want MAJOR.MINOR.PATCH", v)
	}
}
