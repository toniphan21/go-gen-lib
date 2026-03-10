// Package genlib is a generic library for Go code generation that provides
// utilities for creating, managing, and organizing generated source files.

//go:generate find ./pkl -name "*.pkl.go" -type f -delete
//go:generate pkl-gen-go ./pkl/dev/Example.pkl
//go:generate pkl-gen-go ./pkl/gen.pkl
//go:generate pkl-gen-go ./pkl/match.pkl

package genlib
