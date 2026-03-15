package cli

import (
	"log/slog"
	"os"
	"path/filepath"

	genlib "nhatp.com/go/gen-lib"
)

type GenerateLogPoints struct {
	BeforeHandleOutput func(dryRun bool)
	BeforeHandleFile   func(fullPath, relPath, content string, dryRun bool)
	AfterHandleFile    func(fullPath, relPath, content string, dryRun bool)
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

	if err := cb(); err != nil {
		return err
	}

	if c.fm == nil {
		return nil
	}

	if c.logPoints != nil && c.logPoints.BeforeHandleOutput != nil {
		c.logPoints.BeforeHandleOutput(c.DryRun)
	}

	for _, out := range c.fm.Files() {
		content := out.Content()

		if c.logPoints != nil && c.logPoints.BeforeHandleFile != nil {
			c.logPoints.BeforeHandleFile(out.FullPath, out.RelPath, content, c.DryRun)
		}

		if c.DryRun {
			PrintFileWithFunction(out.RelPath, []byte(content), func(l string) {
				c.logger.Info(l)
			})
		} else {
			err := os.WriteFile(out.FullPath, []byte(content), 0644)
			if err != nil {
				return err
			}
		}

		if c.logPoints != nil && c.logPoints.AfterHandleFile != nil {
			c.logPoints.AfterHandleFile(out.FullPath, out.RelPath, content, c.DryRun)
		}
	}

	return nil
}
