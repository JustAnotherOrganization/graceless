package sed

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"

	cmds "github.com/justanotherorganization/graceless/internal/commands"
	"github.com/justanotherorganization/justanotherbotkit/commands"
	"github.com/justanotherorganization/justanotherbotkit/transport"
	"github.com/pkg/errors"
)

func init() {
	cmd := exec.Command("which", "sed")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		fmt.Println(err.Error())
		return
	}

	if str := out.String(); str != "" {
		sedBinary := strings.TrimSuffix(str, "\n")

		_cmd := commands.Command{
			Use:     "sed",
			Aliases: []string{"s"},
			Short:   "pass the following to sed",
			Hidden:  false,
		}

		_cmd.ExecFunc = func(ev *transport.Event) error {
			usage := func() error {
				if err := ev.Transport.SendMessage(ev.Origin.ID, "sed '<sed command>' <input>"); err != nil {
					return errors.Wrap(err, "ev.Transport.SendMessage")
				}

				return nil
			}

			if ev.Body == "" {
				return usage()
			}

			affix := "'"
			if strings.Count(ev.Body, affix) < 2 {
				return usage()
			}

			if !strings.HasPrefix(ev.Body, affix) {
				return usage()
			}

			idx := strings.LastIndex(ev.Body, affix)
			if len(ev.Body) <= idx+1 {
				return usage()
			}

			_cmd := strings.TrimPrefix(ev.Body[0:idx], affix)
			txt := strings.TrimSpace(strings.TrimPrefix(ev.Body[idx:], affix))

			cmd := exec.Command(sedBinary, _cmd)
			stdin, err := cmd.StdinPipe()
			if err != nil {
				return errors.Wrap(err, "cmd.StdinPipe")
			}

			var (
				out  bytes.Buffer
				errb bytes.Buffer
			)
			cmd.Stdout = &out
			cmd.Stderr = &errb

			if err := cmd.Start(); err != nil {
				return errors.Wrap(err, "cmd.Start")
			}

			go func() {
				defer stdin.Close()
				io.WriteString(stdin, txt)
			}()

			if err := cmd.Wait(); err != nil {
				return ev.Transport.SendMessage(ev.Origin.ID, fmt.Sprintf("%s", errb.String()))
			}

			return ev.Transport.SendMessage(ev.Origin.ID, fmt.Sprintf("```%s```", out.String()))
		}

		cmds.Register(&_cmd)
	}
}
