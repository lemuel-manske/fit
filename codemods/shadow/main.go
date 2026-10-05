package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"codemods"

	"go/ast"
	"go/format"
	"go/token"

	"path/filepath"

	"golang.org/x/tools/go/packages"
)

func main() {
	write := flag.Bool("w", false, "write changes to files")
	flag.Parse()

	target := "."
	if flag.NArg() > 0 {
		target = flag.Arg(0)
	}

	target, err := filepath.Abs(target)
	if err != nil {
		codemods.Fatal(err)
	}

	cfg := &packages.Config{
		Mode: packages.LoadSyntax,
		Dir:  target,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		codemods.Fatal(err)
	}

	if n := packages.PrintErrors(pkgs); n != 0 {
		codemods.Fatalf(
			"refusing to rewrite: package loading reported %d error(s)",
			n,
		)
	}

	changedFiles := 0
	changedInitializers := 0

	for _, pkg := range pkgs {
		nFiles, nInitializers, err := rewritePackage(pkg, *write)
		if err != nil {
			codemods.Fatal(err)
		}

		changedFiles += nFiles
		changedInitializers += nInitializers
	}

	if *write {
		fmt.Printf(
			"rewrote %d initializer(s) in %d file(s)\n",
			changedInitializers,
			changedFiles,
		)
	} else {
		fmt.Printf(
			"would rewrite %d initializer(s) in %d file(s)\n",
			changedInitializers,
			changedFiles,
		)
	}
}

func rewritePackage(pkg *packages.Package, write bool) (int, int, error) {
	if pkg.Types == nil || pkg.TypesInfo == nil || pkg.Fset == nil {
		return 0, 0, fmt.Errorf(
			"package %q was loaded without type information",
			pkg.PkgPath,
		)
	}

	changedFiles := 0
	changedInitializers := 0

	for i, file := range pkg.Syntax {
		if i >= len(pkg.CompiledGoFiles) {
			return changedFiles, changedInitializers, fmt.Errorf(
				"package %q: syntax/file mismatch",
				pkg.PkgPath,
			)
		}

		filename := pkg.CompiledGoFiles[i]

		// Avoid rewriting generated files.
		if ast.IsGenerated(file) {
			continue
		}

		n := rewriteFileAST(pkg, file)
		if n == 0 {
			continue
		}

		changedFiles++
		changedInitializers += n

		fmt.Printf("%s: %d rewrite(s)\n", filename, n)

		if !write {
			continue
		}

		if err := writeFormattedFile(pkg.Fset, file, filename); err != nil {
			return changedFiles, changedInitializers, err
		}
	}

	return changedFiles, changedInitializers, nil
}

func rewriteFileAST(pkg *packages.Package, file *ast.File) int {
	changed := 0

	ast.Inspect(file, func(node ast.Node) bool {
		ifStmt, ok := node.(*ast.IfStmt)
		if !ok {
			return true
		}

		assign, ok := ifStmt.Init.(*ast.AssignStmt)
		if !ok || assign.Tok != token.DEFINE {
			return true
		}

		if !codemods.ContainsIdentifier(assign.Lhs, "err") {
			return true
		}

		if !codemods.CanConvertDefineToAssign(pkg, ifStmt, assign) {
			return true
		}

		pos := pkg.Fset.Position(assign.Pos())

		fmt.Printf(
			"  %s:%d:%d: := -> =\n",
			pos.Filename,
			pos.Line,
			pos.Column,
		)

		assign.Tok = token.ASSIGN
		changed++

		return true
	})

	return changed
}

func writeFormattedFile(
	fset *token.FileSet,
	file *ast.File,
	filename string,
) error {
	var buf bytes.Buffer

	if err := format.Node(&buf, fset, file); err != nil {
		return fmt.Errorf("format %s: %w", filename, err)
	}

	info, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("stat %s: %w", filename, err)
	}

	tmp, err := os.CreateTemp(
		filepath.Dir(filename),
		"."+filepath.Base(filename)+".shadowfix-*",
	)
	if err != nil {
		return fmt.Errorf("create temp file for %s: %w", filename, err)
	}

	tmpName := tmp.Name()
	keep := false

	defer func() {
		if !keep {
			_ = os.Remove(tmpName)
		}
	}()

	if err := tmp.Chmod(info.Mode()); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temp file for %s: %w", filename, err)
	}

	if _, err := tmp.Write(buf.Bytes()); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp file for %s: %w", filename, err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file for %s: %w", filename, err)
	}

	if err := os.Rename(tmpName, filename); err != nil {
		return fmt.Errorf("replace %s: %w", filename, err)
	}

	keep = true
	return nil
}
