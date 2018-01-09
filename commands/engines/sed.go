package engines

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/justanotherorganization/graceless/commands"
	"github.com/justanotherorganization/justanotherbotkit/accessors"
	"github.com/pkg/errors"
)

var (
	// FoundSed is used to determine if sed was found before enabling it as a command.
	FoundSed = false

	_sedBinary string
)

type (
	// Sed runs a string through sed.
	Sed struct {
		*commands.Index
		usage string
	}
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
		_sedBinary = strings.TrimSuffix(str, "\n")
		FoundSed = true
	}
}

// NewSedCommand returns a new Sed command.
func NewSedCommand() *Sed {
	return &Sed{
		Index: &commands.Index{
			CmdName: "sed",
			// Even though technically this is an engine command cause it runs
			// outside of graceless we don't want to disable the command prefix
			// so treat it as a generic command.
			CommandType: commands.GenericCommand,
		},
		usage: "Usage: sed '<sed command>' <input>",
	}
}

// HelpShort returns the short help message for the sed command.
func (*Sed) HelpShort() string {
	return "sed : pass the following to sed"
}

// Help returns the command help.
func (s *Sed) Help() string {
	return s.usage
}

// Match matches a command string against the Sed command.
func (*Sed) Match(str string) (string, bool) {
	_str := strings.ToLower(str)

	if prefix := "sed"; strings.HasPrefix(_str, prefix) {
		return strings.TrimSpace(str[len(prefix):]), true
	}

	if prefix := "s"; strings.HasPrefix(_str, prefix) && !(len(str) > len(prefix) && string(str[len(prefix)]) != " ") {
		return strings.TrimSpace(str[len(prefix):]), true
	}

	return str, false
}

// Exec performs the sed command.
func (s *Sed) Exec(acc accessors.Accessor, cmdStr string, ev accessors.MessageEvent) error {
	usage := func() error {
		if err := acc.SendMessage(s.usage, ev.Origin); err != nil {
			return errors.Wrap(err, "acc.SendMessage")
		}

		return nil
	}

	if cmdStr == "" {
		return usage()
	}

	affix := "'"
	if strings.Count(cmdStr, affix) < 2 {
		return usage()
	}

	if !strings.HasPrefix(cmdStr, affix) {
		return usage()
	}

	idx := strings.LastIndex(cmdStr, affix)
	if len(cmdStr) <= idx+1 {
		return usage()
	}

	_cmd := strings.TrimPrefix(cmdStr[0:idx], affix)
	txt := strings.TrimSpace(strings.TrimPrefix(cmdStr[idx:], affix))

	cmd := exec.Command(_sedBinary, _cmd)
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
		return acc.SendMessage(fmt.Sprintf("%s", errb.String()), ev.Origin)
	}

	return acc.SendMessage(fmt.Sprintf("```%s```", out.String()), ev.Origin)
}
