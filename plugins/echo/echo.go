package echo

import (
	"github.com/oklahomer/go-sarah"
	"github.com/oklahomer/go-sarah/gitter"
	"github.com/oklahomer/go-sarah/slack"
	"golang.org/x/net/context"
	"regexp"
)

var (
	identifier   = "echo"
	matchPattern = regexp.MustCompile(`^\.echo`)
)

var commandFnc = func(input sarah.Input) string {
	return sarah.StripMessage(matchPattern, input.Message())
}

// SlackCommandFunc is a sarah.CommandFunc especially designed for Slack adapter.
func SlackCommandFunc(_ context.Context, input sarah.Input) (*sarah.CommandResponse, error) {
	return slack.NewStringResponse(commandFnc(input)), nil
}

// GitterCommandFunc is a sarah.CommandFunc especially designed for gitter adapter.
func GitterCommandFunc(_ context.Context, input sarah.Input) (*sarah.CommandResponse, error) {
	return gitter.NewStringResponse(commandFnc(input)), nil
}

// SlackCommand is a pre-built Command for Slack.
var SlackCommand = sarah.NewCommandBuilder().
	Identifier(identifier).
	MatchPattern(matchPattern).
	Func(SlackCommandFunc).
	InputExample(".echo knock knock").
	MustBuild()

// GitterCommand is a pre-built Command for gitter.
var GitterCommand = sarah.NewCommandBuilder().
	Identifier(identifier).
	MatchPattern(matchPattern).
	Func(GitterCommandFunc).
	InputExample(".echo knock knock")
