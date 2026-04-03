package genlib

import (
	"context"

	"golang.org/x/tools/go/packages"
)

type EmitterContext interface {
	context.Context

	Package() *packages.Package

	GenFile() *GenFile

	FileManager() FileManager

	VarName() NameManager
}

func NewEmitterContext(pkg *packages.Package, fm FileManager, gf *GenFile, varPrefix string) EmitterContext {
	vp := "v"
	if varPrefix != "" {
		vp = varPrefix
	}

	return &emitterContext{
		Context:        context.Background(),
		pkg:            pkg,
		fileManager:    fm,
		genFile:        gf,
		varNameManager: NewNameManager(vp, nil),
	}
}

type emitterContext struct {
	context.Context

	pkg            *packages.Package
	pkgNameManager NameManager
	genFile        *GenFile
	fileManager    FileManager
	varNameManager NameManager
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

func (c *emitterContext) VarName() NameManager {
	return c.varNameManager
}

var _ EmitterContext = (*emitterContext)(nil)
