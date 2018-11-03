package graceless // github.com/justanotherorganization/graceless

import (
	"context"
	"strings"

	cmds "github.com/justanotherorganization/graceless/commands"
	"github.com/justanotherorganization/graceless/config"
	"github.com/justanotherorganization/graceless/internal"
	"github.com/justanotherorganization/justanotherbotkit/commands"
	"github.com/justanotherorganization/justanotherbotkit/proto"
	"github.com/justanotherorganization/justanotherbotkit/transport"
	"github.com/justanotherorganization/l5424"
	"github.com/justanotherorganization/l5424/x5424"
	"github.com/pkg/errors"
)

type (
	// Graceless is a clumsy bot.
	Graceless struct {
		config *config.Config
	}
)

// New creates a new Graceless instance.
func New(conf *config.Config) (*Graceless, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}

	return &Graceless{
		config: conf,
	}, nil
}

// Start our Graceless bot.
func (g *Graceless) Start(ctx context.Context, cancel context.CancelFunc, errCh chan error) {
	startMsg := "Starting Graceless"

	if g.Safemode() {
		startMsg += " in safemode"
	}

	g.config.Log(x5424.Severity, l5424.InfoLvl, startMsg, "\n")

	// Start building our database (as early as possible).
	if !g.Safemode() {
		cmds.SetUserDB(g.config.UserDB)

		go func() {
			// We wouldn't be building the database if we hadn't passed in
			// a backend so treat errors in this instance as fatal.
			users, err := g.config.Transport.GetUsers()
			if err != nil {
				errCh <- err
				cancel()
				return
			}

			for _, user := range users {
				// This probably needs to be expanded a bit.
				if user.GetID() == "USLACKBOT" {
					continue
				}

				_user, err := g.config.UserDB.GetUser(context.Background(), user.GetID())
				if err != nil {
					errCh <- err
					cancel()
					return
				}

				if _user == nil {
					_user, err = g.config.UserDB.CreateUser(context.Background(), user)
					if err != nil {
						errCh <- err
						cancel()
						return
					}
				}

				for _, rootUser := range g.config.RootUsers {
					if _user.GetID() == rootUser ||
						strings.EqualFold(_user.GetName(), rootUser) {
						isSet := false
						for _, p := range _user.GetPermissions() {
							if p == "root" {
								isSet = true
								break
							}
						}

						if !isSet {
							// Update our transport.User.BaseUser with the contents from our database saved user.
							user.BaseUser = _user.(*pb.BaseUser)
							user.Permissions = append(user.Permissions, "root")
							_, err := g.config.UserDB.UpdateUser(context.Background(), user)
							if err != nil {
								errCh <- err
								cancel()
								return
							}
						}
					}
				}
			}
		}()
	}

	eventCh := make(chan *transport.Event)
	go func() {
		g.config.Transport.TunnelEvents(ctx, eventCh, errCh)
	}()

	cmds.Register(
		&commands.Command{
			Use:    "shutdown",
			Short:  "Shutdown the bot",
			Hidden: true,
			Perms:  []string{"shutdown"},
			ExecFunc: func(ev *transport.Event) error {
				cancel()
				return nil
			},
		},
	// &commands.Command{
	// 	Use: "safemode",
	// 	Short: "toggle safemode status",
	// 	Hidden: true,
	// 	Perms: []string{"safemode"},
	// 	ExecFunc: func(ev *transport.Event) error {
	// 		// TODO:
	// 	}
	// }
	)

	for stop := false; !stop; {
		select {
		case <-ctx.Done():
			stop = true
		case ev := <-eventCh:
			// Handle messages in their own goroutines so they don't block.
			go func(ev *transport.Event) {
				g.config.Log(x5424.Severity, l5424.DebugLvl, "%s (%s): %v", ev.Origin.Sender.Name, ev.Origin.Sender.ID, ev.Body)

				if ev.Origin.Sender.ID == "USLACKBOT" {
					return
				}

				user, err := g.config.Transport.GetUser(ev.Origin.Sender.ID)
				if err != nil {
					errCh <- err
					return
				}

				// FIXME:
				// if _user.IsBot {
				// 	return
				// }

				if !g.Safemode() {
					// Ensure that our user exists before operating on it, this shouldn't be an issue
					// _unless_ a command comes in before the above database build has completed.
					_user, err := g.config.UserDB.GetUser(context.Background(), user.GetID())
					if err != nil {
						errCh <- errors.Wrap(err, "config.UserDB.GetUser")
						return
					}

					if _user == nil {
						_, err = g.config.UserDB.CreateUser(context.Background(), user)
						if err != nil {
							errCh <- errors.Wrap(err, "config.UserDB.CreateUser")
							return
						}
					}

					// Update our transport.User.BaseUser with the contents from our database saved user.
					user.BaseUser = _user.(*pb.BaseUser)
					saidHello, err := internal.SayHello(user, g.config)
					if err != nil {
						errCh <- errors.Wrap(err, "internal.SayHello")
						// Don't return, we still want the command to get processed...
					}

					if saidHello {
						user.Permissions = append(user.Permissions, "hello")
						_, err := g.config.UserDB.UpdateUser(context.Background(), user)
						if err != nil {
							errCh <- err
							// Don't return, we still want the command to get processed...
						}
					}
				}

				if prefix := g.config.CmdPrefix; strings.HasPrefix(ev.Body, prefix) {
					ev.Body = strings.TrimSpace(strings.TrimPrefix(ev.Body, prefix))

					if err := cmds.Execute(ev); err != nil {
						errCh <- errors.Wrap(err, "cmds.Execute")
						return
					}
				}
			}(ev)
		}
	}

	g.config.Log(x5424.Severity, l5424.InfoLvl, "Shutting down...\n")
	g.config.Log(x5424.Severity, l5424.InfoLvl, "Goodbye\n")
}

// Safemode returns whather we are currently running in safemode.
func (g *Graceless) Safemode() (b bool) {
	if g.config.Safemode {
		return true
	}

	// This shouldn't be possible if safemode is false, but just in case...
	if g.config.UserDB == nil {
		return true
	}

	return false
}
