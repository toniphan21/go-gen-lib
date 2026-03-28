package genlib

import (
	"strconv"
	"sync"
)

type VarNameGenerator interface {
	Next() string
	Request(desire string) string
	Reserve(name string, additional ...string)
}

type varNameGenerator struct {
	mu            *sync.Mutex
	defaultPrefix string
	counters      map[string]int
	used          map[string]bool
}

func (g *varNameGenerator) Next() string {
	return g.Request(g.defaultPrefix)
}

func (g *varNameGenerator) Request(desire string) string {
	g.mu.Lock()
	defer g.mu.Unlock()

	count := g.counters[desire]
	for {
		result := desire + strconv.Itoa(count)
		count++
		if _, used := g.used[result]; !used {
			g.used[result] = true
			g.counters[desire] = count
			return result
		}
	}
}

func (g *varNameGenerator) Reserve(name string, additional ...string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if name != "" {
		g.used[name] = true
	}
	for _, v := range additional {
		if v != "" {
			g.used[v] = true
		}
	}
}

var _ VarNameGenerator = (*varNameGenerator)(nil)

func NewVarName(prefix string, existing []string) VarNameGenerator {
	used := make(map[string]bool)
	for _, n := range existing {
		used[n] = true
	}
	return &varNameGenerator{
		mu:            new(sync.Mutex),
		counters:      make(map[string]int),
		defaultPrefix: prefix,
		used:          used,
	}
}
