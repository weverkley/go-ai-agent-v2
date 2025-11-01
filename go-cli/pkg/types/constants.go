package types

const (
	GLOB_TOOL_NAME            = "glob"
	WRITE_TODOS_TOOL_NAME     = "write_todos"
	WRITE_FILE_TOOL_NAME      = "write_file"
	WEB_SEARCH_TOOL_NAME      = "google_web_search"
	WEB_FETCH_TOOL_NAME       = "web_fetch"
	EDIT_TOOL_NAME            = "replace"
	SHELL_TOOL_NAME           = "run_shell_command"
	GREP_TOOL_NAME            = "search_file_content"
	READ_MANY_FILES_TOOL_NAME = "read_many_files"
	READ_FILE_TOOL_NAME       = "read_file"
	LS_TOOL_NAME              = "list_directory"
	MEMORY_TOOL_NAME          = "save_memory"
	TASK_COMPLETE_TOOL_NAME   = "task_complete"
)

const (
	ApprovalModeDefault  ApprovalMode = "default"
	ApprovalModeAutoEdit ApprovalMode = "autoEdit"
	ApprovalModeYOLO     ApprovalMode = "yolo"
)

const (
	StreamEventTypeChunk StreamEventType = "CHUNK"
	StreamEventTypeError StreamEventType = "ERROR"
	StreamEventTypeDone  StreamEventType = "DONE"
)

const (
	ToolErrorTypeToolNotRegistered  ToolErrorType = "TOOL_NOT_REGISTERED"
	ToolErrorTypeInvalidToolParams  ToolErrorType = "INVALID_TOOL_PARAMS"
	ToolErrorTypeUnhandledException ToolErrorType = "UNHANDLED_EXCEPTION"
	ToolErrorTypeExecutionFailed    ToolErrorType = "EXECUTION_FAILED"
)

const (
	ToolConfirmationOutcomeProceedAlways    ToolConfirmationOutcome = "PROCEED_ALWAYS"
	ToolConfirmationOutcomeProceedOnce      ToolConfirmationOutcome = "PROCEED_ONCE"
	ToolConfirmationOutcomeCancel           ToolConfirmationOutcome = "CANCEL"
	ToolConfirmationOutcomeModifyWithEditor ToolConfirmationOutcome = "MODIFY_WITH_EDITOR"
)

const (
	AgentTerminateModeError    AgentTerminateMode = "ERROR"
	AgentTerminateModeGoal     AgentTerminateMode = "GOAL"
	AgentTerminateModeMaxTurns AgentTerminateMode = "MAX_TURNS"
	AgentTerminateModeTimeout  AgentTerminateMode = "TIMEOUT"
	AgentTerminateModeAborted  AgentTerminateMode = "ABORTED"
)