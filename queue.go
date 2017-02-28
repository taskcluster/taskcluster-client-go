package taskcluster

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/taskcluster/go-taskcluster-client/tcqueue"
)

// Queue client interface
type Queue interface {
	Task(taskId string) (*tcqueue.TaskDefinitionResponse, error)
	// TODO: All other methods

	WithContext(context.Context) Queue
	WithAuthorizedScopes(...string) Queue
	BuildURL(string, []string, map[string]string) *url.URL
	BuildSignedURL(string, []string, map[string]string, time.Duration) (*url.URL, error)
}

type MockQueue struct {
	mock.Mock
}

func (m *MockQueue) Task(taskId string) (*tcqueue.TaskDefinitionResponse, error) {
	args := m.Called(taskId)
	return args.Get(0).(*tcqueue.TaskDefinitionResponse), args.Get(1).(error)
}

// TODO: Other methods for MockQueue

type queue struct {
	client
}

func NewQueue(options Options) Queue {
	// TODO: Set sane defaults where needed
	return &queue{}
}

func (q *queue) Task(taskId string) (*tcqueue.TaskDefinitionResponse, error) {
	var result tcqueue.TaskDefinitionResponse
	err := q.call("GET", "/task/"+url.QueryEscape(taskId), nil, nil, &result)
	return &result, err
}

// TODO: All other methods

func (q *queue) WithContext(ctx context.Context) Queue {
	n := &queue{}
	*n = *q
	n.Context = ctx
	return n
}

func (q *queue) WithAuthorizedScopes(scopes ...string) Queue {
	n := &queue{}
	*n = *q
	n.Authorizer = q.Authorizer.WithAuthorizedScopes(scopes...)
	return n
}

// BuildURL returns a url for given method, panics if incorrect arguments.
func (q *queue) BuildURL(method string, params []string, query map[string]string) *url.URL {
	if method == "" { // if method isn't a function name...
		panic("Given method doesn't exist")
	}
	if strings.ToLower(method) == "task" {
		if len(params) != 1 {
			panic("expected 1 parameter!")
		}
		if query != nil {
			panic("Unexpected query...")
		}
		u, _ := url.Parse(q.BaseURL + "/task/" + url.QueryEscape(params[0]))
		return u
	}
	// TODO: Add other methods
	return nil
}

// BuildSignedURL is the same as BuildURL, but it creates a signed URL. Hence,
// this only works for GET requests. Just like BuildURL it panics if given
// incorrect arguments.
func (q *queue) BuildSignedURL(signature string, params []string, query map[string]string, expiration time.Duration) (*url.URL, error) {
	return nil, nil
}
