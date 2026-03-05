// Code generated from Pkl module `genlib.gen`. DO NOT EDIT.
package gen

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Gen struct {
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Gen
func LoadFromPath(ctx context.Context, path string) (ret Gen, err error) {
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Gen
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (Gen, error) {
	var ret Gen
	err := evaluator.EvaluateModule(ctx, source, &ret)
	return ret, err
}
