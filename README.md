# go.toshicloud.io/product

The contract every Toshi product implements: the `Manifest` a product
declares and its check, the `Product` a product's package returns, the
`Env` a control plane hands the adapter it hosts, the headers of the
adapter's calls into the core, product stability (`stability`), and
`producttest`, the conformance suite every product passes. A control plane
serves a product through this contract and never depends on a particular
one.

The module depends on nothing of Toshi's own module and imports no API
module, so a product built in another module requires only this one:

```
require go.toshicloud.io/product v0.1.0
```

The module is not served at its import path yet. Until it is, a product
module replaces it with the `product/` directory of a Toshi checkout, and
a distribution built from that checkout names the product's module under
`[[modules]]` with a `dir` (`docs/backends.md` in Toshi's repository,
"Products").

## A product

A product's package has a function that returns its `Product`, and a
distribution manifest names that package and function under
`[control] products`:

```go
func Product() product.Product {
	return product.Product{
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
		Hosted: serve,
	}
}
```

`Hosted`, when set, is the product's adapter, which the control plane runs
in its own process: it serves plain HTTP on `Env.Listener`, calls the
core at `Env.CoreURL` with the token `Env.Token` returns, and returns nil
once its context ends. The control plane holds no settings for a product
built outside Toshi's repository: `Env.Decode` returns an error for it.

## Versions

`VERSION` is the module's version. It is raised when the contract or its
suite changes, once per release of the module, not with every change. A
distribution links one version of this module, the highest any of its
modules requires.

## Conformance

A product runs the suite in its own tests:

```go
func TestConformance(t *testing.T) {
	producttest.Run(t, Product())
}
```

The suite checks that the manifest is one a control plane registers, and,
when the product has an adapter, that the adapter answers HTTP on its
listener and returns nil, with its listener closed, once its context ends.
`producttest.WithDecode` gives the adapter's `Env.Decode`.
