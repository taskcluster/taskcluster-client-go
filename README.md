taskcluster client for golang
=============================

notes:
type Options struct {
    Authorizer
    BaseURL string
    got.Got
    Logger
    Metrics
}

var queue taskcluster.Queue

queue = taskcluster.NewQueue(Options{})
queue = queue.WithContext(context.Background)

mockQueue = &taskcluster.MockQueue{}
mockQueue.On("Ping").Return(nil)

creds := taskcluster.PermanentCredentials{ClientID: "...", AccessToken: "..."}
taskcluster.TemporaryCredentials{taskcluster.PermanentCredentials, Certificate}

taskcluster.CreateTemporaryCredentials(creds, tempClientId, duration, scopes)

type Authorizer interface {
  SignedHeader(method string, URL url.URL, payload json.Message) (string, error)
  SignURL(URL url.URL) (url.URL, error)
  WithAuthorizedScopes(scopes ...string) Authorizer
}

taskcluster.NewAuthorizer(func(ctx context.Context) (taskcluster.TemporaryCredentials, error) {...}) Authorizer
