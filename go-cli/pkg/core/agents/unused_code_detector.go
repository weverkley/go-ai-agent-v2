package agents

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// UnusedFunction represents an unused function or method.
type UnusedFunction struct {
	FilePath string
	Name     string
	Type     string // "function" or "method"
}

// FindUnusedFunctions finds all unused functions and methods in a given directory.
func FindUnusedFunctions(dir string) ([]UnusedFunction, error) {
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	definedFunctions := make(map[string]UnusedFunction)
	calledFunctions := make(map[string]bool)

	for _, pkg := range packages {
		for _, file := range pkg.Files {
			filePath := fset.File(file.Pos()).Name()

			// Collect all function and method declarations
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok {
					funcName := fn.Name.Name
					if fn.Recv != nil {
						// This is a method
						if len(fn.Recv.List) > 0 {
							if ident, ok := fn.Recv.List[0].Type.(*ast.Ident); ok {
								methodName := ident.Name + "." + funcName
								definedFunctions[methodName] = UnusedFunction{
									FilePath: filePath,
									Name:     methodName,
									Type:     "method",
								}
							} else if starExpr, ok := fn.Recv.List[0].Type.(*ast.StarExpr); ok {
								if ident, ok := starExpr.X.(*ast.Ident); ok {
									methodName := ident.Name + "." + funcName
									definedFunctions[methodName] = UnusedFunction{
										FilePath: filePath,
										Name:     methodName,
										Type:     "method",
									}
								}
							}
						}
					} else {
						// This is a function
						definedFunctions[funcName] = UnusedFunction{
							FilePath: filePath,
							Name:     funcName,
							Type:     "function",
						}
					}
				}
			}

			// Collect all function and method calls
			ast.Inspect(file, func(n ast.Node) bool {
				if callExpr, ok := n.(*ast.CallExpr); ok {
					if fun, ok := callExpr.Fun.(*ast.Ident); ok {
						calledFunctions[fun.Name] = true
					} else if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
						// This could be a method call or a qualified identifier (e.g., fmt.Println)
						if ident, ok := selExpr.X.(*ast.Ident); ok {
							// Check if it's a method call (e.g., receiver.Method)
							// This is a simplification, a full type checker would be needed for accuracy
							methodName := ident.Name + "." + selExpr.Sel.Name
							calledFunctions[methodName] = true
						} else if _, ok := selExpr.X.(*ast.CallExpr); ok {
							// Handle cases like (obj.Method()).AnotherMethod()
							// For simplicity, we'll ignore these for now or assume they are used
						}
						// Also consider qualified identifiers (e.g., "fmt.Println")
						calledFunctions[selExpr.Sel.Name] = true
					}
				}
				return true
			})
		}
	}

	var unusedFunctions []UnusedFunction
	for name, fn := range definedFunctions {
		if !calledFunctions[name] {
			unusedFunctions = append(unusedFunctions, fn)
		}
	}

	return unusedFunctions, nil
}

// Helper function to get all Go files in a directory.
func getGoFiles(dir string) ([]string, error) {
	var goFiles []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			goFiles = append(goFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return goFiles, nil
}
