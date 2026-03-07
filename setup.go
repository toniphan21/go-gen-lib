package genlib

import (
	"fmt"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/tools/go/packages"
	"nhatp.com/go/gen-lib/file"
)

func writeSourceFile(out string, f file.File) error {
	path := filepath.Join(out, f.FilePath())
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(path, f.FileContent(), 0644); err != nil {
		return err
	}
	return nil
}

func LoadPackages(dir string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedDeps | packages.NeedImports,
		Fset: token.NewFileSet(),
		Dir:  dir,
	}

	return packages.Load(cfg, "./...")
}

func SetupSourceCode(dir string, files []file.File, additionalFiles ...file.File) error {
	for _, f := range files {
		if err := writeSourceFile(dir, f); err != nil {
			return err
		}
	}

	for _, f := range additionalFiles {
		if err := writeSourceFile(dir, f); err != nil {
			return err
		}
	}
	return nil
}

func RunGoGet(dir string, requires map[string]string) error {
	for pkg, version := range requires {
		cmd := exec.Command("go", "get", fmt.Sprintf("%v@%v", pkg, version))
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}
