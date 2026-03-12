package genlib

import (
	"go/types"
	"strconv"
	"strings"

	"github.com/dave/jennifer/jen"
	"golang.org/x/tools/go/packages"
)

const CommentWidth = 80

func VarToJenCode(v *types.Var) jen.Code {
	if v.Name() == "" {
		return TypeToJenCode(v.Type())
	}
	return jen.Id(v.Name()).Add(TypeToJenCode(v.Type()))
}

func TypeToJenCode(t types.Type) jen.Code {
	switch tt := t.(type) {
	case *types.Basic:
		if tt.Kind() == types.Invalid {
			return jen.Id("interface{}")
		}
		if tt.Kind() == types.UnsafePointer {
			return jen.Id("unsafe.Pointer")
		}
		if tt.Name() == "error" {
			return jen.Error()
		}
		return jen.Id(tt.Name())

	case *types.Pointer:
		return jen.Op("*").Add(TypeToJenCode(tt.Elem()))

	case *types.Named:
		obj := tt.Obj()
		pkg := obj.Pkg()
		if pkg != nil {
			return jen.Qual(pkg.Path(), obj.Name())
		}
		return jen.Id(obj.Name())

	case *types.Slice:
		return jen.Index().Add(TypeToJenCode(tt.Elem()))

	case *types.Array:
		return jen.Index(jen.Lit(tt.Len())).Add(TypeToJenCode(tt.Elem()))

	case *types.Map:
		return jen.Map(TypeToJenCode(tt.Key())).Add(TypeToJenCode(tt.Elem()))

	case *types.Chan:
		elemCode := TypeToJenCode(tt.Elem())
		switch tt.Dir() {
		case types.SendRecv:
			return jen.Chan().Add(elemCode)
		case types.SendOnly:
			return jen.Chan().Op("<-").Add(elemCode)
		case types.RecvOnly:
			return jen.Op("<-").Chan().Add(elemCode)
		}

	case *types.Signature:
		fn := jen.Func()
		fn.Params(funcGroupFromTuple(tt.Params())...)
		if tt.Results().Len() == 1 {
			fn.Params(funcGroupFromTuple(tt.Results())...)
		} else if tt.Results().Len() > 1 {
			fn.Params(funcGroupFromTuple(tt.Results())...)
		}
		return fn

	default:
		return jen.Id(tt.String())
	}
	return nil
}

func funcGroupFromTuple(t *types.Tuple) []jen.Code {
	var res []jen.Code
	for i := 0; i < t.Len(); i++ {
		v := t.At(i)
		typ := TypeToJenCode(v.Type())
		if v.Name() != "" {
			res = append(res, jen.Id(v.Name()).Add(typ))
		} else {
			res = append(res, typ)
		}
	}
	return res
}

func TypeSimpleName(t types.Type) string {
	switch tt := t.(type) {
	case *types.Named:
		return tt.Obj().Name()
	case *types.Basic:
		return tt.Name()
	case *types.Pointer:
		return TypeSimpleName(tt.Elem())
	case *types.Slice:
		return "[]" + TypeSimpleName(tt.Elem())
	case *types.Array:
		return "[" + strconv.FormatInt(tt.Len(), 10) + "]" + TypeSimpleName(tt.Elem())
	case *types.Map:
		return "map[" + TypeSimpleName(tt.Key()) + "]" + TypeSimpleName(tt.Elem())
	case *types.Chan:
		return "chan " + TypeSimpleName(tt.Elem())
	case *types.Interface:
		return "interface{}"
	case *types.Signature:
		return "func"
	default:
		return t.String()
	}
}

func TypeSimpleNameWithPkg(t types.Type, pkg *packages.Package) string {
	switch tt := t.(type) {
	case *types.Named:
		simpleName := tt.Obj().Name()

		obj := tt.Obj()

		actualPkgPath := ""
		if obj.Pkg() != nil {
			actualPkgPath = obj.Pkg().Path()
		}
		if actualPkgPath == pkg.PkgPath {
			return simpleName
		}
		return obj.Pkg().Name() + "." + simpleName

	case *types.Pointer:
		return TypeSimpleNameWithPkg(tt.Elem(), pkg)
	default:
		return TypeSimpleName(t)
	}
}

func WrapComment(comment string) jen.Code {
	lines := WrapText(comment, CommentWidth)
	l := len(lines)
	if l == 1 {
		return jen.Comment(lines[0])
	}

	var code = jen.Add()
	for i, line := range lines {
		code = code.Add(jen.Comment(line))
		if i != l-1 {
			code = code.Line()
		}
	}
	return code
}

func WrapText(s string, width int) []string {
	var lines []string
	words := strings.Fields(s)

	if len(words) == 0 {
		return []string{""}
	}

	var current string
	for _, word := range words {
		if len(current)+len(word)+1 > width {
			lines = append(lines, strings.TrimSpace(current))
			current = word
		} else {
			if current == "" {
				current = word
			} else {
				current += " " + word
			}
		}
	}
	if current != "" {
		lines = append(lines, current)
	}

	return lines
}

func ZeroValueOfType(t types.Type) jen.Code {
	switch tt := t.(type) {
	case *types.Basic:
		switch tt.Kind() {
		case types.Bool:
			return jen.False()
		case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
			types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64, types.Uintptr,
			types.Float32, types.Float64,
			types.Complex64, types.Complex128:
			return jen.Lit(0)
		case types.String:
			return jen.Lit("")
		case types.UnsafePointer:
			return jen.Nil()
		default:
			if tt.Name() == "error" {
				return jen.Nil()
			}
			return jen.Lit(0)
		}

	case *types.Pointer,
		*types.Slice,
		*types.Map,
		*types.Chan,
		*types.Interface,
		*types.Signature:
		return jen.Nil()

	case *types.Named:
		under := tt.Underlying()
		if _, ok := under.(*types.Struct); ok {
			obj := tt.Obj()
			if obj != nil && obj.Pkg() != nil {
				return jen.Qual(obj.Pkg().Path(), obj.Name()).Values()
			}
			return jen.Id(obj.Name()).Values()
		}
		return ZeroValueOfType(under)

	case *types.Struct:
		return jen.Values()

	case *types.Array:
		return jen.Values()

	default:
		return jen.Nil()
	}
}
