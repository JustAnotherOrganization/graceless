package internal

import (
	"fmt"
	"strings"

	"github.com/justanotherorganization/graceless/config"
	"github.com/justanotherorganization/graceless/internal/commands"
	"github.com/justanotherorganization/justanotherbotkit/transport"
	"github.com/pkg/errors"
)

// SayHello ...
func SayHello(user *transport.User, conf *config.Config) (bool, error) {
	// This is a silly way to track the say hello functionality
	// it needs some cleanup.
	known := false
	for _, p := range user.GetPermissions() {
		if p == "hello" {
			known = true
			break
		}
	}

	if known {
		return false, nil
	}

	totalCmds := commands.Count()

	// FIXME: this should probably use a strings builder or something, it's very inefficient.
	introStr := conf.IntroStart
	if totalCmds < 11 {
		introStr = fmt.Sprintf("%s\nUnfortunately I only know a few commands right now...\n", introStr)
	}
	if totalCmds > 10 && totalCmds < 21 {
		introStr = fmt.Sprintf("%s\nI can do several things, I think you'll like...\n", introStr)
	}
	if totalCmds > 20 && totalCmds < 31 {
		introStr = fmt.Sprintf("%s\nI can do a number of things to help with your day...\n", introStr)
	}
	if totalCmds > 30 && totalCmds < 41 {
		introStr = fmt.Sprintf("%s\nI can do a lot a things, it's really cool just how many...\n", introStr)
	}
	if totalCmds > 40 && totalCmds < 51 {
		introStr = fmt.Sprintf("%s\nI can do so many things, like you won't believe your eyes...\n", introStr)
	}
	if totalCmds > 50 && totalCmds < 101 {
		introStr = fmt.Sprintf("%s\nI can do a great many things, it's totally amazing...\n", introStr)
	}
	if totalCmds > 100 {
		introStr = fmt.Sprintf("%s\nI can do way too many things, like seriously...\n", introStr)
	}

	finishStr := strings.Replace(conf.IntroFinish, "[tag]", "`", -1)
	finishStr = strings.Replace(finishStr, "[prefix]", conf.CmdPrefix, -1)
	introStr = fmt.Sprintf("%s%s", introStr, finishStr)

	conversationID, err := user.Transport.GetConversation(user.GetID())
	if err != nil {
		return false, errors.Wrap(err, "Transport.GetConversation")
	}

	return true, user.Transport.SendMessage(conversationID, introStr)
}
