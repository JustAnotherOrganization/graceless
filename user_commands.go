package graceless

import (
	"fmt"
	"strings"

	"github.com/justanotherorganization/justanotherbotkit/accessors"
	"github.com/justanotherorganization/justanotherbotkit/permissions"
	"github.com/pkg/errors"
)

type (
	userIsCommand struct {
		*Index
		usage string
	}
)

func newUserIsCommand() *userIsCommand {
	return &userIsCommand{
		Index: &Index{
			CmdName: "whois",
		},
		usage: "Usage: whois <@user>",
	}
}

func (*userIsCommand) HelpShort() string {
	return "whois : get a user's ID" // FIXME: use a format string for this...
}

func (c *userIsCommand) Help() string {
	return c.usage
}

func (c *userIsCommand) Match(str string) (string, bool) {
	if prefix := c.CmdName; strings.HasPrefix(str, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(str, prefix)), true
	}

	return str, false
}

func (c *userIsCommand) Exec(user *permissions.User, acc accessors.Accessor, cmdStr string, ev accessors.MessageEvent) error {
	if cmdStr == "" || strings.Contains(cmdStr, " ") {
		if err := acc.SendMessage(c.usage, ev.Origin); err != nil {
			return errors.Wrap(err, "acc.SendMessage")
		}

		return nil
	}

	cmdStr = strings.TrimPrefix(strings.TrimSuffix(cmdStr, ">"), "<@")
	if err := acc.SendMessage(fmt.Sprintf("user id %s", cmdStr), ev.Origin); err != nil {
		return errors.Wrap(err, "acc.SendMessage")
	}

	return nil
}
