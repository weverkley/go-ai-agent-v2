read my @go-cli/PLAN.md and continue my execution plan, after implementing each command you must test 
the command, also when creating or updating a new go file, read the source javasctipt file to address 
the logic, and make sure the same logic is applied, it must have the same responses and behaviour.

When implementing a feature and its not fully yet implemented, or depends on another feature that is not done yet, you must create a placeholder for the missing feature, highlight it as a TODO and give it a comment to dont block my linter like "//nolint:staticcheck" and when the missing feature is done, you must replace the placeholder with the missing feature.

i want to replicate the same GUI for the cli which the javascript version already have, for terminal gui tasks i want to use https://github.com/charmbracelet/bubbletea as the main cli gui.

The javascript files reside in the folder docs/gemini-cli-main/packages folder The documentation about each the javascript verion feature redise in the folder docs/gemini-cli-main/docs.

Before continuing, i want you to read my current progress and the files i have implemented already on the folder go-cli, start reading the folders and then proceed to read files.

When finishing migrating a full function from javascript into go, you must run the linter and then build the application.

When the build passes and the linter has no error, you must test the implemented feature.

At first before you start crafting, you need to understand the JavaScript's intent and translate it idiomatically into Go, leveraging the existing Go project structure. To do this, you will need to get a comprehensive understanding of my entire Go project by recursively listing the go-cli directory and then reading relevant files to grasp current implementations.
