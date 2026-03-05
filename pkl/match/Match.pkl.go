// Code generated from Pkl module `genlib.match`. DO NOT EDIT.
package match

import (
	"context"

	"github.com/apple/pkl-go/pkl"
)

type Match struct {
}

// LoadFromPath loads the pkl module at the given path and evaluates it into a Match
func LoadFromPath(ctx context.Context, path string) (ret Match, err error) {
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

// Load loads the pkl module at the given source and evaluates it with the given evaluator into a Match
func Load(ctx context.Context, evaluator pkl.Evaluator, source *pkl.ModuleSource) (Match, error) {
	var ret Match
	err := evaluator.EvaluateModule(ctx, source, &ret)
	return ret, err
}
