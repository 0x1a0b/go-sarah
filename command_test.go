package sarah

import (
	"golang.org/x/net/context"
	"regexp"
	"strings"
	"testing"
	"time"
)

type DummyCommand struct {
	IdentifierValue string

	ExecuteFunc func(context.Context, Input) (*CommandResponse, error)

	InputExampleFunc func() string

	MatchFunc func(string) bool
}

func (command *DummyCommand) Identifier() string {
	return command.IdentifierValue
}

func (command *DummyCommand) Execute(ctx context.Context, input Input) (*CommandResponse, error) {
	return command.ExecuteFunc(ctx, input)
}

func (command *DummyCommand) InputExample() string {
	return command.InputExampleFunc()
}

func (command *DummyCommand) Match(str string) bool {
	return command.MatchFunc(str)
}

func TestInsufficientSettings(t *testing.T) {
	matchPattern := regexp.MustCompile(`^\.echo`)

	builder := NewCommandBuilder().
		Identifier("someID").
		MatchPattern(matchPattern).
		InputExample(".echo knock knock")

	if _, err := builder.Build("/path/"); err == nil {
		t.Error("expected error not given.")
	} else {
		if err != ErrCommandInsufficientArgument {
			t.Errorf("expected error not given. %#v", err)
		}
	}

	builder.Func(func(_ context.Context, input Input) (*CommandResponse, error) {
		return &CommandResponse{
			Content: StripMessage(matchPattern, input.Message()),
		}, nil
	})

	if _, err := builder.Build(""); err != nil {
		t.Errorf("something is wrong with command construction. %#v", err)
	}
}

// TODO switch to use DummyCommand on following commit
type abandonedCommand struct{}

func (abandonedCommand *abandonedCommand) Identifier() string {
	return "arbitraryStringThatWouldNeverBeRecognized"
}

func (abandonedCommand *abandonedCommand) Execute(_ context.Context, _ Input) (*CommandResponse, error) {
	return nil, nil
}

func (abandonedCommand *abandonedCommand) InputExample() string {
	return ""
}

func (abandonedCommand *abandonedCommand) Match(_ string) bool {
	return false
}

type echoCommand struct{}

func (echoCommand *echoCommand) Identifier() string {
	return "echo"
}

func (echoCommand *echoCommand) Execute(_ context.Context, input Input) (*CommandResponse, error) {
	return &CommandResponse{Content: regexp.MustCompile(`^\.echo`).ReplaceAllString(input.Message(), "")}, nil
}

func (echoCommand *echoCommand) InputExample() string {
	return ""
}

func (echoCommand *echoCommand) Match(msg string) bool {
	return strings.HasPrefix(msg, "echo")
}

type echoInput struct{}

func (echoInput *echoInput) SenderKey() string {
	return "uniqueValue"
}

func (echoInput *echoInput) Message() string {
	return "echo foo"
}

func (echoInput *echoInput) SentAt() time.Time {
	return time.Now()
}

func (echoInput *echoInput) ReplyTo() OutputDestination {
	return nil
}

func TestCommands_FindFirstMatched(t *testing.T) {
	commands := NewCommands()
	commands.Append(&abandonedCommand{})
	commands.Append(&echoCommand{})
	commands.Append(&abandonedCommand{})

	echo := commands.FindFirstMatched("echo")
	if echo == nil {
		t.Error("expected command is not found")
		return
	}

	switch echo.(type) {
	case *echoCommand:
	// O.K.
	default:
		t.Errorf("expecting echoCommand's pointer, but was %#v.", echo)
		return
	}

	response, err := commands.ExecuteFirstMatched(context.TODO(), &echoInput{})
	if err != nil {
		t.Errorf("unexpected error on commands execution: %#v", err)
		return
	}

	if response == nil {
		t.Error("response expected, but was not returned")
		return
	}

	switch v := response.Content.(type) {
	case string:
	//OK
	default:
		t.Errorf("expected string, but was %#v", v)
	}
}
