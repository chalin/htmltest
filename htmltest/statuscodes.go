package htmltest

/* Generalization of HTTP status codes, to account for tool-specific errors.

- **Positive (> 0)**: official HTTP status codes
- **Zero (= 0)**: Unchecked/undiscovered (neutral state)
- **Negative (< 0)**: Tool-specific error states, such as timeout, network
  error, DNS failure, etc.

For details, see @docs/tasks/migrate-status-codes.md
*/

const (
	StatusUnchecked = 0
	StatusTimeout   = -10
	// Future: Additional tool-specific error codes (not yet implemented)
	// StatusNetworkError = -20  // DNS failures, connection refused, etc.
	// StatusCertError    = -30  // Certificate validation errors
	// StatusClientError  = -40  // Generic HTTP client errors
)

func IsHTTPStatus(code int) bool {
	return code > 0
}

// IsUnchecked returns true if the link was discovered but not checked
func IsUnchecked(code int) bool {
	return code == 0
}

// IsToolError returns true if the status is a tool-specific error
func IsToolError(code int) bool {
	return code < 0
}
