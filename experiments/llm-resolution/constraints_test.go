package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEXPSC01_TheModuleHasNoDependencies(t *testing.T) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	mod := string(data)
	if !strings.Contains(mod, "module github.com/Roarge/sysml-federation/experiments/llm-resolution\n") {
		t.Errorf("go.mod doesn't declare the experiment's own module:\n%s", mod)
	}
	if strings.Contains(mod, "require") {
		t.Errorf("go.mod requires a dependency:\n%s", mod)
	}
}

// emptyInterfaces reports interface{} and any wherever they are written as
// types in one file.
func emptyInterfaces(fset *token.FileSet, f *ast.File) []string {
	var found []string
	isAny := func(e ast.Expr) bool {
		switch x := e.(type) {
		case *ast.InterfaceType:
			return x.Methods == nil || len(x.Methods.List) == 0
		case *ast.Ident:
			return x.Name == "any"
		}
		return false
	}
	check := func(e ast.Expr) {
		if e != nil && isAny(e) {
			found = append(found, fset.Position(e.Pos()).String())
		}
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Field:
			check(x.Type)
		case *ast.ValueSpec:
			check(x.Type)
		case *ast.TypeSpec:
			check(x.Type)
		case *ast.ArrayType:
			check(x.Elt)
		case *ast.MapType:
			check(x.Key)
			check(x.Value)
		case *ast.ChanType:
			check(x.Value)
		case *ast.Ellipsis:
			check(x.Elt)
		case *ast.StarExpr:
			check(x.X)
		case *ast.CompositeLit:
			check(x.Type)
		case *ast.TypeAssertExpr:
			check(x.Type)
		case *ast.CallExpr:
			if len(x.Args) == 1 {
				check(x.Fun) // a conversion such as any(v)
			}
		}
		return true
	})
	return found
}

func TestEXPSC02_NoEmptyInterfaceInValuePositions(t *testing.T) {
	files, _ := filepath.Glob("*.go")
	if len(files) == 0 {
		t.Fatal("no Go files")
	}
	fset := token.NewFileSet()
	for _, name := range files {
		f, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, at := range emptyInterfaces(fset, f) {
			t.Errorf("empty interface at %s", at)
		}
	}
	// The scan itself finds both forms.
	probe := "package p\nvar a any\nvar b interface{}\nfunc f(x ...interface{}) {}\nvar c = map[string]any{}\n"
	f, err := parser.ParseFile(fset, "probe.go", probe, 0)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(emptyInterfaces(fset, f)); n != 4 {
		t.Errorf("the scan found %d of the 4 empty interfaces in its own probe", n)
	}
}

func TestEXPSC04_ResultsAreIgnoredByGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("no git")
	}
	top, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		t.Skip("not in a git checkout")
	}
	root := strings.TrimSpace(string(top))
	for _, path := range []string{
		"experiments/llm-resolution/results/llm-resolution-2026-09-25T0000Z.jsonl",
		"experiments/llm-resolution/results/llm-resolution-2026-09-25T0000Z.md",
	} {
		if err := exec.Command("git", "-C", root, "check-ignore", "-q", path).Run(); err != nil {
			t.Errorf("git doesn't ignore %s", path)
		}
	}
}
