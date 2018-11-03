package source

import (
	cmds "github.com/justanotherorganization/graceless/commands"
	"github.com/justanotherorganization/justanotherbotkit/commands"
	"github.com/justanotherorganization/justanotherbotkit/transport"
)

func init() {
	cmds.Register(
		&commands.Command{
			Use:   "source",
			Short: "Graceless source code",
			ExecFunc: func(ev *transport.Event) error {
				return ev.Transport.SendMessage(ev.Origin.ID, "https://github.com/justanotherorganization/graceless")
			},
		},
	)
}
