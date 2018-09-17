package main

import (
	"context"
	"flag"
	"os"
	"strings"

	"github.com/justanotherorganization/graceless"
	"github.com/justanotherorganization/graceless/config"
	"github.com/justanotherorganization/justanotherbotkit/transport"
	"github.com/justanotherorganization/justanotherbotkit/transport/slack"
	"github.com/justanotherorganization/justanotherbotkit/users"
	"github.com/justanotherorganization/justanotherbotkit/users/bolt"
	"github.com/sirupsen/logrus"

	// Import the commands we want to be registered.
	_ "github.com/justanotherorganization/graceless/internal/commands/sed"
)

var (
	slackToken string
	dbPath     string
	rootUsers  []string

	log *logrus.Entry
)

func init() {
	log = logrus.NewEntry(logrus.New())

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
		log.Fatal("Slack token must be set")
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
			log.Error(err)
			return
		}
	}

	slack, err := slack.New(&transport.Config{
		Token:       slackToken,
		IgnoreUsers: []string{"keeper"},
	})
	if err != nil {
		log.Error(err)
		return
	}

	g, err := graceless.New(&config.Config{
		RootUsers: rootUsers,
		CmdPrefix: config.DefaultCommandPrefix,
		Log:       log,
		Transport: slack,
		UserDB:    db,
	})
	if err != nil {
		log.Error(err)
		return
	}

	errCh := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for stop := false; !stop; {
			select {
			case <-ctx.Done():
				stop = true
			case err := <-errCh:
				if err != nil {
					// For now treat all errors as non-fatal.
					log.Error(err)
				}
			}
		}
	}()

	g.Start(ctx, cancel, errCh)
}
