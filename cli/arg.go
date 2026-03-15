package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/alexflint/go-arg"
)

type VersionCmd struct{}

type Arguments struct {
	Generate *GenerateCmd `arg:"subcommand:generate" help:"Parse pkl configuration and generate code"`
	Test     *TestCmd     `arg:"subcommand:test" help:"Test generator using markdown files"`
	Version  *VersionCmd  `arg:"subcommand:version" help:"Print version information and exit"`

	NoColor bool `arg:"--no-color" help:"Disable colors" default:"false"`
	Verbose bool `arg:"-v,--verbose" help:"Enable verbose logging"`
}

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
