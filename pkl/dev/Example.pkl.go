// Code generated from Pkl module `genlib.Example`. DO NOT EDIT.
package dev

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Example struct {
	Matcher MatcherUsage `pkl:"matcher"`
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Example
func LoadFromPath(ctx context.Context, path string) (ret Example, err error) {
	evaluator, err := pkl.NewEvaluator(ctx, pkl.PreconfiguredOptions)
	if err != nil {
		return ret, err
	}
	defer func() {
		cerr := evaluator.Close()
		if err == nil {
			err = cerr
		}
	}()
	ret, err = Load(ctx, evaluator, pkl.FileSource(path))
	return ret, err
}

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Example
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (Example, error) {
	var ret Example
	err := evaluator.EvaluateModule(ctx, source, &ret)
	return ret, err
}
