package graceless

import (
	"github.com/justanotherorganization/justanotherbotkit/accessors"
	"github.com/justanotherorganization/justanotherbotkit/permissions"
)

// FIXME: thiu really is terribad...

const (
	// GenericCommand is the default Command type.
	GenericCommand = iota
	// AddCommand should be set as the Command type in an "add" prefixed command.
	AddCommand
	// GetCommand should be set as the Command type in a "get" prefixed command.
	GetCommand
	// DelCommand should be set as the Command type in a "del", "delete", "rm",
	// or "remove" prefixed command.
	DelCommand
)

type (
	// CommandIndex is used to track a Graceless command in a map.
	CommandIndex interface {
		// HelpShort returns the short help message.
		HelpShort() string
		// Match matches a string against a command.
		Match(string) (string, bool)
		// Name returns the command name (for help purposes).
		Name() string
		// NeedsDB tells if a command requires a database connection to function.
		NeedsDB() bool
		// RequiredPerms returns the required permissions for the command.
		RequiredPerms() []string
		// NoSafemode tells if a command can't be ran in safemode.
		NoSafemode() bool
		// Type returns the command type.
		Type() int
	}

	// Command is a Graceless command.
	Command interface {
		// Exec executes a given command returning any errors.
		Exec(user *permissions.User, acc accessors.Accessor, cmdStr string, ev accessors.MessageEvent) error
		// Help returns the commands help message.
		Help() string
	}

	// Index should be embedded into a struct wanting to implement CommandIndex.
	Index struct {
		CmdName       string
		CmdNoSafemode bool
		CmdNeedsDB    bool
		CmdPerms      []string
		CommandType   int
	}
)

// Name returns the command name (for help purposes).
func (ci *Index) Name() string {
	return ci.CmdName
}

// NoSafemode tells if a command can't be ran in safemode.
func (ci *Index) NoSafemode() bool {
	return ci.CmdNoSafemode
}

// NeedsDB tells if a command requires a database connection to function.
func (ci *Index) NeedsDB() bool {
	return ci.CmdNeedsDB
}

// RequiredPerms returns the required permissions for the command.
func (ci *Index) RequiredPerms() []string {
	return ci.CmdPerms
}

// Type returns the command type.
func (ci *Index) Type() int {
	return ci.CommandType
}
