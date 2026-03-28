package genlib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_VarNameGenerate(t *testing.T) {
	t.Run("Next", func(t *testing.T) {
		g := NewVarName("v", nil)

		assert.Equal(t, "v0", g.Next())
		assert.Equal(t, "v1", g.Next())
		assert.Equal(t, "v2", g.Next())
		assert.Equal(t, "v3", g.Next())
		assert.Equal(t, "v4", g.Next())
		assert.Equal(t, "v5", g.Next())
		assert.Equal(t, "v6", g.Next())
		assert.Equal(t, "v7", g.Next())
		assert.Equal(t, "v8", g.Next())
		assert.Equal(t, "v9", g.Next())
		assert.Equal(t, "v10", g.Next())
	})

	t.Run("Reserve", func(t *testing.T) {
		g := NewVarName("v", []string{"v3", "v5"})

		assert.Equal(t, "v0", g.Next())
		assert.Equal(t, "v1", g.Next())
		assert.Equal(t, "v2", g.Next())
		assert.Equal(t, "v4", g.Next())
		assert.Equal(t, "v6", g.Next())

		g.Reserve("v7", "v8", "v9")

		assert.Equal(t, "v10", g.Next())
	})

	t.Run("Request", func(t *testing.T) {
		g := NewVarName("v", []string{"v3", "v5"})

		assert.Equal(t, "v0", g.Next())
		assert.Equal(t, "v1", g.Next())
		assert.Equal(t, "v2", g.Next())
		assert.Equal(t, "v4", g.Next())
		assert.Equal(t, "v6", g.Next())

		g.Reserve("v7", "v8", "v9")

		assert.Equal(t, "v10", g.Request("v1"))
		assert.Equal(t, "v11", g.Next())

		g.Reserve("name2")
	
		assert.Equal(t, "name0", g.Request("name"))
		assert.Equal(t, "name1", g.Request("name"))
		assert.Equal(t, "name3", g.Request("name"))
	})
}
