package sarah

import (
	"context"
	"fmt"
	"golang.org/x/xerrors"
)

// ErrWatcherNotRunning is returned when ConfigWatcher.Unwatch() is called but the context is already canceled.
var ErrWatcherNotRunning = xerrors.New("context is already canceled")

// ErrAlreadySubscribing is returned when duplicated calls to ConfigWatcher.Watch() occur.
var ErrAlreadySubscribing = xerrors.New("already subscribing")

// ConfigNotFoundError is returned when corresponding configuration is not found.
// This is typically returned when the caller tries to see if there is any configuration available via ConfigWatcher.Read().
type ConfigNotFoundError struct {
	BotType BotType
	ID      string
}

// Error returns stringified representation of the error.
func (err *ConfigNotFoundError) Error() string {
	return fmt.Sprintf("no configuration found for %s:%s", err.BotType, err.ID)
}

var _ error = (*ConfigNotFoundError)(nil)

// ConfigWatcher provides an interface for such component that subscribes to any changes for Command and ScheduledTask.
// One example could be watchers.fileWatcher that subscribes to directory changes.
type ConfigWatcher interface {
	// Read reads the latest configuration value and apply that value to configPtr.
	Read(botCtx context.Context, botType BotType, id string, configPtr interface{}) error
	// Watch subscribes to given id's configuration.
	// When a change to the corresponding configuration value occurs, callback is called.
	// A call to callback function triggers go-sarah's core to call Read() to reflect the latest configuration value.
	Watch(botCtx context.Context, botType BotType, id string, callback func()) error
	// Unwatch is called when Bot is stopped and subscription is no longer required.
	Unwatch(botType BotType) error
}

type nullConfigWatcher struct{}

var _ ConfigWatcher = (*nullConfigWatcher)(nil)

func (*nullConfigWatcher) Read(_ context.Context, _ BotType, _ string, _ interface{}) error {
	return nil
}

func (*nullConfigWatcher) Watch(_ context.Context, _ BotType, _ string, _ func()) error {
	return nil
}

func (*nullConfigWatcher) Unwatch(_ BotType) error {
	return nil
}
