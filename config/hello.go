package config

import (
	"fmt"
	"strings"

	"github.com/justanotherorganization/graceless/commands"
	"github.com/justanotherorganization/justanotherbotkit/transport"
	"github.com/pkg/errors"
)

const (
	introStart = `Hi! I just wanted to introduce myself, I'm a graceless chat bot.`
	// FIXME:
	_introFinish = `If you want to know what they are just type [tag][prefix]help[tag] in any channel
and I'll respond to you here.

Lastly, just to warn you, I'm really clumsy!!!`

	temporaryIntroFinish = `Sadly my help command is totally borked currently too, but if you'd like to help
with fixing it you can get the link to my source using [tag][prefix]source[tag].

Lastly, just to warn you, I'm really clumsy!!!`
)

var introFinish = temporaryIntroFinish

func defaultSayHelloF(user *transport.User, conf *Config) error {
	totalCmds := commands.Count()

	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", introStart)
	if totalCmds < 11 {
		fmt.Fprintf(&b, "Unfortunately I only know a few commands right now...\n\n")
	}
	if totalCmds > 10 && totalCmds < 21 {
		fmt.Fprintf(&b, "I can do several things, I think you'll like...\n\n")
	}
	if totalCmds > 20 && totalCmds < 31 {
		fmt.Fprintf(&b, "I can do a number of things to help with your day...\n\n")
	}
	if totalCmds > 30 && totalCmds < 41 {
		fmt.Fprintf(&b, "I can do a lot a things, it's really cool just how many...\n\n")
	}
	if totalCmds > 40 && totalCmds < 51 {
		fmt.Fprintf(&b, "I can do so many things, like you won't believe your eyes...\n\n")
	}
	if totalCmds > 50 && totalCmds < 101 {
		fmt.Fprintf(&b, "I can do a great many things, it's totally amazing...\n\n")
	}
	if totalCmds > 100 {
		fmt.Fprintf(&b, "I can do way too many things, like seriously...\n\n")
	}

	finishStr := strings.Replace(introFinish, "[tag]", "`", -1)
	finishStr = strings.Replace(finishStr, "[prefix]", conf.CmdPrefix, -1)
	fmt.Fprintln(&b, finishStr)

	conversationID, err := user.Transport.GetConversation(user.GetID())
	if err != nil {
		return errors.Wrap(err, "Transport.GetConversation")
	}

	return user.Transport.SendMessage(conversationID, b.String())
}
