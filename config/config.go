package config

import (
	"github.com/justanotherorganization/justanotherbotkit/transport"
	"github.com/justanotherorganization/justanotherbotkit/users"
	"github.com/sirupsen/logrus"
)

// TODO: implement text marshallers for different message types.
const (
	// DefaultCommandPrefix is the default command prefix.
	DefaultCommandPrefix = `.`

	defaultIntroStart  = `Hi! I just wanted to introduce myself, I'm a graceless chat bot.`
	defaultIntroFinish = `If you want to know what they are just type [tag][prefix]help[tag] in any channel
and I'll respond to you here. Lastly, just to warn you, I'm really clumsy!`

	temporaryIntroFinish = `Sadly my help command is totally borked currently too, but if you'd like to help
with fixing it you can get the link to my source using [tag][prefix]source[tag].
Lastly, just to warn you, I'm really clumsy!`
)

type (
	// Config is the config for a graceless bot.
	Config struct {
		// RootUsers contains the valid users who have root access to the bot.
		RootUsers []string
		// CmdPrefix is the prefix to match on for recognizing commands.
		CmdPrefix string
		// Safemode sets the bot into safemode.
		Safemode bool
		// Intro allows for settings a custom bot introduction (if not provided
		// a default will be used).
		IntroStart string
		// IntroFinish is the end of the bot introduction (if not provided a
		// default will be used).
		IntroFinish string
		// DisableIntro allows for disabling the introduction entirely,
		// by default this is false.
		DisableIntro bool
		// Transport is the network transport.
		Transport transport.Transport
		// Log is a logrus.Entry (this will be replaced very very soon).
		Log *logrus.Entry
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

	if c.Log == nil {
		c.Log = logrus.NewEntry(logrus.New())
	}

	if c.UserDB == nil {
		c.Safemode = true
	}

	if c.IntroStart == "" {
		c.IntroStart = defaultIntroStart
	}

	if c.IntroFinish == "" {
		// FIXME:
		//c.IntroFinish = defaultIntroFinish
		c.IntroFinish = temporaryIntroFinish
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
