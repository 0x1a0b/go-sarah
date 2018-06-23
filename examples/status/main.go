/*
Package main provides an example that uses Runner.Status()
to return current sarah.Runner and its belonging Bot status via HTTP server.

In this example two bots, slack and nullBot, are registered to sarah.Runner and become subject to supervise.
See handler.go for Runner.Status() usage.
*/
package main

import (
	"flag"
	"github.com/oklahomer/go-sarah"
	"github.com/oklahomer/go-sarah/log"
	"github.com/oklahomer/go-sarah/slack"
	"golang.org/x/net/context"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	// Parse and check command line flags
	var path = flag.String("config", "", "path to application configuration file.")
	flag.Parse()
	if *path == "" {
		panic("./path/to/executable -config=/path/to/config/app.yml")
	}

	// Initialize config with given file content
	cfg, err := readConfig(*path)
	if err != nil {
		panic(err)
	}

	// A handy struct that stores all sarah.RunnerOption to be passed to sarah.Runner
	runnerOptions := sarah.NewRunnerOptions()

	// Setup a bot
	nullBot := &nullBot{}
	runnerOptions.Append(sarah.WithBot(nullBot))

	// Setup another bot
	slackBot, err := setupSlackBot(cfg)
	if err != nil {
		panic(err)
	}
	runnerOptions.Append(sarah.WithBot(slackBot))

	// Setup a Runner to run and supervise above bots
	runner, err := sarah.NewRunner(cfg.Runner, runnerOptions.Arg())
	if err != nil {
		panic(err)
	}

	// Run sarah.Runner and a HTTP server that returns sarah.Runner's status.
	// See handler.go for detail and example of HTTP server.
	run(runner)
}

func setupSlackBot(cfg *config) (sarah.Bot, error) {
	storage := sarah.NewUserContextStorage(cfg.ContextCache)
	slackAdapter, err := slack.NewAdapter(cfg.Slack)
	if err != nil {
		return nil, err
	}
	slackBot, err := sarah.NewBot(slackAdapter, sarah.BotWithStorage(storage))
	if err != nil {
		return nil, err
	}
	return slackBot, nil
}

func run(runner sarah.Runner) {
	log.Infof("start pid %d\n", os.Getpid())

	ctx, cancel := context.WithCancel(context.Background())

	go runner.Run(ctx)

	mux := http.NewServeMux()
	setStatusHandler(mux, runner)
	server := newServer(mux)
	go server.Run(ctx)

	// Wait til signal reception
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	log.Info("Stopping due to signal reception.")
	cancel()
	time.Sleep(1 * time.Second) // Wait a bit til things finish
}
