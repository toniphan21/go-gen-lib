package gentest

import (
	"embed"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nhatp.com/go/gen-lib/file"
)

func Test_ParseMarkdownCode(t *testing.T) {
	cases := []struct {
		name     string
		content  []string
		expected []MarkdownTestCase
	}{
		{
			name: "simple content with pkl file content",
			content: []string{
				`## MD file can be documentation and golden test as the same time`,
				`let's setup a module`,
				"",
				"```go.mod",
				"module nhatp.com/go/example",
				"go 1.24",
				"```",
				"",
				`and you have a golang source code`,
				"",
				"```go",
				`// file: main.go`,
				`package main`,
				"```",
				"",
				`with configuration`,
				"```pkl",
				"// your pkl config here",
				"```",
				"",
				"the expected output is",
				"```go",
				`// golden-file: gen.go`,
				`package main`,
				"```",
			},
			expected: []MarkdownTestCase{
				{
					SourceFiles: []file.File{
						file.New("go.mod", []byte("module nhatp.com/go/example\ngo 1.24\n")),
						file.New("main.go", []byte("package main\n")),
					},
					GoldenFiles: []file.File{
						file.New("gen.go", []byte("package main\n")),
					},
					GoModFileContent:  []byte("module nhatp.com/go/example\ngo 1.24\n"),
					PklDevFileContent: []byte("// your pkl config here\n"),
				},
			},
		},

		{
			name: "simple content with pkl file",
			content: []string{
				`## MD file can be documentation and golden test as the same time`,
				`let's setup a module`,
				"",
				"```go.mod",
				"module nhatp.com/go/example",
				"go 1.24",
				"```",
				"",
				`and you have a golang source code`,
				"",
				"```go",
				`// file: main.go`,
				`package main`,
				"```",
				"",
				`with configuration`,
				"```pkl",
				"// file: config.pkl",
				"// your pkl config here",
				"```",
				"",
				"the expected output is",
				"```go",
				`// golden-file: gen.go`,
				`package main`,
				"```",
			},
			expected: []MarkdownTestCase{
				{
					SourceFiles: []file.File{
						file.New("go.mod", []byte("module nhatp.com/go/example\ngo 1.24\n")),
						file.New("main.go", []byte("package main\n")),
						file.New("config.pkl", []byte("// your pkl config here\n")),
					},
					GoldenFiles: []file.File{
						file.New("gen.go", []byte("package main\n")),
					},
					GoModFileContent: []byte("module nhatp.com/go/example\ngo 1.24\n"),
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := ParseMarkdown([]byte(strings.Join(tc.content, "\n")))
			var result []MarkdownTestCase
			for _, v := range out {
				item := v
				item.Name = ""
				item.Content = ""
				item.Headers = nil
				result = append(result, item)
			}

			assert.Equal(t, tc.expected, result)
		})
	}
}

//go:embed testdata/*.md
var testData embed.FS

func TestParseMarkdownFiles(t *testing.T) {
	cases := []struct {
		file     string
		expected []MarkdownTestCase
	}{
		{
			file: "testdata/readme.md",
			expected: []MarkdownTestCase{
				{
					SourceFiles: []file.File{
						file.New("go.mod", []byte("module nhatp.com/go/example\ngo 1.24\n")),
						file.New("main.go", []byte("package main\n")),
					},
					GoldenFiles: []file.File{
						file.New("gen.go", []byte("package main\n")),
					},
					GoModFileContent:  []byte("module nhatp.com/go/example\ngo 1.24\n"),
					PklDevFileContent: []byte("// your pkl config here\n"),
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			content, err := testData.ReadFile(tc.file)
			require.NoError(t, err)

			out := ParseMarkdown(content)
			var result []MarkdownTestCase
			for _, v := range out {
				item := v
				item.Name = ""
				item.Content = ""
				item.Headers = nil
				result = append(result, item)
			}

			assert.Equal(t, tc.expected, result)
		})
	}
}
