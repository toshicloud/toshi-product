// SPDX-License-Identifier: Apache-2.0

// Package producttest is the conformance suite every product passes: its
// manifest is one a control plane registers, and its adapter, when it has
// one, serves on the listener the control plane hands it and stops when
// the control plane stops it.
package producttest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.toshicloud.io/product"
)

// Option changes how Run checks a product.
type Option func(*options)

type options struct {
	decode func(v any) error
}

// WithDecode gives the adapter's Env.Decode. Without it, Decode returns an
// error for every type, as a control plane does for a product built
// outside Toshi's repository.
func WithDecode(decode func(v any) error) Option {
	return func(o *options) { o.decode = decode }
}

// stopWithin bounds how long the suite waits for the adapter to answer and
// to stop, so that a hung adapter fails the suite instead of hanging it.
const stopWithin = 30 * time.Second

// Run checks p: its manifest passes Manifest.Check, and its hosted
// adapter, when it has one, answers HTTP on its listener and returns nil
// once its context ends, with its listener closed.
func Run(t *testing.T, p product.Product, opts ...Option) {
	t.Helper()
	o := options{decode: func(v any) error {
		return fmt.Errorf("product %q: the control plane holds no settings of type %T", p.Manifest.ID, v)
	}}
	for _, opt := range opts {
		opt(&o)
	}

	t.Run("manifest", func(t *testing.T) {
		if err := p.Manifest.Check(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("hosted adapter", func(t *testing.T) {
		if p.Hosted == nil {
			t.Skip("the product has no hosted adapter")
		}
		hosted(t, p, o)
	})
}

// hosted serves p's adapter as a control plane does, against a core that
// answers no call, asks it once over HTTP, and stops it.
func hosted(t *testing.T, p product.Product, o options) {
	core := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(core.Close)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	adapterURL := "http://" + ln.Addr().String()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- p.Hosted(ctx, product.Env{
			Listener: ln,
			CoreURL:  core.URL,
			Token:    func() string { return "producttest-token" },
			Issuer:   core.URL,
			Audience: adapterURL,
			Decode:   o.decode,
			Log:      slog.New(slog.DiscardHandler),
		})
	}()

	if err := answers(adapterURL, done); err != nil {
		t.Fatal(err)
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("the adapter stopped with %v, want nil once its context ended", err)
		}
	case <-time.After(stopWithin):
		t.Fatalf("the adapter did not stop within %s of its context ending", stopWithin)
	}
	if c, err := net.Dial("tcp", ln.Addr().String()); err == nil {
		_ = c.Close()
		t.Fatal("the adapter's listener still accepts connections after it stopped")
	}
}

// answers returns nil once the adapter at url answers an HTTP request with
// any status, and an error when it stops first or does not answer in time.
func answers(url string, done <-chan error) error {
	client := &http.Client{Timeout: time.Second}
	deadline := time.Now().Add(stopWithin)
	for {
		resp, err := client.Get(url + "/")
		if err == nil {
			_ = resp.Body.Close()
			return nil
		}
		select {
		case err := <-done:
			return errors.Join(errors.New("the adapter stopped before it answered"), err)
		default:
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("the adapter did not answer within %s: %w", stopWithin, err)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
