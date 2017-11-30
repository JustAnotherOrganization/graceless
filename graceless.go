package graceless

import (
	"errors"
	"strings"
	"sync"

	"github.com/justanotherorganization/justanotherbotkit/accessors"
	"github.com/justanotherorganization/justanotherbotkit/permissions"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultCommandPrefix is the default command prefix.
	DefaultCommandPrefix = `.`
)

type (
	// Graceless is a clumsy bot.
	Graceless struct {
		accessor accessors.Accessor
		log      *logrus.Entry
		config   *Config
		pm       *permissions.Manager

		genCmdMu  *sync.RWMutex
		genCmdIdx []CommandIndex
		genCmdMap map[CommandIndex]Command

		addCmdMu  *sync.RWMutex
		addCmdIdx []CommandIndex
		addCmdMap map[CommandIndex]Command

		getCmdMu  *sync.RWMutex
		getCmdIdx []CommandIndex
		getCmdMap map[CommandIndex]Command

		delCmdMu  *sync.RWMutex
		delCmdIdx []CommandIndex
		delCmdMap map[CommandIndex]Command
	}

	// Config is the config for a graceless bot.
	Config struct {
		// RootIDs contains the IDs of the users who can shutdown the bot
		// (if no IDs are provided anyone will be able to shutdown the bot).
		RootIDs []string
		// CmdPrefix is the prefix to match on for recognizing commands.
		CmdPrefix string
		// Safemode sets the bot into safemode.
		Safemode bool
	}
)

// New creates a new Graceless instance.
func New(accessor accessors.Accessor, log *logrus.Entry, config *Config, pm *permissions.Manager) (*Graceless, error) {
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
		config.CmdPrefix = DefaultCommandPrefix
	}

	if pm == nil {
		config.Safemode = true
		log.Info("Starting in safemode")
	}

	return &Graceless{
		accessor: accessor,
		log:      log,
		config:   config,
		pm:       pm,

		addCmdMu:  &sync.RWMutex{},
		addCmdIdx: []CommandIndex{},
		addCmdMap: make(map[CommandIndex]Command),

		getCmdMu:  &sync.RWMutex{},
		getCmdIdx: []CommandIndex{},
		getCmdMap: make(map[CommandIndex]Command),

		delCmdMu:  &sync.RWMutex{},
		delCmdIdx: []CommandIndex{},
		delCmdMap: make(map[CommandIndex]Command),

		genCmdMu:  &sync.RWMutex{},
		genCmdIdx: []CommandIndex{},
		genCmdMap: make(map[CommandIndex]Command),
	}, nil
}

// Start our Graceless bot.
func (g *Graceless) Start(errCh, stopCh chan error) {
	// Register some default commands.
	g.RegisterCommand(newShutdownIdx(), newShutdownCommand(stopCh))
	g.RegisterCommand(newSafemodeIdx(), newSafemodeCommand(g.config))
	if permsAddCmd, err := newAddPerms(g.pm); err != nil {
		errCh <- err
	} else {
		g.RegisterCommand(newAddPermsIdx(), permsAddCmd)
	}
	if permsGetCmd, err := newGetPerms(g.pm); err != nil {
		errCh <- err
	} else {
		g.RegisterCommand(newGetPermsIdx(), permsGetCmd)
	}
	if permsDelCmd, err := newDelPerms(g.pm); err != nil {
		errCh <- err
	} else {
		g.RegisterCommand(newDelPermsIdx(), permsDelCmd)
	}

	eventCh := make(chan accessors.MessageEvent)

	go func() {
		g.accessor.TunnelEvents(eventCh, errCh, stopCh)
	}()

	preCheck := func(idx CommandIndex, user *permissions.User) bool {
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

out:
	for {
		select {
		case <-stopCh:
			break out
		case msg := <-eventCh:
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
				} else {
					user = &permissions.User{
						ID: msg.Sender.Id,
					}
				}

				if strings.HasPrefix(msg.Body, g.config.CmdPrefix) {
					cmdStr := strings.TrimPrefix(msg.Body, g.config.CmdPrefix)

					if prefix := "help"; strings.HasPrefix(cmdStr, prefix) {
						cmd := strings.TrimSpace(strings.TrimPrefix(cmdStr, prefix))
						if cmd == "" {
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

							var fields []string
							for _, idx := range g.addCmdIdx {
								if !preCheck(idx, user) {
									continue
								}

								if idx.Name() != "hidden" {
									fields = append(fields, idx.HelpShort())
								}
							}
							for _, idx := range g.getCmdIdx {
								if !preCheck(idx, user) {
									continue
								}

								if idx.Name() != "hidden" {
									fields = append(fields, idx.HelpShort())
								}
							}
							for _, idx := range g.delCmdIdx {
								if !preCheck(idx, user) {
									continue
								}

								if idx.Name() != "hidden" {
									fields = append(fields, idx.HelpShort())
								}
							}
							for _, idx := range g.genCmdIdx {
								if !preCheck(idx, user) {
									continue
								}

								if idx.Name() != "hidden" {
									fields = append(fields, idx.HelpShort())
								}
							}

							// FIXME: help messages should always be returned in direct message.
							if err := g.accessor.SendMessage(strings.Join(fields, "\n"), msg.Origin); err != nil {
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
							// FIXME: help messages should always be returned in direct message.
							if err := g.accessor.SendMessage(helpMsg, msg.Origin); err != nil {
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
								if err := g.addCmdMap[idx].Exec(user, g.accessor, str, msg); err != nil {
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
								if err := g.getCmdMap[idx].Exec(user, g.accessor, str, msg); err != nil {
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
									if err := g.delCmdMap[idx].Exec(user, g.accessor, str, msg); err != nil {
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
						if str, ok := idx.Match(cmdStr); ok {
							if err := g.genCmdMap[idx].Exec(user, g.accessor, str, msg); err != nil {
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
func (g *Graceless) RegisterCommand(ci CommandIndex, cmd Command) {
	switch ci.Type() {
	case AddCommand:
		g.addCmdMu.Lock()
		defer g.addCmdMu.Unlock()

		g.addCmdIdx = append(g.addCmdIdx, ci)
		g.addCmdMap[ci] = cmd
	case GetCommand:
		g.getCmdMu.Lock()
		defer g.getCmdMu.Unlock()

		g.getCmdIdx = append(g.getCmdIdx, ci)
		g.getCmdMap[ci] = cmd
	case DelCommand:
		g.delCmdMu.Lock()
		defer g.delCmdMu.Unlock()

		g.delCmdIdx = append(g.delCmdIdx, ci)
		g.delCmdMap[ci] = cmd
	default:
		g.genCmdMu.Lock()
		defer g.genCmdMu.Unlock()

		g.genCmdIdx = append(g.genCmdIdx, ci)
		g.genCmdMap[ci] = cmd
	}
}

func (g *Graceless) isRoot(user *permissions.User) bool {
	// Not RootIDs are set, everyone can call shutdown.
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
