package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"codemods"

	"io/fs"

	"go/ast"
	"go/format"
	"go/parser"
	"go/token"

	"path/filepath"
)

func main() {
	write := flag.Bool("w", false, "write changes to files")
	flag.Parse()

	root := "."
	if flag.NArg() > 0 {
		root = flag.Arg(0)
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			switch d.Name() {
			case ".git":
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(path) != ".go" {
			return nil
		}

		changed, out, err := rewriteFile(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if !changed {
			return nil
		}

		fmt.Println(path)

		if *write {
			return os.WriteFile(path, out, 0o644)
		}

		_, err = os.Stdout.Write(out)
		return err
	})
	if err != nil {
		panic(err)
	}
}

func rewriteFile(path string) (bool, []byte, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return false, nil, err
	}

	fset := token.NewFileSet()

	file, err := parser.ParseFile(
		fset,
		path,
		src,
		parser.ParseComments,
	)
	if err != nil {
		return false, nil, err
	}

	changed := false

	ast.Inspect(file, func(n ast.Node) bool {
		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}

		if rewriteBlock(block) {
			changed = true
		}

		return true
	})

	if !changed {
		return false, src, nil
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return false, nil, err
	}

	return true, buf.Bytes(), nil
}

func rewriteBlock(block *ast.BlockStmt) bool {
	changed := false

	for i := 0; i+1 < len(block.List); i++ {
		assign, ok := block.List[i].(*ast.AssignStmt)
		if !ok {
			continue
		}

		ifStmt, ok := block.List[i+1].(*ast.IfStmt)
		if !ok {
			continue
		}

		name, ok := matches(assign, ifStmt)
		if !ok {
			continue
		}

		// `:=` introduces a new variable. Moving it into the if
		// narrows its scope, so don't rewrite if it is needed later.
		//
		// `=` assigns an already-existing variable, so moving the
		// assignment into IfStmt.Init does not change its scope.
		if assign.Tok == token.DEFINE &&
			codemods.NameUsedInStatements(name, block.List[i+2:]) {
			continue
		}

		ifStmt.Init = assign

		// Remove the standalone assignment.
		block.List = append(
			block.List[:i],
			block.List[i+1:]...,
		)

		changed = true
		i--
	}

	return changed
}

func matches(assign *ast.AssignStmt, ifStmt *ast.IfStmt) (string, bool) {
	// Match either:
	//
	//     err := something()
	//
	// or:
	//
	//     err = something()
	//
	if assign.Tok != token.DEFINE && assign.Tok != token.ASSIGN {
		return "", false
	}

	if len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return "", false
	}

	id, ok := assign.Lhs[0].(*ast.Ident)
	if !ok {
		return "", false
	}

	name := id.Name

	if ifStmt.Init != nil {
		return "", false
	}

	// Match:
	//
	//     if err != nil {
	if !codemods.IsNotNil(ifStmt.Cond, name) {
		return "", false
	}

	// Match:
	//
	//     return err
	if len(ifStmt.Body.List) != 1 {
		return "", false
	}

	ret, ok := ifStmt.Body.List[0].(*ast.ReturnStmt)
	if !ok || len(ret.Results) != 1 {
		return "", false
	}

	returned, ok := ret.Results[0].(*ast.Ident)
	if !ok || returned.Name != name {
		return "", false
	}

	return name, true
}
