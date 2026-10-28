// SPDX-License-Identifier: Apache-2.0

package product

import (
	"slices"
	"strings"

	"go.toshicloud.io/product/stability"
)

// Manifest is what a product declares: its API, what it needs from the
// core, its config section, what it stores and meters, and its features.
type Manifest struct {
	// ID names the product: its config section [products.<id>] and the
	// prefix of its feature keys and permissions. Lower-case letters,
	// digits and underscores, starting with a letter.
	ID string
	// Name is the product's name as people read it, such as "Runners".
	Name string
	// Description says in a sentence what the product does.
	Description string
	// Version is the product's version.
	Version string
	// Docs is where the product's documentation lives: a URL, or a path
	// in Toshi's repository for a product that ships in it.
	Docs string
	// Stability is how mature the product is. A deployment's
	// products.min_stability turns off a product below it.
	Stability stability.Stability
	// Procedures are the calls of the product's API, each with the
	// permission a caller needs. A call only the operator may make carries
	// PermissionOperator: no workspace permission admits it.
	Procedures []Procedure
	// CoreProcedures are the core's calls the product makes, as procedure
	// names ("/package.Service/Method").
	CoreProcedures []string
	// Requires is what the product cannot run without: a deployment whose
	// build lacks any of it does not start with the product enabled.
	Requires Needs
	// Uses is what the product uses when the build has it.
	Uses Needs
	// ConfigSection is the product's section in the control plane's
	// configuration.
	ConfigSection string
	// Tables are the store tables the product owns. They outlive the
	// product: turning it off or leaving it out of a build keeps them, and
	// only a purge empties them.
	Tables []string
	// Shared are the product's rows in tables it shares with the core or
	// other products, each named by a column and the values it holds. A
	// purge deletes them with the Tables.
	Shared []SharedRows
	// Meters are the usage meters the product reports.
	Meters []string
	// Features are the product's typed keys; <id>.enabled, a bool, is the
	// product's own switch.
	Features []Feature
}

// DisplayName is the product's name as people read it: its Name, or its
// ID when the manifest names none.
func (m Manifest) DisplayName() string {
	if m.Name == "" {
		return m.ID
	}
	return m.Name
}

// APIPath is the path the control plane serves the product's API at,
// "/<package.Service>/" of its first procedure; empty when it has none.
func (m Manifest) APIPath() string {
	for _, p := range m.Procedures {
		if service, _, ok := strings.Cut(strings.TrimPrefix(p.Name, "/"), "/"); ok {
			return "/" + service + "/"
		}
	}
	return ""
}

// Procedure is one call of a product's API.
type Procedure struct {
	// Name is the procedure's name, "/package.Service/Method".
	Name string
	// Permission is what a caller needs: <id>.read or <id>.write, a
	// permission of its own when the call needs less or more, or
	// PermissionOperator.
	Permission string
	// Stop marks a call that stops, cancels or revokes work the product
	// runs. A product that takes no new work still answers it, as it
	// answers its reads, while that work runs.
	Stop bool
}

// Class is what a call of a product's API does to the product's work,
// which decides whether a product that takes no new work answers it.
type Class string

// The classes of a product's calls.
const (
	// ClassNewWork is a call that starts or changes work, or anything else
	// that neither reads nor stops: only a product that takes new work
	// answers it.
	ClassNewWork Class = "new_work"
	// ClassRead is a call that only reads.
	ClassRead Class = "read"
	// ClassStop is a call that stops, cancels or revokes work the product
	// runs.
	ClassStop Class = "stop"
)

// Class returns p's class: ClassStop when p is marked Stop, ClassRead when
// its permission's level is read (<area>.read), and ClassNewWork otherwise.
func (p Procedure) Class() Class {
	switch {
	case p.Stop:
		return ClassStop
	case strings.HasSuffix(p.Permission, ".read"):
		return ClassRead
	default:
		return ClassNewWork
	}
}

// PermissionOperator is the Permission of a call only the operator may
// make, with the operator token.
const PermissionOperator = "operator"

// Permissions returns the permissions m's procedures name, but
// PermissionOperator, sorted and without repeats: those a role may give a
// caller of the product's API.
func (m Manifest) Permissions() []string {
	var out []string
	for _, p := range m.Procedures {
		if p.Permission != PermissionOperator {
			out = append(out, p.Permission)
		}
	}
	slices.Sort(out)
	return slices.Compact(out)
}

// SharedRows names a product's rows in a shared table: the rows of Table
// whose Column holds one of Values, every name the product's rows were
// written under.
type SharedRows struct {
	Table  string
	Column string
	Values []string
}

// Needs are primitives and workload kinds.
type Needs struct {
	Primitives    []Primitive
	WorkloadKinds []WorkloadKind
}

// Primitive is a generic part of the core that products are built from.
// Job, sandbox and service are workload kinds, not primitives.
type Primitive string

// The core's primitives.
const (
	// PrimitiveWorkspaces is workspaces, their projects and environments.
	PrimitiveWorkspaces Primitive = "workspaces"
	// PrimitivePlans is plans, limits and meters.
	PrimitivePlans Primitive = "plans"
	// PrimitiveExecutions is exec and file calls into a running workload.
	PrimitiveExecutions Primitive = "executions"
	// PrimitiveCacheVolumes is cache volumes.
	PrimitiveCacheVolumes Primitive = "cache_volumes"
	// PrimitiveStateVolumes is state volumes.
	PrimitiveStateVolumes Primitive = "state_volumes"
	// PrimitiveArtifacts is artifacts and recordings.
	PrimitiveArtifacts Primitive = "artifacts"
	// PrimitiveChannels is a workload's Link, its forwards and channels.
	PrimitiveChannels Primitive = "channels"
	// PrimitiveSecrets is workspace secrets.
	PrimitiveSecrets Primitive = "secrets"
	// PrimitiveWorkloadIdentity is the workload identity token.
	PrimitiveWorkloadIdentity Primitive = "workload_identity"
	// PrimitiveBindings is bindings between workloads and services.
	PrimitiveBindings Primitive = "bindings"
	// PrimitiveEvents is the event log and its webhooks.
	PrimitiveEvents Primitive = "events"
	// PrimitiveDisplays is display sessions.
	PrimitiveDisplays Primitive = "displays"
	// PrimitiveImages is VM images.
	PrimitiveImages Primitive = "images"
)

// WorkloadKind is a kind of workload the core runs, as the Workspace API
// names it.
type WorkloadKind string

// The workload kinds.
const (
	// WorkloadJob starts, runs its command and ends on its own.
	WorkloadJob WorkloadKind = "job"
	// WorkloadSandbox runs until its owner stops it.
	WorkloadSandbox WorkloadKind = "sandbox"
	// WorkloadService is a long-running service.
	WorkloadService WorkloadKind = "service"
)

// Feature is a typed key a product declares, named <id>.<name>.
type Feature struct {
	// Key is <id>.<name>.
	Key  string
	Type FeatureType
	// Values are an enum's values, from the narrowest to the widest; empty
	// for the other types.
	Values []string
	// Default is a workspace's entitlement when its plan gives the feature
	// no value, of the feature's type.
	Default Value
	// Narrowable reports whether a workspace's settings may narrow it
	// within its entitlements: turn it off, or set a lower limit. A string
	// cannot be narrowed.
	Narrowable bool
	// Max bounds a count, or both ends of a range: no value of the feature
	// is higher. 0 bounds nothing.
	Max int64
	// PlanKey names the plan key a count the feature replaces was read
	// from, such as preview_max_connections: a plan that sets that key and
	// gives the feature no value of its own entitles a workspace to the
	// key's value. Empty for a feature no plan key came before.
	PlanKey string
}

// IsLimit reports whether f is a limit: a count or a range.
func (f Feature) IsLimit() bool { return f.Type == FeatureCount || f.Type == FeatureRange }

// FeatureType is the type of a feature's value: a bool, a count (an upper
// limit), a range of integers, one of an enum's values, or a string.
type FeatureType string

// The feature types.
const (
	FeatureBool   FeatureType = "bool"
	FeatureCount  FeatureType = "count"
	FeatureRange  FeatureType = "range"
	FeatureEnum   FeatureType = "enum"
	FeatureString FeatureType = "string"
)

// EnabledKey is the key of a product's own switch, a bool feature.
func EnabledKey(id string) string { return id + ".enabled" }
