package gentest

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"nhatp.com/go/gen-lib/file"
)

type MarkdownTestCase struct {
	Name              string
	Content           string
	Headers           []string
	SourceFiles       []file.File
	GoldenFiles       []file.File
	GoModFileContent  []byte
	GoSumFileContent  []byte
	PklDevFileContent []byte
}

var mdParser = &markdownParser{}

func ParseMarkdown(data []byte) []MarkdownTestCase {
	roots := mdParser.parseTree(data)
	var allPaths [][]mdBlock
	for _, root := range roots {
		mdParser.depthFirstSearch(root, []mdBlock{}, &allPaths)
	}

	var result []MarkdownTestCase
	for _, path := range allPaths {
		tc := mdParser.parsePath(path)
		if tc != nil {
			result = append(result, *tc)
		}
	}
	return result
}

type mdBlock struct {
	Name     string
	Content  string
	Children []*mdBlock
}

type mdCodeBlock struct {
	Language string
	Content  string
}

type mdCodeBlockInfo struct {
	IsFile           bool
	IsGoldenFile     bool
	Path             string
	RemainingContent string
}

type mdStackItem struct {
	block *mdBlock
	level int
}

type markdownParser struct{}

func (p *markdownParser) parsePath(path []mdBlock) *MarkdownTestCase {
	var names []string
	var contents []string
	for i, v := range path {
		names = append(names, v.Name)
		contents = append(contents, fmt.Sprintf("%v %v\n", strings.Repeat("#", i+2), v.Name))
		contents = append(contents, v.Content)
	}
	name := strings.Join(names, " > ")
	content := strings.Join(contents, "\n")
	codeBlocks := p.parseCodeBlocks([]byte(content))

	tc := MarkdownTestCase{
		Name:    name,
		Headers: names,
		Content: content,
	}
	hasGoldenFile := false
	for _, cb := range codeBlocks {
		switch cb.Language {
		case "go.mod":
			cbc := []byte(cb.Content)
			tc.SourceFiles = append(tc.SourceFiles, file.New("go.mod", cbc))
			tc.GoModFileContent = cbc

		case "go.sum":
			cbc := []byte(cb.Content)
			tc.SourceFiles = append(tc.SourceFiles, file.New("go.sum", cbc))
			tc.GoSumFileContent = cbc

		case "pkl":
			cbi := p.getCodeBlockInfo(cb)
			if !cbi.IsFile {
				tc.PklDevFileContent = []byte(cb.Content)
			} else {
				tc.SourceFiles = append(tc.SourceFiles, file.New(cbi.Path, []byte(cbi.RemainingContent)))
			}

		case "go":
			cbi := p.getCodeBlockInfo(cb)
			if cbi.IsFile {
				tc.SourceFiles = append(tc.SourceFiles, file.New(cbi.Path, []byte(cbi.RemainingContent)))
				continue
			}

			if cbi.IsGoldenFile {
				hasGoldenFile = true
				tc.GoldenFiles = append(tc.GoldenFiles, file.New(cbi.Path, []byte(cbi.RemainingContent)))
			}
		}
	}

	if !hasGoldenFile {
		return nil
	}
	return &tc
}

func (p *markdownParser) parseTree(data []byte) []*mdBlock {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var roots []*mdBlock
	var stack []mdStackItem
	var activeBlock *mdBlock

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// 1. Detect Header Level
		level := 0
		if strings.HasPrefix(trimmedLine, "#") {
			for _, char := range trimmedLine {
				if char == '#' {
					level++
				} else {
					break
				}
			}
			// Headers must be followed by a space (e.g., "## Name")
			if len(trimmedLine) > level && trimmedLine[level] != ' ' {
				level = 0
			}
		}

		// 2. Handle Headers
		if level > 0 {
			name := strings.TrimSpace(trimmedLine[level:])
			newBlock := &mdBlock{Name: name}

			// Pop the stack until we find a parent with a lower level
			for len(stack) > 0 && stack[len(stack)-1].level >= level {
				stack = stack[:len(stack)-1]
			}

			if len(stack) == 0 {
				// This is a top-level block (e.g., the first ## found)
				roots = append(roots, newBlock)
			} else {
				// Add as a child to the current parent on the stack
				parent := stack[len(stack)-1].block
				parent.Children = append(parent.Children, newBlock)
			}

			// Push this new block onto the stack to potentially become a parent
			stack = append(stack, mdStackItem{block: newBlock, level: level})
			activeBlock = newBlock
			continue
		}

		// 3. Handle Content
		// Only attach content if we have an active block and the line isn't empty
		// (This ensures "content only contains itself" because as soon as a
		// new header is found, activeBlock changes).
		if activeBlock != nil {
			if activeBlock.Content != "" {
				activeBlock.Content += "\n"
			}
			activeBlock.Content += line
		}
	}

	return roots
}

func (p *markdownParser) depthFirstSearch(current *mdBlock, currentPath []mdBlock, allPaths *[][]mdBlock) {
	newPath := append(currentPath, *current)

	if len(current.Children) == 0 {
		pathCopy := make([]mdBlock, len(newPath))
		copy(pathCopy, newPath)
		*allPaths = append(*allPaths, pathCopy)
		return
	}

	for _, child := range current.Children {
		p.depthFirstSearch(child, newPath, allPaths)
	}
}

func (p *markdownParser) parseCodeBlocks(data []byte) []mdCodeBlock {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	var blocks []mdCodeBlock
	var currentBlock *mdCodeBlock
	var contentBuilder strings.Builder
	inBlock := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			if !inBlock {
				// Entering a code block
				inBlock = true
				lang := strings.TrimPrefix(trimmed, "```")
				currentBlock = &mdCodeBlock{
					Language: strings.ToLower(strings.TrimSpace(lang)),
				}
				contentBuilder.Reset()
				continue
			}

			if currentBlock != nil {
				// Exiting a code block
				inBlock = false
				currentBlock.Content = contentBuilder.String()
				blocks = append(blocks, *currentBlock)
				currentBlock = nil
				continue
			}
		}

		if inBlock {
			contentBuilder.WriteString(line + "\n")
		}
	}

	return blocks
}

func (p *markdownParser) getCodeBlockInfo(cb mdCodeBlock) mdCodeBlockInfo {
	var info mdCodeBlockInfo

	firstLine := cb.Content
	idx := strings.Index(cb.Content, "\n")
	if idx == -1 {
		return info
	}

	if idx+1 < len(cb.Content) {
		info.RemainingContent = cb.Content[idx+1:]
	}

	firstLine = cb.Content[:idx]

	line := strings.TrimSpace(firstLine)

	const goldenPrefix = "// golden-file:"
	const filePrefix = "// file:"

	if strings.HasPrefix(line, goldenPrefix) {
		info.IsGoldenFile = true
		info.Path = strings.TrimSpace(strings.TrimPrefix(line, goldenPrefix))
		return info
	}

	if strings.HasPrefix(line, filePrefix) {
		info.IsFile = true
		info.Path = strings.TrimSpace(strings.TrimPrefix(line, filePrefix))
		return info
	}

	return info
}
