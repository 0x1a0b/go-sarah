package sarah

import (
	"fmt"
	"path"
)

// commandFunc is a function type that represents command function
type taskFunc func(ScheduledTaskConfig) (*Message, error)

type ScheduledTaskConfig interface {
	Schedule() string
}

type Task interface {
	Identifier() string

	Execute() (*Message, error)
}

type scheduledTask struct {
	identifier string

	taskFunc taskFunc

	config ScheduledTaskConfig
}

func (task *scheduledTask) Identifier() string {
	return task.identifier
}

func (task *scheduledTask) Execute() (*Message, error) {
	return task.taskFunc(task.config)
}

type scheduledTaskBuilder struct {
	identifier string
	taskFunc   taskFunc
	config     ScheduledTaskConfig
}

func NewScheduledTaskBuilder() *scheduledTaskBuilder {
	return &scheduledTaskBuilder{}
}

func (builder *scheduledTaskBuilder) Identifier(id string) *scheduledTaskBuilder {
	builder.identifier = id
	return builder
}

func (builder *scheduledTaskBuilder) Func(function taskFunc) *scheduledTaskBuilder {
	builder.taskFunc = function
	return builder
}

func (builder *scheduledTaskBuilder) ConfigStruct(config ScheduledTaskConfig) *scheduledTaskBuilder {
	builder.config = config
	return builder
}

func (builder *scheduledTaskBuilder) build(configDir string) (*scheduledTask, error) {
	if builder.identifier == "" {
		return nil, NewTaskInsufficientArgumentError("Identifier")
	}
	if builder.taskFunc == nil {
		return nil, NewTaskInsufficientArgumentError("Func")
	}
	if builder.config == nil {
		return nil, NewTaskInsufficientArgumentError("ConfigStruct")
	}

	taskConfig := builder.config
	fileName := builder.identifier + ".yaml"
	configPath := path.Join(configDir, fileName)
	err := readConfig(configPath, taskConfig)
	if err != nil {
		return nil, err
	}

	return &scheduledTask{
		identifier: builder.identifier,
		taskFunc:   builder.taskFunc,
		config:     builder.config,
	}, nil
}

type TaskInsufficientArgumentError struct {
	Err string
}

func (e *TaskInsufficientArgumentError) Error() string {
	return e.Err
}

func NewTaskInsufficientArgumentError(argName string) *TaskInsufficientArgumentError {
	return &TaskInsufficientArgumentError{Err: fmt.Sprintf("% must be set.", argName)}
}