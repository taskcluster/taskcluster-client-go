package tcqueue

import "time"

type TaskDefinitionResponse struct {
	Created      time.Time              `json:"created"`
	Deadline     time.Time              `json:"deadline"`
	Dependencies []string               `json:"dependencies"`
	Expires      time.Time              `json:"expires,omitempty"`
	Extra        map[string]interface{} `json:"extra"`
	Metadata     struct {
		Description string `json:"description"`
		Name        string `json:"name"`
		Owner       string `json:"owner"`
		Source      string `json:"source"`
	} `json:"metadata"`
	Payload       map[string]interface{} `json:"payload"`
	Priority      string                 `json:"priority"`
	ProvisionerID string                 `json:"provisionerId"`
	Requires      string                 `json:"requires"`
	Retries       int                    `json:"retries"`
	Routes        []string               `json:"routes"`
	SchedulerID   string                 `json:"schedulerId"`
	Scopes        []string               `json:"scopes"`
	Tags          map[string]string      `json:"tags"`
	TaskGroupID   string                 `json:"taskGroupId"`
	WorkerType    string                 `json:"workerType"`
}
