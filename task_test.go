package sarah

import (
	"golang.org/x/net/context"
	"testing"
)

type DummyScheduledTaskConfig struct {
	ScheduleValue    string `yaml:"schedule"`
	DestinationValue OutputDestination
}

func (config *DummyScheduledTaskConfig) Schedule() string {
	return config.ScheduleValue
}

func (config *DummyScheduledTaskConfig) DefaultDestination() OutputDestination {
	return config.DestinationValue
}

func TestNewScheduledTaskBuilder(t *testing.T) {
	builder := NewScheduledTaskBuilder()

	if builder == nil {
		t.Fatal("Returned builder instance is nil.")
	}
}

func TestScheduledTaskBuilder_Identifier(t *testing.T) {
	id := "overWhelmedWithTasks"
	builder := &ScheduledTaskBuilder{}
	builder.Identifier(id)

	if builder.identifier != id {
		t.Fatal("Supplied id is not set.")
	}
}

func TestScheduledTaskBuilder_Func(t *testing.T) {
	res := "dummyResponse"
	taskFunc := func(_ context.Context) ([]*ScheduledTaskResult, error) {
		return []*ScheduledTaskResult{{Content: res}}, nil
	}
	builder := &ScheduledTaskBuilder{}
	builder.Func(taskFunc)

	actualRes, err := builder.taskFunc(context.TODO())
	if err != nil {
		t.Fatalf("Unexpected error returned: %s.", err.Error())
	}

	if actualRes[0].Content != res {
		t.Fatal("Supplied func is not set.")
	}
}

func TestScheduledTaskBuilder_ConfigurableFunc(t *testing.T) {
	config := &DummyScheduledTaskConfig{}
	taskFunc := func(_ context.Context, _ TaskConfig) ([]*ScheduledTaskResult, error) {
		return nil, nil
	}
	builder := &ScheduledTaskBuilder{}
	builder.ConfigurableFunc(config, taskFunc)

	if builder.config != config {
		t.Fatal("Supplied config is not set.")
	}
	if builder.taskFunc == nil {
		t.Fatal("Supplied function is not set.")
	}
}

func TestScheduledTaskBuilder_Build(t *testing.T) {
	// Schedule is manually set at first.
	dummySchedule := "dummy"
	config := &DummyScheduledTaskConfig{ScheduleValue: dummySchedule}
	taskFunc := func(_ context.Context, _ TaskConfig) ([]*ScheduledTaskResult, error) {
		return nil, nil
	}

	builder := &ScheduledTaskBuilder{}
	_, err := builder.Build("dummy")
	if err != ErrTaskInsufficientArgument {
		t.Fatalf("Expected error is not returned: %#v.", err)
	}

	id := "scheduled"
	builder.Identifier(id)
	builder.ConfigurableFunc(config, taskFunc)

	// When corresponding configuration file is not found, then manually set schedule must stay.
	_, err = builder.Build("no/corresponding/config")
	if err != nil {
		t.Fatal("Error on task construction with no config file.")
	}
	if config.Schedule() != dummySchedule {
		t.Errorf("Config value changed: %s.", config.DefaultDestination())
	}

	task, err := builder.Build("testdata/taskbuilder")

	if err != nil {
		t.Fatalf("Error returned on task build: %#v.", err)
	}

	if task.Identifier() != id {
		t.Errorf("Supplied id is not returned: %s", task.Identifier())
	}

	if builder.taskFunc == nil {
		t.Fatal("Supplied func is not set.")
	}

	if config.Schedule() == dummySchedule {
		t.Errorf("Config value is not overridden: %s.", config.DefaultDestination())
	}
}

func TestScheduledTask_Identifier(t *testing.T) {
	id := "myTask"
	task := &scheduledTask{identifier: id}

	if task.Identifier() != id {
		t.Fatalf("ID is not properly returned: %s.", task.Identifier())
	}
}

func TestScheduledTask_Execute(t *testing.T) {
	returningContent := "abc"
	taskFunc := func(_ context.Context, _ ...TaskConfig) ([]*ScheduledTaskResult, error) {
		return []*ScheduledTaskResult{
			{Content: returningContent},
		}, nil
	}
	task := &scheduledTask{taskFunc: taskFunc}
	results, err := task.Execute(context.TODO())

	if err != nil {
		t.Fatalf("Unexpected error is returned: %#v.", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result to return, but was: %d.", len(results))

	}
	if results[0].Content != returningContent {
		t.Errorf("Unexpected content is returned: %s", results[0].Content)
	}
}
