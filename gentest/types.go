package gentest

import (
	"fmt"
	"go/token"
	"go/types"
	"strings"
)

func Param(name string, typ string) *types.Var {
	return types.NewParam(token.NoPos, nil, name, Type(typ))
}

func Result(args ...string) []*types.Var {
	var out []*types.Var
	for _, arg := range args {
		out = append(out, types.NewParam(token.NoPos, nil, "", Type(arg)))
	}
	return out
}

func NamedResult(name string, typ string) *types.Var {
	return types.NewParam(token.NoPos, nil, name, Type(typ))
}

func Type(typ string) types.Type {
	typ = strings.TrimSpace(typ)

	// handle pointer
	if strings.HasPrefix(typ, "*") {
		return types.NewPointer(Type(typ[1:]))
	}

	// handle slice
	if strings.HasPrefix(typ, "[]") {
		return types.NewSlice(Type(typ[2:]))
	}

	// handle map
	if strings.HasPrefix(typ, "map[") {
		inner := typ[4:] // strip "map["
		key, value, ok := parseMapKeyValue(inner)
		if !ok {
			panic(fmt.Sprintf("invalid map type: %q", typ))
		}
		return types.NewMap(Type(key), Type(value))
	}

	// handle named / universe types
	pkgPath, typeName := parseTypeString(typ)
	if pkgPath == "" {
		lookup := types.Universe.Lookup(typeName)
		if lookup != nil {
			return lookup.Type()
		}
		panic(fmt.Sprintf("unknown universe type: %q", typeName))
	}

	pkgName := pkgPath
	if i := strings.LastIndex(pkgPath, "/"); i >= 0 {
		pkgName = pkgPath[i+1:]
	}

	pkg := types.NewPackage(pkgPath, pkgName)
	obj := types.NewTypeName(token.NoPos, pkg, typeName, nil)
	return types.NewNamed(obj, types.Typ[types.Invalid], nil)
}

func parseTypeString(s string) (pkgPath string, typeName string) {
	lastSlash := strings.LastIndex(s, "/")
	separatorIndex := strings.LastIndex(s, ".")

	if separatorIndex > lastSlash {
		pkgPath = s[:separatorIndex]
		s = s[separatorIndex+1:]
	} else {
		pkgPath = ""
	}
	typeName = s

	return pkgPath, typeName
}

func parseMapKeyValue(s string) (key, value string, ok bool) {
	depth := 0
	for i, c := range s {
		switch c {
		case '[':
			depth++
		case ']':
			if depth == 0 {
				return s[:i], s[i+1:], true
			}
			depth--
		}
	}
	return "", "", false
}
