package commands

import (
	"fmt"
	"strings"

	"github.com/justanotherorganization/graceless/config"
	"github.com/justanotherorganization/justanotherbotkit/accessors"
	"github.com/pkg/errors"
)

type (
	// ShutdownCommand like the first command you should add right?
	ShutdownCommand struct {
		*Index
		stopCh chan error
	}

	// SafemodeCommand gives us the ability to restrict commands with a safemode flag.
	SafemodeCommand struct {
		*Index
		config *config.Config
	}
)

// NewShutdownCommand returns a new shutdown command.
func NewShutdownCommand(stopCh chan error) *ShutdownCommand {
	return &ShutdownCommand{
		Index: &Index{
			CmdName:  "hidden",
			CmdPerms: []string{"shutdown"},
		},
		stopCh: stopCh,
	}
}

// HelpShort is a dummy cause this is a hidden command.
func (*ShutdownCommand) HelpShort() string {
	return ""
}

// Match a command against the shutdown command.
func (*ShutdownCommand) Match(str string) (string, bool) {
	if strings.Compare(str, "shutdown") == 0 {
		return str, true
	}

	return "", false
}

// Exec performs a shutdown.
func (sh *ShutdownCommand) Exec(acc accessors.Accessor, cmdStr string, ev accessors.MessageEvent) error {
	sh.stopCh <- errors.New("shutdown")
	return nil
}

// Help is a dummy cause this is a hidden command.
func (*ShutdownCommand) Help() string {
	return ""
}

// NewSafemodeCommand returns a new SafemodeCommand.
func NewSafemodeCommand(conf *config.Config) *SafemodeCommand {
	return &SafemodeCommand{
		Index: &Index{
			CmdName:  "hidden",
			CmdPerms: []string{"safemode"},
		},
		config: conf,
	}
}

// HelpShort is a dummy cause this is a hidden command.
func (*SafemodeCommand) HelpShort() string {
	return ""
}

// Match matches a command string against the safemode command.
func (s *SafemodeCommand) Match(str string) (string, bool) {
	if prefix := "safemode"; strings.HasPrefix(str, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(str, prefix)), true
	}

	return str, false
}

// Exec performs the safemode command.
func (s *SafemodeCommand) Exec(acc accessors.Accessor, cmdStr string, ev accessors.MessageEvent) error {
	if strings.Compare(cmdStr, "true") == 0 {
		s.config.Safemode = true
	}
	if strings.Compare(cmdStr, "false") == 0 {
		s.config.Safemode = false
	}

	return acc.SendMessage(fmt.Sprintf("safemode: %v", s.config.Safemode), ev.Origin)
}

// Help is a dummy cause this is a hidden command.
func (*SafemodeCommand) Help() string {
	return ""
}
