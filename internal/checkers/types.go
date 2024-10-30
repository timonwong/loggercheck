package checkers

import "go/types"

type commonAlias interface {
	Obj() *types.TypeName
}

func isTypeVariadicEmptyInterface(typ types.Type) bool {
	sliceTyp, ok := typ.(*types.Slice)
	if !ok {
		return false
	}

	typ = sliceTyp.Elem()
	for i := 0; i < 2; i++ {
		switch iface := typ.(type) {
		case *types.Interface:
			return iface.Empty()
		case *types.Alias:
			typ = iface.Underlying()
		default:
			return false
		}
	}
	return false
}
