package graceless

import (
	"fmt"
	"strings"

	"github.com/justanotherorganization/justanotherbotkit/accessors"
	"github.com/justanotherorganization/justanotherbotkit/permissions"
	"github.com/pkg/errors"
)

type (
	// Perms because who doesn't like permissions?
	addPermsIndex struct {
		*Index
	}

	addPerms struct {
		usage string
		pm    *permissions.Manager
	}

	getPermsIndex struct {
		*Index
	}

	getPerms struct {
		usage string
		pm    *permissions.Manager
	}

	delPermsIndex struct {
		*Index
	}

	delPerms struct {
		usage string
		pm    *permissions.Manager
	}
)

func newAddPermsIdx() *addPermsIndex {
	return &addPermsIndex{
		Index: &Index{
			CmdName:     "add perm",
			CmdNeedsDB:  true,
			CommandType: AddCommand,
		},
	}
}

func newAddPerms(pm *permissions.Manager) (*addPerms, error) {
	if pm == nil {
		return nil, errors.New("pm must not be nil")
	}

	return &addPerms{
		pm:    pm,
		usage: "Usage: add perm <user id> <perm>",
	}, nil
}

func (*addPermsIndex) HelpShort() string {
	return "add perm : add permissions to a user" // FIXME: use a format string for this...
}

func (*addPermsIndex) Match(str string) (string, bool) {
	_str := strings.ToLower(str)

	if prefix := "add perm"; strings.HasPrefix(_str, prefix) {
		return strings.TrimSpace(str[len(prefix):]), true
	}

	if prefix := "add permission"; strings.HasPrefix(_str, prefix) {
		return strings.TrimSpace(str[len(prefix):]), true
	}

	return str, false
}

func (ap *addPerms) Exec(user *permissions.User, acc accessors.Accessor, cmdStr string, ev accessors.MessageEvent) error {
	if ap.pm == nil {
		return errors.New("pm must not be nil")
	}

	ok, err := user.GetPerm("perms")
	if err != nil {
		return errors.Wrap(err, "user.GetPerm")
	}

	if ok {
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
	}

	return nil
}

func (ap *addPerms) Help() string {
	return ap.usage
}

func newGetPermsIdx() *getPermsIndex {
	return &getPermsIndex{
		Index: &Index{
			CmdName:     "get perms",
			CmdNeedsDB:  true,
			CommandType: GetCommand,
		},
	}
}

func newGetPerms(pm *permissions.Manager) (*getPerms, error) {
	if pm == nil {
		return nil, errors.New("pm must not be nil")
	}

	return &getPerms{
		pm:    pm,
		usage: "Usage: get perms <user id>",
	}, nil
}

func (*getPermsIndex) HelpShort() string {
	return "get perms : get permissions for a user" // FIXME: use a format string for this.
}

func (*getPermsIndex) Match(str string) (string, bool) {
	_str := strings.ToLower(str)

	if prefix := "get perms"; strings.HasPrefix(_str, prefix) {
		return strings.TrimSpace(str[len(prefix):]), true
	}

	if prefix := "get permissions"; strings.HasPrefix(_str, prefix) {
		return strings.TrimSpace(str[len(prefix):]), true
	}

	return str, false
}

func (gp *getPerms) Exec(user *permissions.User, acc accessors.Accessor, cmdStr string, ev accessors.MessageEvent) error {
	if gp.pm == nil {
		return errors.New("pm must not be nil")
	}

	ok, err := user.GetPerm("perms")
	if err != nil {
		return errors.Wrap(err, "user.GetPerm")
	}

	if ok {
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
	}

	return nil
}

func (gp *getPerms) Help() string {
	return gp.usage
}

func newDelPermsIdx() *delPermsIndex {
	return &delPermsIndex{
		Index: &Index{
			CmdName:     "del perm",
			CmdNeedsDB:  true,
			CommandType: DelCommand,
		},
	}
}

func newDelPerms(pm *permissions.Manager) (*delPerms, error) {
	if pm == nil {
		return nil, errors.New("pm must not be nil")
	}

	return &delPerms{
		pm:    pm,
		usage: "Usage: del perm <user id> <perm>",
	}, nil
}

func (*delPermsIndex) HelpShort() string {
	return "del perm : delete permissions from a user" // FIXME: use a format string...
}

func (*delPermsIndex) Match(str string) (string, bool) {
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

func (dp *delPerms) Exec(user *permissions.User, acc accessors.Accessor, cmdStr string, ev accessors.MessageEvent) error {
	if dp.pm == nil {
		return errors.New("pm must not be nil")
	}

	ok, err := user.GetPerm("perms")
	if err != nil {
		return errors.Wrap(err, "user.GetPerm")
	}

	if ok {
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
	}

	return nil
}

func (dp *delPerms) Help() string {
	return dp.usage
}