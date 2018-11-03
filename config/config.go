package config

import (
	"github.com/justanotherorganization/justanotherbotkit/transport"
	"github.com/justanotherorganization/justanotherbotkit/users"
	"github.com/justanotherorganization/l5424"
)

// TODO: implement text marshallers for different message types.
const (
	// DefaultCommandPrefix is the default command prefix.
	DefaultCommandPrefix = `.`
)

type (
	// Config is the config for a graceless bot.
	Config struct {
		l5424.Logger
		// RootUsers contains the valid users who have root access to the bot.
		RootUsers []string
		// CmdPrefix is the prefix to match on for recognizing commands.
		CmdPrefix string
		// Safemode sets the bot into safemode.
		Safemode bool
		// DisableIntro allows for disabling the introduction entirely,
		// by default this is false.
		DisableIntro bool
		// HelloFunc is the function executed to say hello to new users.
		HelloFunc func(user *transport.User, conf *Config) error
		// Transport is the network transport.
		Transport transport.Transport
		// UserDB is a users.DB.
		UserDB users.DB
	}
)

// Validate a configuration and apply defaults (if not set) where possible.
// This does not set a CmdPrefix to allow for it to be un-set.
func (c *Config) Validate() error {
	if c.Transport == nil {
		return transport.ErrNilTransport
	}

	if c.Logger == nil {
		c.Logger = &l5424.NoOpLogger{}
	}

	if c.UserDB == nil {
		c.Safemode = true
	}

	if !c.DisableIntro {
		if c.HelloFunc == nil {
			c.HelloFunc = defaultSayHelloF
		}
	}

	return nil
}

// MarshalMessage marshals a string message using a specific format.
// func MarshalMessage(wrap uint, str string) string {
// 	fields := strings.Fields(str)

// 	var written uint
// 	line := []string{}
// 	final := []string{}
// 	for _, f := range fields {
// 		// TODO: better respect for existing new lines?
// 		if written == wrap {
// 			final = append(final, strings.Join(line, " "))
// 			line = []string{}
// 			written = 0
// 		}

// 		if (written + uint(len(f))) >= wrap {
// 			final = append(final, strings.Join(line, " "))
// 			line = []string{}
// 			written = 0
// 		}

// 		line = append(line, f)
// 		written += uint(len(f))
// 	}

// 	return strings.Join(final, "\n")
// }
