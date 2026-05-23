package engine_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestProductionPromptLiteralsDeclarePresentation(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to locate current test file")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
	internalDir := filepath.Join(root, "internal")

	fset := token.NewFileSet()
	var missing []string
	err := filepath.WalkDir(internalDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := filepath.Base(path)
			if base == "vendor" || strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.CompositeLit)
			if !ok || !isModelPromptLiteral(lit.Type) {
				return true
			}
			if !promptLiteralHasPresentation(lit) {
				pos := fset.Position(lit.Pos())
				rel, _ := filepath.Rel(root, pos.Filename)
				missing = append(missing, rel+":"+itoa(pos.Line))
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(missing) > 0 {
		t.Fatalf("model.Prompt literals missing Presentation:\n%s", strings.Join(missing, "\n"))
	}
}

func isModelPromptLiteral(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.SelectorExpr:
		if t.Sel.Name != "Prompt" {
			return false
		}
		x, ok := t.X.(*ast.Ident)
		return ok && x.Name == "model"
	case *ast.Ident:
		return t.Name == "Prompt"
	default:
		return false
	}
}

func promptLiteralHasPresentation(lit *ast.CompositeLit) bool {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if ok && key.Name == "Presentation" {
			return true
		}
	}
	return false
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
