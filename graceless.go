package graceless

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	//"github.com/justanotherorganization/graceless/db"

	"github.com/justanotherorganization/graceless/commands"
	"github.com/justanotherorganization/graceless/commands/engines/golang"
	"github.com/justanotherorganization/graceless/commands/engines/js"
	"github.com/justanotherorganization/graceless/config"
	"github.com/justanotherorganization/justanotherbotkit/accessors"
	"github.com/justanotherorganization/justanotherbotkit/permissions"
	"github.com/sirupsen/logrus"
)

type (
	// Graceless is a clumsy bot.
	Graceless struct {
		accessor accessors.Accessor
		log      *logrus.Entry
		config   *config.Config
		pm       *permissions.Manager
		//_db      *db.DB

		genCmdMu  *sync.RWMutex
		genCmdIdx []commands.CommandIndex
		genCmdMap map[commands.CommandIndex]commands.Command

		addCmdMu  *sync.RWMutex
		addCmdIdx []commands.CommandIndex
		addCmdMap map[commands.CommandIndex]commands.Command

		getCmdMu  *sync.RWMutex
		getCmdIdx []commands.CommandIndex
		getCmdMap map[commands.CommandIndex]commands.Command

		delCmdMu  *sync.RWMutex
		delCmdIdx []commands.CommandIndex
		delCmdMap map[commands.CommandIndex]commands.Command

		engineCmdMu  *sync.RWMutex
		engineCmdIdx []commands.CommandIndex
		engineCmdMap map[commands.CommandIndex]commands.Command
	}
)

// New creates a new Graceless instance.
func New(accessor accessors.Accessor, log *logrus.Entry, conf *config.Config, pm *permissions.Manager) (*Graceless, error) {
	if accessor == nil {
		return nil, errors.New("acceessor cannot be nil")
	}

	if log == nil {
		log = logrus.NewEntry(logrus.New())
	}

	config.SetDefaults(conf)

	if pm == nil {
		conf.Safemode = true
		log.Info("Starting in safemode")
	}

	return &Graceless{
		accessor: accessor,
		log:      log,
		config:   conf,
		pm:       pm,

		addCmdMu:  &sync.RWMutex{},
		addCmdIdx: []commands.CommandIndex{},
		addCmdMap: make(map[commands.CommandIndex]commands.Command),

		getCmdMu:  &sync.RWMutex{},
		getCmdIdx: []commands.CommandIndex{},
		getCmdMap: make(map[commands.CommandIndex]commands.Command),

		delCmdMu:  &sync.RWMutex{},
		delCmdIdx: []commands.CommandIndex{},
		delCmdMap: make(map[commands.CommandIndex]commands.Command),

		genCmdMu:  &sync.RWMutex{},
		genCmdIdx: []commands.CommandIndex{},
		genCmdMap: make(map[commands.CommandIndex]commands.Command),

		engineCmdMu:  &sync.RWMutex{},
		engineCmdIdx: []commands.CommandIndex{},
		engineCmdMap: make(map[commands.CommandIndex]commands.Command),
	}, nil
}

// Start our Graceless bot.
func (g *Graceless) Start(errCh, stopCh chan error) {
	// Listen to stop on it's own goroutine to allow us to stop early if needed.
	var stop bool
	go func() {
		for range stopCh {
			stop = true
			close(stopCh)
		}
	}()

	// Start building our database (as early as possible.
	if !stop && g.pm != nil && !g.config.Safemode {
		go func() {
			// We wouldn't be building the database if we hadn't passed in
			// a backend so treat errors in this instance as fatal.
			users, err := g.accessor.GetUsers()
			if err != nil {
				errCh <- err
				stopCh <- err
				return
			}

			for _, user := range users {
				if user.Id == "USLACKBOT" || user.IsBot || user.Deleted {
					continue
				}

				_user, err := g.pm.GetUser(user.Id)
				if err != nil {
					errCh <- err
					stopCh <- err
					return
				}

				if _user == nil {
					_user, err = g.pm.NewUser(user.Id, user.Name)
					if err != nil {
						errCh <- err
						stopCh <- err
						return
					}
				}
			}
		}()
	}

	// Register some default commands.
	if !stop {
		go func() {
			if err := registerDefaultCommands(g, stopCh); err != nil {
				errCh <- err
			}
		}()
	}

	preCheck := func(idx commands.CommandIndex, user *permissions.User) bool {
		if user.ID == "USLACKBOT" {
			return false
		}

		if idx.Disabled() {
			return false
		}

		_user, err := g.accessor.GetUser(user.ID)
		if err != nil {
			errCh <- err
			return false
		}

		if _user.IsBot {
			return false
		}

		// If we're in safemode don't run commands that have
		// been marked flagged to not be reachable in safemode.
		if g.config.Safemode && idx.NoSafemode() {
			return false
		}

		// If no permission manager is set don't run
		// commands that have been flagged to require
		// a db connection.
		if g.pm == nil && idx.NeedsDB() {
			return false
		}

		// Check the users perms before running any permissions
		// based commands.
		if perms := idx.RequiredPerms(); perms != nil {
			// Root user's don't need individual perms.
			if g.isRoot(user) {
				return true
			}

			l := len(perms)

			_perms, err := user.GetPerms()
			if err != nil {
				errCh <- err
				return false
			}

			for _, perm := range perms {
				for _, _perm := range _perms {
					if _perm == perm {
						l--
					}
				}
			}

			if l > 0 {
				return false
			}
		}

		return true
	}

	eventCh := make(chan accessors.MessageEvent)
	if !stop {
		go func() {
			g.accessor.TunnelEvents(eventCh, errCh, stopCh)
		}()
	}

out:
	for {
		select {
		case <-stopCh:
			break out
		case msg := <-eventCh:
			// Handle messages in their own goroutines so they don't block.
			go func(msg accessors.MessageEvent) {
				g.log.Debugf("%s (%s): %v", msg.Sender.Name, msg.Sender.Id, msg.Body)

				var user *permissions.User
				if g.pm != nil {
					var err error
					user, err = g.pm.GetUser(msg.Sender.Id)
					if err != nil {
						errCh <- err
						return
					}
					if user == nil {
						user, err = g.pm.NewUser(msg.Sender.Id, msg.Sender.Name)
						if err != nil {
							errCh <- err
							return
						}
					}

					// This is a silly way to track the say hello functionality
					// it needs some cleanup.
					ok, err := user.GetPerm("hello")
					if err != nil {
						errCh <- err
						return
					}

					if !ok {
						if err := g.sayHello(&accessors.User{
							Id: user.ID,
						}); err != nil {
							errCh <- err
							return
						}

						if err := user.AddPerms("hello"); err != nil {
							errCh <- err
							return
						}
					}
				} else {
					user = &permissions.User{
						ID: msg.Sender.Id,
					}
				}

				if strings.HasPrefix(msg.Body, g.config.CmdPrefix) {
					cmdStr := strings.TrimPrefix(msg.Body, g.config.CmdPrefix)

					if prefix := "help"; strings.HasPrefix(cmdStr, prefix) {
						cmd := strings.TrimSpace(strings.TrimPrefix(cmdStr, prefix))
						g.addCmdMu.RLock()
						g.getCmdMu.RLock()
						g.delCmdMu.RLock()
						g.genCmdMu.RLock()
						defer func() {
							g.addCmdMu.RUnlock()
							g.getCmdMu.RUnlock()
							g.delCmdMu.RUnlock()
							g.genCmdMu.RUnlock()
						}()

						if cmd == "" {
							var fields []string
							for _, idx := range g.addCmdIdx {
								if !preCheck(idx, user) {
									continue
								}

								if idx.Name() != "hidden" {
									fields = append(fields, g.addCmdMap[idx].HelpShort())
								}
							}
							for _, idx := range g.getCmdIdx {
								if !preCheck(idx, user) {
									continue
								}

								if idx.Name() != "hidden" {
									fields = append(fields, g.getCmdMap[idx].HelpShort())
								}
							}
							for _, idx := range g.delCmdIdx {
								if !preCheck(idx, user) {
									continue
								}

								if idx.Name() != "hidden" {
									fields = append(fields, g.delCmdMap[idx].HelpShort())
								}
							}
							for _, idx := range g.genCmdIdx {
								if !preCheck(idx, user) {
									continue
								}

								if idx.Name() != "hidden" {
									fields = append(fields, g.genCmdMap[idx].HelpShort())
								}
							}

							conversationID, err := g.accessor.GetConversation(msg.Sender.Id)
							if err != nil {
								errCh <- err
								return
							}

							if err := g.accessor.SendMessage(strings.Join(fields, "\n"), conversationID); err != nil {
								errCh <- err
							}

							return
						}

						var helpMsg string
						for _, idx := range g.addCmdIdx {
							if !preCheck(idx, user) {
								continue
							}

							if name := idx.Name(); name != "hidden" && strings.Compare(cmd, name) == 0 {
								helpMsg = g.addCmdMap[idx].Help()
								break
							}
						}
						if helpMsg == "" {
							for _, idx := range g.getCmdIdx {
								if !preCheck(idx, user) {
									continue
								}

								if name := idx.Name(); name != "hidden" && strings.Compare(cmd, name) == 0 {
									helpMsg = g.getCmdMap[idx].Help()
								}
							}
						}
						if helpMsg == "" {
							for _, idx := range g.delCmdIdx {
								if !preCheck(idx, user) {
									continue
								}

								if name := idx.Name(); name != "hidden" && strings.Compare(cmd, name) == 0 {
									helpMsg = g.delCmdMap[idx].Help()
								}
							}
						}
						if helpMsg == "" {
							for _, idx := range g.genCmdIdx {
								if !preCheck(idx, user) {
									continue
								}

								if name := idx.Name(); name != "hidden" && strings.Compare(cmd, name) == 0 {
									helpMsg = g.genCmdMap[idx].Help()
								}
							}
						}

						if helpMsg != "" {
							conversationID, err := g.accessor.GetConversation(msg.Sender.Id)
							if err != nil {
								errCh <- err
								return
							}

							if err := g.accessor.SendMessage(helpMsg, conversationID); err != nil {
								errCh <- err
							}
						}
					}

					if prefix := "add "; strings.HasPrefix(cmdStr, prefix) {
						g.addCmdMu.RLock()
						defer g.addCmdMu.RUnlock()

						for _, idx := range g.addCmdIdx {
							if !preCheck(idx, user) {
								return
							}

							if str, ok := idx.Match(cmdStr); ok {
								if err := g.addCmdMap[idx].Exec(g.accessor, str, msg); err != nil {
									errCh <- err
								}

								return
							}
						}
					}

					if prefix := "get "; strings.HasPrefix(cmdStr, prefix) {
						g.getCmdMu.RLock()
						defer g.getCmdMu.RUnlock()

						for _, idx := range g.getCmdIdx {
							if !preCheck(idx, user) {
								return
							}

							if str, ok := idx.Match(cmdStr); ok {
								if err := g.getCmdMap[idx].Exec(g.accessor, str, msg); err != nil {
									errCh <- err
								}

								return
							}
						}
					}

					for _, prefix := range []string{"del", "delete", "rm", "remove"} {
						if strings.HasPrefix(cmdStr, prefix) {
							g.delCmdMu.RLock()
							defer g.delCmdMu.RUnlock()

							for _, idx := range g.delCmdIdx {
								if !preCheck(idx, user) {
									return
								}

								if str, ok := idx.Match(cmdStr); ok {
									if err := g.delCmdMap[idx].Exec(g.accessor, str, msg); err != nil {
										errCh <- err
									}

									return
								}
							}
						}
					}

					g.genCmdMu.RLock()
					defer g.genCmdMu.RUnlock()

					for _, idx := range g.genCmdIdx {
						if !preCheck(idx, user) {
							return
						}
						if str, ok := idx.Match(cmdStr); ok {
							if err := g.genCmdMap[idx].Exec(g.accessor, str, msg); err != nil {
								errCh <- err
							}
						}
					}
				} else {
					// No command prefix provided, check engine commands
					g.engineCmdMu.RLock()
					defer g.engineCmdMu.RUnlock()

					for _, idx := range g.engineCmdIdx {
						if !preCheck(idx, user) {
							return
						}
						if str, ok := idx.Match(msg.Body); ok {
							if err := g.engineCmdMap[idx].Exec(g.accessor, str, msg); err != nil {
								errCh <- err
							}
						}
					}
				}
			}(msg)
		}
	}

	g.log.Info("Shutting down...")
	g.log.Info("Goodbye")
}

// RegisterCommand registers a command with Graceless.
func (g *Graceless) RegisterCommand(ci commands.CommandIndex, cmd commands.Command) {
	switch ci.Type() {
	case commands.AddCommand:
		g.addCmdMu.Lock()
		defer g.addCmdMu.Unlock()

		g.addCmdIdx = append(g.addCmdIdx, ci)
		g.addCmdMap[ci] = cmd
	case commands.GetCommand:
		g.getCmdMu.Lock()
		defer g.getCmdMu.Unlock()

		g.getCmdIdx = append(g.getCmdIdx, ci)
		g.getCmdMap[ci] = cmd
	case commands.DelCommand:
		g.delCmdMu.Lock()
		defer g.delCmdMu.Unlock()

		g.delCmdIdx = append(g.delCmdIdx, ci)
		g.delCmdMap[ci] = cmd
	case commands.EngineCommand:
		g.engineCmdMu.Lock()
		defer g.engineCmdMu.Unlock()

		g.engineCmdIdx = append(g.engineCmdIdx, ci)
		g.engineCmdMap[ci] = cmd
	default:
		g.genCmdMu.Lock()
		defer g.genCmdMu.Unlock()

		g.genCmdIdx = append(g.genCmdIdx, ci)
		g.genCmdMap[ci] = cmd
	}
}

func (g *Graceless) isRoot(user *permissions.User) bool {
	// No RootIDs are set, everyone can call shutdown.
	if g.config.RootIDs == nil {
		return true
	}

	for _, id := range g.config.RootIDs {
		if user.ID == id {
			return true
		}
	}

	return false
}

func (g *Graceless) sayHello(user *accessors.User) error {
	introStr := g.config.IntroStart
	totalCmds := len(g.addCmdIdx) + len(g.getCmdIdx) + len(g.delCmdIdx) + len(g.genCmdIdx)
	if totalCmds < 11 {
		introStr = fmt.Sprintf("%s\nUnfortunately I only know a few commands right now...\n", introStr)
	}
	if totalCmds > 10 && totalCmds < 21 {
		introStr = fmt.Sprintf("%s\nI can do several things, I think you'll like...\n", introStr)
	}
	if totalCmds > 20 && totalCmds < 31 {
		introStr = fmt.Sprintf("%s\nI can do a number of things to help with your day...\n", introStr)
	}
	if totalCmds > 30 && totalCmds < 41 {
		introStr = fmt.Sprintf("%s\nI can do a lot a thing, it's really cool just how many...\n", introStr)
	}
	if totalCmds > 40 && totalCmds < 51 {
		introStr = fmt.Sprintf("%s\nI can do so many things, like you won't believe your eyes...\n", introStr)
	}
	if totalCmds > 50 && totalCmds < 101 {
		introStr = fmt.Sprintf("%s\nI can do a great many things, it's totally amazing...\n", introStr)
	}
	if totalCmds > 100 {
		introStr = fmt.Sprintf("%s\nI can do way too many thints, like seriously...\n", introStr)
	}

	finishStr := strings.Replace(g.config.IntroFinish, "[tag]", "`", -1)
	finishStr = strings.Replace(finishStr, "[prefix]", g.config.CmdPrefix, -1)
	introStr = fmt.Sprintf("%s%s", introStr, finishStr)

	//introStr = config.MarshalMessage(65, introStr)

	conversationID, err := g.accessor.GetConversation(user.Id)
	if err != nil {
		return err
	}

	return g.accessor.SendMessage(introStr, conversationID)
}

func registerDefaultCommands(g *Graceless, stopCh chan error) error {
	if shutdownCmd := commands.NewShutdownCommand(stopCh); shutdownCmd != nil {
		g.RegisterCommand(shutdownCmd, shutdownCmd)
	}
	if safemodeCmd := commands.NewSafemodeCommand(g.config); safemodeCmd != nil {
		g.RegisterCommand(safemodeCmd, safemodeCmd)
	}
	if whoisCmd := commands.NewUserIsCommand(); whoisCmd != nil {
		g.RegisterCommand(whoisCmd, whoisCmd)
	}

	permsAddCmd, err := commands.NewAddPerms(g.pm)
	if err != nil {
		return err
	}
	g.RegisterCommand(permsAddCmd, permsAddCmd)

	permsGetCmd, err := commands.NewGetPerms(g.pm)
	if err != nil {
		return err
	}
	g.RegisterCommand(permsGetCmd, permsGetCmd)

	permsDelCmd, err := commands.NewDelPerms(g.pm)
	if err != nil {
		return err
	}
	g.RegisterCommand(permsDelCmd, permsDelCmd)

	if g.config.WithGoEngine {
		if goengineCmd := golang.NewEngineCommand(); goengineCmd != nil {
			g.RegisterCommand(goengineCmd, goengineCmd)
		}
	}

	if g.config.WithJSEngine {
		if jsengineCmd := js.NewEngineCommand(); jsengineCmd != nil {
			g.RegisterCommand(jsengineCmd, jsengineCmd)
		}
	}

	return nil
}
