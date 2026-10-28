// SPDX-License-Identifier: Apache-2.0

package producttest_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"go.toshicloud.io/product"
	"go.toshicloud.io/product/producttest"
	"go.toshicloud.io/product/stability"
)

// A product built outside Toshi's repository: a manifest with its switch
// and one call, and an adapter that serves until its context ends.
func TestRunPassesAProductBuiltElsewhere(t *testing.T) {
	producttest.Run(t, product.Product{
		Manifest: product.Manifest{
			ID:        "notes",
			Version:   "0.1.0",
			Stability: stability.Of(stability.Alpha),
			Procedures: []product.Procedure{
				{Name: "/example.notes.v1.NoteService/ListNotes", Permission: "notes.read"},
			},
			Features: []product.Feature{
				{Key: product.EnabledKey("notes"), Type: product.FeatureBool, Default: product.Bool(true)},
			},
		},
		Hosted: func(ctx context.Context, env product.Env) error {
			srv := &http.Server{Handler: http.NotFoundHandler()}
			go func() {
				<-ctx.Done()
				_ = srv.Close()
			}()
			if err := srv.Serve(env.Listener); !errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return nil
		},
	})
}
