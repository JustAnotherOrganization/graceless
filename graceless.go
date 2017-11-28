package graceless

import (
	"errors"
	"strings"

	"github.com/justanotherorganization/justanotherbotkit/accessors"
	"github.com/sirupsen/logrus"
)

const (
	defaultCommandPrefix = `.`
)

type (
	// Graceless is a clumsy bot.
	Graceless struct {
		accessor accessors.Accessor
		log      *logrus.Entry
		config   *Config
	}

	// Config is the config for a graceless bot.
	Config struct {
		// RootIDs contains the IDs of the users who can shutdown the bot
		// (if no IDs are provided anyone will be able to shutdown the bot).
		RootIDs []string
		// CmdPrefix is the prefix to match on for recognizing commands.
		CmdPrefix string
	}
)

// New creates a new Graceless instance.
func New(accessor accessors.Accessor, log *logrus.Entry, config *Config) (*Graceless, error) {
	if accessor == nil {
		return nil, errors.New("acceessor cannot be nil")
	}

	if log == nil {
		log = logrus.NewEntry(logrus.New())
	}

	if config == nil {
		config = &Config{}
	}

	if config.CmdPrefix == "" {
		config.CmdPrefix = defaultCommandPrefix
	}

	return &Graceless{
		accessor: accessor,
		log:      log,
		config:   config,
	}, nil
}

// Start our Graceless bot.
func (g *Graceless) Start(errCh, stopCh chan error) {
	eventCh := make(chan accessors.MessageEvent)

	go func() {
		g.accessor.TunnelEvents(eventCh, errCh, stopCh)
	}()

	isRoot := func(id string) bool {
		// Not RootIDs are set, everyone can call shutdown.
		if g.config.RootIDs == nil {
			return true
		}

		for _, _id := range g.config.RootIDs {
			if id == _id {
				return true
			}
		}

		return false
	}

out:
	for {
		select {
		case <-stopCh:
			break out
		case msg := <-eventCh:
			g.log.Debugf("%s (%s): %v", msg.Sender.Name, msg.Sender.Id, msg.Body)

			if strings.HasPrefix(msg.Body, g.config.CmdPrefix) {
				cmd := strings.TrimPrefix(msg.Body, g.config.CmdPrefix)
				switch {
				case cmd == "shutdown":
					if isRoot(msg.Sender.Id) {
						g.log.Info("shutdown called")
						stopCh <- errors.New("shutdown")
						break out
					}
				}
			}
		}
	}

	g.log.Info("Shutting down...")
	g.accessor.WG().Wait()
	g.log.Info("Goodbye")
}
