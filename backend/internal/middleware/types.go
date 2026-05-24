package middleware

// contextKey is used for storing values in request context
type contextKey string

const (
	// UserIDKey is the context key for storing authenticated user ID
	UserIDKey contextKey = "user_id"
	// UsernameKey is the context key for storing authenticated username
	UsernameKey contextKey = "username"
)

// usernameKey is an unexported key for storing username in context
// It avoids collision with the exported UsernameKey constant
var usernameKey contextKey = "username"
