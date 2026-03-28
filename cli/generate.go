package cli

import (
	"log/slog"
	"os"

	genlib "nhatp.com/go/gen-lib"
)

type GenerateLogPoints struct {
	BeforeHandleOutput func(dryRun bool)
	BeforeHandleFile   func(fullPath, relPath, content string, dryRun bool)
	AfterHandleFile    func(fullPath, relPath, content string, dryRun bool)
}

type GenerateRunner struct {
	Generate    func() error
	DryRun      bool
	FileManager genlib.FileManager
	Logger      *slog.Logger
	LogPoints   *GenerateLogPoints
}

func (r *GenerateRunner) Run() error {
	if err := r.Generate(); err != nil {
		return err
	}

	if r.FileManager == nil {
		return nil
	}

	if r.LogPoints != nil && r.LogPoints.BeforeHandleOutput != nil {
		r.LogPoints.BeforeHandleOutput(r.DryRun)
	}

	for _, out := range r.FileManager.Files() {
		content := out.Content()

		if r.LogPoints != nil && r.LogPoints.BeforeHandleFile != nil {
			r.LogPoints.BeforeHandleFile(out.FullPath, out.RelPath, content, r.DryRun)
		}

		if r.DryRun {
			if r.Logger != nil {
				PrintFileWithFunction(out.RelPath, []byte(content), func(l string) {
					r.Logger.Info(l)
				})
			}
		} else {
			err := os.WriteFile(out.FullPath, []byte(content), 0644)
			if err != nil {
				return err
			}
		}

		if r.LogPoints != nil && r.LogPoints.AfterHandleFile != nil {
			r.LogPoints.AfterHandleFile(out.FullPath, out.RelPath, content, r.DryRun)
		}
	}

	return nil
}
