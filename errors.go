package taskcluster

import "fmt"

// Error represents a taskcluster API request error.
type Error struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int
	Attempts   int
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
