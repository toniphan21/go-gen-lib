package genlib

import (
	"go/types"
)

func IsIdentical(t1, t2 types.Type) bool {
	// unwrap pointers recursively
	p1, ok1 := t1.(*types.Pointer)
	p2, ok2 := t2.(*types.Pointer)
	if ok1 != ok2 {
		return false
	}
	if ok1 && ok2 {
		return IsIdentical(p1.Elem(), p2.Elem())
	}

	// handle slices
	s1, ok1 := t1.(*types.Slice)
	s2, ok2 := t2.(*types.Slice)
	if ok1 != ok2 {
		return false
	}
	if ok1 && ok2 {
		return IsIdentical(s1.Elem(), s2.Elem())
	}

	// handle maps
	m1, ok1 := t1.(*types.Map)
	m2, ok2 := t2.(*types.Map)
	if ok1 != ok2 {
		return false
	}
	if ok1 && ok2 {
		return IsIdentical(m1.Key(), m2.Key()) && IsIdentical(m1.Elem(), m2.Elem())
	}

	// handle named types (cross-package safe)
	n1, ok1 := t1.(*types.Named)
	n2, ok2 := t2.(*types.Named)
	if ok1 != ok2 {
		return false
	}
	if ok1 && ok2 {
		obj1, obj2 := n1.Obj(), n2.Obj()
		path1, path2 := "", ""
		if obj1.Pkg() != nil {
			path1 = obj1.Pkg().Path()
		}
		if obj2.Pkg() != nil {
			path2 = obj2.Pkg().Path()
		}
		return path1 == path2 && obj1.Name() == obj2.Name()
	}

	return types.Identical(t1, t2)
}
