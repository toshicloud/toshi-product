// SPDX-License-Identifier: Apache-2.0

// Package product defines the contract a product of Toshi implements. A
// product is a part of Toshi a deployment turns on and off, such as
// Runners: it declares a Manifest (its API and the permission each call
// needs, what it requires from the core, the tables it owns, its meters
// and its features) and may bring an adapter the control plane runs in
// its own process (Serve).
//
// A product's package returns its Product from a function, and a
// distribution manifest names that package and function under its
// control plane's products; the builder links them into the control
// plane. A product built in another module requires only this one.
//
// Subpackage stability is how mature a product is, and subpackage
// producttest is the conformance suite every product passes.
package product
