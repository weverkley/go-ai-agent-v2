package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/generative-ai-go/genai"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Go AI Agent project",
	Long:  `The init command analyzes the project and creates a tailored GEMINI.md file for instructional context.`, //nolint:staticcheck
	Run: func(cmd *cobra.Command, args []string) {
		projectRoot, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			return
		}
		geminiMdPath := filepath.Join(projectRoot, "GEMINI.md")

		// Check if GEMINI.md already exists
		if _, err := os.Stat(geminiMdPath); !os.IsNotExist(err) {
			fmt.Printf("A GEMINI.md file already exists in this directory. No changes were made.\n")
			return
		}

		// Create an empty GEMINI.md file
		err = os.WriteFile(geminiMdPath, []byte(""), 0644)
		if err != nil {
			fmt.Printf("Error creating GEMINI.md file: %v\n", err)
			return
		}

		fmt.Println("Empty GEMINI.md created. Now analyzing the project to populate it.")

		prompt := `
You are an AI agent that brings the power of Gemini directly into the terminal. Your task is to analyze the current directory and generate a comprehensive GEMINI.md file to be used as instructional context for future interactions.

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

**GEMINI.md Content Generation:**

**For a Code Project:**

*   **Project Overview:** Write a clear and concise summary of the project's purpose, main technologies, and architecture.
*   **Building and Running:** Document the key commands for building, running, and testing the project. Infer these from the files you've read (e.g., 'scripts' in 'package.json', 'Makefile', etc.). If explicit commands cannot be inferred, state that and suggest common commands for the detected project type (e.g., 'npm install && npm start' for Node.js projects).
*   **Development Conventions:** Describe any coding styles, testing practices, or contribution guidelines you can infer from the codebase.

**For a Non-Code Project:**

*   **Directory Overview:** Describe the purpose and contents of the directory. What is it for? What kind of information does it hold?
*   **Key Files:** List the most important files and briefly explain what they contain.
*   **Usage:** Explain how the contents of this directory are intended to be used.

**Final Output:**

Write the complete content to the 'GEMINI.md' file. The output must be well-formatted Markdown.
`

		resp, err := executor.GenerateContent(
			&genai.Content{
				Parts: []genai.Part{genai.Text(prompt)},
				Role:  "user",
			},
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating GEMINI.md content: %v\n", err)
			os.Exit(1)
		}

		generatedContent := ""
		for _, part := range resp.Candidates[0].Content.Parts {
			if text, ok := part.(genai.Text); ok {
				generatedContent += string(text)
			}
		}

		err = os.WriteFile(geminiMdPath, []byte(generatedContent), 0644)
		if err != nil {
			fmt.Printf("Error writing generated content to GEMINI.md: %v\n", err)
			return
		}

		fmt.Println("GEMINI.md populated with generated content.")
	},
}
