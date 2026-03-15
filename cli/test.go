package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	genlib "nhatp.com/go/gen-lib"
	"nhatp.com/go/gen-lib/file"
	"nhatp.com/go/gen-lib/gentest"
)

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

type TestCmd struct {
	Files     []string `arg:"positional" help:"Markdown file(s) to test" placeholder:"FILE"`
	Name      string   `arg:"-n,--name" help:"Run test which has matched name (case insensitive)" default:""`
	ShowSetup bool     `arg:"-s,--show-setup" help:"Show test setup steps" default:"false"`
	TabSize   int      `arg:"-t,--tab-size" help:"Number of spaces to use in tab size" default:"8"`
	EmitCode  string   `arg:"-e,--emit-code" help:"Emit to code if the test passed. If emit is empty looking for path in Markdown comment." default:""`

	logger *slog.Logger
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

func (c *TestCmd) matchName(name, term string) bool {
	if strings.TrimSpace(term) != "" {
		search := strings.ToLower(strings.TrimSpace(term))
		return strings.Index(search, strings.ToLower(name)) != -1
	}
	return false
}

func (c *TestCmd) PrintError(msg string, args ...any) {
	c.logger.Error(msg, args...)
}

func (c *TestCmd) PrintWarn(msg string, args ...any) {
	c.logger.Warn(msg, args...)
}

func (c *TestCmd) PrintSetup(msg string, args ...any) {
	if !c.ShowSetup {
		return
	}
	c.logger.Info(msg, args...)
}

func (c *TestCmd) PrintSetupVerbose(msg string, args ...any) {
	if !c.ShowSetup {
		return
	}
	c.logger.Debug(msg, args...)
}

func (c *TestCmd) Print(msg string) {
	c.logger.Info(msg)
}

func (c *TestCmd) TestFiles() ([]testFile, int) {
	var count int
	var result []testFile
	for _, inputFile := range c.Files {
		stat, err := os.Stat(inputFile)
		if err != nil {
			c.PrintError(ColorRed(err.Error()))
			continue
		}

		if stat.IsDir() {
			c.PrintWarn(ColorBlue(inputFile) + " is a directory, skipped")
			continue
		}

		content, err := os.ReadFile(inputFile)
		if err != nil {
			c.PrintError(ColorRed(err.Error()))
			continue
		}

		var matched []gentest.MarkdownTestCase
		tcs := gentest.ParseMarkdown(content)
		if strings.TrimSpace(c.Name) != "" {
			for _, v := range tcs {
				if c.matchName(v.Name, c.Name) {
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

func (c *TestCmd) Execute(executeTestCase func(testCase TestCase, options map[string]any) genlib.FileManager) {
	SetTabSize(c.TabSize)

	var total, passed, failed int
	var failedTests []failedTest
	var tempDirs []string
	defer func() {
		if len(tempDirs) > 0 {
			c.PrintSetup("deleting temporary directories")
			for _, dir := range tempDirs {
				c.PrintSetupVerbose("\tdeleted temporary directory " + dir)
				_ = os.RemoveAll(dir)
			}
			c.PrintSetup("")
		}

		if total == passed {
			c.Print(ColorGreen(fmt.Sprintf("Result: passed all %d total tests", passed)))
		} else {
			c.Print(ColorRedBold(fmt.Sprintf("Result: %d failed, passed %d/%d total tests.", failed, passed, total)))
			c.Print("")
			c.Print("Run failed test command:")
			c.Print("")
			for _, ft := range failedTests {
				c.Print("\t" + ft.makeRunCmd())
			}
		}
		c.Print("")
	}()

	testFiles, count := c.TestFiles()

	total = count

	for _, tf := range testFiles {
		c.Print(ColorBlue(tf.fileName))

		for i, md := range tf.mdTestCases {
			c.PrintSetup("\t" + ColorCyan(md.Name))
			tempDir, err := os.MkdirTemp("", "gen-test-*")
			if err != nil {
				c.PrintError("\tError creating temp dir:", slog.Any("error", err))
			}
			c.PrintSetup("\tcreated temporary directory at " + ColorWhite(tempDir))

			setupOk := true
			for _, sf := range md.SourceFiles {
				fn := sf.FilePath()
				fc := sf.FileContent()
				if err := c.writeTestFile(tempDir, fn, fc); err != nil {
					c.PrintError(ColorRed(err.Error()))
					setupOk = false
					continue
				}
				c.PrintSetup("\tcreated source file " + ColorWhite(fn))

				PrintFileWithFunction("", sf.FileContent(), func(s string) {
					c.PrintSetupVerbose("\t" + s)
				})

				if fn == "go.mod" {
					directDependencies, err := file.ParseGoModDirectDependencies(fc)
					if err != nil {
						c.PrintError("\t" + ColorRed(err.Error()))
						setupOk = false
						continue
					}

					if len(directDependencies) > 0 {
						if err = genlib.RunGoGet(tempDir, directDependencies); err != nil {
							c.PrintError("\t" + ColorRed(err.Error()))
							setupOk = false
							continue
						}
						c.PrintSetup("\tinstalled dependencies")
					}
				}
			}

			var fm genlib.FileManager
			if !setupOk {
				c.PrintSetup("\tfailed to setup test")
			} else {
				tc := tf.makeTestCase(md, tempDir)
				opts := c.parseOptions(md.Content)

				fm = executeTestCase(tc, opts)
			}

			// handle test result by comparing golden files in the fm and md.GoldenFile
			isSuccess := true
			if fm != nil {
				for _, f := range fm.Files() {
					err := c.writeTestFile(tempDir, f.RelPath, []byte(f.Content()))
					if err != nil {
						c.PrintError(err.Error())
						isSuccess = false
					}
				}

				for _, gf := range md.GoldenFiles {
					fn := gf.FilePath()
					fc := gf.FileContent()
					out, err := c.readTestFile(tempDir, fn)
					if err != nil {
						if errors.Is(err, os.ErrNotExist) {
							c.PrintSetup(ColorRed(fmt.Sprintf("\texpected golden file %v but file does not exist", fn)))
							isSuccess = false
							continue
						}

						c.PrintError(err.Error())
						continue
					}

					if !c.compareFileContent(out, fc) {
						c.PrintSetup(ColorRed("\tgolden file content does not match expectation"))
						PrintDiffWithFunction("expected", fc, "generated", out, func(s string) {
							c.PrintSetup("\t" + s)
						})
						isSuccess = false

						_ = c.writeTestFile(tempDir, fn+".golden", fc)
						continue
					}

					c.PrintSetupVerbose("\tpassed with " + ColorYellow("golden file ") + ColorWhite(fn))
					PrintFileWithFunction("", fc, func(s string) {
						c.PrintSetupVerbose("\t" + s)
					})
				}
			}

			if isSuccess {
				if err := c.emitTestCode(md, tempDir); err != nil {
					c.PrintError("\t" + ColorRed(err.Error()))
				}

				tempDirs = append(tempDirs, tempDir)
				c.Print(ColorGreen("\t\u2714 passed ") + md.Name)
				passed++

				if i != len(tf.mdTestCases)-1 {
					c.PrintSetup("")
				}
				continue
			}

			failedTests = append(failedTests, failedTest{
				fileName: tf.fileName,
				testName: md.Name,
			})
			failed++
			c.Print(ColorRed("\t\u2718 failed ") + md.Name)

			if i != len(tf.mdTestCases)-1 {
				c.PrintSetup("")
			}
		}

		c.Print("")
	}
}

func (c *TestCmd) readTestFile(testDir string, filePath string) ([]byte, error) {
	fp := filepath.Join(testDir, filePath)
	return os.ReadFile(fp)
}

func (c *TestCmd) compareFileContent(left, right []byte) bool {
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

func (c *TestCmd) writeTestFile(testDir string, filePath string, fileContent []byte) error {
	fp := filepath.Join(testDir, filePath)
	dir := filepath.Dir(fp)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	return os.WriteFile(fp, fileContent, 0600)
}

func (c *TestCmd) copyDir(src string, dst string) error {
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

		return c.copyFile(path, targetPath)
	})
}

func (c *TestCmd) copyFile(srcFile, dstFile string) error {
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

func (c *TestCmd) parseOptions(content string) map[string]any {
	var result map[string]any
	var rawOptions string
	lines := strings.Split(content, "\n")
	for _, v := range lines {
		line := strings.TrimSpace(v)
		if strings.HasPrefix(line, "[//]: # (Options:") && strings.HasSuffix(line, ")") {
			line = strings.TrimPrefix(line, "[//]: # (Options:")
			line = strings.TrimSuffix(line, ")")
			rawOptions = line
			continue
		}
	}

	if rawOptions != "" {
		_ = json.Unmarshal([]byte(rawOptions), &result)
	}
	return result
}

func (c *TestCmd) emitTestCode(md gentest.MarkdownTestCase, dir string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	var readme []string
	var emittedPath = c.EmitCode

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

		if err = c.copyDir(dir, dst); err != nil {
			return err
		}

		_ = c.writeTestFile(dst, "README.md", []byte(strings.Join(readme, "\n")))
		c.PrintSetup("\temit code to " + emittedPath)
	}
	return nil
}
