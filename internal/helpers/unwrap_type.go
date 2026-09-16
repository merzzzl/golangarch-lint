package helpers

import "go/types"

func UnwrapType(t types.Type) types.Type {
	for {
		switch v := t.(type) {
		case *types.Pointer:
			t = v.Elem()
		case *types.Slice:
			t = v.Elem()
		case *types.Array:
			t = v.Elem()
		default:
			return t
		}
	}
}
