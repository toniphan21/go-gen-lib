package gentest

import (
	"fmt"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nhatp.com/go/gen-lib"
	"nhatp.com/go/gen-lib/file"
)

func Test_Type(t *testing.T) {
	cases := []struct {
		name         string
		typ          string
		dependencies map[string]string // module -> version
		imports      map[string]string // pkg path -> alias
		sourceType   string
	}{
		{name: "string"},
		{name: "*string"},
		{name: "[]string"},
		{name: "map[string]int"},
		{name: "map[string]map[int]string"},
		{name: "map[[]string]int"},
		{name: "time.Time", imports: map[string]string{"time": ""}},
		{name: "context.Context", imports: map[string]string{"context": ""}},
		// ---
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			typ := tc.typ
			if typ == "" {
				typ = tc.name
			}

			sourceType := tc.sourceType
			if sourceType == "" {
				sourceType = typ
			}

			// ---

			lines := []string{
				`package test`,
				``,
			}
			if tc.imports != nil {
				for pkgPath, alias := range tc.imports {
					if alias == "" {
						lines = append(lines, fmt.Sprintf(`import "%s"`, pkgPath))
					} else {
						lines = append(lines, fmt.Sprintf(`import %s "%s"`, alias, pkgPath))
					}
				}
			}
			lines = append(lines, "")

			lines = append(lines, fmt.Sprintf("var target %s", sourceType))

			goMod := &file.GoModFile{Requires: tc.dependencies}
			files := []file.File{
				file.Go("input.go", lines...),
				goMod,
			}

			dir := t.TempDir()
			require.NoError(t, genlib.SetupSourceCode(dir, files))

			deps, err := goMod.DirectDependencies()
			require.NoError(t, err)
			if len(deps) > 0 {
				require.NoError(t, genlib.RunGoGet(dir, deps))
			}

			pkgs, err := genlib.LoadPackages(dir)
			require.NoError(t, err)

			obj := pkgs[0].Types.Scope().Lookup("target")
			require.NotNil(t, obj)

			v, ok := obj.(*types.Var)
			require.True(t, ok)

			result := Type(typ)
			assert.True(t, genlib.IsIdentical(result, v.Type()))
		})
	}
}
