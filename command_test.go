package sarah

import (
	"strings"
	"testing"
	"time"
)

func TestInsufficientSettings(t *testing.T) {
	builder := NewCommandBuilder()

	if _, err := builder.build("/path/"); err == nil {
		t.Error("expected error not given.")
	} else {
		switch err.(type) {
		case *CommandInsufficientArgumentError:
		// O.K.
		default:
			t.Errorf("expected error not given. %#v", err)
		}
	}

	builder.Identifier("someID")
	if _, err := builder.build("/path/"); err == nil {
		t.Error("expected error not given.")
	} else {
		switch err.(type) {
		case *CommandInsufficientArgumentError:
		// O.K.
		default:
			t.Errorf("expected error not given. %#v", err)
		}
	}

	builder.Constructor(
		func(conf CommandConfig) Command {
			return nil
		},
	)
	if _, err := builder.build("/path/"); err == nil {
		t.Error("expected error not given.")
	} else {
		switch err.(type) {
		case *CommandInsufficientArgumentError:
		// O.K.
		default:
			t.Errorf("expected error not given. %#v", err)
		}
	}

	builder.ConfigStruct(NullConfig)
	if _, err := builder.build("/path/"); err != nil {
		t.Errorf("something is wrong with command construction. %#v", err)
	}
}

type abandonedCommand struct{}

func (_ *abandonedCommand) Identifier() string {
	return "arbitraryStringThatWouldNeverBeRecognized"
}

func (_ *abandonedCommand) Execute(_ BotInput) (*CommandResponse, error) {
	return nil, nil
}

func (_ *abandonedCommand) Example() string {
	return ""
}

func (_ *abandonedCommand) Match(_ string) bool {
	return false
}

func (_ *abandonedCommand) StripCommand(_ string) string {
	return ""
}

type echoCommand struct{}

func (_ *echoCommand) Identifier() string {
	return "echo"
}

func (_ *echoCommand) Execute(input BotInput) (*CommandResponse, error) {
	return &CommandResponse{ResponseContent: input.GetMessage()}, nil
}

func (_ *echoCommand) Example() string {
	return ""
}

func (_ *echoCommand) Match(msg string) bool {
	return strings.HasPrefix(msg, "echo")
}

func (_ *echoCommand) StripCommand(msg string) string {
	return strings.TrimPrefix(msg, "echo")
}

type echoInput struct{}

func (_ *echoInput) GetSenderID() string {
	return ""
}

func (_ *echoInput) GetMessage() string {
	return "echo foo"
}

func (_ *echoInput) GetSentAt() time.Time {
	return time.Now()
}

func (_ *echoInput) GetRoomID() string {
	return ""
}

func TestCommands_FindFirstMatched(t *testing.T) {
	commands := NewCommands()
	commands.Append(&abandonedCommand{})
	commands.Append(&echoCommand{})
	commands.Append(&abandonedCommand{})

	if echo := commands.FindFirstMatched("echo"); echo == nil {
		t.Errorf("expected command is not found")
		return
	} else {
		switch echo.(type) {
		case *echoCommand:
		// O.K.
		default:
			t.Errorf("expecting echoCommand's pointer, but was %#v.", echo)
			return
		}
	}

	response, err := commands.ExecuteFirstMatched(&echoInput{})
	if err != nil {
		t.Errorf("unexpected error on commands execution: %#v", err)
		return
	}

	if response == nil {
		t.Error("response expected, but was not returned")
		return
	}

	switch v := response.ResponseContent.(type) {
	case string:
		//OK
	default:
		t.Errorf("expected string, but was %#v", v)
	}
}
