package genlib

import (
	"strconv"
	"sync"
)

type NameManager interface {
	Next() string
	Request(desire string) string
	Reserve(name string, additional ...string)
}

type nameManager struct {
	mu            *sync.Mutex
	defaultPrefix string
	counters      map[string]int
	used          map[string]bool
}

func (g *nameManager) Next() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	count := g.counters[g.defaultPrefix]
	for {
		result := g.defaultPrefix + strconv.Itoa(count)
		count++
		if _, used := g.used[result]; !used {
			g.used[result] = true
			g.counters[g.defaultPrefix] = count
			return result
		}
	}
}

func (g *nameManager) Request(desire string) string {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, used := g.used[desire]; !used {
		g.used[desire] = true
		return desire
	}

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

func (g *nameManager) Reserve(name string, additional ...string) {
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

var _ NameManager = (*nameManager)(nil)

func NewNameManager(prefix string, existing []string) NameManager {
	used := make(map[string]bool)
	for _, n := range existing {
		used[n] = true
	}
	return &nameManager{
		mu:            new(sync.Mutex),
		counters:      make(map[string]int),
		defaultPrefix: prefix,
		used:          used,
	}
}
