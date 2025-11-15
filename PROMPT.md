read my @go-cli/PLAN.md and continue my execution plan, after implementing each command you must test 
the command, also when creating or updating a new go file, read the source javasctipt file to address 
the logic, and make sure the same logic is applied, it must have the same responses and behaviour.

When implementing a feature and its not fully yet implemented, or depends on another feature that is not done yet, you must create a placeholder for the missing feature, highlight it as a TODO and give it a comment to dont block my linter like "//nolint:staticcheck" and when the missing feature is done, you must replace the placeholder with the missing feature.

i want to replicate the same GUI for the cli which the javascript version already have, for terminal gui tasks i want to use https://github.com/charmbracelet/bubbletea as the main cli gui.

The javascript files reside in the folder docs/go-ai-agent-main/packages folder The documentation about each the javascript verion feature redise in the folder docs/go-ai-agent-main/docs.

Before continuing, i want you to read my current progress and the files i have implemented already on the folder go-cli, start reading the folders and then proceed to read files.

When finishing migrating a full function from javascript into go, you must run the linter and then build the application.

When the build passes and the linter has no error, you must test the implemented feature.

At first before you start crafting, you need to understand the JavaScript's intent and translate it idiomatically into Go, leveraging the existing Go project structure. To do this, you will need to get a comprehensive understanding of my entire Go project by recursively listing the go-cli directory and then reading relevant files to grasp current implementations.

My mock executor is a new feature that i started implementing, it must mock calls to AI, i will use it to
test my full execution flow. Soon i will implement multiple executors other than only gemini and mock, like qwen, openai.

-----------------------

Based on the updated PLAN.md and README.md, here's a prioritized plan for migrating your JavaScript CLI
   to Go:

  High-Priority:
   1. Implement Mock Executor and Executor Factory: Essential for robust testing and future AI model
      extensibility. Design interfaces, implement a basic mock, and set up the architecture for multiple
      executors.
   2. Improve Error Handling and User Feedback: Standardize error types, implement consistent reporting,
      and focus on user-friendly messages.
   3. Comprehensive Testing: Write unit and integration tests for all migrated commands and core services,
       leveraging the mock executor.

  Mid-Priority:
   4. Enhance Interactive UI: Expand charmbracelet/bubbletea UI to other commands like code-guide and
      grep-code for improved user experience.
   5. Tool Integration for AI Commands: Ensure AI commands like code-guide and grep-code can dynamically
      use GitService, FileSystemService, and ShellExecutionService by registering them as AI tools.
   6. Implement Secure API Key Storage/Clearing: Research and implement OS-specific secure storage for API
       keys.

  Lower-Priority / Future Enhancements:
   7. Remaining JavaScript CLI Commands: Review the original codebase for any missed functionality.
   8. Other Features: Prioritize IDE integration, theme changing, terminal keybinding, external editor,
      usage statistics, restore functionality, folder trust management, and MCP server management based on
       user demand.

  General Recommendations:
   * Maintain iterative development, regular linting, and updated documentation.
   * Continue using clear commit messages.

  What would you like to tackle first?
