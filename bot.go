package sarah

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/oklahomer/go-sarah/worker"
	"time"
)

var (
	stashedCommandBuilder = map[BotType][]*commandBuilder{}
)

// BotType indicates what bot implementation a particular BotAdapter/Plugin is corresponding to.
type BotType string

// String returns a stringified form of BotType
func (botType BotType) String() string {
	return string(botType)
}

/*
BotAdapter defines interface that each Bot implementation has to satisfy.
Its instance can be fed to Bot to start bot interaction.
*/
type BotAdapter interface {
	GetBotType() BotType
	Run(chan<- BotInput)
	SendResponse(*CommandResponse)
	Stop()
}

/*
botProperty stashes some properties for each bot implementation.

Since each bot implementation, BotAdapter, is not responsible for implementing and storing its commands -- this is managed by Bot --
Bot needs to internally store each BotAdapter, corresponding Commands, and miscellaneous properties/attributes at one place.
This is to increase Bot's implementation handiness, so this struct is never meant to be exposed.
*/
type botProperty struct {
	adapter         BotAdapter
	commands        *Commands
	pluginConfigDir string
}

/*
newBotProperty creates and return new botProperty to store each bot implementation.
*/
func newBotProperty(adapter BotAdapter, configDir string) *botProperty {
	return &botProperty{
		adapter:         adapter,
		commands:        NewCommands(),
		pluginConfigDir: configDir,
	}
}

/*
Bot is the core of sarah.
Developers can register desired BotAdapter and Commands to create own bot.
*/
type BotRunner struct {
	botProperties []*botProperty
	workerPool    *worker.Pool
	stopAll       chan struct{}
}

// NewBotRunner creates and return new Bot instance.
func NewBotRunner() *BotRunner {
	return &BotRunner{
		botProperties: []*botProperty{},
		workerPool:    worker.NewPool(10),
		stopAll:       make(chan struct{}),
	}
}

/*
AddAdapter allows developer to register desired BotAdapter implementation.
Bot and each adapter mainly communicate via designated channels to pass incoming message and outgoing response.
*/
func (runner *BotRunner) AddAdapter(adapter BotAdapter, pluginConfigDir string) {
	for _, botProperty := range runner.botProperties {
		if botProperty.adapter.GetBotType() == adapter.GetBotType() {
			panic(fmt.Sprintf("BotType (%s) conflicted with stored BotAdapter.", adapter.GetBotType()))
		}
	}

	// New adapter. Append to stored ones.
	runner.botProperties = append(runner.botProperties, newBotProperty(adapter, pluginConfigDir))
}

/*
Run starts Bot interaction.

At this point bot starts its internal workers, runs each BotAdapter, and starts listening to incoming messages.
*/
func (runner *BotRunner) Run() {
	go runner.runWorkers()
	for _, botProperty := range runner.botProperties {
		// build commands with stashed builder settings
		builders, ok := stashedCommandBuilder[botProperty.adapter.GetBotType()]
		if !ok {
			// No command builder is stashed for this bot type.
			continue
		}
		commands := buildCommands(builders, botProperty.pluginConfigDir)
		for _, command := range commands {
			botProperty.commands.Append(command)
		}

		// Prepare a channel to pass around receiving messages, and run with it.
		receiver := make(chan BotInput)
		botProperty.adapter.Run(receiver)
		go runner.respondMessage(botProperty, receiver)
	}
}

/*
respondMessage listens to incoming messages via channel.

Each BotAdapter enqueues incoming messages to runner's listening channel, and respondMessage receives them.
When corresponding command is found, command is executed and the result can be passed to BotAdapter's SendResponse method.
*/
func (runner *BotRunner) respondMessage(botProperty *botProperty, receiver <-chan BotInput) {
	for {
		select {
		case <-runner.stopAll:
			return
		case botInput := <-receiver:
			logrus.Debugf("responding to %#v", botInput)
			runner.EnqueueJob(func() {
				res, err := botProperty.commands.ExecuteFirstMatched(botInput)
				if err != nil {
					logrus.Errorf("error on message handling. botInput: %s. error: %#v.", botInput, err.Error())
				}

				if res != nil {
					botProperty.adapter.SendResponse(res)
				}
			})
		}
	}
}

// Stop can be called to stop all bot interaction including each BotAdapter.
func (runner *BotRunner) Stop() {
	close(runner.stopAll)
	for _, botProperty := range runner.botProperties {
		botProperty.adapter.Stop()
	}
}

// runWorkers starts BotRunner's internal workers.
func (runner *BotRunner) runWorkers() {
	runner.workerPool.Run()
	defer runner.workerPool.Stop()

	<-runner.stopAll
}

// EnqueueJob can be used to enqueue task to BotRunner's internal workers.
func (runner *BotRunner) EnqueueJob(job func()) {
	runner.workerPool.EnqueueJob(job)
}

/*
AppendCommandBuilder appends given commandBuilder to internal stash.
Stashed builder is used to configure and build Command instance on BotRunner's initialization.
*/
func AppendCommandBuilder(botType BotType, builder *commandBuilder) {
	logrus.Infof("appending command builder for %s. builder %#v.", botType, builder)
	_, ok := stashedCommandBuilder[botType]
	if !ok {
		stashedCommandBuilder[botType] = make([]*commandBuilder, 0)
	}

	stashedCommandBuilder[botType] = append(stashedCommandBuilder[botType], builder)
}

/*
buildCommands configures and creates Command instances with given stashed CommandBuilders
*/
func buildCommands(builders []*commandBuilder, configDir string) []Command {
	commands := []Command{}
	for _, builder := range builders {
		command, err := builder.build(configDir)
		if err != nil {
			logrus.Errorf(fmt.Sprintf("can't configure plugin: %s. error: %s.", builder.identifier, err.Error()))
			continue
		}
		commands = append(commands, command)
	}

	return commands
}

// BotInput defines interface that each incoming message must satisfy.
type BotInput interface {
	GetSenderID() string

	GetMessage() string

	GetSentAt() time.Time

	GetRoomID() string
}
