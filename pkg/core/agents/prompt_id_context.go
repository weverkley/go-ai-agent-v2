package agents

import "context"

type promptIDContextKeyType string

const promptIDContextKey promptIDContextKeyType = "promptID"

// WithPromptID returns a new context with the given promptID.
func WithPromptID(ctx context.Context, promptID string) context.Context {
	return context.WithValue(ctx, promptIDContextKey, promptID)
}

// GetPromptID returns the promptID from the context, or an empty string if not found.
func GetPromptID(ctx context.Context) string {
	if promptID, ok := ctx.Value(promptIDContextKey).(string); ok {
		return promptID
	}
	return ""
}
