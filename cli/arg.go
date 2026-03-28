package cli

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/alexflint/go-arg"
	genlib "nhatp.com/go/gen-lib"
)

type VersionCmd struct{}

type Arguments struct {
	Generate *GenerateCmd `arg:"subcommand:generate" help:"Parse pkl configuration and generate code"`
	Test     *TestCmd     `arg:"subcommand:test" help:"Test generator using markdown files"`
	Version  *VersionCmd  `arg:"subcommand:version" help:"Print version information and exit"`

	NoColor bool `arg:"--no-color" help:"Disable colors" default:"false"`
	Verbose bool `arg:"-v,--verbose" help:"Enable verbose logging"`
}

type GenerateCmd struct {
	WorkingDir     string `arg:"-w,--working-dir" help:"Working directory" default:"." placeholder:"WORKING_DIR"`
	ConfigFileName string `arg:"-c,--config" help:"Config file name" placeholder:"FILE_NAME"`
	DryRun         bool   `arg:"-d,--dry-run" help:"Preview changes without writing to disk"`

	logger    *slog.Logger
	fm        genlib.FileManager
	logPoints *GenerateLogPoints
}

func (c *GenerateCmd) Logger() *slog.Logger {
	return c.logger
}

func (c *GenerateCmd) ResolveWorkingDir() string {
	if c.WorkingDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		return wd
	}

	absPath, err := filepath.Abs(c.WorkingDir)
	if err != nil {
		panic(err)
	}
	return absPath
}

func (c *GenerateCmd) ConfigFilePath(defaultName string) string {
	var fn string
	if c.ConfigFileName == "" {
		fn = defaultName
	} else {
		fn = c.ConfigFileName
	}
	return filepath.Join(c.ResolveWorkingDir(), fn)
}

func (c *GenerateCmd) FileManager(options ...genlib.FileManagerOption) genlib.FileManager {
	if c.fm == nil {
		c.fm = genlib.NewFileManager(c.ResolveWorkingDir(), options...)
	}
	return c.fm
}

func (c *GenerateCmd) Execute(cb func() error, options ...ExecuteCmdOption) error {
	for _, opt := range options {
		opt.generate(c)
	}

	runner := &GenerateRunner{
		Generate:    cb,
		DryRun:      c.DryRun,
		FileManager: c.fm,
		Logger:      c.logger,
		LogPoints:   c.logPoints,
	}
	return runner.Run()
}

type TestCmd struct {
	Files     []string `arg:"positional" help:"Markdown file(s) to test" placeholder:"FILE"`
	Name      string   `arg:"-n,--name" help:"Run test which has matched name (case insensitive)" default:""`
	ShowSetup bool     `arg:"-s,--show-setup" help:"Show test setup steps" default:"false"`
	TabSize   int      `arg:"-t,--tab-size" help:"Number of spaces to use in tab size" default:"8"`
	EmitCode  string   `arg:"-e,--emit-code" help:"Emit to code if the test passed. If empty looking for path in Markdown comment." default:""`

	logger           *slog.Logger
	filePathResolver func(string) string
	runner           *TestRunner
}

func (c *TestCmd) Runner() *TestRunner {
	if c.runner == nil {
		c.runner = &TestRunner{
			Files:            c.Files,
			Name:             c.Name,
			TabSize:          c.TabSize,
			ShowSetup:        c.ShowSetup,
			EmitPath:         c.EmitCode,
			Logger:           c.logger,
			FilePathResolver: c.filePathResolver,
		}
	}
	return c.runner
}

func (c *TestCmd) Execute(
	executeTestCase func(testCase TestCase, options map[string]any) (genlib.FileManager, error),
	options ...ExecuteCmdOption,
) {
	for _, opt := range options {
		opt.test(c)
	}

	runner := c.Runner()
	runner.RunTestCase = executeTestCase
	runner.Run()
}

func (c *TestCmd) PrintError(msg string, args ...any) {
	c.Runner().PrintError(msg, args...)
}

func (c *TestCmd) PrintWarn(msg string, args ...any) {
	c.Runner().PrintWarn(msg, args...)
}

func (c *TestCmd) PrintSetup(msg string, args ...any) {
	c.Runner().PrintSetup(msg, args...)
}

func (c *TestCmd) PrintSetupVerbose(msg string, args ...any) {
	c.Runner().PrintSetupVerbose(msg, args...)
}

func (c *TestCmd) PrintFilePathResolved(in, out string) {
	c.Runner().PrintFilePathResolved(in, out)
}

func (c *TestCmd) Print(msg string) {
	c.Runner().Print(msg)
}

// ---

func ParseArgs() (*Arguments, *arg.Parser) {
	var args Arguments
	p := arg.MustParse(&args)
	if p.Subcommand() == nil {
		args.Generate = &GenerateCmd{}
	}

	return &args, p
}

// ---

type ExecuteCmdOption interface {
	generate(*GenerateCmd)
	test(*TestCmd)
}

type executeCmdOption struct {
	applyGenerateCmd func(*GenerateCmd)
	applyTestCmd     func(*TestCmd)
}

func (e *executeCmdOption) generate(cmd *GenerateCmd) {
	e.applyGenerateCmd(cmd)
}

func (e *executeCmdOption) test(cmd *TestCmd) {
	e.applyTestCmd(cmd)
}

var _ ExecuteCmdOption = (*executeCmdOption)(nil)

func WithLogger(logger *slog.Logger) ExecuteCmdOption {
	return &executeCmdOption{
		applyGenerateCmd: func(cmd *GenerateCmd) {
			cmd.logger = logger
		},
		applyTestCmd: func(cmd *TestCmd) {
			cmd.logger = logger
		},
	}
}

func WithGenerateLogPoints(logPoints *GenerateLogPoints) ExecuteCmdOption {
	return &executeCmdOption{
		applyGenerateCmd: func(cmd *GenerateCmd) {
			cmd.logPoints = logPoints
		},
	}
}

func WithTestFilePathResolver(resolver func(string) string) ExecuteCmdOption {
	return &executeCmdOption{
		applyTestCmd: func(cmd *TestCmd) {
			cmd.filePathResolver = resolver
		},
	}
}

type Subcommand struct {
	Setup       func(args *Arguments) error
	HandleError func(err error, logger *slog.Logger)
	Version     func(logger *slog.Logger)
	Generate    func(cmd GenerateCmd) error
	Test        func(cmd TestCmd) error
}

func Run(cb Subcommand) {
	handleError := func(err error, logger *slog.Logger) {
		if err != nil && cb.HandleError != nil {
			cb.HandleError(err, logger)
		}
	}

	args, p := ParseArgs()
	logger := NewLogger(args)

	if args.NoColor {
		DisableColor()
	}

	if cb.Setup != nil {
		handleError(cb.Setup(args), logger)
	}

	switch {
	case args.Version != nil:
		if cb.Version != nil {
			cb.Version(logger)
			os.Exit(0)
		}
		fmt.Println("there is no Version callback")
		os.Exit(1)

	case args.Generate != nil:
		cmd := *args.Generate
		cmd.logger = logger

		if cb.Generate != nil {
			handleError(cb.Generate(cmd), logger)
			os.Exit(0)
		}
		fmt.Println("there is no Generate callback")
		os.Exit(1)

	case args.Test != nil:
		cmd := *args.Test
		cmd.logger = logger
		cmd.filePathResolver = VanityURLFilePathResolver

		if cb.Test != nil {
			handleError(cb.Test(cmd), logger)
			os.Exit(0)
		}
		fmt.Println("there is no Test callback")
		os.Exit(1)

	default:
		p.WriteHelp(os.Stderr)
		os.Exit(0)
	}
}
