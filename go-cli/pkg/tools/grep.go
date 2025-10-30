package tools

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/core/agents"

	"github.com/google/generative-ai-go/genai"
)

// GrepTool represents the grep tool.
type GrepTool struct{}

// NewGrepTool creates a new instance of GrepTool.
func NewGrepTool() *GrepTool {
	return &GrepTool{}
}

// Name returns the name of the tool.
func (t *GrepTool) Name() string {
	return "grep"
}

// Definition returns the tool's definition for the Gemini API.
func (t *GrepTool) Definition() *genai.Tool {
	return &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{
				Name:        t.Name(),
				Description: "Searches for a regular expression pattern within files in a specified directory.",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"pattern": {
							Type:        genai.TypeString,
							Description: "The regular expression (regex) pattern to search for.",
						},
						"path": {
							Type:        genai.TypeString,
							Description: "Optional: The path to the directory to search within. Defaults to the current directory.",
						},
						"include": {
							Type:        genai.TypeString,
							Description: "Optional: A glob pattern to filter which files are searched (e.g., '*.js', 'src/**').",
						},
					},
					Required: []string{"pattern"},
				},
			},
		},
	}
}

// GrepMatch represents a single grep match.
type GrepMatch struct {
	FilePath   string
	LineNumber int
	Line       string
}

// Execute performs a grep search.
func (t *GrepTool) Execute(args map[string]any) (agents.ToolResult, error) {
	// Extract arguments
	pattern, ok := args["pattern"].(string)
	if !ok {
		return agents.ToolResult{}, fmt.Errorf("invalid or missing 'pattern' argument")
	}

	searchPath := "." // Default to current directory
	if p, ok := args["path"].(string); ok && p != "" {
		searchPath = p
	}

	var include string
	if i, ok := args["include"].(string); ok {
		include = i
	}

	// Compile the regex pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		return agents.ToolResult{}, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Resolve the search path
	absSearchPath, err := filepath.Abs(searchPath)
	if err != nil {
		return agents.ToolResult{}, fmt.Errorf("failed to resolve absolute path for %s: %w", searchPath, err)
	}

	var allMatches []GrepMatch
	matchesByFile := make(map[string][]GrepMatch)

	err = filepath.Walk(absSearchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Filter by include glob pattern if provided
		if include != "" {
			match, err := filepath.Match(include, info.Name())
			if err != nil {
				return fmt.Errorf("invalid include pattern: %w", err)
			}
			if !match {
				return nil // Skip file if it doesn't match include pattern
			}
		}

		file, err := os.Open(path)
		if err != nil {
			// Silently ignore files that can't be opened
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNumber := 1
		for scanner.Scan() {
			line := scanner.Text()
			if re.MatchString(line) {
				match := GrepMatch{
					FilePath:   path,
					LineNumber: lineNumber,
					Line:       line,
				}
				allMatches = append(allMatches, match)
				matchesByFile[path] = append(matchesByFile[path], match)
			}
			lineNumber++
		}

		if err := scanner.Err(); err != nil {
			// Silently ignore read errors
			return nil
		}

		return nil
	})

	if err != nil {
		return agents.ToolResult{}, fmt.Errorf("error walking the path %s: %w", absSearchPath, err)
	}

	if len(allMatches) == 0 {
		return agents.ToolResult{
			LLMContent:    fmt.Sprintf("No matches found for pattern \"%s\" in path \"%s\"", pattern, searchPath),
			ReturnDisplay: fmt.Sprintf("No matches found for pattern \"%s\" in path \"%s\"", pattern, searchPath),
		}, nil
	}

	var llmContent strings.Builder
	llmContent.WriteString(fmt.Sprintf("Found %d matches for pattern \"%s\" in path \"%s\":\n---\n", len(allMatches), pattern, searchPath))

	for filePath, matches := range matchesByFile {
		llmContent.WriteString(fmt.Sprintf("File: %s\n", filePath))
		for _, match := range matches {
			llmContent.WriteString(fmt.Sprintf("L%d: %s\n", match.LineNumber, strings.TrimSpace(match.Line)))
		}
		llmContent.WriteString("---\n\n")
	}

	return agents.ToolResult{
		LLMContent:    llmContent.String(),
		ReturnDisplay: llmContent.String(),
	}, nil
}
