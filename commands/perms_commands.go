package commands

import (
	"fmt"
	"strings"

	"github.com/justanotherorganization/justanotherbotkit/accessors"
	"github.com/justanotherorganization/justanotherbotkit/permissions"
	"github.com/pkg/errors"
)

type (
	// AddPerms because who doesn't like permissions?
	AddPerms struct {
		*Index
		usage string
		pm    *permissions.Manager
	}

	// GetPerms that you added.
	GetPerms struct {
		*Index
		usage string
		pm    *permissions.Manager
	}

	// DelPerms that you added also.
	DelPerms struct {
		*Index
		usage string
		pm    *permissions.Manager
	}
)

// NewAddPerms returns a new AddPerms command.
func NewAddPerms(pm *permissions.Manager) (*AddPerms, error) {
	if pm == nil {
		return nil, errors.New("pm must not be nil")
	}

	return &AddPerms{
		Index: &Index{
			CmdName:     "add perm",
			CmdNeedsDB:  true,
			CommandType: AddCommand,
			CmdPerms:    []string{"perms"},
		},
		pm:    pm,
		usage: "Usage: add perm <user id> <perm>",
	}, nil
}

// HelpShort returns the short help message for the add perms command.
func (*AddPerms) HelpShort() string {
	return "add perm : add permissions to a user" // FIXME: use a format string for this...
}

// Match matches a command string against the AddPerms command.
func (*AddPerms) Match(str string) (string, bool) {
	_str := strings.ToLower(str)

	if prefix := "add perm"; strings.HasPrefix(_str, prefix) {
		return strings.TrimSpace(str[len(prefix):]), true
	}

	if prefix := "add permission"; strings.HasPrefix(_str, prefix) {
		return strings.TrimSpace(str[len(prefix):]), true
	}

	return str, false
}

// Exec performs the add perms command.
func (ap *AddPerms) Exec(acc accessors.Accessor, cmdStr string, ev accessors.MessageEvent) error {
	if ap.pm == nil {
		return errors.New("pm must not be nil")
	}

	fields := strings.Fields(cmdStr)
	if len(fields) != 2 {
		if err := acc.SendMessage(ap.usage, ev.Origin); err != nil {
			return errors.Wrap(err, "acc.SendMessage")
		}

		return nil
	}

	_user, err := ap.pm.GetUser(fields[0])
	if err != nil {
		return errors.Wrap(err, "pm.GetUser")
	}

	if _user == nil {
		if err := acc.SendMessage("User %s not in database", fields[0]); err != nil {
			return errors.Wrap(err, "acc.SendMessage")
		}
	}

	if err := _user.AddPerms(fields[1]); err != nil {
		return errors.Wrap(err, "_user.AddPerms")
	}

	return nil
}

// Help returns the add perms usage string.
func (ap *AddPerms) Help() string {
	return ap.usage
}

// NewGetPerms returns a new get perms command.
func NewGetPerms(pm *permissions.Manager) (*GetPerms, error) {
	if pm == nil {
		return nil, errors.New("pm must not be nil")
	}

	return &GetPerms{
		Index: &Index{
			CmdName:     "get perms",
			CmdNeedsDB:  true,
			CommandType: GetCommand,
			CmdPerms:    []string{"perms"},
		},
		pm:    pm,
		usage: "Usage: get perms <user id>",
	}, nil
}

// HelpShort returns the short help message.
func (*GetPerms) HelpShort() string {
	return "get perms : get permissions for a user" // FIXME: use a format string for this.
}

// Match matches a command string against the perms command.
func (*GetPerms) Match(str string) (string, bool) {
	_str := strings.ToLower(str)

	if prefix := "get perms"; strings.HasPrefix(_str, prefix) {
		return strings.TrimSpace(str[len(prefix):]), true
	}

	if prefix := "get permissions"; strings.HasPrefix(_str, prefix) {
		return strings.TrimSpace(str[len(prefix):]), true
	}

	return str, false
}

// Exec performs the get perms command.
func (gp *GetPerms) Exec(acc accessors.Accessor, cmdStr string, ev accessors.MessageEvent) error {
	if gp.pm == nil {
		return errors.New("pm must not be nil")
	}

	if cmdStr == "" || strings.Contains(cmdStr, " ") {
		if err := acc.SendMessage(gp.usage, ev.Origin); err != nil {
			return errors.Wrap(err, "acc.SendMessage")
		}

		return nil
	}

	_user, err := gp.pm.GetUser(cmdStr)
	if err != nil {
		return errors.Wrap(err, "pm.GetUser")
	}

	if _user == nil {
		if err := acc.SendMessage(fmt.Sprintf("User %s not in database", cmdStr), ev.Origin); err != nil {
			return errors.Wrap(err, "acc.SendMessage")
		}
	}

	perms, err := _user.GetPerms()
	if err != nil {
		return errors.Wrap(err, "_user.GetPerms")
	}

	if err := acc.SendMessage(fmt.Sprintf("Permissions for %s : %v", _user.ID, perms), ev.Origin); err != nil {
		return errors.Wrap(err, "acc.SendMessage")
	}

	return nil
}

// Help returns the get perms usage string.
func (gp *GetPerms) Help() string {
	return gp.usage
}

// NewDelPerms returns a new del perm command.
func NewDelPerms(pm *permissions.Manager) (*DelPerms, error) {
	if pm == nil {
		return nil, errors.New("pm must not be nil")
	}

	return &DelPerms{
		Index: &Index{
			CmdName:     "del perm",
			CmdNeedsDB:  true,
			CommandType: DelCommand,
			CmdPerms:    []string{"perms"},
		},
		pm:    pm,
		usage: "Usage: del perm <user id> <perm>",
	}, nil
}

// HelpShort returns the short help message.
func (*DelPerms) HelpShort() string {
	return "del perm : delete permissions from a user" // FIXME: use a format string...
}

// Match matches a command string against the del perm command
func (*DelPerms) Match(str string) (string, bool) {
	_str := strings.ToLower(str)

	if prefix := "del perm"; strings.HasPrefix(_str, prefix) {
		return strings.TrimSpace(str[len(prefix):]), true
	}

	if prefix := "delete perm"; strings.HasPrefix(_str, prefix) {
		return strings.TrimSpace(str[len(prefix):]), true
	}

	if prefix := "rm perm"; strings.HasPrefix(_str, prefix) {
		return strings.TrimSpace(str[len(prefix):]), true
	}

	if prefix := "remove perm"; strings.HasPrefix(_str, prefix) {
		return strings.TrimSpace(str[len(prefix):]), true
	}

	return str, false
}

// Exec performs the del perm command.
func (dp *DelPerms) Exec(acc accessors.Accessor, cmdStr string, ev accessors.MessageEvent) error {
	if dp.pm == nil {
		return errors.New("pm must not be nil")
	}

	fields := strings.Fields(cmdStr)
	if len(fields) != 2 {
		if err := acc.SendMessage(dp.usage, ev.Origin); err != nil {
			return errors.Wrap(err, "acc.SendMessage")
		}

		return nil
	}

	_user, err := dp.pm.GetUser(fields[0])
	if err != nil {
		return errors.Wrap(err, "pm.GetUser")
	}

	if _user == nil {
		if err := acc.SendMessage("User %s not in database", fields[0]); err != nil {
			return errors.Wrap(err, "acc.SendMessage")
		}
	}

	if err := _user.DelPerms(fields[1]); err != nil {
		return errors.Wrap(err, "_user.DelPerms")
	}

	return nil
}

// Help returns the del perm usage string.
func (dp *DelPerms) Help() string {
	return dp.usage
}
