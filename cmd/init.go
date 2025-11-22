package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"go-ai-agent-v2/go-cli/pkg/core"
	"go-ai-agent-v2/go-cli/pkg/types"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Go AI Agent project",
	Long:  `The init command analyzes the project and creates a tailored GOAIAGENT.md file for instructional context.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		projectRoot, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			return
		}
		geminiMdPath := filepath.Join(projectRoot, "GOAIAGENT.md")

		// Check if GOAIAGENT.md already exists
		if _, err := os.Stat(geminiMdPath); !os.IsNotExist(err) {
			fmt.Printf("A GOAIAGENT.md file already exists in this directory. No changes were made.\n")
			return
		}

		var generatedContent string

		// Check if the global executor is a mock executor.
		if _, ok := executor.(*core.MockExecutor); ok {
			fmt.Println("Mock executor detected. Generating mock GOAIAGENT.md content.")
			generatedContent = `# Project Overview

This is a Go project for the Go AI Agent. It appears to be a command-line interface tool for interacting with AI models.

## Building and Running

- To build the project, run: ` + "`go build`" + `
- To run tests, use: ` + "`go test ./...`" + `

## Development Conventions

- The project uses Go modules for dependency management (` + "`go.mod`" + `).
- Code seems to be organized into ` + "`cmd`" + ` for main applications and ` + "`pkg`" + ` for shared libraries.
- Standard Go formatting is expected.
`
		} else {
			fmt.Println("Empty GOAIAGENT.md created. Now analyzing the project to populate it.")

			prompt := `
You are an AI agent that brings the power of Gemini directly into the terminal. Your task is to analyze the current directory and generate a comprehensive GOAIAGENT.md file to be used as instructional context for future interactions.

**Analysis Process:**

1.  **Initial Exploration:**
    *   Start by listing the files and directories to get a high-level overview of the structure.
        *   Read the README file (e.g., 'README.md', 'README.txt') if it exists. This is often the best place to start.

    2.  **Iterative Deep Dive (up to 10 files):**
        *   Based on your initial findings, select a few files that seem most important (e.g., configuration files, main source files, documentation).
        *   Read them. As you learn more, refine your understanding and decide which files to read next. You don't need to decide all 10 files at once. Let your discoveries guide your exploration.

    3.  **Identify Project Type:**
        *   **Code Project:** Look for clues like 'package.json', 'requirements.txt', 'pom.xml', 'go.mod', 'Cargo.toml', 'build.gradle', or a 'src' directory. If you find them, this is likely a software project.
        *   **Non-Code Project:** If you don't find code-related files, this might be a directory for documentation, research papers, notes, or something else.

**GOAIAGENT.md Content Generation:**

**For a Code Project:**

*   **Project Overview:** Write a clear and concise summary of the project's purpose, main technologies, and architecture.
*   **Building and Running:** Document the key commands for building, running, and testing the project. Infer these from the files you've read (e.g., 'scripts' in 'package.json', 'Makefile', etc.). If explicit commands cannot be inferred, state that and suggest common commands for the detected project type (e.g., 'npm install && npm start' for Node.js projects).
*   **Development Conventions:** Describe any coding styles, testing practices, or contribution guidelines you can infer from the codebase.

**For a Non-Code Project:**

*   **Directory Overview:** Describe the purpose and contents of the directory. What is it for? What kind of information does it hold?
*   **Key Files:** List the most important files and briefly explain what they contain.
*   **Usage:** Explain how the contents of this directory are intended to be used.

**Final Output:**

Write the complete content to the 'GOAIAGENT.md' file. The output must be well-formatted Markdown.
`

			resp, err := executor.GenerateContent(
				&types.Content{
					Parts: []types.Part{{Text: prompt}},
					Role:  "user",
				},
			)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating GOAIAGENT.md content: %v\n", err)
				os.Exit(1)
			}

			if resp != nil && len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
				for _, part := range resp.Candidates[0].Content.Parts {
					if part.Text != "" { // Directly access Text field
						generatedContent += part.Text
					}
				}
			}
		}

		err = os.WriteFile(geminiMdPath, []byte(generatedContent), 0644)
		if err != nil {
			fmt.Printf("Error writing generated content to GOAIAGENT.md: %v\n", err)
			return
		}

		fmt.Println("GOAIAGENT.md populated with generated content.")
	},
}
