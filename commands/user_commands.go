package commands

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/justanotherorganization/justanotherbotkit/accessors"
	"github.com/pkg/errors"
)

type (
	// UserIsCommand is a whois command.
	UserIsCommand struct {
		*Index
		usage string
	}
)

// NewUserIsCommand returns a new whois commands.
func NewUserIsCommand() *UserIsCommand {
	return &UserIsCommand{
		Index: &Index{
			CmdName: "whois",
		},
		usage: "Usage: whois <@user>",
	}
}

// HelpShort returns the short help message.
func (*UserIsCommand) HelpShort() string {
	return "whois : get a user's ID" // FIXME: use a format string for this...
}

// Help returns the whois usage string.
func (c *UserIsCommand) Help() string {
	return c.usage
}

// Match matches a command against the whois command.
func (c *UserIsCommand) Match(str string) (string, bool) {
	if prefix := c.CmdName; strings.HasPrefix(str, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(str, prefix)), true
	}

	return str, false
}

// Exec performs a whois lookup.
func (c *UserIsCommand) Exec(acc accessors.Accessor, cmdStr string, ev accessors.MessageEvent) error {
	if cmdStr == "" || strings.Contains(cmdStr, " ") {
		if err := acc.SendMessage(c.usage, ev.Origin); err != nil {
			return errors.Wrap(err, "acc.SendMessage")
		}

		return nil
	}

	cmdStr = strings.TrimPrefix(strings.TrimSuffix(cmdStr, ">"), "<@")
	_user, err := acc.GetUser(cmdStr)
	if err != nil {
		return errors.Wrap(err, "acc.GetUser")
	}

	m := jsonpb.Marshaler{
		Indent: "  ",
	}
	userStr, err := m.MarshalToString(_user)
	if err != nil {
		return errors.Wrap(err, "m.MarshaltoString")
	}

	conversationID, err := acc.GetConversation(ev.Sender.Id)
	if err != nil {
		return errors.Wrap(err, "acc.GetConversation")
	}

	if err := acc.SendMessage(fmt.Sprintf("```%s```", userStr), conversationID); err != nil {
		return errors.Wrap(err, "acc.SendMessage")
	}

	return nil
}
