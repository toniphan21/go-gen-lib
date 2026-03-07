// Code generated from Pkl module `genlib.Example`. DO NOT EDIT.
package dev

import "nhatp.com/go/gen-lib/pkl/gen"

type MatcherUsage struct {
	Single gen.Matcher `pkl:"single"`

	List []gen.Matcher `pkl:"list"`

	Union any `pkl:"union"`

	InvalidType string `pkl:"invalid_type"`
}
