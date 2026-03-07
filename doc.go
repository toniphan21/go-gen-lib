// Package genlib

//go:generate find ./pkl -name "*.pkl.go" -type f -delete
//go:generate pkl-gen-go ./pkl/dev/Example.pkl
//go:generate pkl-gen-go ./pkl/gen.pkl
//go:generate pkl-gen-go ./pkl/match.pkl

package genlib
