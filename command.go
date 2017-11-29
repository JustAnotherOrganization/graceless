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
		Match(string) bool
		// NoSafemode tells if a command can't be ran in safemode.
		NoSafemode() bool
		// NeedsDB tells if a command requires a database connection to function.
		NeedsDB() bool
		// NeedsRoot tells if the command requires root privs to run.
		NeedsRoot() bool // FIXME: replace with perms slice
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
		noSafemode  bool
		needsDB     bool
		needsRoot   bool
		commandType int
	}
)

// NoSafemode tells if a command can't be ran in safemode.
func (ci *Index) NoSafemode() bool {
	return ci.noSafemode
}

// NeedsDB tells if a command requires a database connection to function.
func (ci *Index) NeedsDB() bool {
	return ci.needsDB
}

// NeedsRoot tells if the command requires root privs to run.
func (ci *Index) NeedsRoot() bool {
	return ci.needsRoot
}

// Type returns the command type.
func (ci *Index) Type() int {
	return ci.commandType
}
