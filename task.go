package sarah

import (
	"errors"
	"golang.org/x/net/context"
	"path"
)

var (
	ErrTaskInsufficientArgument = errors.New("Identifier, Func and ConfigStruct must be set.")
)

// commandFunc is a function type that represents command function
type taskFunc func(context.Context, ScheduledTaskConfig) (*CommandResponse, error)

type ScheduledTaskConfig interface {
	Schedule() string

	Destination() OutputDestination
}

type scheduledTask struct {
	identifier string

	taskFunc taskFunc

	config ScheduledTaskConfig
}

func (task *scheduledTask) Identifier() string {
	return task.identifier
}

func (task *scheduledTask) Execute(ctx context.Context) (*CommandResponse, error) {
	return task.taskFunc(ctx, task.config)
}

type ScheduledTaskBuilder struct {
	identifier string
	taskFunc   taskFunc
	config     ScheduledTaskConfig
}

func NewScheduledTaskBuilder() *ScheduledTaskBuilder {
	return &ScheduledTaskBuilder{}
}

func (builder *ScheduledTaskBuilder) Identifier(id string) *ScheduledTaskBuilder {
	builder.identifier = id
	return builder
}

func (builder *ScheduledTaskBuilder) Func(function taskFunc) *ScheduledTaskBuilder {
	builder.taskFunc = function
	return builder
}

func (builder *ScheduledTaskBuilder) ConfigStruct(config ScheduledTaskConfig) *ScheduledTaskBuilder {
	builder.config = config
	return builder
}

func (builder *ScheduledTaskBuilder) build(configDir string) (*scheduledTask, error) {
	if builder.identifier == "" || builder.taskFunc == nil || builder.config == nil {
		return nil, ErrTaskInsufficientArgument
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
