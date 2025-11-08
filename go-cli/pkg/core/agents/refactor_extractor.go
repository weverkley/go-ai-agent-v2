package agents

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"strings"
)

// ExtractedFunction represents the details of an extracted function/method.
type ExtractedFunction struct {
	NewCode string
	OldCode string
}

// ExtractFunction extracts a code block into a new function or method.
func ExtractFunction(
	filePath string,
	startLine, endLine int,
	newFunctionName string,
	receiver string, // e.g., "s *MyStruct" for a method, empty for a function
) (*ExtractedFunction, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// Find the statements to extract
	var (
		targetStmts []ast.Stmt
	)

	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		// Check if the node is a statement and within the target line range
		if stmt, ok := n.(ast.Stmt); ok {
			start := fset.Position(stmt.Pos()).Line
			end := fset.Position(stmt.End()).Line

			if start >= startLine && end <= endLine {
				targetStmts = append(targetStmts, stmt)
			}
		}
		return true
	})

	if len(targetStmts) == 0 {
		return nil, fmt.Errorf("no statements found in the specified line range")
	}

	// Analyze dependencies (simplified: just collect identifiers)
	// A real implementation would need type checking to distinguish variables, functions, methods, etc.
	usedVars := make(map[string]bool)
	declaredVars := make(map[string]bool)

	for _, stmt := range targetStmts {
		ast.Inspect(stmt, func(n ast.Node) bool {
			if ident, ok := n.(*ast.Ident); ok {
				// Ignore blank identifier
				if ident.Name == "_" {
					return true
				}

				// Check if it's a declaration within the extracted block
				if assign, isAssign := ident.Obj.Decl.(*ast.AssignStmt); isAssign && assign.Tok == token.DEFINE {
					for _, lhs := range assign.Lhs {
						if lhsIdent, isLHSIdent := lhs.(*ast.Ident); isLHSIdent {
							declaredVars[lhsIdent.Name] = true
						}
					}
				} else if valSpec, isValSpec := ident.Obj.Decl.(*ast.ValueSpec); isValSpec {
					for _, name := range valSpec.Names {
						declaredVars[name.Name] = true
					}
				} else if field, isField := ident.Obj.Decl.(*ast.Field); isField {
					for _, name := range field.Names {
						declaredVars[name.Name] = true
					}
				}

				// If not declared within the block, it's a dependency
				if _, isDeclared := declaredVars[ident.Name]; !isDeclared {
					usedVars[ident.Name] = true
				}
			}
			return true
		})
	}

	// Determine parameters and return values (very simplified)
	// This needs significant improvement with type information
	var params []string
	var returns []string
	inferredTypes := make(map[string]string)

	// Basic type inference
	for _, stmt := range targetStmts {
		ast.Inspect(stmt, func(n ast.Node) bool {
			switch expr := n.(type) {
			case *ast.BinaryExpr:
				if expr.Op == token.ADD || expr.Op == token.SUB || expr.Op == token.MUL || expr.Op == token.QUO {
					if left, ok := expr.X.(*ast.Ident); ok {
						if _, declared := declaredVars[left.Name]; !declared {
							inferredTypes[left.Name] = "int" // Assume int for arithmetic
						}
					}
					if right, ok := expr.Y.(*ast.Ident); ok {
						if _, declared := declaredVars[right.Name]; !declared {
							inferredTypes[right.Name] = "int" // Assume int for arithmetic
						}
					}
				}
			case *ast.CallExpr:
				if fun, ok := expr.Fun.(*ast.SelectorExpr); ok {
					if x, ok := fun.X.(*ast.Ident); ok {
						if x.Name == "strings" { // Basic check for strings package functions
							for _, arg := range expr.Args {
								if argIdent, isIdent := arg.(*ast.Ident); isIdent {
									if _, declared := declaredVars[argIdent.Name]; !declared {
										inferredTypes[argIdent.Name] = "string"
									}
								}
							}
						}
					}
				}
			}
			return true
		})
	}

	for varName := range usedVars {
		inferredType, ok := inferredTypes[varName]
		if !ok {
			inferredType = "interface{}" // Default to interface{} if no type inferred
		}
		params = append(params, varName+" "+inferredType)
	}

	// Construct the new function/method
	var newFuncBuilder strings.Builder
	if receiver != "" {
		newFuncBuilder.WriteString(fmt.Sprintf("func (%s) %s(", receiver, newFunctionName))
	} else {
		newFuncBuilder.WriteString(fmt.Sprintf("func %s(", newFunctionName))
	}
	newFuncBuilder.WriteString(strings.Join(params, ", "))
	newFuncBuilder.WriteString(") ")

	if len(returns) > 0 {
		newFuncBuilder.WriteString(fmt.Sprintf("(%s) ", strings.Join(returns, ", ")))
	}

	newFuncBuilder.WriteString("{\n")
	for _, stmt := range targetStmts {
		var buf bytes.Buffer
		format.Node(&buf, fset, stmt)
		newFuncBuilder.WriteString(buf.String())
		newFuncBuilder.WriteString("\n")
	}
	newFuncBuilder.WriteString("}\n\n")

	// Replace the original statements with a call to the new function/method
	var callArgs []string
	for varName := range usedVars {
		callArgs = append(callArgs, varName)
	}
	callString := fmt.Sprintf("%s(%s)", newFunctionName, strings.Join(callArgs, ", "))

	// Read the original file content
	originalContentBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read original file: %w", err)
	}
	originalContent := string(originalContentBytes)
	lines := strings.Split(originalContent, "\n")

	// Build the modified content
	var modifiedLines []string
	for i, line := range lines {
		if i+1 < startLine || i+1 > endLine {
			modifiedLines = append(modifiedLines, line)
		} else if i+1 == startLine {
			// Indent the call to match the first line of the extracted block
			indent := getIndent(line)
			modifiedLines = append(modifiedLines, indent+callString)
		}
	}

	modifiedContent := strings.Join(modifiedLines, "\n")

	// Insert the new function/method at the end of the file (simplified)
	modifiedContent += "\n" + newFuncBuilder.String()

	return &ExtractedFunction{
		NewCode: modifiedContent,
		OldCode: originalContent,
	}, nil
}

// getIndent returns the leading whitespace of a string.
func getIndent(s string) string {
	for i, r := range s {
		if r != ' ' && r != '\t' {
			return s[:i]
		}
	}
	return s // All whitespace
}
