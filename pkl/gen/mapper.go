package gen

import "nhatp.com/go/gen-lib"

func (o *Output) ToOutput() genlib.Output {
	output := genlib.Output{}
	if o.Package != nil {
		output.PackageName = *o.Package
	}

	if o.Source != nil {
		output.SourceFileName = *o.Source
	}

	if o.Test != nil {
		output.TestFileName = *o.Test
	}
	return output
}
