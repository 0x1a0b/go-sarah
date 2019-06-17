package gitter

import (
	"context"
	"github.com/oklahomer/go-sarah"
	"github.com/oklahomer/go-sarah/log"
	"github.com/oklahomer/go-sarah/retry"
	"golang.org/x/xerrors"
)

const (
	// GITTER is a dedicated BotType for gitter implementation.
	GITTER sarah.BotType = "gitter"
)

// AdapterOption defines function signature that Adapter's functional option must satisfy.
type AdapterOption func(adapter *Adapter) error

// Adapter stores REST/Streaming API clients' instances to let users interact with gitter.
type Adapter struct {
	config          *Config
	apiClient       APIClient
	streamingClient StreamingClient
}

// NewAdapter creates and returns new Adapter instance.
func NewAdapter(config *Config, options ...AdapterOption) (*Adapter, error) {
	adapter := &Adapter{
		config:          config,
		apiClient:       NewRestAPIClient(config.Token),
		streamingClient: NewStreamingAPIClient(config.Token),
	}

	for _, opt := range options {
		err := opt(adapter)
		if err != nil {
			return nil, xerrors.Errorf("failed to apply options: %w", err)
		}
	}

	return adapter, nil
}

// BotType returns gitter designated BotType.
func (adapter *Adapter) BotType() sarah.BotType {
	return GITTER
}

// Run fetches all belonging Room and connects to them.
func (adapter *Adapter) Run(ctx context.Context, enqueueInput func(sarah.Input) error, notifyErr func(error)) {
	// Get belonging rooms.
	var rooms *Rooms
	err := retry.WithPolicy(adapter.config.RetryPolicy, func() (e error) {
		rooms, e = adapter.apiClient.Rooms(ctx)
		return e
	})
	if err != nil {
		notifyErr(sarah.NewBotNonContinuableError(err.Error()))
		return
	}

	// Connect to each room.
	for _, room := range *rooms {
		go adapter.runEachRoom(ctx, room, enqueueInput)
	}
}

// SendMessage let Bot send message to gitter.
func (adapter *Adapter) SendMessage(ctx context.Context, output sarah.Output) {
	switch content := output.Content().(type) {
	case string:
		room, ok := output.Destination().(*Room)
		if !ok {
			log.Errorf("Destination is not instance of Room. %#v.", output.Destination())
			return
		}
		_, err := adapter.apiClient.PostMessage(ctx, room, content)
		log.Errorf("Failed posting message to %s: %+v", room.ID, err)

	default:
		log.Warnf("Unexpected output %#v", output)

	}
}

func (adapter *Adapter) runEachRoom(ctx context.Context, room *Room, enqueueInput func(sarah.Input) error) {
	for {
		select {
		case <-ctx.Done():
			return

		default:
			log.Infof("Connecting to room: %s", room.ID)

			var conn Connection
			err := retry.WithPolicy(adapter.config.RetryPolicy, func() (e error) {
				conn, e = adapter.streamingClient.Connect(ctx, room)
				return e
			})
			if err != nil {
				log.Warnf("Could not connect to room: %s. Error: %+v", room.ID, err)
				return
			}

			connErr := receiveMessageRecursive(conn, enqueueInput)
			_ = conn.Close()

			// TODO: Intentional connection close such as context.cancel also comes here.
			// It would be nice if we could detect such event to distinguish intentional behaviour and unintentional connection error.
			// But, the truth is, given error is just a privately defined error instance given by http package.
			// var errRequestCanceled = errors.New("net/http: request canceled")
			// For now, let error log appear and proceed to next loop, select case with ctx.Done() will eventually return.
			log.Errorf("Disconnected from room %s: %+v", room.ID, connErr)

		}
	}
}

func receiveMessageRecursive(messageReceiver MessageReceiver, enqueueInput func(sarah.Input) error) error {
	log.Infof("Start receiving message")
	for {
		message, err := messageReceiver.Receive()

		var malformedErr *MalformedPayloadError
		if xerrors.Is(err, ErrEmptyPayload) {
			// https://developer.gitter.im/docs/streaming-api
			// Parsers must be tolerant of occasional extra newline characters placed between messages.
			// These characters are sent as periodic "keep-alive" messages to tell clients and NAT firewalls
			// that the connection is still alive during low message volume periods.
			continue

		} else if xerrors.As(err, &malformedErr) {
			log.Warnf("Skipping malformed input: %+v", err)
			continue

		} else if err != nil {
			// At this point, assume connection is unstable or is closed.
			// Let caller proceed to reconnect or quit.
			return xerrors.Errorf("failed to receive input: %w", err)

		}

		_ = enqueueInput(message)
	}
}

// NewStringResponse creates new sarah.CommandResponse instance with given string.
func NewStringResponse(responseContent string) *sarah.CommandResponse {
	return &sarah.CommandResponse{
		Content:     responseContent,
		UserContext: nil,
	}
}

// NewStringResponseWithNext creates new sarah.CommandResponse instance with given string and next function to continue
func NewStringResponseWithNext(responseContent string, next sarah.ContextualFunc) *sarah.CommandResponse {
	return &sarah.CommandResponse{
		Content:     responseContent,
		UserContext: sarah.NewUserContext(next),
	}
}

// APIClient is an interface that Rest API client must satisfy.
// This is mainly defined to ease tests.
type APIClient interface {
	Rooms(context.Context) (*Rooms, error)
	PostMessage(context.Context, *Room, string) (*Message, error)
}

// StreamingClient is an interface that HTTP Streaming client must satisfy.
// This is mainly defined to ease tests.
type StreamingClient interface {
	Connect(context.Context, *Room) (Connection, error)
}
