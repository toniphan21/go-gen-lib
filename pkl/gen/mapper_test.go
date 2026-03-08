package gen_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nhatp.com/go/gen-lib/pkl/dev"
	"nhatp.com/go/gen-lib/pkl/gen"
)

func Test_ToMatcher(t *testing.T) {
	ctx := context.Background()
	cfg, err := dev.LoadFromPath(ctx, "./testdata/matcher.pkl")
	require.NoError(t, err)

	t.Run("function", func(t *testing.T) {
		t.Run("invalid type", func(t *testing.T) {
			result, err := gen.ToMatcher(cfg.Matcher.InvalidType)
			assert.Nil(t, result)
			assert.Error(t, err)
			assert.Equal(t, "unable to convert to matcher", err.Error())
		})

		t.Run("single, case sensitive", func(t *testing.T) {
			result, err := gen.ToMatcher(cfg.Matcher.Single)
			require.NoError(t, err)

			assert.Equal(t, `match "value" case sensitive`, result.String())
		})

		t.Run("union, case insensitive", func(t *testing.T) {
			result, err := gen.ToMatcher(cfg.Matcher.Union)
			require.NoError(t, err)

			assert.Equal(t, `match "value-ignore-case" ignore case`, result.String())
		})

		t.Run("list", func(t *testing.T) {
			result, err := gen.ToMatcher(cfg.Matcher.List)
			require.NoError(t, err)

			assert.Equal(t, `match "string" case sensitive, match "string" ignore case, match by regex "regex-pattern"`, result.String())
		})

		t.Run("can handle matcher in []any", func(t *testing.T) {
			var input []any
			for _, i := range cfg.Matcher.List {
				input = append(input, i)
			}

			result, err := gen.ToMatcher(input)
			require.NoError(t, err)

			assert.Equal(t, `match "string" case sensitive, match "string" ignore case, match by regex "regex-pattern"`, result.String())
		})
	})

	t.Run("method", func(t *testing.T) {
		t.Run("string, case sensitive", func(t *testing.T) {
			result, err := cfg.Matcher.Single.ToMatcher()
			require.NoError(t, err)

			assert.Equal(t, `match "value" case sensitive`, result.String())
		})

		t.Run("string, case insensitive", func(t *testing.T) {
			result, err := cfg.Matcher.List[1].ToMatcher()
			require.NoError(t, err)

			assert.Equal(t, `match "string" ignore case`, result.String())
		})

		t.Run("regex", func(t *testing.T) {
			result, err := cfg.Matcher.List[2].ToMatcher()
			require.NoError(t, err)

			assert.Equal(t, `match by regex "regex-pattern"`, result.String())
		})
	})
}
