package tools

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go-ai-agent-v2/go-cli/pkg/types"
)

// ReadFileTool represents the read-file tool.
type ReadFileTool struct {
	*types.BaseDeclarativeTool
}

// NewReadFileTool creates a new instance of ReadFileTool.
func NewReadFileTool() *ReadFileTool {
	return &ReadFileTool{
		types.NewBaseDeclarativeTool(
			"read_file",
			"read_file",
			"Reads and returns the content of a specified file. If the file is large, the content will be truncated. The tool's response will clearly indicate if truncation has occurred and will provide details on how to read more of the file using the 'offset' and 'limit' parameters. Handles text, images (PNG, JPG, GIF, WEBP, SVG, BMP), and PDF files. For text files, it can read specific line ranges.",
			types.KindOther, // Assuming KindOther for now
			types.JsonSchemaObject{
				Type: "object",
				Properties: map[string]types.JsonSchemaProperty{
					"absolute_path": {
						Type:        "string",
						Description: "The absolute path to the file to read (e.g., '/home/user/project/file.txt'). Relative paths are not supported. You must provide an absolute path.",
					},
					"offset": {
						Type:        "number",
						Description: "Optional: For text files, the 0-based line number to start reading from. Requires 'limit' to be set. Use for paginating through large files.",
					},
					"limit": {
						Type:        "number",
						Description: "Optional: For text files, maximum number of lines to read. Use with 'offset' to paginate through large files. If omitted, reads the entire file (if feasible, up to a default limit).",
					},
				},
				Required: []string{"absolute_path"},
			},
			false, // isOutputMarkdown
			false, // canUpdateOutput
			nil,   // MessageBus
		),
	}
}

// Execute performs a read-file operation.
func (t *ReadFileTool) Execute(args map[string]any) (types.ToolResult, error) {
	absolutePath, ok := args["absolute_path"].(string)
	if !ok || absolutePath == "" {
		return types.ToolResult{}, fmt.Errorf("invalid or missing 'absolute_path' argument")
	}

	var offset int
	if o, ok := args["offset"].(float64); ok {
		offset = int(o)
	}

	var limit int
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	// Check if file exists
	info, err := os.Stat(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			return types.ToolResult{}, fmt.Errorf("file not found: %s", absolutePath)
		}
		return types.ToolResult{}, fmt.Errorf("failed to get file info for %s: %w", absolutePath, err)
	}

	// Check if it's a directory
	if info.IsDir() {
		return types.ToolResult{}, fmt.Errorf("path is a directory, not a file: %s", absolutePath)
	}

	// For now, assume all are text files.
	// TODO: Implement image/pdf handling.
	ext := strings.ToLower(filepath.Ext(absolutePath))
	if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".bmp" || ext == ".pdf" {
		return types.ToolResult{
			LLMContent:    fmt.Sprintf("Content of %s (binary file, not displayed)", absolutePath),
			ReturnDisplay: fmt.Sprintf("Content of %s (binary file, not displayed)", absolutePath),
		}, nil
	}

	file, err := os.Open(absolutePath)
	if err != nil {
		return types.ToolResult{}, fmt.Errorf("failed to open file %s: %w", absolutePath, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return types.ToolResult{}, fmt.Errorf("error reading file %s: %w", absolutePath, err)
	}

	originalLineCount := len(lines)
	isTruncated := false
	linesShownStart := 0
	linesShownEnd := originalLineCount

	if offset > 0 {
		linesShownStart = offset
	}
	if limit > 0 {
		linesShownEnd = linesShownStart + limit
	}

	if linesShownStart >= originalLineCount {
		return types.ToolResult{
			LLMContent:    fmt.Sprintf("Offset %d is beyond the end of the file (total lines: %d)", offset, originalLineCount),
			ReturnDisplay: fmt.Sprintf("Offset %d is beyond the end of the file (total lines: %d)", offset, originalLineCount),
		}, nil
	}

	if linesShownEnd > originalLineCount {
		linesShownEnd = originalLineCount
	}

	if linesShownStart > 0 || linesShownEnd < originalLineCount {
		isTruncated = true
	}

	var contentBuilder strings.Builder
	for i := linesShownStart; i < linesShownEnd; i++ {
		contentBuilder.WriteString(lines[i])
		contentBuilder.WriteString("\n")
	}

	var llmContent strings.Builder
	if isTruncated {
		nextOffset := linesShownEnd
		llmContent.WriteString("\nIMPORTANT: The file content has been truncated.\n")
		llmContent.WriteString(fmt.Sprintf("Status: Showing lines %d-%d of %d total lines.\n", linesShownStart+1, linesShownEnd, originalLineCount))
		llmContent.WriteString(fmt.Sprintf("Action: To read more of the file, you can use the 'offset' and 'limit' parameters in a subsequent 'read_file' call. For example, to read the next section of the file, use offset: %d.\n", nextOffset))
		llmContent.WriteString("\n--- FILE CONTENT (truncated) ---\n")
	}
	llmContent.WriteString(contentBuilder.String())

	return types.ToolResult{
		LLMContent:    llmContent.String(),
		ReturnDisplay: llmContent.String(), // For now, same as LLMContent
	}, nil
}
