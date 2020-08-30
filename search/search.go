// Package search allows for searching a collection of entries.
package search

// Query represents a query to a collection of entries.
type Query []Constraint

// Constraint represents a constraint, something which selectively allows
// some entries and disallows others.
type Constraint struct{}
