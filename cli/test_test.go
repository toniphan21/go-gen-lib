package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_matchName(t *testing.T) {
	cases := []struct {
		name string
		term string
		want bool
	}{
		{name: "Minimum config > Return zero value", term: "", want: false},
		{name: "Minimum config > Return zero value", term: "drop", want: false},
		{name: "Minimum config > Return zero value", term: "minimum", want: true},
		{name: "Minimum config > Return zero value", term: "config", want: true},
		{name: "Minimum config > Return zero value", term: "return", want: true},
		{name: "Minimum config > Return zero value", term: "zero", want: true},
		{name: "Minimum config > Return zero value", term: "value", want: true},
		{name: "Minimum config > Return zero value", term: "val", want: true},
		{name: "Minimum config > Return zero value", term: "minimum config > return zero value", want: true},
	}

	for _, tc := range cases {
		runner := TestRunner{}
		t.Run(tc.name+" "+tc.term, func(t *testing.T) {
			out := runner.matchName(tc.name, tc.term)
			assert.Equal(t, tc.want, out)
		})
	}
}

func Test_readFileContent(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	runner := &TestRunner{
		FilePathResolver: VanityURLFilePathResolver,
	}
	fileContent, err := runner.readFileContent(filepath.Join(wd, "testdata", "file.md"))
	require.NoError(t, err)
	assert.Equal(t, "# just for testing\n", string(fileContent))

	dirContent, err := runner.readFileContent(filepath.Join(wd, "testdata"))
	assert.Error(t, err)
	assert.Nil(t, dirContent)

	_, err = runner.readFileContent("https://raw.githubusercontent.com/toniphan21/go-mapper-gen/refs/heads/")
	assert.Error(t, err)

	remoteFile, err := runner.readFileContent("https://raw.githubusercontent.com/toniphan21/go-mapper-gen/refs/heads/main/README.md")
	require.NoError(t, err)
	assert.NotNil(t, remoteFile)

	useVanityBase, err := runner.readFileContent("nhatp.com/go/mapper-gen/README.md")
	require.NoError(t, err)
	assert.NotNil(t, useVanityBase)
}

func Test_VanityURLFilePathResolver(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{input: "/root/something", want: "/root/something"},
		{input: "https://other.com/mapper-gen", want: "https://other.com/mapper-gen"},
		{input: "http://nhatp.com/java/mapper-gen", want: "http://nhatp.com/java/mapper-gen"},
		{
			input: "http://nhatp.com/go/mapper-gen",
			want:  "https://raw.githubusercontent.com/toniphan21/go-mapper-gen",
		},
		{
			input: "https://nhatp.com/go/mapper-gen",
			want:  "https://raw.githubusercontent.com/toniphan21/go-mapper-gen",
		},
		{

			input: "nhatp.com/go/mapper-gen",
			want:  "https://raw.githubusercontent.com/toniphan21/go-mapper-gen",
		},
		{
			input: "nhatp.com/go/mapper-gen/README.md",
			want:  "https://raw.githubusercontent.com/toniphan21/go-mapper-gen/refs/heads/main/README.md",
		},
		{
			input: "nhatp.com/go/mapper-gen/features/basic.md",
			want:  "https://raw.githubusercontent.com/toniphan21/go-mapper-gen/refs/heads/main/features/basic.md",
		},
		{
			input: "http://nhatp.com/go/mapper-gen/features/basic.md",
			want:  "https://raw.githubusercontent.com/toniphan21/go-mapper-gen/refs/heads/main/features/basic.md",
		},
		{
			input: "https://nhatp.com/go/mapper-gen/features/basic.md",
			want:  "https://raw.githubusercontent.com/toniphan21/go-mapper-gen/refs/heads/main/features/basic.md",
		},
		{
			input: "nhatp.com/go/stringer-gen/testdata/tags.md",
			want:  "https://raw.githubusercontent.com/toniphan21/go-stringer-gen/refs/heads/main/testdata/tags.md",
		},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			want := VanityURLFilePathResolver(tc.input)

			assert.Equal(t, tc.want, want)
		})
	}
}

func Test_PrefixMappedFilePathResolver(t *testing.T) {
	cases := []struct {
		name     string
		replaces map[string]string
		input    string
		output   string
	}{
		{
			name:   "no replaces",
			input:  "/root/something",
			output: "/root/something",
		},
		{
			name:   "no replaces - no vanity",
			input:  "https://nhatp.com/go/stringer-gen/testdata/tags.md",
			output: "https://nhatp.com/go/stringer-gen/testdata/tags.md",
		},
		{
			name:     "one prefix - no match",
			replaces: map[string]string{"repo://": "https://nhatp.com/"},
			input:    "/anything",
			output:   "/anything",
		},
		{
			name:     "one prefix - match",
			replaces: map[string]string{"repo://": "https://nhatp.com/"},
			input:    "repo://anything",
			output:   "https://nhatp.com/anything",
		},
		{
			name:     "two prefixes - no match",
			replaces: map[string]string{"repo://": "https://nhatp.com/", "repo-refs://": "https://nhatp.com/refs/"},
			input:    "/anything",
			output:   "/anything",
		},
		{
			name:     "two prefixes - match",
			replaces: map[string]string{"repo://": "https://nhatp.com/", "repo-refs://": "https://nhatp.com/refs/"},
			input:    "repo://anything",
			output:   "https://nhatp.com/anything",
		},
		{
			name:     "two prefixes - match 2nd",
			replaces: map[string]string{"repo://": "https://nhatp.com/", "repo-refs://": "https://nhatp.com/refs/"},
			input:    "repo-refs://anything",
			output:   "https://nhatp.com/refs/anything",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fn := PrefixMappedFilePathResolver(tc.replaces)
			assert.Equal(t, tc.output, fn(tc.input))
		})
	}
}

func Test_WithVanityURLFilePathResolver(t *testing.T) {
	cases := []struct {
		name     string
		replaces map[string]string
		input    string
		output   string
	}{
		{
			name:   "no replaces - resolved by builtin VanityURLFilePathResolver - file path",
			input:  "/root/something",
			output: "/root/something",
		},
		{
			name:   "no replaces - resolved by builtin VanityURLFilePathResolver - resolved",
			input:  "https://nhatp.com/go/stringer-gen/testdata/tags.md",
			output: "https://raw.githubusercontent.com/toniphan21/go-stringer-gen/refs/heads/main/testdata/tags.md",
		},

		{
			name:     "one prefix - no match",
			replaces: map[string]string{"repo://": "https://nhatp.com/"},
			input:    "/anything",
			output:   "/anything",
		},
		{
			name:     "one prefix - no match - use vanity",
			replaces: map[string]string{"repo://": "https://nhatp.com/"},
			input:    "nhatp.com/go/stringer-gen/testdata/tags.md",
			output:   "https://raw.githubusercontent.com/toniphan21/go-stringer-gen/refs/heads/main/testdata/tags.md",
		},
		{
			name:     "one prefix - match",
			replaces: map[string]string{"repo://": "https://nhatp.com/"},
			input:    "repo://anything",
			output:   "https://nhatp.com/anything",
		},
		{
			name:     "two prefixes - no match",
			replaces: map[string]string{"repo://": "https://nhatp.com/", "repo-refs://": "https://nhatp.com/refs/"},
			input:    "/anything",
			output:   "/anything",
		},
		{
			name:     "two prefixes - match",
			replaces: map[string]string{"repo://": "https://nhatp.com/", "repo-refs://": "https://nhatp.com/refs/"},
			input:    "repo://anything",
			output:   "https://nhatp.com/anything",
		},
		{
			name:     "two prefixes - match 2nd",
			replaces: map[string]string{"repo://": "https://nhatp.com/", "repo-refs://": "https://nhatp.com/refs/"},
			input:    "repo-refs://anything",
			output:   "https://nhatp.com/refs/anything",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fn := WithVanityURLFilePathResolver(tc.replaces)
			assert.Equal(t, tc.output, fn(tc.input))
		})
	}
}
