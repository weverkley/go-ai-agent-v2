# Implementation Plan for New Tools and Sub-Agents

This document outlines the step-by-step implementation plan for the new tools and sub-agents proposed to enhance this agent's software engineering capabilities. The new features are designed to align with modern software development principles like TDD, SOLID, and safe refactoring practices.

## Guiding Principles

-   **Test-Driven Development (TDD/BDD)**: Natively support a tight "test-code-refactor" loop.
-   **SOLID & KISS**: Favor code-aware, intelligent tools over simple text manipulation for safer and simpler refactoring.
-   **DDD Insights**: Provide tools to visualize and understand the codebase as a domain model.
-   **Workflow Automation**: Integrate common development steps like building, linting, and committing into the agent's core capabilities.

---

## Part 1: Foundational Engineering Tools

These tools provide the basic, reusable building blocks for more advanced workflows and agents.

### 1.1. Tool: `run_tests`

*   **Description**: A dedicated tool to execute project-specific tests. It will intelligently discover and use the correct test command for the project (e.g., `go test`, `npm test`, `pytest`). This is a safer and more abstract alternative to using the generic `execute_command`.
*   **Tool Definition**:
    *   **Name**: `run_tests`
    *   **Parameters**:
        *   `target` (string, optional): The specific test file, directory, or test name/pattern to run. Defaults to all tests.
        *   `coverage` (boolean, optional): If true, generates and includes a code coverage report in the output.

*   **Implementation Steps**:
    1.  **Create `pkg/tools/run_tests.go`**:
        *   Define a `RunTestsTool` struct that depends on the `services.ShellExecutionService`.
        *   Create a `NewRunTestsTool` constructor that defines the name, description, and parameter schema.
        *   Implement the `Execute` method. Its logic should:
            1.  Check for project markers like `go.mod` or `package.json` to determine the test command.
            2.  Construct the command string (e.g., `go test [target] -cover` or `npm test -- [target]`).
            3.  Use the `shellService` to execute the command.
            4.  Return the stdout and stderr from the test run as the result.
    2.  **Create `pkg/tools/run_tests_test.go`**:
        *   Create unit tests for `RunTestsTool`.
        *   Mock the `shellService` dependency.
        *   Include test cases for different project types (Go, Node.js) and the handling of `target` and `coverage` parameters.
    3.  **Register in `pkg/tools/register.go`**:
        *   In `RegisterAllTools`, add a line to register an instance of the new tool: `registry.Register(NewRunTestsTool(shellService))`.

### 1.2. Tool: `git_commit`

*   **Description**: Stages files and creates a Git commit, integrating a core version control step into the agent's workflow.
*   **Tool Definition**:
    *   **Name**: `git_commit`
    *   **Parameters**:
        *   `message` (string, required): The commit message.
        *   `files_to_stage` (array of strings, optional): A list of specific files to stage. If not provided, all tracked and modified files will be staged (`git add .`).
*   **Implementation Steps**:
    1.  **Enhance `pkg/services/git_service.go`**:
        *   Add `StageFiles(dir string, files []string) error` and `Commit(dir, message string) error` methods to the `GitService`.
    2.  **Create `pkg/tools/git_commit.go`**:
        *   Define the `GitCommitTool` struct with a dependency on `services.GitService`.
        *   Implement the `Execute` method to parse arguments, call the new `gitService` methods, and return a success or failure message.
    3.  **Create `pkg/tools/git_commit_test.go`**:
        *   Write unit tests, mocking the `gitService` to verify that `StageFiles` and `Commit` are called with the correct parameters.
    4.  **Register in `pkg/tools/register.go`**.

---

## Part 2: Advanced AST-Based Refactoring Tools

These tools require parsing the source code's Abstract Syntax Tree (AST) for safe and precise modifications, moving beyond simple text replacement. They are essential for advanced refactoring.

### 2.1. Tool: `find_references`

*   **Description**: Finds all usages of a specific symbol (function, variable, etc.) in the codebase. This is a crucial, read-only analysis tool for assessing the impact of a change.
*   **Tool Definition**:
    *   **Name**: `find_references`
    *   **Parameters**:
        *   `file_path` (string, required): Path to the file containing the symbol's definition.
        *   `line` (integer, required): The line number of the symbol.
        *   `column` (integer, required): The column number of the symbol.
*   **Implementation Steps**:
    1.  **Backend Logic**:
        *   Create a new package `pkg/analysis`.
        *   Inside, create a function `FindSymbolReferences(filePath, line, column)`. This is a complex task. For Go, this can be implemented robustly using the `go/types` and `golang.org/x/tools/go/ssa` packages to build a type-checked and fully resolved representation of the code.
    2.  **Create `pkg/tools/find_references.go`**:
        *   Define the `FindReferencesTool` that calls the new backend analysis function.
    3.  **Register in `pkg/tools/register.go`**.

### 2.2. Tool: `rename_symbol`

*   **Description**: Safely renames a symbol (variable, function, struct, etc.) across all its usages in the codebase.
*   **Tool Definition**:
    *   **Name**: `rename_symbol`
    *   **Parameters**:
        *   `file_path`, `line`, `column` (string/integer, required): Location of the symbol to rename.
        *   `new_name` (string, required): The new name for the symbol.
*   **Implementation Steps**:
    1.  **Backend Logic in `pkg/analysis`**:
        *   Create a `RenameSymbol(filePath, line, column, newName)` function.
        *   This would first use the logic from `FindSymbolReferences` to locate all usages.
        *   It would then iterate through each file containing a reference and perform a precise, text-based replacement *only* at the locations identified. This is safer than a global find-and-replace.
    2.  **Create `pkg/tools/rename_symbol.go`**:
        *   Define the `RenameSymbolTool` that calls the backend logic.
    3.  **Register in `pkg/tools/register.go`**.

---

## Part 3: Autonomous Sub-Agents

These agents leverage the foundational tools to perform complex, multi-step tasks, following the robust sub-agent pattern already present in `@pkg/core/agents`.

### 3.1. Agent: `TestWriterAgent`

*   **Description**: An agent dedicated to TDD. Given a function, it generates a complete unit test file, including setup, mocks, and assertions.
*   **Agent Definition**:
    *   **Name**: `test_writer_agent`
    *   **InputConfig**: `source_file_path` (string), `symbol_name` (string).
    *   **Internal Tools**: `read_file`, `find_references`, `write_file`, and the new `run_tests`.
*   **Implementation Steps**:
    1.  **Create `pkg/core/agents/test_writer_prompts.md`**:
        *   Write a detailed system prompt guiding the agent's thought process. Example steps: "1. Read the target function. 2. Identify its dependencies. 3. List edge cases and common scenarios to test. 4. Write one test case for a success scenario. 5. Write the test to a file. 6. Run the test and confirm it passes or fails as expected."
    2.  **Create `pkg/core/agents/test_writer.go`**:
        *   Define the `TestWriterAgent` as an `AgentDefinition` struct, loading its prompts from the `.md` file.
        *   Configure its `ToolConfig` to grant access to its required tools.
    3.  **Register in `pkg/core/agents/registry.go`**:
        *   In `loadBuiltInAgents`, add logic to register `TestWriterAgent`, ideally controlled by a setting.

### 3.2. Agent: `RefactorAgent`

*   **Description**: A powerful agent that performs complex, multi-step refactoring while ensuring no functionality is broken by continuously running tests.
*   **Agent Definition**:
    *   **Name**: `refactor_agent`
    *   **InputConfig**: `target_path` (string), `refactoring_goal` (string, e.g., "Simplify the `processOrder` function by extracting helper methods").
    *   **Internal Tools**: `codebase_investigator` (as a first step), `run_tests`, `extract_function`, `rename_symbol`, `find_references`.
*   **Implementation Steps**:
    1.  **Create `pkg/core/agents/refactor_prompts.md`**:
        *   Write a system prompt that emphasizes safety and verification. Example steps: "1. Use `codebase_investigator` to fully understand the target code. 2. Formulate a step-by-step refactoring plan. 3. Before making any change, run all tests to establish a baseline. 4. Perform *one* small, atomic refactoring step (e.g., a single function extraction). 5. After the change, run tests again to ensure no regressions were introduced. 6. If tests fail, revert the change. If they pass, continue to the next step. 7. Only complete the task when the goal is achieved and all tests pass."
    2.  **Create `pkg/core/agents/refactor.go`**:
        *   Define the `RefactorAgent` `AgentDefinition`.
    3.  **Register in `pkg/core/agents/registry.go`**.

By implementing these features, this agent will evolve from a command-executor into a true software engineering assistant capable of complex, safe, and context-aware development tasks.
