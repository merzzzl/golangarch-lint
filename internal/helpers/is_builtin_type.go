package helpers

var isBuiltinTypeSet = map[string]bool{
	"string": true, "error": true, "bool": true, "byte": true, "rune": true,
	"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
	"uintptr": true, "float32": true, "float64": true,
	"complex64": true, "complex128": true, "any": true,
}

func IsBuiltinType(t string) bool {
	return isBuiltinTypeSet[t]
}
