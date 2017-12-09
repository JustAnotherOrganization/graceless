package graceless

import (
	"fmt"
	"strings"

	"github.com/justanotherorganization/justanotherbotkit/accessors"
	"github.com/justanotherorganization/justanotherbotkit/permissions"
	"github.com/pkg/errors"
)

type (
	// Shutdown, like the first command you should add right?
	shutdownCommand struct {
		*Index
		stopCh chan error
	}

	// Safemode gives us the ability to restrict commands with a safemode flag.
	safemodeCommand struct {
		*Index
		config *Config
	}
)

func newShutdownCommand(stopCh chan error) *shutdownCommand {
	return &shutdownCommand{
		Index: &Index{
			CmdName:  "hidden",
			CmdPerms: []string{"shutdown"},
		},
		stopCh: stopCh,
	}
}

func (*shutdownCommand) HelpShort() string {
	return ""
}

func (*shutdownCommand) Match(str string) (string, bool) {
	if strings.Compare(str, "shutdown") == 0 {
		return str, true
	}

	return "", false
}

func (sh *shutdownCommand) Exec(user *permissions.User, acc accessors.Accessor, cmdStr string, ev accessors.MessageEvent) error {
	sh.stopCh <- errors.New("shutdown")
	return nil
}

func (*shutdownCommand) Help() string {
	return ""
}

func newSafemodeCommand(config *Config) *safemodeCommand {
	return &safemodeCommand{
		Index: &Index{
			CmdName:  "hidden",
			CmdPerms: []string{"safemode"},
		},
		config: config,
	}
}

func (*safemodeCommand) HelpShort() string {
	return ""
}

func (s *safemodeCommand) Match(str string) (string, bool) {
	if prefix := "safemode"; strings.HasPrefix(str, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(str, prefix)), true
	}

	return str, false
}

func (s *safemodeCommand) Exec(user *permissions.User, acc accessors.Accessor, cmdStr string, ev accessors.MessageEvent) error {
	if strings.Compare(cmdStr, "true") == 0 {
		s.config.Safemode = true
	}
	if strings.Compare(cmdStr, "false") == 0 {
		s.config.Safemode = false
	}

	return acc.SendMessage(fmt.Sprintf("safemode: %v", s.config.Safemode), ev.Origin)
}

func (*safemodeCommand) Help() string {
	return ""
}
