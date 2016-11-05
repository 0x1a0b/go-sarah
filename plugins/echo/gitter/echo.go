package gitter

import (
	"github.com/oklahomer/go-sarah"
	"github.com/oklahomer/go-sarah/gitter"
	"golang.org/x/net/context"
	"regexp"
)

var (
	identifier   = "echo"
	matchPattern = regexp.MustCompile(`^\.echo`)
)

func echo(_ context.Context, input sarah.Input, _ sarah.CommandConfig) (*sarah.PluginResponse, error) {
	return gitter.NewStringResponse(sarah.StripMessage(matchPattern, input.Message())), nil
}

func init() {
	builder := sarah.NewCommandBuilder().
		Identifier(identifier).
		ConfigStruct(sarah.NullConfig).
		MatchPattern(matchPattern).
		Func(echo).
		InputExample(".echo knock knock")
	sarah.AppendCommandBuilder(gitter.GITTER, builder)
}
