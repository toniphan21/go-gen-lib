package file

import (
	"fmt"
	"strings"

	"golang.org/x/mod/modfile"
)

type GoFile struct {
	Path    string
	Content string
	Lines   []string
}

func (g *GoFile) FilePath() string {
	return g.Path
}

func (g *GoFile) FileContent() []byte {
	if g.Lines == nil {
		return []byte(g.Content)
	}
	return []byte(strings.Join(g.Lines, "\n"))
}

var _ File = (*GoFile)(nil)

func Go(path string, lines ...string) *GoFile {
	return &GoFile{Path: path, Lines: lines}
}

// ---

const DefaultGoModule = "github.com/gen/dev"
const DefaultGoVersion = "1.25"

type GoModFile struct {
	Module   string
	Version  string
	Requires map[string]string // package => version
}

func (g *GoModFile) FilePath() string {
	return "go.mod"
}

func (g *GoModFile) FileContent() []byte {
	var lines []string

	module := g.Module
	if strings.TrimSpace(module) == "" {
		module = DefaultGoModule
	}

	version := g.Version
	if strings.TrimSpace(version) == "" {
		version = DefaultGoVersion
	}

	lines = append(lines, fmt.Sprintf("module %s\n", module))
	lines = append(lines, fmt.Sprintf("go %s\n", version))
	for pkg, ver := range g.Requires {
		lines = append(lines, fmt.Sprintf("require %s %s", pkg, ver))
	}

	return []byte(strings.Join(lines, "\n"))
}

func (g *GoModFile) DirectDependencies() (map[string]string, error) {
	f, err := modfile.Parse("go.mod", g.FileContent(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse go.mod: %w", err)
	}

	dependencies := make(map[string]string)
	for _, req := range f.Require {
		if !req.Indirect {
			dependencies[req.Mod.Path] = req.Mod.Version
		}
	}
	return dependencies, nil
}

var _ File = (*GoModFile)(nil)

// GoMod creates a GoModFile file with the given dependencies and versions.
// Args should be provided as alternating dependency-version pairs, e.g.:
//
// GoMod("dependency-a", "v1.0.0", "dependency-b", "v2.3.1")
//
// Panics if an odd number of arguments is provided.
func GoMod(args ...string) *GoModFile {
	if len(args)%2 != 0 {
		panic("GoMod: args must be provided as alternating dependency-version pairs")
	}

	requires := make(map[string]string) // dep -> version
	for i := 0; i < len(args); i += 2 {
		dep := args[i]
		version := args[i+1]
		requires[dep] = version
	}

	return &GoModFile{Requires: requires}
}

// ---

type GoSumFile struct {
	Lines []string
}

func (g *GoSumFile) FilePath() string {
	return "go.sum"
}

func (g *GoSumFile) FileContent() []byte {
	if g.Lines == nil {
		return []byte("")
	}
	return []byte(strings.Join(g.Lines, "\n"))
}

func GoSum(lines ...string) *GoSumFile {
	return &GoSumFile{Lines: lines}
}

var _ File = (*GoSumFile)(nil)
