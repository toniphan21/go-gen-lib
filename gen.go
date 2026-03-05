package genlib

//go:generate find ./pkg -name "*.pkl.go" -type f -delete
//go:generate pkl-gen-go ./pkl/gen.pkl
//go:generate pkl-gen-go ./pkl/match.pkl
