package config

// TODO: implement text marshallers for different message types.
const (
	// DefaultCommandPrefix is the default command prefix.
	DefaultCommandPrefix = `.`

	defaultIntroStart  = `Hi! I just wanted to introduce myself, I'm a graceless chat bot.`
	defaultIntroFinish = `If you want to know what they are just type [tag][prefix]help[tag] in any channel
and I'll respond to you here. Lastly, just to warn you, I'm really clumsy!`
)

type (
	// Config is the config for a graceless bot.
	Config struct {
		// RootIDs contains the IDs of the users who can shutdown the bot
		// (if no IDs are provided anyone will be able to shutdown the bot).
		RootIDs []string
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
		// WithGoEngine can be used to enable support for running go code
		// from a channel (this is highly experimental and very unsafe,
		// default is false).
		WithGoEngine bool
		// WithJSEngine can be used to enable support for running js code
		// from a channel (this is highly experimental, default is false).
		WithJSEngine bool
	}
)

// SetDefaults sets any unset config values to their defaults.
func SetDefaults(config *Config) {
	if config == nil {
		config = &Config{}
	}

	if config.CmdPrefix == "" {
		config.CmdPrefix = DefaultCommandPrefix
	}

	if config.IntroStart == "" && !config.DisableIntro {
		config.IntroStart = defaultIntroStart
	}

	if config.IntroFinish == "" && !config.DisableIntro {
		config.IntroFinish = defaultIntroFinish
	}
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
