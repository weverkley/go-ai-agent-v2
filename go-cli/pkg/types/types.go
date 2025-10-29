package types

// ApprovalMode defines the approval mode for tool calls.
type ApprovalMode string

const (
	ApprovalModeDefault  ApprovalMode = "default"
	ApprovalModeAutoEdit ApprovalMode = "autoEdit"
	ApprovalModeYOLO     ApprovalMode = "yolo"
)
