package file

import "strings"

type PklFile struct {
	Path    string
	Content string
	Lines   []string
}

func (p *PklFile) FilePath() string {
	return p.Path
}

func (p *PklFile) FileContent() []byte {
	if p.Lines == nil {
		return []byte(p.Content)
	}
	return []byte(strings.Join(p.Lines, "\n"))
}

var _ File = (*PklFile)(nil)
