// SPDX-License-Identifier: Apache-2.0

package product

import (
	"context"
	"log/slog"
	"net"
)

// Product is a product a control plane can serve: its manifest, and the
// adapter the control plane hosts for it, if any. A product's package
// returns its own from a function, and a build names the products it
// serves by those packages and functions.
type Product struct {
	Manifest Manifest
	// Hosted runs the product's adapter in the control plane's process;
	// nil for a product whose calls the core serves itself.
	Hosted Serve
}

// Serve runs a product's adapter in env until ctx ends. It returns nil
// once ctx has ended and the adapter has stopped.
type Serve func(ctx context.Context, env Env) error

// Env is what the control plane hands the adapter it runs in its own
// process: the adapter is the same one a product's own binary serves,
// reached on loopback and reaching the core on loopback, with a product
// token the control plane mints for it.
type Env struct {
	// Listener is the adapter's, bound on a loopback address; the adapter
	// serves plain HTTP on it and closes it when it stops.
	Listener net.Listener
	// CoreURL is where the control plane serves the product's core calls,
	// on loopback.
	CoreURL string
	// Token returns the product token the adapter presents to the core at
	// each call; the control plane rotates it.
	Token func() string
	// Issuer is the assertions' issuer, whose key set is served at its JWKS
	// path on loopback, and Audience the adapter's.
	Issuer   string
	Audience string
	// Decode fills v, a pointer, from the control plane's configuration.
	// Toshi's own products read the control plane's configuration type
	// from it; the control plane holds no settings for a product built
	// elsewhere, and Decode returns an error for any other type.
	Decode func(v any) error
	Log    *slog.Logger
}

// The headers an adapter's call into the core carries besides its bearer
// product token.
const (
	// AssertionHeader carries the forward assertion of the call the
	// control plane passed to the adapter.
	AssertionHeader = "Toshi-Forward-Assertion"
	// NonceHeader carries the adapter process's instance nonce.
	NonceHeader = "Toshi-Adapter-Nonce"
)
