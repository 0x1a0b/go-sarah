package sarah

import (
	"errors"
	"fmt"
	"github.com/oklahomer/go-sarah/log"
	"golang.org/x/net/context"
	"os"
	"path"
)

var (
	// ErrTaskInsufficientArgument is returned when required parameters are not set.
	ErrTaskInsufficientArgument = errors.New("Identifier and Func must be set.")

	// ErrTaskScheduleNotGiven is returned when schedule is provided by neither ScheduledTaskBuilder's parameter nor config.
	ErrTaskScheduleNotGiven = errors.New("Task schedule is not set or given from config struct.")
)

// ScheduledTaskResult is a struct that ScheduledTask returns on its execution.
type ScheduledTaskResult struct {
	Content     interface{}
	Destination OutputDestination
}

// taskFunc is a function type that represents scheduled task.
type taskFunc func(context.Context, ...TaskConfig) ([]*ScheduledTaskResult, error)

// TaskConfig provides an interface that every task configuration must satisfy, which actually means empty.
type TaskConfig interface{}

// ScheduledConfig defines an interface that config with schedule MUST satisfy.
// When no execution schedule is set with ScheduledTaskBuilder.Schedule, this value is taken as default on ScheduledTaskBuilder.Build.
type ScheduledConfig interface {
	Schedule() string
}

// DestinatedConfig defines an interface that config with default destination MUST satisfy.
// When no default output destination is set with ScheduledTaskBuilder.DefaultDestination, this value is taken as default on ScheduledTaskBuilder.Build.
type DestinatedConfig interface {
	DefaultDestination() OutputDestination
}

// ScheduledTask defines interface that all scheduled task MUST satisfy.
// As long as a struct satisfies this interface, the struct can be registered as ScheduledTask via Runner.RegisterScheduledTask.
type ScheduledTask interface {
	Identifier() string
	Execute(context.Context) ([]*ScheduledTaskResult, error)
	DefaultDestination() OutputDestination
	Schedule() string
}

type scheduledTask struct {
	identifier         string
	taskFunc           taskFunc
	schedule           string
	defaultDestination OutputDestination
	config             TaskConfig
}

// Identifier returns unique ID of this task.
func (task *scheduledTask) Identifier() string {
	return task.identifier
}

// Execute executes scheduled task and returns *slice* of results.
//
// Note that scheduled task may result in sending messages to multiple destinations;
// sending taking-out-trash alarm to #dady-chores room while sending go-to-school alarm to #daughter room.
//
// When sending messages to multiple destinations, create and return results as many as destinations.
// If output destination is nil, caller tries to find corresponding destination from config struct.
func (task *scheduledTask) Execute(ctx context.Context) ([]*ScheduledTaskResult, error) {
	return task.taskFunc(ctx, task.config)
}

// Schedule returns execution schedule.
func (task *scheduledTask) Schedule() string {
	return task.schedule
}

// DefaultDestination returns the default output destination of this task.
// OutputDestination returned by task execution has higher priority.
func (task *scheduledTask) DefaultDestination() OutputDestination {
	return task.defaultDestination
}

// ScheduledTaskBuilder is a helper that creates scheduledTask with given parameter.
type ScheduledTaskBuilder struct {
	identifier         string
	taskFunc           taskFunc
	schedule           string
	defaultDestination OutputDestination
	config             TaskConfig
}

// NewScheduledTaskBuilder creates and returns ScheduledTaskBuilder instance.
func NewScheduledTaskBuilder() *ScheduledTaskBuilder {
	return &ScheduledTaskBuilder{}
}

// Identifier sets unique ID of this task.
// This is used to identify re-configure tasks and replace old ones.
func (builder *ScheduledTaskBuilder) Identifier(id string) *ScheduledTaskBuilder {
	builder.identifier = id
	return builder
}

// Func sets function to be called on task execution.
// To set function that requires some sort of configuration struct, use ConfigurableFunc.
func (builder *ScheduledTaskBuilder) Func(fn func(context.Context) ([]*ScheduledTaskResult, error)) *ScheduledTaskBuilder {
	builder.config = nil
	builder.taskFunc = func(ctx context.Context, cfg ...TaskConfig) ([]*ScheduledTaskResult, error) {
		return fn(ctx)
	}
	return builder
}

// Schedule sets execution schedule.
// Representation spec. is identical to that of github.com/robfig/cron.
func (builder *ScheduledTaskBuilder) Schedule(schedule string) *ScheduledTaskBuilder {
	builder.schedule = schedule
	return builder
}

// DefaultDestination sets default output destination of this task.
// OutputDestination returned by task execution has higher priority.
func (builder *ScheduledTaskBuilder) DefaultDestination(dest OutputDestination) *ScheduledTaskBuilder {
	builder.defaultDestination = dest
	return builder
}

// ConfigurableFunc sets function for ScheduledTask with configuration struct.
// Passed configuration struct is passed to function as a third argument.
//
// When this builder is stashed via StashScheduledTaskBuilder and Runner runs with Config.PluginConfigRoot,
// configuration struct gets updated automatically when corresponding configuration file is modified.
func (builder *ScheduledTaskBuilder) ConfigurableFunc(config TaskConfig, fn func(context.Context, TaskConfig) ([]*ScheduledTaskResult, error)) *ScheduledTaskBuilder {
	builder.config = config
	builder.taskFunc = func(ctx context.Context, cfg ...TaskConfig) ([]*ScheduledTaskResult, error) {
		return fn(ctx, cfg[0])
	}
	return builder
}

// Build creates instance of scheduledTask with given parameters.
func (builder *ScheduledTaskBuilder) Build(configDir string) (ScheduledTask, error) {
	if builder.identifier == "" || builder.taskFunc == nil {
		return nil, ErrTaskInsufficientArgument
	}

	// If path to the configuration files' directory is given, corresponding configuration file MAY exist.
	// If exists, read and map to given config struct; if file does not exist, assume the config struct is already configured by developer.
	taskConfig := builder.config
	if configDir != "" && taskConfig != nil {
		fileName := builder.identifier + ".yaml"
		configPath := path.Join(configDir, fileName)
		err := readConfig(configPath, taskConfig)
		if err != nil && os.IsNotExist(err) {
			log.Infof("config struct is set, but there was no corresponding setting file at %s. "+
				"assume config struct is already filled with appropriate value and keep going. command ID: %s.",
				configPath, builder.identifier)
		} else if err != nil {
			// File was there, but could not read.
			return nil, err
		}
	}

	// Setup execution schedule
	schedule := builder.schedule
	if taskConfig != nil {
		if scheduledConfig, ok := (taskConfig).(ScheduledConfig); ok {
			if s := scheduledConfig.Schedule(); s != "" {
				schedule = s
			}
		}
	}
	if schedule == "" {
		return nil, ErrTaskScheduleNotGiven
	}

	// Setup default destination
	// This can be nil since each task execution may return corresponding destination
	dest := builder.defaultDestination
	if taskConfig != nil {
		if destConfig, ok := (taskConfig).(DestinatedConfig); ok {
			if d := destConfig.DefaultDestination(); d != nil {
				dest = d
			}
		}
	}

	return &scheduledTask{
		identifier:         builder.identifier,
		taskFunc:           builder.taskFunc,
		schedule:           schedule,
		defaultDestination: dest,
		config:             builder.config,
	}, nil
}

// MustBuild is like Build but panics if any error occurs on Build.
// It simplifies safe initialization of global variables holding built Command instances.
func (builder *ScheduledTaskBuilder) MustBuild() ScheduledTask {
	task, err := builder.Build("")
	if err != nil {
		panic(fmt.Sprintf("Error on building task: %s", err.Error()))
	}

	return task
}
