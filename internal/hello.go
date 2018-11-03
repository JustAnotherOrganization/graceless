package internal

import (
	"github.com/justanotherorganization/graceless/config"
	"github.com/justanotherorganization/justanotherbotkit/transport"
)

// SayHello ...
func SayHello(user *transport.User, conf *config.Config) (bool, error) {
	if conf.DisableIntro || conf.HelloFunc == nil {
		return false, nil
	}

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

	if err := conf.HelloFunc(user, conf); err != nil {
		return false, err
	}

	return true, nil
}
