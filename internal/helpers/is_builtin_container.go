package helpers

import "go/types"

func IsBuiltinContainer(t types.Type) bool {
	switch t.(type) {
	case *types.Map, *types.Chan, *types.Signature:
		return true
	default:
		return false
	}
}
