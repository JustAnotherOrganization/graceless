package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"

	"github.com/justanotherorganization/graceless"
	"github.com/justanotherorganization/graceless/config"
	"github.com/justanotherorganization/justanotherbotkit/transport"
	"github.com/justanotherorganization/justanotherbotkit/transport/slack"
	"github.com/justanotherorganization/justanotherbotkit/users"
	"github.com/justanotherorganization/justanotherbotkit/users/bolt"
	"github.com/justanotherorganization/l5424"
	"github.com/justanotherorganization/l5424/x5424"

	// Import the commands we want to be registered.
	_ "github.com/justanotherorganization/graceless/commands/sed"
)

var (
	slackToken string
	dbPath     string
	rootUsers  []string
	logger     *x5424.Logger
)

func init() {
	logger = x5424.New(l5424.InfoLvl.String(), nil)

	flag.StringVar(&slackToken, "st", "", "Slack token")
	flag.StringVar(&dbPath, "db", "", "User database path")
	var _rootUsers string
	flag.StringVar(&_rootUsers, "ru", "", "Root users")
	flag.Parse()

	if slackToken == "" {
		slackToken = os.Getenv("SLACKTOKEN")
	}

	if dbPath == "" {
		dbPath = os.Getenv("USERDBPATH")
	}

	if _rootUsers == "" {
		_rootUsers = os.Getenv("ROOTUSERS")
	}

	if slackToken == "" {
		logger.Log(x5424.Severity, l5424.EmergencyLvl, "Slack token must be set\n")
		os.Exit(1)
	}

	if _rootUsers != "" {
		rootUsers = strings.Split(_rootUsers, ",")
	}
}

func main() {
	var (
		db  users.DB
		err error
	)

	if dbPath != "" {
		db, err = bolt.New(&bolt.Config{
			File: dbPath,
		})
		if err != nil {
			logger.Log(x5424.Severity, l5424.EmergencyLvl, err)
			return
		}
	}

	slack, err := slack.New(&transport.Config{
		Token:       slackToken,
		IgnoreUsers: []string{"keeper"},
	})
	if err != nil {
		logger.Log(x5424.Severity, l5424.EmergencyLvl, err)
		return
	}

	g, err := graceless.New(&config.Config{
		RootUsers: rootUsers,
		CmdPrefix: config.DefaultCommandPrefix,
		Logger:    logger,
		Transport: slack,
		UserDB:    db,
	})
	if err != nil {
		logger.Log(x5424.Severity, l5424.EmergencyLvl, err)
		return
	}

	errCh := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())

	// Start signal handler
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-signals:
				cancel()
				return
			case err := <-errCh:
				if err != nil {
					// For now treat all errors as non-fatal.
					logger.Log(x5424.Severity, l5424.ErrorLvl, err, "\n")
				}
			}
		}
	}()

	g.Start(ctx, cancel, errCh)
}
