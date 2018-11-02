package commands

import (
	"github.com/justanotherorganization/justanotherbotkit/commands"
	"github.com/justanotherorganization/justanotherbotkit/transport"
	"github.com/justanotherorganization/justanotherbotkit/users"
)

var _root commands.Command

// Register a new command.
func Register(c ...*commands.Command) {
	for _, _c := range c {
		_root.AddCommand(_c)
	}
}

// Execute a command.
func Execute(ev *transport.Event) error {
	return _root.Execute(ev)
}

// Count the number of base commands present in root.
func Count() int {
	return len(_root.Children())
}

// SetUserDB sets the userDB on the root command.
func SetUserDB(db users.DB) {
	_root.UserDB = db
}
