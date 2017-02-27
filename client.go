package taskcluster

import (
	"context"
	"net/http"
	"net/url"

	got "github.com/taskcluster/go-got"
)

// Logger is a simple logging interface. This is already satisfied by the
// builtin log package as well as logrus.
type Logger interface {
	Println(v ...interface{})
}

// Options for creating an API client.
type Options struct {
	Authorizer
	BaseURL string
	*http.Client
	Retries int
	*got.BackOff
	Logger
	MetricsCollector
}

type client struct {
	Authorizer
	BaseURL string
	*got.Got
	Logger
	MetricsCollector
	context.Context
}

// call is an auxiliary method for doing an API call
func (c *client) call(method, route string, query url.Values, payload, result interface{}) error {
	return nil
}
