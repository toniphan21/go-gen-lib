package cli

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	genlib "nhatp.com/go/gen-lib"
	"nhatp.com/go/gen-lib/file"
	"nhatp.com/go/gen-lib/gentest"
	"nhatp.com/go/gen-lib/internal/util"
)

func PrefixMappedFilePathResolver(replaces map[string]string) func(string) string {
	return func(path string) string {
		if replaces == nil {
			return path
		}

		for prefix, replaced := range replaces {
			if strings.HasPrefix(path, prefix) {
				return replaced + strings.TrimPrefix(path, prefix)
			}
		}

		return path
	}
}

func WithVanityURLFilePathResolver(replaces map[string]string) func(string) string {
	prefixes := PrefixMappedFilePathResolver(replaces)
	return func(path string) string {
		out := prefixes(path)
		if out != path {
			return out
		}
		return VanityURLFilePathResolver(path)
	}
}

func VanityURLFilePathResolver(input string) string {
	vanityBase := "nhatp.com/go/"
	prefixes := []string{
		vanityBase,
		"http://" + vanityBase,
		"https://" + vanityBase,
	}

	var base string
	for _, prefix := range prefixes {
		if !strings.HasPrefix(input, prefix) {
			continue
		}

		base = strings.TrimPrefix(input, prefix)
		break
	}

	if base == "" {
		return input
	}

	var outBase = "https://raw.githubusercontent.com/toniphan21/go-"
	idx := strings.Index(base, "/")
	if idx == -1 {
		return outBase + base
	}
	return outBase + base[:idx] + "/refs/heads/main/" + base[idx+1:]
}

type failedTest struct {
	fileName string
	testName string
}

func (t *failedTest) makeRunCmd() string {
	execPath := os.Args[0]
	cmd := execPath
	if strings.Index(execPath, "go-build") != -1 {
		cmd = "go run ./cmd/" + filepath.Base(execPath)
	}
	cmd += " test " + t.fileName + " -n='" + strings.ToLower(t.testName) + "'"
	return cmd
}

type testFile struct {
	fileName    string
	mdTestCases []gentest.MarkdownTestCase
}

func (t *testFile) makeTestCase(md gentest.MarkdownTestCase, dir string) TestCase {
	return TestCase{
		TestFileName:      t.fileName,
		TestDir:           dir,
		Name:              md.Name,
		Content:           md.Content,
		Headers:           md.Headers,
		SourceFiles:       md.SourceFiles,
		GoldenFiles:       md.GoldenFiles,
		GoModFileContent:  md.GoModFileContent,
		GoSumFileContent:  md.GoSumFileContent,
		PklDevFileContent: md.PklDevFileContent,
	}
}

type TestCase struct {
	TestFileName      string
	TestDir           string
	Name              string
	Content           string
	Headers           []string
	SourceFiles       []file.File
	GoldenFiles       []file.File
	GoModFileContent  []byte
	GoSumFileContent  []byte
	PklDevFileContent []byte
}

func (tc *TestCase) PklDevLines() []string {
	if tc.PklDevFileContent != nil {
		return []string{string(tc.PklDevFileContent)}
	}
	return nil
}

// ---

type TestRunner struct {
	RunTestCase      func(testCase TestCase, options map[string]any) (genlib.FileManager, error)
	Files            []string
	Name             string
	TabSize          int
	ShowSetup        bool
	EmitPath         string
	Logger           *slog.Logger
	FilePathResolver func(string) string
}

func (r *TestRunner) matchName(name, term string) bool {
	if strings.TrimSpace(term) != "" {
		search := strings.ToLower(strings.TrimSpace(term))
		return strings.Index(strings.ToLower(name), search) != -1
	}
	return false
}

func (r *TestRunner) PrintError(msg string, args ...any) {
	r.Logger.Error(msg, args...)
}

func (r *TestRunner) PrintWarn(msg string, args ...any) {
	r.Logger.Warn(msg, args...)
}

func (r *TestRunner) PrintSetup(msg string, args ...any) {
	if !r.ShowSetup {
		return
	}
	r.Logger.Info(msg, args...)
}

func (r *TestRunner) PrintSetupVerbose(msg string, args ...any) {
	if !r.ShowSetup {
		return
	}
	r.Logger.Debug(msg, args...)
}

func (r *TestRunner) PrintFilePathResolved(in, out string) {
	if in == out || r.Logger == nil {
		return
	}

	r.Print(fmt.Sprintf("   input: %s", ColorWhite(in)))
	r.Print(fmt.Sprintf("resolved: %s", ColorWhite(out)))
	r.Print("")
}

func (r *TestRunner) Print(msg string) {
	r.Logger.Info(msg)
}

func (r *TestRunner) Run() {
	SetTabSize(r.TabSize)

	var total, passed, failed int
	var failedTests []failedTest
	var tempDirs []string
	defer func() {
		if len(tempDirs) > 0 {
			r.PrintSetup("deleting temporary directories")
			for _, dir := range tempDirs {
				r.PrintSetupVerbose("\tdeleted temporary directory " + dir)
				_ = os.RemoveAll(dir)
			}
			r.PrintSetup("")
		}

		if total == passed {
			r.Print(ColorGreen(fmt.Sprintf("Result: passed all %d total tests", passed)))
		} else {
			r.Print(ColorRedBold(fmt.Sprintf("Result: %d failed, passed %d/%d total tests.", failed, passed, total)))
			r.Print("")
			r.Print("Run failed test command:")
			r.Print("")
			for _, ft := range failedTests {
				r.Print("\t" + ft.makeRunCmd())
			}
		}
		r.Print("")
	}()

	testFiles, count := r.testFiles()

	total = count

	for _, tf := range testFiles {
		r.Print(ColorBlue(tf.fileName))

		for i, md := range tf.mdTestCases {
			r.PrintSetup("\t" + ColorCyan(md.Name))
			tempDir, err := os.MkdirTemp("", "gen-test-*")
			if err != nil {
				r.PrintError("\tError creating temp dir:", slog.Any("error", err))
			}
			r.PrintSetup("\tcreated temporary directory at " + ColorWhite(tempDir))

			setupOk := true
			for _, sf := range md.SourceFiles {
				fn := sf.FilePath()
				fc := sf.FileContent()
				if err := r.writeTestFile(tempDir, fn, fc); err != nil {
					r.PrintError(ColorRed(err.Error()))
					setupOk = false
					continue
				}
				r.PrintSetup("\tcreated source file " + ColorWhite(fn))

				PrintFileWithFunction("", sf.FileContent(), func(s string) {
					r.PrintSetupVerbose("\t" + s)
				})

				if fn == "go.mod" {
					directDependencies, err := file.ParseGoModDirectDependencies(fc)
					if err != nil {
						r.PrintError("\t" + ColorRed(err.Error()))
						setupOk = false
						continue
					}

					if len(directDependencies) > 0 {
						if err = genlib.RunGoGet(tempDir, directDependencies); err != nil {
							r.PrintError("\t" + ColorRed(err.Error()))
							setupOk = false
							continue
						}
						r.PrintSetup("\tinstalled dependencies")
					}
				}
			}

			isSuccess := true
			var fm genlib.FileManager
			if !setupOk {
				r.PrintSetup("\tfailed to setup test")
			} else {
				tc := tf.makeTestCase(md, tempDir)
				opts := util.ParseOptionsInMarkdownFile(md.Content)

				fm, err = r.RunTestCase(tc, opts)
				if err != nil {
					isSuccess = false
					r.PrintError("\t" + ColorRed(err.Error()))
				}
			}

			// handle test result by comparing golden files in the fm and md.GoldenFile
			if fm != nil {
				for _, f := range fm.Files() {
					err := r.writeTestFileFullPath(f.FilePath(), f.FileContent())
					if err != nil {
						r.PrintError(err.Error())
						isSuccess = false
					}
				}

				for _, gf := range md.GoldenFiles {
					fn := gf.FilePath()
					fc := gf.FileContent()
					out, err := r.readTestFile(tempDir, fn)
					if err != nil {
						if errors.Is(err, os.ErrNotExist) {
							r.PrintSetup(ColorRed(fmt.Sprintf("\texpected golden file %v but file does not exist", fn)))
							isSuccess = false
							continue
						}

						r.PrintError(err.Error())
						continue
					}

					if !r.compareFileContent(out, fc) {
						r.PrintSetup(ColorRed("\tgolden file content does not match expectation"))
						PrintDiffWithFunction("expected", fc, "generated", out, func(s string) {
							r.PrintSetup("\t" + s)
						})
						isSuccess = false

						_ = r.writeTestFile(tempDir, fn+".golden", fc)
						continue
					}

					r.PrintSetupVerbose("\tpassed with " + ColorYellow("golden file ") + ColorWhite(fn))
					PrintFileWithFunction("", fc, func(s string) {
						r.PrintSetupVerbose("\t" + s)
					})
				}
			}

			if isSuccess {
				if err := r.emitTestCode(md, tempDir); err != nil {
					r.PrintError("\t" + ColorRed(err.Error()))
				}

				tempDirs = append(tempDirs, tempDir)
				r.Print(ColorGreen("\t\u2714 passed ") + md.Name)
				passed++

				if i != len(tf.mdTestCases)-1 {
					r.PrintSetup("")
				}
				continue
			}

			failedTests = append(failedTests, failedTest{
				fileName: tf.fileName,
				testName: md.Name,
			})
			failed++
			r.Print(ColorRed("\t\u2718 failed ") + md.Name)

			if i != len(tf.mdTestCases)-1 {
				r.PrintSetup("")
			}
		}

		r.Print("")
	}
}

func (r *TestRunner) testFiles() ([]testFile, int) {
	var count int
	var result []testFile
	for _, inputFile := range r.Files {
		content, err := r.readFileContent(inputFile)
		if err != nil {
			r.PrintError(ColorRed(err.Error()))
			continue
		}

		var matched []gentest.MarkdownTestCase
		tcs := gentest.ParseMarkdown(content)
		if strings.TrimSpace(r.Name) != "" {
			for _, v := range tcs {
				if r.matchName(v.Name, r.Name) {
					matched = append(matched, v)
				}
			}
		} else {
			matched = tcs
		}

		count = count + len(matched)
		result = append(result, testFile{
			fileName:    inputFile,
			mdTestCases: matched,
		})
	}
	return result, count
}

func (r *TestRunner) readFileContent(path string) ([]byte, error) {
	resolved := path
	if r.FilePathResolver != nil {
		resolved = r.FilePathResolver(path)
	}
	r.PrintFilePathResolved(path, resolved)

	if strings.HasPrefix(resolved, "http") {
		resp, err := http.Get(resolved)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("http: %v", resp.Status)
		}

		return io.ReadAll(resp.Body)
	}

	stat, err := os.Stat(resolved)
	if err != nil {
		return nil, err
	}

	if stat.IsDir() {
		return nil, errors.New(resolved + " is a directory")
	}

	return os.ReadFile(resolved)
}

func (r *TestRunner) emitTestCode(md gentest.MarkdownTestCase, dir string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	var readme []string
	var emittedPath = r.EmitPath

	lines := strings.Split(md.Content, "\n")
	for _, v := range lines {
		line := strings.TrimSpace(v)
		if emittedPath == "" {
			if strings.HasPrefix(line, "[//]: # (EmitCode:") && strings.HasSuffix(line, ")") {
				line = strings.TrimPrefix(line, "[//]: # (EmitCode:")
				line = strings.TrimSuffix(line, ")")
				emittedPath = line
				continue
			}
		}
		readme = append(readme, v)
	}

	if emittedPath != "" {
		// copy all files in temp to
		dst := filepath.Join(wd, emittedPath)
		if err = os.RemoveAll(dst); err != nil {
			return err
		}

		if err = r.copyDir(dir, dst); err != nil {
			return err
		}

		_ = r.writeTestFile(dst, "README.md", []byte(strings.Join(readme, "\n")))
		r.PrintSetup("\temit code to " + emittedPath)
	}
	return nil
}

func (r *TestRunner) readTestFile(testDir string, filePath string) ([]byte, error) {
	fp := filepath.Join(testDir, filePath)
	return os.ReadFile(fp)
}

func (r *TestRunner) compareFileContent(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for i := 0; i < len(left); i++ {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func (r *TestRunner) writeTestFile(testDir string, filePath string, fileContent []byte) error {
	fp := filepath.Join(testDir, filePath)
	return r.writeTestFileFullPath(fp, fileContent)
}

func (r *TestRunner) writeTestFileFullPath(filePath string, fileContent []byte) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	return os.WriteFile(filePath, fileContent, 0600)
}

func (r *TestRunner) copyDir(src string, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		return r.copyFile(path, targetPath)
	})
}

func (r *TestRunner) copyFile(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer out.Close()

	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer in.Close()

	_, err = io.Copy(out, in)
	return err
}
