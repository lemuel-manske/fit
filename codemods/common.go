// Package codemods provides a set of functions to perform code modifications on Go source files.
package codemods

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

func IsNotNil(expr ast.Expr, name string) bool {
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok || bin.Op != token.NEQ {
		return false
	}

	left, leftOK := bin.X.(*ast.Ident)
	right, rightOK := bin.Y.(*ast.Ident)

	return leftOK &&
		rightOK &&
		left.Name == name &&
		right.Name == "nil"
}

func NameUsedInStatements(name string, stmts []ast.Stmt) bool {
	used := false

	for _, stmt := range stmts {
		ast.Inspect(stmt, func(n ast.Node) bool {
			if used {
				return false
			}

			id, ok := n.(*ast.Ident)
			if ok && id.Name == name {
				used = true
				return false
			}

			return true
		})

		if used {
			return true
		}
	}

	return false
}

func ContainsIdentifier(exprs []ast.Expr, name string) bool {
	for _, expr := range exprs {
		id, ok := expr.(*ast.Ident)
		if ok && id.Name == name {
			return true
		}
	}

	return false
}

func CanConvertDefineToAssign(
	pkg *packages.Package,
	ifStmt *ast.IfStmt,
	assign *ast.AssignStmt,
) bool {
	// go/types creates a scope for the entire if statement,
	// including variables declared by its Init.
	ifScope := pkg.TypesInfo.Scopes[ifStmt]
	if ifScope == nil {
		return false
	}

	// We specifically need variables that existed BEFORE entering
	// the if's own scope.
	outerScope := ifScope.Parent()
	if outerScope == nil {
		return false
	}

	for _, expr := range assign.Lhs {
		id, ok := expr.(*ast.Ident)
		if !ok {
			return false
		}

		if id.Name == "_" {
			continue
		}

		scope, obj := outerScope.LookupParent(id.Name, assign.Pos())
		if obj == nil || scope == nil {
			// No existing variable at this point.
			//
			// Example:
			//
			//     if err := foo(); err != nil {
			//
			// Here `err` MUST remain :=.
			return false
		}

		// Only an actual variable can appear on the LHS of `=`.
		if _, ok := obj.(*types.Var); !ok {
			return false
		}

		// Optional policy:
		// don't treat package-level variables as an existing local.
		//
		// This means:
		//
		//     var err error
		//
		//     func foo() {
		//         if err := bar(); ...
		//     }
		//
		// remains :=.
		if scope == pkg.Types.Scope() {
			return false
		}
	}

	return true
}
