package gen

import (
	"errors"

	"nhatp.com/go/gen-lib"
)

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

func (m *Matcher) ToMatcher() (genlib.Matcher, error) {
	return mapSingleMatcher(*m)
}

func mapSingleMatcher(i any) (genlib.Matcher, error) {
	m, ok := i.(Matcher)
	if !ok {
		return nil, errors.New("unable to convert to matcher")
	}

	switch m.Type {
	case "string":
		ignoredCase := false
		if m.Options != nil {
			caseSensitive, exists := m.Options["case_sensitive"]
			if exists {
				b, ok := caseSensitive.(bool)
				ignoredCase = ok && b
			}
		}
		return genlib.NewStringMatcher(m.Pattern, ignoredCase), nil

	case "regex":
		return genlib.NewRegexMatcher(m.Pattern)
	}
	return nil, errors.New("unable to convert to matcher")
}

func castListMatcher(i any) ([]Matcher, bool) {
	list, ok := i.([]Matcher)
	if ok {
		return list, true
	}

	slice, ok := i.([]any)
	if !ok {
		return nil, false
	}

	var result []Matcher
	for _, v := range slice {
		ele, ok := v.(Matcher)
		if ok {
			result = append(result, ele)
		}
	}
	return result, true
}

func ToMatcher(i any) (genlib.Matcher, error) {
	list, ok := castListMatcher(i)
	if !ok {
		return mapSingleMatcher(i)
	}
	var matchers []genlib.Matcher
	for _, v := range list {
		o, err := mapSingleMatcher(v)
		if err != nil {
			return nil, err
		}
		matchers = append(matchers, o)
	}

	return genlib.NewListMatcher(matchers...), nil
}
