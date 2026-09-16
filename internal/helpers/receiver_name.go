package helpers

import "go/ast"

func ReceiverName(expr ast.Expr) string {
	for {
		switch t := expr.(type) {
		case *ast.Ident:
			return t.Name
		case *ast.StarExpr:
			expr = t.X
		case *ast.IndexExpr:
			expr = t.X
		case *ast.IndexListExpr:
			expr = t.X
		case *ast.ParenExpr:
			expr = t.X
		default:
			return ""
		}
	}
}
