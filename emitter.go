package genlib

import (
	"context"
	"fmt"

	"golang.org/x/tools/go/packages"
)

type EmitterContext interface {
	context.Context

	Package() *packages.Package

	GenFile() *GenFile

	FileManager() FileManager

	NextVarName() string
}

func NewEmitterContext(pkg *packages.Package, fm FileManager, gf *GenFile, varTemplate string) EmitterContext {
	return &emitterContext{
		Context:     context.Background(),
		pkg:         pkg,
		fileManager: fm,
		genFile:     gf,
		varTemplate: varTemplate,
	}
}

type emitterContext struct {
	context.Context

	pkg             *packages.Package
	genFile         *GenFile
	fileManager     FileManager
	varTemplate     string
	currentVarCount int
}

func (c *emitterContext) Package() *packages.Package {
	return c.pkg
}

func (c *emitterContext) GenFile() *GenFile {
	return c.genFile
}

func (c *emitterContext) FileManager() FileManager {
	return c.fileManager
}

func (c *emitterContext) NextVarName() string {
	template := "v%d"
	if c.varTemplate != "" {
		template = c.varTemplate
	}

	v := fmt.Sprintf(template, c.currentVarCount)
	c.currentVarCount++
	return v
}

var _ EmitterContext = (*emitterContext)(nil)
